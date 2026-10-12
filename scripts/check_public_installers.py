#!/usr/bin/env python3
"""Exercise real public release URLs without credentials or local asset fixtures."""
import hashlib
import json
import os
import platform
import shutil
import subprocess
import tempfile
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
VERSION = "v0.0.2"
REPO = "https://github.com/bibleit-labs/bibleit-cli"


def main():
    env = {k: v for k, v in os.environ.items() if "TOKEN" not in k.upper() and "SECRET" not in k.upper() and not k.startswith(("GIT_", "BIBLEIT_"))}
    env.update(GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_SYSTEM=os.devnull, GIT_TERMINAL_PROMPT="0")
    windows = platform.system() == "Windows"
    engines = ["pwsh", "powershell"] if windows else ["bash"]
    evidence = {"version": VERSION, "platform": platform.platform(), "machine": platform.machine(), "checks": []}
    with tempfile.TemporaryDirectory(prefix="bibleit-public-") as temporary:
        work = Path(temporary)
        for engine in engines:
            assert shutil.which(engine), engine
            destination = work / engine / "bin"
            for mode in ("latest", "explicit"):
                name = "install.ps1" if windows else "install.sh"
                url = REPO + ("/releases/latest/download/" if mode == "latest" else "/releases/download/" + VERSION + "/") + name
                with urllib.request.urlopen(url, timeout=120) as response:
                    data = response.read()
                    resolved_url = response.geturl()
                assert hashlib.sha256(data).digest() == hashlib.sha256((ROOT / name).read_bytes()).digest()
                script = work / name
                script.write_bytes(data)
                if windows:
                    command = [engine, "-NoProfile", "-File", str(script), "-InstallDir", str(destination)]
                    if mode == "explicit":
                        command += ["-Version", VERSION]
                else:
                    env["BIBLEIT_INSTALL_DIR"] = str(destination)
                    command = [engine, str(script)] + ([VERSION] if mode == "explicit" else [])
                subprocess.run(command, env=env, check=True, timeout=300)
                binary = destination / ("bibleit.exe" if windows else "bibleit")
                result = subprocess.run([str(binary), "version"], env=env, check=True, capture_output=True, text=True)
                assert result.stdout.strip() == "bibleit CLI version " + VERSION[1:]
                subprocess.run([str(binary), "--help"], env=env, check=True, capture_output=True)
                for notice in ("LICENSE", "THIRD_PARTY_NOTICES"):
                    assert (destination / "bibleit-licenses" / notice).read_bytes() == (ROOT / notice).read_bytes()
                evidence["checks"].append({"engine": engine, "mode": mode, "installer_url": url, "resolved_url": resolved_url.split("?")[0], "installer_sha256": hashlib.sha256(data).hexdigest(), "version_help_notices": "passed"})
                print(engine + " " + mode + ": public download, install/replacement, version, help and notices passed", flush=True)
    report = ROOT / "build/validation/public-installers.json"
    report.parent.mkdir(parents=True, exist_ok=True)
    report.write_text(json.dumps(evidence, indent=2) + "\n")


if __name__ == "__main__":
    main()
