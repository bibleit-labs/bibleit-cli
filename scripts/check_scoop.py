#!/usr/bin/env python3
"""Native Windows Scoop checks using a disposable root and loopback archives."""
import argparse
import functools
import http.server
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import threading


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--assets', required=True)
    parser.add_argument('--scoop-source', required=True, help='Existing Scoop source checkout; bin/scoop.ps1 must exist')
    parser.add_argument('--report', default='build/validation/scoop.json')
    args = parser.parse_args()
    if os.name != 'nt':
        parser.error('native Windows is required; Linux PowerShell is not Scoop validation')
    assets = Path(args.assets).resolve()
    source = Path(args.scoop_source).resolve()
    if not (source / 'bin/scoop.ps1').is_file():
        parser.error('invalid Scoop source checkout')
    powershell = shutil.which('pwsh') or shutil.which('powershell')
    if not powershell:
        parser.error('PowerShell is required')
    evidence = []
    class Handler(http.server.SimpleHTTPRequestHandler):
        def log_message(self, *args):
            pass
    with http.server.ThreadingHTTPServer(('127.0.0.1', 0), functools.partial(Handler, directory=str(assets))) as server:
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            with tempfile.TemporaryDirectory(prefix='bibleit-scoop-') as scratch:
                scratch = Path(scratch)
                root = scratch / 'scoop'
                current = root / 'apps/scoop/current'
                shutil.copytree(source, current, ignore=shutil.ignore_patterns('.git'))
                (root / 'buckets').mkdir()
                # Portable config and application directories are all temporary.
                (root / 'config.json').write_text(json.dumps({'root_path': str(root), 'global_path': str(scratch / 'global'), 'cache_path': str(scratch / 'cache'), 'use_isolated_path': True, 'aria2-enabled': False, 'use_external_7zip': False, 'use_sqlite_cache': False}))
                env = dict(os.environ, SCOOP=str(root), SCOOP_GLOBAL=str(scratch / 'global'), SCOOP_CACHE=str(scratch / 'cache'))
                manifest = json.loads((assets / 'bibleit.json').read_text())
                version = manifest['version']
                # Only URLs change; both architecture hashes and shim declaration
                # are the real generated manifest. Downloads remain hash-checked.
                for item in manifest['architecture'].values():
                    name = item['url'].split('/')[-1]
                    item['url'] = f'http://127.0.0.1:{server.server_port}/{name}'
                fixture = scratch / 'bibleit.json'
                fixture.write_text(json.dumps(manifest, indent=2))
                scoop = current / 'bin/scoop.ps1'
                def run(*command, succeeds=True):
                    result = subprocess.run([powershell, '-NoProfile', '-File', str(scoop), *command], env=env, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
                    evidence.append({'command': ['scoop', *command], 'exit_code': result.returncode, 'output': result.stdout})
                    if succeeds is not None and (result.returncode == 0) != succeeds:
                        raise RuntimeError(result.stdout)
                    if succeeds is not None:
                        print(f'PASS: scoop {" ".join(command)}', flush=True)
                    return result.stdout
                def verify_binary():
                    shim = root / 'shims/bibleit.exe'
                    assert shim.is_file(), 'Scoop executable shim missing'
                    reported = subprocess.check_output([str(shim), 'version'], env=env, text=True).strip()
                    assert reported == f'bibleit CLI version {version}', reported
                    assert (root / 'apps/bibleit/current/LICENSE').is_file()
                run('install', '--no-update-scoop', '--independent', str(fixture))
                verify_binary()
                run('reset', 'bibleit')
                verify_binary()
                run('uninstall', 'bibleit')
                assert not (root / 'shims/bibleit.exe').exists()
                run('install', '--no-update-scoop', '--independent', str(fixture))
                verify_binary()
                run('uninstall', 'bibleit')
                assert not (root / 'apps/bibleit').exists()
                # A wrong archive digest must fail before executable/shim creation.
                for item in manifest['architecture'].values():
                    item['hash'] = '0' * 64
                fixture.write_text(json.dumps(manifest, indent=2))
                # Scoop can return zero after an install_app hash failure inside
                # its pipeline. Assert the rejection and durable state directly.
                output = run('install', '--no-update-scoop', '--independent', '--no-cache', str(fixture), succeeds=None)
                assert 'hash check failed' in output.lower(), output
                assert not (root / 'shims/bibleit.exe').exists()
                assert not (root / 'apps/bibleit/current').exists()
                assert not list((root / 'apps/bibleit').rglob('bibleit.exe'))
                print('PASS: Scoop rejected incorrect hash without installing an executable or shim', flush=True)
        finally:
            server.shutdown()
            thread.join()
    report = Path(args.report)
    report.parent.mkdir(parents=True, exist_ok=True)
    report.write_text(json.dumps({'platform': 'Windows', 'architecture': os.environ.get('PROCESSOR_ARCHITECTURE'), 'version': version, 'isolated_root': True, 'loopback_assets': True, 'checks': evidence}, indent=2) + '\n')
    print(f'PASS: native Scoop installation/shim/reset/reinstall/uninstall/checksum rejection; report {report}')


if __name__ == '__main__':
    main()
