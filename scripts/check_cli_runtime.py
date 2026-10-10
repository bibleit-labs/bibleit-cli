#!/usr/bin/env python3
"""Run the built CLI on this OS; fixtures are local, disposable and offline."""
import argparse
import json
import os
from pathlib import Path
import platform
import select
import shutil
import subprocess
import tempfile
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


def run(binary, args, env, code=0):
    result = subprocess.run([str(binary), *args], env=env, input="", text=True,
                            capture_output=True, timeout=20)
    assert result.returncode == code, (args, result.returncode, result.stdout, result.stderr)
    return result


def shells(binary, env, directory, requested):
    cases = {
        "bibleit li": ["live"],
        "bibleit live L1 st": ["stack", "start", "stats", "stop"],
        "bibleit live L1 stack p": ["pop", "push"],
        "bibleit live L1 secret ": ["create", "delete", "rotate"],
        "bibleit --transport s": ["ssh"],
        "bibleit profile ": [],
        "bibleit --format j": ["json"],
        "bibleit read KJV ": [],
        "bibleit -- search KJV --": [],
    }
    for shell in requested:
        assert shutil.which(shell), f"required shell unavailable: {shell}"
        script = directory / ("completion." + shell)
        script.write_text(run(binary, ["completion", shell], env).stdout)
        subprocess.run([shell, "-n", str(script)], check=True, env=env)
        for command, expected in cases.items():
            words = command.split(" ")[1:]
            # Cases contain only literal ASCII words; no fixture content is shell code.
            if shell == "fish":
                source = f'source "$BIBLEIT_COMPLETION_FILE"; complete -C "{command}"'
            elif shell == "bash":
                quoted = " ".join("'" + word + "'" for word in ["bibleit", *words])
                source = f'source "$BIBLEIT_COMPLETION_FILE"; COMP_WORDS=({quoted}); COMP_CWORD={len(words)}; _bibleit_complete; printf "%s\\n" "${{COMPREPLY[@]}}"'
            else:
                quoted = " ".join("'" + word + "'" for word in ["bibleit", *words])
                source = f'compdef() {{ :; }}; compadd() {{ shift; printf "%s\\n" "$@"; }}; source "$BIBLEIT_COMPLETION_FILE"; words=({quoted}); CURRENT={len(words)+1}; _bibleit; true'
            result = subprocess.run([shell, "-c", source], env={**env, "BIBLEIT_COMPLETION_FILE": str(script)},
                                    text=True, capture_output=True, timeout=20)
            assert result.returncode == 0 and result.stderr == "", (shell, command, result.stderr)
            got = sorted(line.split("\t")[0] for line in result.stdout.splitlines() if line)
            assert got == sorted(expected), (shell, command, got, expected)
        print(f"{shell}: syntax and {len(cases)} actual completion cases pass", flush=True)


def terminal_checks(binary, env, requests):
    if os.name == "nt":
        print("Windows process checks use redirected handles; real console detection is covered by TestWindowsConsoleHandles", flush=True)
        return
    import pty
    for answer, code, delta in (("yes", 1, 0), ("", 1, 0), ("delete all", 0, 1)):
        master, slave = pty.openpty()
        before = len(requests)
        child = subprocess.Popen([str(binary), "live", "delete", "all"], env=env,
                                 stdin=slave, stderr=slave, stdout=subprocess.PIPE)
        os.close(slave)
        try:
            output = b""
            deadline = time.monotonic() + 10
            while b'Type "delete all" to confirm: ' not in output and time.monotonic() < deadline:
                if select.select([master], [], [], 0.5)[0]:
                    output += os.read(master, 4096)
            assert b'Type "delete all" to confirm: ' in output, output
            assert b"http" in output and b"localhost:8080" in output and b"bt_runtime" not in output, output
            os.write(master, (answer + "\n").encode())
            stdout, _ = child.communicate(timeout=10)
            assert child.returncode == code and len(requests) == before + delta, (answer, child.returncode, requests)
            if code != 0:
                assert stdout == b"", stdout
        finally:
            if child.poll() is None:
                child.kill()
                child.wait()
            os.close(master)
    print("Native PTY: scoped bulk prompt, rejection, empty answer and confirmed dispatch pass", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--shells", nargs="*", choices=["bash", "zsh", "fish"], default=[])
    args = parser.parse_args()
    source_binary = args.binary.resolve(strict=True)
    requests = []

    class Handler(BaseHTTPRequestHandler):
        def log_message(self, *_):
            pass

        def do_POST(self):
            body = self.rfile.read(int(self.headers.get("Content-Length", "0")))
            requests.append((self.path, self.headers.get("Authorization"), body))
            command = json.loads(body).get("command")
            if self.path != "/api/cli/command" or self.headers.get("Authorization") != "Bearer bt_runtime":
                self.send_response(403)
                self.end_headers()
                return
            reply = b"OK pong=true\n" if command == "PING" else b"OK event=deleted\n"
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.send_header("Content-Length", str(len(reply)))
            self.end_headers()
            self.wfile.write(reply)

    server = ThreadingHTTPServer(("127.0.0.1", 8080), Handler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        with tempfile.TemporaryDirectory(prefix="bibleit runtime ") as tmp:
            directory = Path(tmp)
            binary = directory / ("bibleit.exe" if os.name == "nt" else "bibleit")
            shutil.copy2(source_binary, binary)
            config = directory / "credentials with spaces.json"
            env = {**os.environ, "BIBLEIT_CONFIG": str(config), "BIBLEIT_PROFILE": "",
                   "BIBLEIT_TOKEN": "bt_runtime", "BIBLEIT_ENDPOINT": "https://other.example", "PATH": str(directory) + os.pathsep + os.environ["PATH"]}
            config.write_text(json.dumps({"version": 2, "endpoint": "http://localhost:8080", "ssh_server": "127.0.0.1:2222", "access_token": "bt_saved"}))
            before = len(requests)
            for invalid_args in (["profile", "list"], ["--profile", "old", "ping"], ["--endpoint", "https://other.example", "ping"]):
                run(binary, invalid_args, env, code=2)
            assert len(requests) == before
            store = json.loads(config.read_text())
            assert store["version"] == 2 and "profiles" not in store
            assert "bt_runtime" not in config.read_text()
            response = json.loads(run(binary, ["ping", "--format", "json"], env).stdout)
            assert response["schema_version"] == 1 and response["records"][0]["fields"]["pong"] is True
            before = len(requests)
            result = run(binary, ["live", "delete", "all", "--format", "json"], env, code=2)
            assert result.stdout == "" and json.loads(result.stderr)["error"]["code"] == "confirmation_required"
            assert len(requests) == before
            run(binary, ["live", "delete", "all"], env, code=2)
            assert len(requests) == before
            response = json.loads(run(binary, ["live", "delete", "all", "--yes", "--format", "json"], env).stdout)
            assert response["ok"] is True and len(requests) == before + 1
            assert json.loads(requests[-1][2])["command"] == "LIVE DELETE ALL"
            print(f"Native {platform.system()}/{platform.machine()}: single configuration, spaced paths, loopback HTTP, JSON and destructive guards pass", flush=True)
            terminal_checks(binary, env, requests)
            before_config, before_requests = config.read_bytes(), len(requests)
            shells(binary, env, directory, args.shells)
            assert config.read_bytes() == before_config and len(requests) == before_requests
            print("Completion: no fixture requests or configuration writes", flush=True)
    finally:
        server.shutdown()
        server.server_close()
        thread.join(timeout=5)


if __name__ == "__main__":
    main()
