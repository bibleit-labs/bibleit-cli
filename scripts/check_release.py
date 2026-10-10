#!/usr/bin/env python3
"""Validate archive contents/checksums and exercise the native installer offline."""
import argparse
import hashlib
import json
import os
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
        assert set(files) == {binary, 'LICENSE', 'README.md', 'completions/bibleit.bash', 'completions/_bibleit', 'completions/bibleit.fish'}
        assert b'Permission is hereby granted' in files['LICENSE']
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
        for path in copied.glob('*.tar.gz' if os.name != 'nt' else '*.zip'):
            with path.open('ab') as out:
                out.write(b'corruption')
        install(version, False)
        assert binary.read_bytes() == previous
    print('PASS: six archive layouts, eight checksums, package manifests, native installation/replacement, invalid versions, missing/duplicate checksums and corrupt archives; Bash latest/download path verified with offline curl fixture where applicable.')


if __name__ == '__main__':
    main()
