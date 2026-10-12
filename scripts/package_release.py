#!/usr/bin/env python3
"""Build release candidates locally; never tag, push, or publish."""
import argparse
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
from urllib.parse import urlparse
import zipfile

ROOT = Path(__file__).resolve().parents[1]
TARGETS = [(system, arch) for system in ("darwin", "linux", "windows") for arch in ("amd64", "arm64")]
VERSION = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*)?\Z")


def run(*args, env=None):
    return subprocess.check_output(args, cwd=ROOT, env=env, text=True).strip()


def archive(path, files, windows):
    if windows:
        with zipfile.ZipFile(path, "w", compression=zipfile.ZIP_DEFLATED) as out:
            for name, (data, mode) in sorted(files.items()):
                info = zipfile.ZipInfo(name, (1980, 1, 1, 0, 0, 0))
                info.external_attr = mode << 16
                info.compress_type = zipfile.ZIP_DEFLATED
                out.writestr(info, data)
    else:
        with path.open("wb") as raw, gzip.GzipFile(filename="", mode="wb", fileobj=raw, mtime=0) as gz, tarfile.open(fileobj=gz, mode="w") as out:
            for name, (data, mode) in sorted(files.items()):
                info = tarfile.TarInfo(name)
                info.size, info.mode, info.mtime = len(data), mode, 0
                out.addfile(info, io.BytesIO(data))


def formula(version, hashes):
    lines = ['class BibleitCli < Formula', '  desc "Command-line client for Bibleit"', '  homepage "https://github.com/bibleit-labs/bibleit-cli"', f'  version "{version}"', '  license "MIT"']
    for system, section in (("darwin", "on_macos"), ("linux", "on_linux")):
        lines.append(f"  {section} do")
        for arch, condition in (("arm64", "Hardware::CPU.arm?"), ("amd64", "Hardware::CPU.intel?")):
            name = f"bibleit_{version}_{system}_{arch}.tar.gz"
            lines.extend([f"    if {condition}", f'      url "https://github.com/bibleit-labs/bibleit-cli/releases/download/v{version}/{name}"', f'      sha256 "{hashes[name]}"', '    end'])
        lines.append('  end')
    lines.extend(['  def install', '    bin.install "bibleit"', '    pkgshare.install "LICENSE", "THIRD_PARTY_NOTICES"', '    bash_completion.install "completions/bibleit.bash" => "bibleit"', '    zsh_completion.install "completions/_bibleit"', '    fish_completion.install "completions/bibleit.fish"', '  end', '  test do', '    assert_match version.to_s, shell_output("#{bin}/bibleit version")', '  end', 'end', ''])
    return '\n'.join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', required=True, help='Candidate version, e.g. v0.1.0-dev.1; no Git tag is created')
    parser.add_argument('--local', action='store_true', help='Prepare development-only artifacts using the unpublished sibling client')
    parser.add_argument('--output', default='build/releases')
    args = parser.parse_args()
    if not VERSION.fullmatch(args.version):
        parser.error('version must be vMAJOR.MINOR.PATCH with an optional prerelease')
    env = dict(os.environ, CGO_ENABLED='0', GOTOOLCHAIN='local')
    env.pop('GOOS', None)
    env.pop('GOARCH', None)
    if not args.local:
        env['GOWORK'] = 'off'
        if re.search(r'\bbibleit-go\s+v0\.0\.0(?:\s|$)', (ROOT / 'go.mod').read_text()):
            parser.error('release blocked: pin a published Go client; use --local for development candidates')
        client = json.loads(run('go', 'list', '-m', '-json', 'github.com/bibleit-labs/bibleit-go', env=env))
        if client.get('Replace'):
            parser.error('release blocked: replaced Go client dependency')
        if not client.get('Dir'):
            run('go', 'mod', 'download', 'github.com/bibleit-labs/bibleit-go', env=env)
            client = json.loads(run('go', 'list', '-m', '-json', 'github.com/bibleit-labs/bibleit-go', env=env))
        client_source = (Path(client['Dir']) / 'client.go').read_text()
        endpoint_match = re.search(r'ServerEndpoint\s*=\s*"([^"]+)"', client_source)
        endpoint = urlparse(endpoint_match[1] if endpoint_match else '')
        ssh_match = re.search(r'defaultSSHServer\s*=\s*"([^"]+)"', (ROOT / 'main.go').read_text())
        ssh = urlparse('ssh://' + (ssh_match[1] if ssh_match else ''))
        def production_host(host):
            return host and host.lower() != 'localhost' and not host.lower().endswith('.localhost') and not host.startswith('127.') and host not in ('::1', '0.0.0.0', '::')
        if endpoint.scheme != 'https' or not production_host(endpoint.hostname) or not production_host(ssh.hostname):
            parser.error('release blocked: set production HTTPS/SSH addresses first')
    output = Path(args.output).resolve()
    if output.exists():
        parser.error('output already exists; select a new directory to preserve previous candidates')
    if run('go', 'version', env=env).split()[2] != 'go1.27.1':
        parser.error('Go 1.27.1 is required')
    version = args.version[1:]
    commit = run('git', 'rev-parse', 'HEAD')
    dirty = bool(run('git', 'status', '--porcelain'))
    if dirty and not args.local:
        parser.error('release blocked: commit the working tree first')
    hashes = {}
    # Stage everything before creating the destination; a failed build leaves no partial release.
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(dir=output.parent) as scratch:
        stage = Path(scratch)
        native = stage / ('native.exe' if os.name == 'nt' else 'native')
        run('go', 'build', '-trimpath', '-buildvcs=false', '-ldflags', f'-s -w -X main.cliVersion={version}', '-o', str(native), '.', env=env)
        common = {name: ((ROOT / name).read_bytes(), 0o644) for name in ('LICENSE', 'THIRD_PARTY_NOTICES', 'README.md')}
        for shell, name in (('bash', 'bibleit.bash'), ('zsh', '_bibleit'), ('fish', 'bibleit.fish')):
            common[f'completions/{name}'] = (run(str(native), 'completion', shell).encode() + b'\n', 0o644)
        for system, arch in TARGETS:
            binary = 'bibleit.exe' if system == 'windows' else 'bibleit'
            target = stage / binary
            target_env = dict(env, GOOS=system, GOARCH=arch)
            run('go', 'build', '-trimpath', '-buildvcs=false', '-ldflags', f'-s -w -X main.cliVersion={version}', '-o', str(target), '.', env=target_env)
            name = f'bibleit_{version}_{system}_{arch}.' + ('zip' if system == 'windows' else 'tar.gz')
            archive(stage / name, dict(common, **{binary: (target.read_bytes(), 0o755)}), system == 'windows')
            hashes[name] = hashlib.sha256((stage / name).read_bytes()).hexdigest()
            target.unlink()
        native.unlink()
        for name in ('install.sh', 'install.ps1'):
            (stage / name).write_bytes((ROOT / name).read_bytes())
            hashes[name] = hashlib.sha256((stage / name).read_bytes()).hexdigest()
        (stage / 'checksums.txt').write_text(''.join(f'{digest}  {name}\n' for name, digest in sorted(hashes.items())))
        (stage / 'bibleit-cli.rb').write_text(formula(version, hashes))
        scoop = dict(version=version, description='Command-line client for Bibleit', homepage='https://github.com/bibleit-labs/bibleit-cli', license='MIT', bin='bibleit.exe', architecture={})
        for arch, scoop_arch in (('amd64', '64bit'), ('arm64', 'arm64')):
            name = f'bibleit_{version}_windows_{arch}.zip'
            scoop['architecture'][scoop_arch] = dict(url=f'https://github.com/bibleit-labs/bibleit-cli/releases/download/v{version}/{name}', hash=hashes[name])
        (stage / 'bibleit.json').write_text(json.dumps(scoop, indent=2) + '\n')
        module = json.loads(run('go', 'list', '-m', '-json', 'github.com/bibleit-labs/bibleit-go', env=env))
        client_metadata = {key: module[key] for key in ('Path', 'Version', 'GoVersion', 'Sum') if key in module}
        client_metadata['workspace'] = bool(module.get('Main'))
        if module.get('Main') and module.get('Dir'):
            client_metadata['commit'] = run('git', '-C', module['Dir'], 'rev-parse', 'HEAD')
            client_metadata['dirty'] = bool(run('git', '-C', module['Dir'], 'status', '--porcelain'))
        (stage / 'manifest.json').write_text(json.dumps(dict(version=args.version, commit=commit, dirty=dirty, local_only=args.local, go='1.27.1', client=client_metadata, artifacts=hashes), indent=2) + '\n')
        stage.rename(output)
    print(f'Prepared {len(TARGETS)} archives at {output}; nothing published.')


if __name__ == '__main__':
    main()
