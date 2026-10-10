#!/usr/bin/env python3
"""Run fixed-endpoint integration in a disposable network namespace, without host ports."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
import time
import uuid

ROOT = Path(__file__).resolve().parents[1]


def run(command, **kwargs):
    return subprocess.run(command, check=True, text=True, **kwargs)


def copy_sources(source, target):
    for name in run(['git', '-C', str(source), 'ls-files', '-c', '-o', '--exclude-standard'], capture_output=True).stdout.splitlines():
        path = source / name
        if path.is_file() and not path.is_symlink():
            destination = target / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(path, destination)


def fingerprints(source):
    return {str(p.relative_to(source)): hashlib.sha256(p.read_bytes()).hexdigest()
            for directory in ('src', 'test') for p in (source / directory).rglob('*.erl')}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--server', type=Path, default=ROOT.parent / 'private-service')
    parser.add_argument('--go-client', type=Path, default=ROOT.parent / 'bibleit-go')
    parser.add_argument('--baseline', type=Path)
    args = parser.parse_args()
    server = args.server.resolve()
    with tempfile.TemporaryDirectory(prefix='bibleit-isolated-integration-') as directory:
        temporary = Path(directory)
        sources = temporary / 'sources'
        sources.mkdir()
        copy_sources(ROOT, sources / 'bibleit-cli')
        copy_sources(args.go_client.resolve(), sources / 'bibleit-go')
        (sources / 'bibleit-cli/go.work').write_text('go 1.27.1\n\nuse (\n .\n ../bibleit-go\n)\n')
        before = fingerprints(server)
        snapshot = sources / 'private-service'
        # No .env, .git, databases, keys, uploaded data, or host-native binaries.
        for relative in ('src', 'test', 'c_src', 'libbibleit/src', 'libbibleit/include',
                         'libbibleit/config', 'priv/config', 'priv/sql', 'priv/i18n',
                         'priv/templates', 'priv/static', '_build/default/lib'):
            source = server / relative
            shutil.copytree(source, snapshot / relative, ignore=lambda folder, names: [
                name for name in names if (Path(folder) / name).is_symlink()
                or name.endswith(('.o', '.a', '.so', '.so.1', '.dylib'))
                or name in ('.git', '.env')])
        shutil.copy2(server / 'Makefile', snapshot / 'Makefile')
        shutil.copy2(server / 'libbibleit/Makefile', snapshot / 'libbibleit/Makefile')
        if before != fingerprints(server) or before != fingerprints(snapshot):
            raise RuntimeError('server source changed during snapshot; rerun')
        metadata = {
            'head': run(['git', '-C', str(server), 'rev-parse', 'HEAD'], capture_output=True).stdout.strip(),
            'working_tree_changes': bool(run(['git', '-C', str(server), 'status', '--porcelain'], capture_output=True).stdout.strip()),
        }
        (sources / 'server-metadata.json').write_text(json.dumps(metadata))
        build = temporary / 'tools'
        build.mkdir()
        (build / 'Dockerfile').write_text('''FROM golang:1.27.1-bookworm AS go
FROM erlang:29-alpine
RUN apk add --no-cache build-base python3 openssh-client
COPY --from=go /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:$PATH" GOTOOLCHAIN=local GOCACHE=/tmp/go-cache GOMODCACHE=/tmp/go-modcache
''')
        image = 'bibleit-client-integration:local'
        run(['docker', 'build', '-t', image, str(build)])
        # Both processes share this disposable namespace. localhost:8080 belongs
        # only to the fixture; the developer's localhost:8080 is never reachable.
        name = 'bibleit-client-db-' + uuid.uuid4().hex[:12]
        started = False
        results = temporary / 'results'
        results.mkdir()
        try:
            run(['docker', 'run', '--pull=never', '--detach', '--name', name,
                 '--tmpfs', '/var/lib/postgresql/data', '--env', 'POSTGRES_USER=bibleit_test',
                 '--env', 'POSTGRES_PASSWORD=bibleit-test-only', '--env', 'POSTGRES_DB=bibleit_test',
                 'postgres:17-alpine'], stdout=subprocess.DEVNULL)
            started = True
            for _ in range(80):
                ready = subprocess.run(['docker', 'exec', name, 'pg_isready', '-h', '127.0.0.1', '-U', 'bibleit_test', '-d', 'bibleit_test'],
                                       stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                if ready.returncode == 0:
                    break
                time.sleep(.25)
            else:
                raise RuntimeError('disposable database did not become ready')
            run(['docker', 'run', '--rm', '--network', 'container:' + name,
                 '--mount', f'type=bind,src={sources},dst=/sources,readonly',
                 '--mount', f'type=bind,src={results},dst=/results',
                 '--env', 'BIBLEIT_TEST_DATABASE_URL=postgresql://bibleit_test:bibleit-test-only@127.0.0.1:5432/bibleit_test?sslmode=disable',
                 image, 'python3', '/sources/bibleit-cli/scripts/prepare_server_container.py'])
            if args.baseline:
                shutil.copy2(results / 'server-baseline.json', args.baseline)
        finally:
            if started:
                run(['docker', 'rm', '--force', name], stdout=subprocess.DEVNULL)


if __name__ == '__main__':
    main()
