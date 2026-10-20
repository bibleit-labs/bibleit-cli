#!/usr/bin/env python3
"""Opt-in production command qualification using individually owned test Lives.

Requires explicit operator authorization. Collection deletion is allowed only
when the initial scope was empty and every current Live belongs to this run.
Never revokes an existing login or removes an initially enabled translation. Captured
responses stay in memory; evidence contains templates and outcomes only.
"""
import argparse
import collections
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import tempfile
import time
import uuid


class CheckFailure(Exception):
    pass


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--report', type=Path, required=True)
    parser.add_argument('--authorize-production-mutations', action='store_true', required=True)
    parser.add_argument('--credentials', type=Path, help='Dedicated test credential document; contents are never logged.')
    parser.add_argument('--ssh-identity', type=Path, help='Matching registered private key; contents are never logged.')
    parser.add_argument('--live-transport', choices=['http', 'ssh'], default='http')
    parser.add_argument('--skip-resource-mutations', action='store_true', help='Run independent checks while test capacity is unavailable.')
    args = parser.parse_args()
    if args.live_transport == 'ssh' and not args.ssh_identity:
        parser.error('--live-transport ssh requires --ssh-identity')
    binary = str(args.binary.resolve())
    rows, owned = [], set()
    initial_ids = None
    initial_library = None
    added_translation = None
    marker = 'CLI qualification ' + uuid.uuid4().hex[:12]
    syntax = re.findall(r'^\s*\{"([^"\n]+)",', (Path(__file__).resolve().parents[1] / 'registry.go').read_text(), re.M)

    with tempfile.TemporaryDirectory(prefix='bibleit-production-commands-') as directory:
        env = dict(os.environ, BIBLEIT_PREFERENCES=str(Path(directory) / 'config.yaml'))
        if args.credentials:
            env['BIBLEIT_CREDENTIALS'] = str(args.credentials.resolve())
            env.pop('BIBLEIT_TOKEN', None)
        last_call = 0.0
        last_ssh = 0.0

        def invoke(command, label, *, transport='http', expected_error=None, yes=True, extra=(), verify=None):
            nonlocal last_call, last_ssh
            # Qualify behavior without exhausting production's request budget.
            delay = max(0, 1.1 - (time.monotonic() - last_call))
            if transport == 'ssh':
                # Production admits 20 new SSH connections per IP per minute.
                delay = max(delay, 3.3 - (time.monotonic() - last_ssh))
            time.sleep(delay)
            last_call = time.monotonic()
            if transport == 'ssh':
                last_ssh = last_call
            row = {'command': label, 'transport': transport, 'status': 'failed'}
            rows.append(row)
            argv = [binary, '--transport', transport, '--format', 'json']
            if transport == 'ssh' and args.ssh_identity:
                argv += ['--identity', str(args.ssh_identity.expanduser().resolve())]
            if yes:
                argv.append('--yes')
            try:
                result = subprocess.run([*argv, *extra, *command], env=env, capture_output=True, text=True, timeout=45, stdin=subprocess.DEVNULL)
                row['exit_code'] = result.returncode
                value = json.loads(result.stderr if result.returncode else result.stdout)
                if expected_error:
                    if result.returncode == 0 or value.get('ok') is not False or value.get('error', {}).get('code') != expected_error:
                        raise CheckFailure('expected_error_not_returned')
                    row['error_code'] = expected_error
                    data = None
                else:
                    if result.returncode:
                        row['error_code'] = value.get('error', {}).get('code', 'unknown')
                        raise CheckFailure('command_failed')
                    if transport == 'http':
                        if value.get('ok') is not True or value.get('schema_version') != 2:
                            raise CheckFailure('invalid_output_envelope')
                        data = value['data']
                    else:
                        if value.get('schema_version') != 1 or not isinstance(value.get('records'), list):
                            raise CheckFailure('invalid_ssh_output_envelope')
                        data = value['records']
                    if verify and not verify(data):
                        raise CheckFailure('domain_readback_failed')
                row['status'] = 'passed'
                return data
            except (ValueError, subprocess.TimeoutExpired):
                row['reason'] = 'invalid_json_or_timeout; private output suppressed'
                raise CheckFailure(row['reason']) from None
            except CheckFailure as error:
                row['reason'] = str(error)
                raise

        def attempt(action):
            try:
                return action()
            except CheckFailure:
                return None

        def call(*command, label=None, **kwargs):
            return invoke(list(command), label or ' '.join(command[:2]), **kwargs)

        def live_ids():
            return {item['id'] for item in call('live', 'list')['data']}

        def library():
            return {item['slug'].lower() for item in call('translation', 'list')['data']}

        def create(name):
            data = call('live', 'create', *([name] if name is not None else []), label='live create [name...]', transport=args.live_transport)
            if args.live_transport == 'ssh':
                data = data[0]['fields']
            identifier = data.get('id')
            if not identifier:
                raise CheckFailure('created_live_id_missing')
            owned.add(identifier)
            return identifier

        def live(identifier, *command, **kwargs):
            if identifier not in owned:
                raise CheckFailure('resource_is_not_owned_by_this_test')
            label = 'live <id> ' + ' '.join(command[:2])
            verify = kwargs.pop('verify', None)
            result = call('live', identifier, *command, label=label, transport=args.live_transport,
                          verify=verify if args.live_transport == 'http' else None, **kwargs)
            if verify and args.live_transport == 'ssh':
                return call('live', identifier, *command, label=label + ' HTTP domain readback', verify=verify)
            return result

        # Preferences and completion use a private preferences file, never the
        # operator's saved preferences. Offline commands have native local output.
        for command in [['help'], ['help', 'read'], ['help', 'live', 'stack', 'pop'], ['version'],
                        ['--version'], ['-v'], ['completion', 'bash'], ['completion', 'zsh'],
                        ['completion', 'fish'], ['config', 'path'], ['config', 'show'],
                        ['config', 'get', 'transport'], ['config', 'set', 'transport', 'http'],
                        ['config', 'unset', 'identity']]:
            result = subprocess.run([binary, *command], env=env, capture_output=True, text=True,
                                    timeout=10, stdin=subprocess.DEVNULL)
            rows.append({'command': ' '.join(command), 'transport': 'local',
                         'status': 'passed' if result.returncode == 0 else 'failed', 'exit_code': result.returncode})
        attempt(lambda: call('live', 'stack', 'info', label='missing Live ID returns usage guidance', expected_error='usage'))
        attempt(lambda: call('live', 'synthetic-nonexistent-id', 'info', label='missing Live returns safe not_found', expected_error='not_found'))
        attempt(lambda: call('live', 'synthetic-nonexistent-id', 'secret', 'synthetic-legacy-secret',
                            label='live <id> secret <value>', expected_error='usage'))

        try:
            initial_ids = live_ids()
            initial_library = library()
            if not initial_library:
                raise CheckFailure('enabled_translation_required')
            slug = 'bpml' if 'bpml' in initial_library else sorted(initial_library)[0]
            for command in [['ping'], ['server', 'info'], ['server', 'help'], ['server', 'help', 'live'],
                            ['server', 'help', 'account'], ['auth', 'info'], ['auth', 'whoami'],
                            ['account', 'info'], ['account', 'quotas'], ['account', 'tokens'],
                            ['account', 'quota', 'list'], ['account', 'token', 'list'],
                            ['organization', 'list'], ['organization', 'list', 'o:00000000-0000-0000-0000-000000000000']]:
                attempt(lambda command=command: invoke(command, ' '.join(command[:3])))
            for action in ['info', 'catalog']:
                metadata = attempt(lambda action=action: call('translation', action, slug, label=f'translation {action} <translation>'))
                if action == 'catalog':
                    named_book = next((x['name'] for x in (metadata or {}).get('data', []) if str(x.get('book')) == '31'), '31')
            organizations = call('organization', 'list')
            application = next((x.get('application_id') for x in organizations.get('data', []) if x.get('application_id')), None)
            if application:
                attempt(lambda: call('organization', 'application', application, 'status',
                                    label='organization application <id> status'))
            source = call('read', slug, '31', '1', '1', label='derive bounded search query from source verse')
            words = re.findall(r'[^\W\d_]{4,}', source['verses'][0]['text'], flags=re.UNICODE)
            if not words:
                raise CheckFailure('source_verse_has_no_search_word')
            query = words[0]
            if args.ssh_identity:
                commands = [['ping'], ['server', 'info'], ['server', 'help'], ['server', 'help', 'live'],
                            ['server', 'help', 'account'], ['auth', 'info'], ['auth', 'whoami'],
                            ['account', 'info'], ['account', 'quotas'], ['account', 'tokens'],
                            ['account', 'quota', 'list'], ['account', 'token', 'list'],
                            ['organization', 'list'], ['organization', 'list', 'o:00000000-0000-0000-0000-000000000000'],
                            ['translation', 'list'], ['live', 'list'],
                            ['translation', 'info', slug], ['translation', 'catalog', slug],
                            ['read', slug, named_book], ['read', slug, named_book, '1'],
                            ['read', slug, named_book, '1', '1'], ['read', slug, named_book, '1:1'], ['read', slug, '31', '1', '1'],
                            ['search', slug, query]]
                if application:
                    commands.append(['organization', 'application', application, 'status'])
                for command in commands:
                    # Labels omit resource identifiers, translation slugs and queries.
                    label = ' '.join(command[:2])
                    if command[:2] == ['organization', 'application']:
                        label = 'organization application <id> status'
                    if command[0] in ['read', 'search']:
                        label = command[0] + ' <translation> <reference-or-query>'
                    attempt(lambda command=command, label=label: invoke(command, label, transport='ssh'))
            for reference in [[named_book], [named_book, '1'], [named_book, '1', '1'], [named_book, '1:1'], ['31', '1', '1']]:
                attempt(lambda reference=reference: call('read', slug, *reference,
                    label='read <translation> <book> [chapter] [verse]', verify=lambda data: bool(data.get('verses'))))
            attempt(lambda: call('search', slug, query, label='search <translation> <query>', verify=lambda d: bool(d.get('verses'))))
            if args.skip_resource_mutations:
                raise CheckFailure('resource_mutations_deferred_until_test_capacity_is_available')
            # Library writes target an initially disabled public-domain translation only.
            candidate = next((s for s in ['web', 'kjv'] if s not in initial_library), None)
            if candidate:
                added_translation = candidate  # Cleanup also covers ambiguous process failure.
                if attempt(lambda: call('translation', 'add', candidate, label='translation add <translation>', transport=args.live_transport)) is not None:
                    attempt(lambda: call('translation', 'list', verify=lambda data: candidate in {x['slug'].lower() for x in data['data']}))
                    attempt(lambda: call('translation', 'remove', candidate, label='translation remove <translation>', transport=args.live_transport))
                    if candidate not in library():
                        added_translation = None

            for name in [None, marker]:
                identifier = create(name)
                if name is None:
                    live(identifier, 'delete')
                    owned.remove(identifier)
                    continue
                attempt(lambda: live(identifier, 'info', verify=lambda d: d.get('name') == marker))
                attempt(lambda: live(identifier, 'stats'))
                for action, expected in [('start', 'running'), ('stop', 'stopped'), ('start', 'running')]:
                    attempt(lambda action=action: live(identifier, action))
                    attempt(lambda expected=expected: live(identifier, 'info', verify=lambda d: d.get('status') == expected))
                for action, paused in [('pause', True), ('resume', False)]:
                    attempt(lambda action=action: live(identifier, action))
                    attempt(lambda paused=paused: live(identifier, 'info', verify=lambda d: d.get('paused') is paused))
                for key, value in [('name', marker + ' updated'), ('reference', 'Obadiah 1:1'), ('translations', slug)]:
                    attempt(lambda key=key, value=value: live(identifier, 'set', key, value))
                attempt(lambda: live(identifier, 'info', verify=lambda d: d.get('name') == marker + ' updated'))
                for reference in [[slug, named_book, '1', '1'], [named_book], [named_book, '1'], [named_book, '1', '1'], [named_book, '1:1'], [slug, '31', '1', '1'], [slug, '31', '1:1']]:
                    attempt(lambda reference=reference: live(identifier, 'stack', 'push', *reference))
                stack = attempt(lambda: live(identifier, 'stack', 'info', verify=lambda d: len(d.get('data', [])) >= 7))
                count = len(stack['data']) if stack else None
                attempt(lambda: live(identifier, 'stack', 'pop', yes=False))
                if count is not None:
                    attempt(lambda: live(identifier, 'stack', 'info', verify=lambda d: len(d.get('data', [])) == count - 1))
                attempt(lambda: live(identifier, 'stack', 'pop', '2', yes=False))
                if count is not None:
                    attempt(lambda: live(identifier, 'stack', 'info', verify=lambda d: len(d.get('data', [])) == count - 3))
                attempt(lambda: live(identifier, 'stack', 'clear'))
                attempt(lambda: live(identifier, 'stack', 'info', verify=lambda d: not d.get('data')))
                attempt(lambda: live(identifier, 'stack', 'push', slug, '31', '1', '1'))
                attempt(lambda: live(identifier, 'clear'))
                attempt(lambda: live(identifier, 'stack', 'info', verify=lambda d: not d.get('data')))
                for action in ['create', 'rotate']:
                    secret_path = Path(directory) / ('audience-' + action)
                    attempt(lambda action=action, secret_path=secret_path: live(identifier, 'secret', action,
                        extra=['--secret-output', str(secret_path)]))
                    valid = secret_path.exists() and secret_path.stat().st_mode & 0o777 == 0o600 and len(secret_path.read_text().strip()) == 48
                    rows.append({'command': 'protected audience secret delivery', 'transport': 'local', 'status': 'passed' if valid else 'failed'})
                attempt(lambda: live(identifier, 'secret', 'delete'))
                attempt(lambda: live(identifier, 'secret', 'synthetic-legacy-secret', expected_error='usage'))
                # Exercise real deployed long polling, a changed snapshot, and SIGINT.
                for fmt in ['ndjson', 'table']:
                    with tempfile.TemporaryFile() as output, tempfile.TemporaryFile() as errors:
                        argv = [binary, '--transport', args.live_transport, '--format', fmt]
                        if args.live_transport == 'ssh':
                            argv += ['--identity', str(args.ssh_identity.expanduser().resolve())]
                        process = subprocess.Popen([*argv, 'live', identifier, 'subscribe'],
                                                   env=env, stdout=output, stderr=errors, stdin=subprocess.DEVNULL)
                        try:
                            time.sleep(2)
                            attempt(lambda: live(identifier, 'pause'))
                            time.sleep(2)
                            process.send_signal(signal.SIGINT)
                            process.wait(timeout=5)
                            output.seek(0)
                            raw = output.read()
                            if fmt == 'ndjson':
                                events = [json.loads(line) for line in raw.splitlines()]
                                version = 2 if args.live_transport == 'http' else 1
                                valid = len(events) >= 2 and all(e.get('schema_version') == version for e in events)
                            else:
                                heading = b'Watching Live snapshots.' if args.live_transport == 'http' else b'Watching Live changes.'
                                valid = heading in raw
                                if args.live_transport == 'http':
                                    valid = valid and b'paused=true' in raw
                            rows.append({'command': 'live <id> subscribe', 'transport': args.live_transport, 'format': fmt,
                                         'status': 'passed' if valid and process.returncode == 0 else 'failed'})
                        finally:
                            if process.poll() is None:
                                process.kill()
                                process.wait()
                    attempt(lambda: live(identifier, 'resume'))
                if not initial_ids and live_ids() == owned:
                    call('live', 'delete', 'all', label='live delete all', transport=args.live_transport)
                    owned.clear()
                else:
                    live(identifier, 'delete')
                    owned.remove(identifier)
                attempt(lambda: call('live', 'list', verify=lambda d: identifier not in {x['id'] for x in d['data']}))
        except CheckFailure as error:
            rows.append({'command': 'campaign prerequisites', 'transport': 'http', 'status': 'blocked', 'reason': str(error)})
        finally:
            for identifier in list(owned):
                try:
                    live(identifier, 'delete')
                    owned.remove(identifier)
                except CheckFailure:
                    pass
            if added_translation and initial_library is not None and added_translation not in initial_library:
                attempt(lambda: call('translation', 'remove', added_translation, label='translation remove <test-translation> cleanup'))
            if initial_ids is not None:
                attempt(lambda: call('live', 'list', label='initial Live resources preserved; no test Lives remain',
                                    verify=lambda d: {x['id'] for x in d['data']} == initial_ids))
            if initial_library is not None:
                attempt(lambda: call('translation', 'list', label='initial translation library restored',
                                    verify=lambda d: {x['slug'].lower() for x in d['data']} == initial_library))

        reasons = {
            'live delete all': 'Requires dedicated isolated production account; existing Lives must be preserved.',
            'auth login [ssh]': 'New browser consent or a registered production SSH identity is required.',
            'auth logout': 'Requires a newly issued test login; existing user credential must remain valid.',
            'organization application <id> status': 'Requires an existing organization application owned by the test account.',
        }
        if any(row['command'] == 'organization application <id> status' and row['status'] == 'passed' for row in rows):
            reasons.pop('organization application <id> status')
        if any(row['command'] == 'live delete all' and row['status'] == 'passed' for row in rows):
            reasons.pop('live delete all')
        if any(row.get('error_code') == 'quota_exceeded' for row in rows):
            reasons['Live lifecycle, settings, stack, secrets, subscription'] = 'Test Live creation blocked by account quota_exceeded.'
        if args.skip_resource_mutations:
            reasons['resource mutations'] = 'Deferred in this independent-check run; see the separate mutation campaign evidence.'
        counts = dict(collections.Counter(row['status'] for row in rows))
        report = {'environment': 'production HTTP/SSH; existing personal credential; individually owned temporary test resources',
                  'production_mutations_authorized': True, 'checks': rows, 'counts': counts,
                  'unexercised_success_paths': reasons, 'all_commands_passed': False,
                  'registered_command_syntaxes': syntax, 'remaining_test_lives': len(owned),
                  'credentials_or_user_content_recorded': False, 'homebrew_modified': False}
        args.report.parent.mkdir(parents=True, exist_ok=True)
        args.report.write_text(json.dumps(report, indent=2) + '\n')
        print(json.dumps({'counts': counts, 'remaining_test_lives': len(owned), 'all_commands_passed': False}))
        return 1 if counts.get('failed') or owned else (2 if counts.get('blocked') else 0)


if __name__ == '__main__':
    raise SystemExit(main())
