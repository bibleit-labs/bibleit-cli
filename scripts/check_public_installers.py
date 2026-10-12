#!/usr/bin/env python3
"""Exercise real public release URLs without credentials or local asset fixtures."""
import argparse
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
REPO = "https://github.com/bibleit-labs/bibleit-cli"


def main():
    from package_release import VERSION
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', required=True)
    args = parser.parse_args()
    if not VERSION.fullmatch(args.version):
        parser.error('expected vMAJOR.MINOR.PATCH with optional prerelease')
    version = args.version
    approved = json.loads((ROOT / 'docs/security' / (version + '-release-validation.json')).read_text())
    manifest = approved['manifest']
    assert manifest['version'] == version and not manifest['dirty'] and not manifest['local_only']
    machine = platform.machine().lower()
    architecture = 'arm64' if machine in ('arm64', 'aarch64') else 'amd64' if machine in ('amd64', 'x86_64') else None
    expected = {
        'ubuntu-24.04': ('Linux', 'amd64'), 'ubuntu-24.04-arm': ('Linux', 'arm64'),
        'macos-15': ('Darwin', 'arm64'), 'macos-15-intel': ('Darwin', 'amd64'),
        'windows-2025': ('Windows', 'amd64'), 'windows-11-arm': ('Windows', 'arm64'),
    }
    runner = os.environ.get('EXPECTED_INSTALLER_RUNNER')
    if runner:
        assert (platform.system(), architecture) == expected[runner], 'native host does not match runner'
    env = {k: v for k, v in os.environ.items() if "TOKEN" not in k.upper() and "SECRET" not in k.upper() and not k.startswith(("GIT_", "BIBLEIT_"))}
    # A Python process launched by PowerShell 7 inherits its module search path.
    # Let each child PowerShell engine initialize its own standard modules.
    env = {k: v for k, v in env.items() if k.upper() not in ("PSMODULEPATH", "PSHOME")}
    env.update(GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_SYSTEM=os.devnull, GIT_TERMINAL_PROMPT="0")
    windows = platform.system() == "Windows"
    engines = ["pwsh", "powershell"] if windows else ["bash"]
    evidence = {"version": version, "commit": manifest['commit'], "platform": platform.platform(), "machine": platform.machine(), "expected_runner": runner, "credentials": False, "local_archive_fixture": False, "checks": []}
    with tempfile.TemporaryDirectory(prefix="bibleit-public-") as temporary:
        work = Path(temporary)
        for engine in engines:
            assert shutil.which(engine), engine
            destination = work / engine / "bin"
            for mode in ("latest", "explicit"):
                name = "install.ps1" if windows else "install.sh"
                url = REPO + ("/releases/latest/download/" if mode == "latest" else "/releases/download/" + version + "/") + name
                with urllib.request.urlopen(url, timeout=120) as response:
                    data = response.read()
                    resolved_url = response.geturl()
                # Compare with the frozen source, regardless of Windows CRLF
                # checkout conversion or later changes to the default branch.
                canonical = subprocess.check_output(["git", "show", manifest['commit'] + ":" + name], cwd=ROOT, env=env)
                assert hashlib.sha256(data).hexdigest() == manifest['artifacts'][name]
                assert data == canonical
                script = work / name
                script.write_bytes(data)
                if windows:
                    command = [engine, "-NoProfile", "-File", str(script), "-InstallDir", str(destination)]
                    if mode == "explicit":
                        command += ["-Version", version]
                else:
                    env["BIBLEIT_INSTALL_DIR"] = str(destination)
                    command = [engine, str(script)] + ([version] if mode == "explicit" else [])
                subprocess.run(command, env=env, check=True, timeout=300)
                binary = destination / ("bibleit.exe" if windows else "bibleit")
                result = subprocess.run([str(binary), "version"], env=env, check=True, capture_output=True, text=True)
                assert result.stdout.strip() == "bibleit CLI version " + version[1:]
                help_result = subprocess.run([str(binary), "--help"], env=env, check=True, capture_output=True, text=True)
                assert 'Bibleit CLI: read Scripture' in help_result.stdout
                config = work / 'legacy-config.json'
                config.write_text('{"version":0}\n')
                original_config = config.read_bytes()
                env['BIBLEIT_CONFIG'] = str(config)
                auth = subprocess.run([str(binary), 'auth'], env=env, check=True, capture_output=True, text=True)
                assert 'bibleit auth login' in auth.stdout and not auth.stderr
                assert config.read_bytes() == original_config
                for notice in ("LICENSE", "THIRD_PARTY_NOTICES"):
                    assert (destination / "bibleit-licenses" / notice).read_bytes() == subprocess.check_output(["git", "show", manifest['commit'] + ":" + notice], cwd=ROOT, env=env)
                binary_sha256 = hashlib.sha256(binary.read_bytes()).hexdigest()
                system = platform.system().lower()
                assert binary_sha256 == approved['binary_sha256'][system + '_' + architecture]
                evidence["checks"].append({"engine": engine, "mode": mode, "installer_url": url, "resolved_url": resolved_url.split("?")[0], "installer_sha256": hashlib.sha256(data).hexdigest(), "binary_sha256": binary_sha256, "version_help_notices_auth_legacy_config": "passed"})
                print(engine + " " + mode + ": public download, install/replacement, version, help and notices passed", flush=True)
    report = ROOT / "build/validation/public-installers.json"
    report.parent.mkdir(parents=True, exist_ok=True)
    report.write_text(json.dumps(evidence, indent=2) + "\n")


if __name__ == "__main__":
    main()
