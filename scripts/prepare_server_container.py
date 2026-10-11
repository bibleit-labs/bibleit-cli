#!/usr/bin/env python3
"""Build fixture NIFs for the container OS, then execute the real-server harness."""
from pathlib import Path
import argparse
import os
import socket
import threading
from fixture_tls import HOST, MATERIAL, Relay, verify_boundary
import shutil
import subprocess
import tempfile


def run(command, **kwargs):
    return subprocess.run(command, check=True, text=True, **kwargs)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--imports', action='store_true')
    parser.add_argument('--catalog-only', action='store_true')
    args = parser.parse_args()
    # Refuse a routed namespace before mapping the real service name. This
    # gateway is confined to the disposable Linux container, never the host.
    if not Path('/.dockerenv').exists() or any(line.split()[1] == '00000000' for line in Path('/proc/net/route').read_text().splitlines()[1:]):
        raise SystemExit('fixture requires an egress-disabled Docker namespace')
    with Path('/etc/hosts').open('a') as hosts:
        hosts.write('\n127.0.0.1 ' + HOST + '\n')
    if {item[4][0] for item in socket.getaddrinfo(HOST, 443)} != {'127.0.0.1'}:
        raise SystemExit('fixed service hostname must resolve exclusively to fixture loopback')
    os.environ['SSL_CERT_FILE'] = str(MATERIAL / 'server.pem')
    os.environ['SSL_CERT_DIR'] = '/tmp/bibleit-empty-trust'
    Path(os.environ['SSL_CERT_DIR']).mkdir(exist_ok=True)
    os.environ['BIBLEIT_FIXED_ENDPOINT_FIXTURE'] = '1'
    for name in ('HTTP_PROXY', 'HTTPS_PROXY', 'ALL_PROXY', 'http_proxy', 'https_proxy', 'all_proxy'):
        os.environ.pop(name, None)
    relay = Relay(443, 8080)
    threading.Thread(target=relay.serve_forever, daemon=True).start()
    verify_boundary(443)
    with tempfile.TemporaryDirectory(prefix='bibleit-linux-snapshot-') as directory:
        workspace = Path(directory) / 'workspace'
        shutil.copytree('/sources', workspace)
        server = workspace / 'private-service'
        erl_root = run(['erl', '-noshell', '-eval', 'io:format("~s",[code:root_dir()]),halt().'], capture_output=True).stdout
        erl_include = str(Path(erl_root) / 'usr/include')
        dependencies = server / '_build/default/lib'
        jargon = dependencies / 'jargon'
        argon = jargon / 'argon2'
        (jargon / 'priv').mkdir(exist_ok=True)
        argon_sources = [argon / 'src' / name for name in
                         ('argon2.c', 'core.c', 'blake2/blake2b.c', 'thread.c', 'encoding.c', 'ref.c')]
        run(['cc', '-shared', '-fPIC', '-O2', '-pthread', '-I' + erl_include,
             '-I' + str(argon / 'include'), '-I' + str(argon / 'src'),
             str(jargon / 'c_src/jargon.c'), *map(str, argon_sources),
             '-o', str(jargon / 'priv/jargon.so')])
        jiffy = dependencies / 'jiffy'
        objects = []
        for source in sorted((jiffy / 'c_src').rglob('*')):
            if source.suffix not in ('.c', '.cc'):
                continue
            output = source.with_suffix('.o')
            compiler = 'cc' if source.suffix == '.c' else 'c++'
            run([compiler, '-fPIC', '-O2', '-I' + erl_include, '-I' + str(jiffy / 'c_src'),
                 '-c', str(source), '-o', str(output)])
            objects.append(str(output))
        (jiffy / 'priv').mkdir(exist_ok=True)
        run(['c++', '-shared', *objects, '-o', str(jiffy / 'priv/jiffy.so')])
        run(['make', 'nif'], cwd=server)
        cli = workspace / 'bibleit-cli'
        if args.catalog_only:
            run(['python3', 'scripts/check_server_catalog.py',
                 '--source-metadata', str(workspace / 'server-metadata.json'),
                 '--report', '/results/server-baseline.json'], cwd=cli)
            return
        run(['python3', 'scripts/check_server.py', '--integration', '--external-disposable-database',
             '--source-metadata', str(workspace / 'server-metadata.json'),
             '--baseline', '/results/server-baseline.json',
             *(['--imports'] if args.imports else [])], cwd=cli)


if __name__ == '__main__':
    main()
