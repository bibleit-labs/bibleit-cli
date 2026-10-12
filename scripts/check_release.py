#!/usr/bin/env python3
"""Validate archive contents/checksums and exercise the native installer offline."""
import argparse
import hashlib
import io
import json
import os
import platform
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--assets', required=True)
    parser.add_argument('--powershell', help='Windows installer runtime, e.g. powershell for 5.1 or pwsh for 7')
    args = parser.parse_args()
    assets = Path(args.assets).resolve()
    manifest = json.loads((assets / 'manifest.json').read_text())
    version = manifest['version']
    machine = platform.machine().lower()
    architecture = 'arm64' if machine in ('arm64', 'aarch64') else 'amd64' if machine in ('amd64', 'x86_64') else None
    assert architecture, f'unsupported native architecture: {machine}'
    expected_runner = os.environ.get('EXPECTED_INSTALLER_RUNNER')
    if expected_runner:
        expected = {
            'ubuntu-24.04': ('Linux', 'amd64'),
            'ubuntu-24.04-arm': ('Linux', 'arm64'),
            'macos-15': ('Darwin', 'arm64'),
            'macos-15-intel': ('Darwin', 'amd64'),
            'windows-2025': ('Windows', 'amd64'),
            'windows-11-arm': ('Windows', 'arm64'),
        }
        assert (platform.system(), architecture) == expected[expected_runner], 'native host does not match runner'
    hashes = {}
    for line in (assets / 'checksums.txt').read_text().splitlines():
        digest, name = line.split('  ')
        assert name not in hashes, f'duplicate checksum: {name}'
        assert hashlib.sha256((assets / name).read_bytes()).hexdigest() == digest, name
        hashes[name] = digest
    assert hashes == manifest['artifacts']
    assert len(hashes) == 8, 'expected six archives and two installers'
    for name in hashes:
        if name.endswith('.zip'):
            with zipfile.ZipFile(assets / name) as archive:
                files = {entry.filename: archive.read(entry) for entry in archive.infolist()}
                assert len(files) == len(archive.infolist())
            binary = 'bibleit.exe'
            assert files[binary].startswith(b'MZ')
        elif name.endswith('.tar.gz'):
            with tarfile.open(assets / name) as archive:
                members = archive.getmembers()
                assert all(member.isfile() for member in members)
                files = {member.name: archive.extractfile(member).read() for member in members}
                assert len(files) == len(members)
                assert archive.getmember('bibleit').mode == 0o755
            binary = 'bibleit'
            if '_linux_' in name:
                assert files[binary].startswith(b'\x7fELF')
        else:
            continue
        assert set(files) == {binary, 'LICENSE', 'THIRD_PARTY_NOTICES', 'README.md', 'completions/bibleit.bash', 'completions/_bibleit', 'completions/bibleit.fish'}
        assert files['LICENSE'] == (ROOT / 'LICENSE').read_bytes()
        assert files['THIRD_PARTY_NOTICES'] == (ROOT / 'THIRD_PARTY_NOTICES').read_bytes()
        assert files['README.md'] == (ROOT / 'README.md').read_bytes()
        assert b'bibleit' + b'-server' not in files[binary], 'private service name embedded in executable'
    for entry in json.loads((assets / 'bibleit.json').read_text())['architecture'].values():
        assert hashes[entry['url'].split('/')[-1]] == entry['hash']
    formula_path = assets / 'bibleit-cli.rb'
    if not formula_path.exists():
        formula_path = assets / 'bibleit.rb'  # Previously generated local candidates.
    formula = formula_path.read_text()
    for name, digest in hashes.items():
        if name.endswith('.tar.gz'):
            assert name in formula and digest in formula
    with tempfile.TemporaryDirectory() as scratch:
        scratch = Path(scratch)
        copied = scratch / 'assets'
        shutil.copytree(assets, copied)
        destination = scratch / 'installed'
        if os.name == 'nt':
            installer = shutil.which(args.powershell) if args.powershell else shutil.which('pwsh') or shutil.which('powershell')
            assert installer, 'PowerShell required on Windows'
            def command(requested):
                return [installer, '-NoProfile', '-File', str(ROOT / 'install.ps1'), '-Version', requested, '-AssetsDir', str(copied), '-InstallDir', str(destination)]
            binary = destination / 'bibleit.exe'
            env = dict(os.environ)
            # CI launches Python from PowerShell 7. Let the selected runtime
            # initialize its own module paths, especially Windows PowerShell 5.1.
            for key in list(env):
                if key.lower() == 'psmodulepath':
                    del env[key]
        else:
            def command(requested):
                return ['bash', str(ROOT / 'install.sh'), requested]
            binary = destination / 'bibleit'
            env = dict(os.environ, BIBLEIT_INSTALL_ASSETS=str(copied), BIBLEIT_INSTALL_DIR=str(destination))
        def install(requested, success):
            result = subprocess.run(command(requested), env=env, text=True, capture_output=True)
            assert (result.returncode == 0) == success, result.stdout + result.stderr
        install(version, True)
        assert subprocess.check_output([str(binary), 'version'], text=True).strip() == f'bibleit CLI version {version[1:]}'
        # Exercise the shipped executable with disposable credentials, including
        # the legacy configuration that prompted this release.
        config = scratch / 'legacy-config.json'
        config.write_text('{"version":0,"endpoint":"https://legacy.example"}\n')
        original_config = config.read_bytes()
        ux_env = dict(env, BIBLEIT_CONFIG=str(config))
        help_result = subprocess.run([str(binary), '--help'], env=ux_env, text=True, capture_output=True, timeout=30)
        assert help_result.returncode == 0 and not help_result.stderr
        assert 'Bibleit CLI: read Scripture' in help_result.stdout
        for format in ('table', 'json'):
            result = subprocess.run([str(binary), 'auth', '--format', format], env=ux_env, text=True, capture_output=True, timeout=30)
            assert result.returncode == 0 and not result.stderr, result.stdout + result.stderr
            assert 'bibleit auth login' in result.stdout
            assert config.read_bytes() == original_config, 'auth help modified saved credentials'
        for notice in ('LICENSE', 'THIRD_PARTY_NOTICES'):
            assert (destination / 'bibleit-licenses' / notice).read_bytes() == (ROOT / notice).read_bytes()
        previous = binary.read_bytes()
        install(version, True)  # Upgrade/replacement path.
        assert binary.read_bytes() == previous
        if os.name != 'nt':
            # Exercise the online/latest branch offline, including HTTPS flags and
            # one concrete release tag for both archive and checksum requests.
            tools = scratch / 'tools'
            tools.mkdir()
            curl = tools / 'curl'
            curl.write_text('''#!/usr/bin/env python3
import os, pathlib, shutil, sys
args = sys.argv[1:]
assert args[args.index('--proto') + 1] == '=https'
assert args[args.index('--proto-redir') + 1] == '=https'
url = next(value for value in args if value.startswith('https://'))
repo = 'https://github.com/bibleit-labs/bibleit-cli'
if '-w' in args:
    assert url == repo + '/releases/latest'
    print(repo + '/releases/tag/' + os.environ['BIBLEIT_TEST_VERSION'], end='')
else:
    prefix = repo + '/releases/download/' + os.environ['BIBLEIT_TEST_VERSION'] + '/'
    assert url.startswith(prefix)
    name = url[len(prefix):]
    assert '/' not in name
    shutil.copyfile(pathlib.Path(os.environ['BIBLEIT_TEST_ASSETS']) / name, args[args.index('-o') + 1])
''')
            curl.chmod(0o755)
            online_env = dict(env, PATH=str(tools) + os.pathsep + env['PATH'], BIBLEIT_TEST_ASSETS=str(copied), BIBLEIT_TEST_VERSION=version)
            online_env.pop('BIBLEIT_INSTALL_ASSETS')
            subprocess.run(['bash', str(ROOT / 'install.sh')], env=online_env, check=True, capture_output=True)
            assert binary.read_bytes() == previous
        install('../invalid', False)
        (copied / 'checksums.txt').write_text('')
        install(version, False)
        assert binary.read_bytes() == previous
        checksums = (assets / 'checksums.txt').read_text()
        (copied / 'checksums.txt').write_text(checksums + checksums)
        install(version, False)
        assert binary.read_bytes() == previous
        (copied / 'checksums.txt').write_text(checksums)
        # A correctly checksummed binary without its redistribution notices
        # must be rejected before replacing an existing installation.
        system = platform.system().lower()
        changed = {}
        for path in copied.glob(f'*_{system}_*' + ('.zip' if os.name == 'nt' else '.tar.gz')):
            if os.name == 'nt':
                with zipfile.ZipFile(path) as source:
                    entries = [(entry, source.read(entry)) for entry in source.infolist() if entry.filename != 'THIRD_PARTY_NOTICES']
                with zipfile.ZipFile(path, 'w') as target:
                    for entry, data in entries:
                        target.writestr(entry, data)
            else:
                with tarfile.open(path) as source:
                    entries = [(entry, source.extractfile(entry).read()) for entry in source.getmembers() if entry.name != 'THIRD_PARTY_NOTICES']
                with tarfile.open(path, 'w:gz') as target:
                    for entry, data in entries:
                        target.addfile(entry, io.BytesIO(data))
            changed[path.name] = hashlib.sha256(path.read_bytes()).hexdigest()
        assert changed
        (copied / 'checksums.txt').write_text(''.join(f'{changed.get(name, digest)}  {name}\n' for name, digest in hashes.items()))
        install(version, False)
        assert binary.read_bytes() == previous
        for name in changed:
            shutil.copyfile(assets / name, copied / name)
        (copied / 'checksums.txt').write_text(checksums)
        for path in copied.glob('*.tar.gz' if os.name != 'nt' else '*.zip'):
            with path.open('ab') as out:
                out.write(b'corruption')
        install(version, False)
        assert binary.read_bytes() == previous
    report = ROOT / 'build/validation' / ('installer-' + (args.powershell or 'default') + '.json')
    report.parent.mkdir(parents=True, exist_ok=True)
    report.write_text(json.dumps({
        'version': version, 'commit': manifest['commit'],
        'system': platform.system(), 'architecture': architecture,
        'expected_runner': expected_runner, 'powershell': args.powershell,
        'binary_sha256': hashlib.sha256(previous).hexdigest(),
        'checks': ['archive_layouts', 'exact_readme_and_notices', 'checksums',
                   'native_install_and_replace', 'root_help', 'bare_auth_table_and_json',
                   'legacy_config_preserved', 'invalid_version', 'missing_notices',
                   'missing_and_duplicate_checksums', 'corrupt_archives'],
        'result': 'passed', 'local_archive_fixture': True,
    }, indent=2) + '\n')
    print('PASS: native installation/replacement, release contents/checksums, auth help with unchanged legacy config, notice and corrupt-archive rejection; report ' + str(report))


if __name__ == '__main__':
    main()
