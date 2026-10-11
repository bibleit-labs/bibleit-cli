#!/usr/bin/env python3
"""Qualify the redirect-alias boot repair without repeating client integration."""
import argparse
import datetime
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

from check_server_container import fingerprints, native_fingerprints

ROOT = Path(__file__).resolve().parents[1]


def run(command, **kwargs):
    return subprocess.run(command, check=True, text=True, **kwargs)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--server', type=Path, default=ROOT.parent / 'bibleit-server')
    parser.add_argument('--revision', required=True, help='assessed committed repair')
    parser.add_argument('--previous-baseline', type=Path,
                        default=ROOT / 'contract/server-baseline.json')
    parser.add_argument('--report', type=Path, required=True)
    args = parser.parse_args()
    server = args.server.resolve()
    revision = run(['git', '-C', str(server), 'rev-parse', args.revision],
                   capture_output=True).stdout.strip()
    previous = json.loads(args.previous_baseline.read_text())
    before = fingerprints(server)
    native = native_fingerprints(server / 'libbibleit')
    native_head = run(['git', '-C', str(server / 'libbibleit'), 'rev-parse', 'HEAD'],
                      capture_output=True).stdout.strip()
    if (native != previous['native_library']['files_sha256']
            or native_head != previous['native_library']['head']):
        raise SystemExit('native inputs changed; broader qualification is required')
    source = {p: h for p, h in before.items() if p.startswith('src/')}
    changed = sorted(p for p in source.keys() | previous['files_sha256'].keys()
                     if source.get(p) != previous['files_sha256'].get(p))
    if changed != ['src/bibleit_server_env.erl']:
        raise SystemExit('runtime changes exceed the alias parser repair')
    # Inherited integration evidence is valid only for this exact one-line repair.
    old_parser = run(['git', '-C', str(server), 'show',
                      previous['head'] + ':src/bibleit_server_env.erl'],
                     capture_output=True).stdout
    expected = old_parser.replace(
        '     {<<"BIBLEIT_HTTP_TRUSTED_PROXY_IPS">>, str},\n',
        '     {<<"BIBLEIT_HTTP_TRUSTED_PROXY_IPS">>, str},\n'
        '     {<<"BIBLEIT_HTTP_REDIRECT_HOSTS">>, str},\n')
    if expected == old_parser or (server / changed[0]).read_text() != expected:
        raise SystemExit('runtime repair differs from the assessed parser registration')
    for path, digest in before.items():
        committed = subprocess.run(['git', '-C', str(server), 'show',
                                    revision + ':' + path], check=True,
                                   capture_output=True).stdout
        if hashlib.sha256(committed).hexdigest() != digest:
            raise SystemExit('working-tree Erlang input differs from assessed revision: ' + path)
    with tempfile.TemporaryDirectory(prefix='bibleit-alias-boot-') as directory:
        snapshot = Path(directory)
        for relative in ('src', 'test', '_build/default/lib/dotenv_config/ebin'):
            shutil.copytree(server / relative, snapshot / relative,
                            ignore=lambda folder, names: [n for n in names
                                if (Path(folder) / n).is_symlink()])
        if before != fingerprints(snapshot) or before != fingerprints(server):
            raise SystemExit('Erlang inputs changed during snapshot')
        # Only configure(), never start(): synthetic DB settings, no listener,
        # no .env, cleared child environment, empty cwd and no external route.
        evaluation = ('case eunit:test(bibleit_server_boot_tests,[verbose]) of '
                      'ok -> halt(0); _ -> halt(1) end.')
        command = ['sh', '-ec',
                   'mkdir /tmp/ebin; '
                   'erlc -pa /snapshot/_build/default/lib/dotenv_config/ebin '
                   '-o /tmp/ebin /snapshot/src/bibleit_server_env.erl '
                   '/snapshot/src/bibleit_server_boot.erl '
                   '/snapshot/src/http/bibleit_http_client_ip.erl '
                   '/snapshot/src/http/bibleit_http_canonical_host.erl '
                   '/snapshot/test/unit/bibleit_server_boot_tests.erl; '
                   'exec erl -noshell -pa /tmp/ebin '
                   '-pa /snapshot/_build/default/lib/dotenv_config/ebin '
                   '-eval "$1"', 'boot-check', evaluation]
        run(['docker', 'run', '--rm', '--pull=never', '--network', 'none',
             '--read-only', '--tmpfs', '/tmp', '--workdir', '/tmp',
             '--mount', f'type=bind,src={snapshot},dst=/snapshot,readonly',
             'bibleit-client-integration:local', *command])
    if before != fingerprints(server) or native != native_fingerprints(server / 'libbibleit'):
        raise SystemExit('qualified inputs changed during boot validation')
    report = {
        'reviewed_at': datetime.datetime.now(datetime.timezone.utc).date().isoformat(),
        'head': revision,
        'working_tree_changes': bool(run(['git', '-C', str(server), 'status', '--porcelain'],
                                         capture_output=True).stdout.strip()),
        'files_sha256': source,
        'test_files_sha256': {p: h for p, h in before.items() if p.startswith('test/')},
        'native_library': previous['native_library'],
        'verification': {
            'direct_boot_eunit_cases': 1,
            'configurations': ['normalized_configured_aliases', 'absent_aliases', 'wildcard_rejected'],
            'egress_disabled': True,
            'cleared_child_environment': True,
            'empty_child_working_directory': True,
            'integration_repeated': False,
            'previous_integration_head': previous['head'],
            'previous_baseline_sha256': hashlib.sha256(args.previous_baseline.read_bytes()).hexdigest(),
            'changed_runtime_files': changed,
        },
    }
    args.report.write_text(json.dumps(report, indent=2) + '\n')
    print('Direct boot regression passed; source/native equivalence verified; prior integration evidence retained.')


if __name__ == '__main__':
    main()
