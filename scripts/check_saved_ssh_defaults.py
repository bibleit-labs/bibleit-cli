#!/usr/bin/env python3
"""Qualify the installed CLI using saved SSH defaults and no connection flags.

Production writes require explicit authorization. Only this run's Lives and an
initially disabled translation are mutated. Auth/config writes use a copy of the
connection defaults, with no copied HTTP credential.
"""
import argparse
import collections
import hashlib
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import tempfile
import time
import uuid


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--report', type=Path, required=True)
    parser.add_argument('--authorize-production-mutations', action='store_true', required=True)
    args = parser.parse_args()
    binary = str(args.binary.resolve())
    rows, owned = [], set()
    env = dict(os.environ)
    last = 0.0
    initial_ids = initial_library = None
    added = None
    marker = 'SSH defaults test ' + uuid.uuid4().hex[:10]

    def check(label, valid, **fields):
        rows.append(dict(command=label, status='passed' if valid else 'failed', **fields))
        if not valid:
            print('FAIL: ' + label, flush=True)
        return valid

    def run(command, label, *, fmt='json', network=True, yes=False, error=None, test_env=None, extra=()):
        nonlocal last
        argv = [binary, '--format', fmt, *(['--yes'] if yes else []), *extra, *command]
        assert not any(a == '--transport' or a.startswith('--transport=') or a == '--identity' or a.startswith('--identity=') for a in argv)
        if network:
            time.sleep(max(0, 4.0 - (time.monotonic() - last)))
            last = time.monotonic()
        try:
            result = subprocess.run(argv, env=test_env or env, capture_output=True, text=True,
                                    stdin=subprocess.DEVNULL, timeout=40)
        except subprocess.TimeoutExpired:
            check(label, False, reason='timeout; private output suppressed')
            raise RuntimeError('command timed out') from None
        if error:
            if fmt == 'json':
                value = json.loads(result.stderr)
                valid = (result.returncode == error[0] and not result.stdout and value.get('schema_version') == 1
                         and value.get('ok') is False and value.get('error', {}).get('code') == error[1])
                message = value.get('error', {}).get('message', '')
            else:
                valid = result.returncode == error[0] and not result.stdout and result.stderr.startswith('bibleit: ')
                message = result.stderr.removeprefix('bibleit: ').strip()
            check(label, valid and bool(message), format=fmt, exit_code=result.returncode, expected_error=error[1])
            return message
        if not check(label, result.returncode == 0 and not result.stderr, format=fmt, exit_code=result.returncode):
            raise RuntimeError('CLI failure; private output suppressed')
        if fmt == 'json':
            value = json.loads(result.stdout)
            if not check(label + ' JSON envelope', value.get('schema_version') == 1 and value.get('ok') is True and isinstance(value.get('records'), list)):
                raise RuntimeError('invalid SSH JSON envelope')
            return value['records']
        return result.stdout

    def fields(records):
        return [x['fields'] for x in records]

    def one(records):
        return fields(records)[0]

    def lives():
        return {x['id'] for x in fields(run(['live', 'list'], 'live list')) if 'id' in x}

    def library():
        records = fields(run(['translation', 'list'], 'translation list'))
        result = set()
        for x in records:
            if isinstance(x.get('translations'), list):
                result.update(x['translations'])
            if 'slug' in x:
                result.add(x['slug'])
        return result

    def live(identifier, *command, **kwargs):
        assert identifier in owned
        return run(['live', identifier, *command], 'live <id> ' + ' '.join(command[:2]), **kwargs)

    def stack(identifier):
        records = fields(live(identifier, 'stack', 'info'))
        return next((int(x['count']) for x in records if 'count' in x), len([x for x in records if 'position' in x]))

    defaults = one(run(['config', 'show'], 'config show', network=False))
    pref_path = Path(defaults['path'])
    protected = [pref_path, Path.home() / '.bibleit/config.json']
    before = {str(p): hashlib.sha256(p.read_bytes()).hexdigest() if p.exists() else None for p in protected}
    assert defaults['transport'] == 'ssh' and defaults['transport_source'] == 'config.yaml'
    assert defaults['identity_source'] == 'config.yaml' and Path(defaults['identity']).expanduser() == Path.home() / '.ssh/bibleit'
    check('saved defaults selected from real config.yaml', True)

    with tempfile.TemporaryDirectory(prefix='bibleit-saved-ssh-defaults-') as scratch:
        scratch = Path(scratch)
        isolated_preferences = scratch / 'config.yaml'
        isolated_preferences.write_text('version: 2\ntransport: ssh\nidentity: ' + json.dumps(defaults['identity']) + '\n')
        isolated_preferences.chmod(0o600)
        isolated = dict(env, BIBLEIT_CONFIG=str(isolated_preferences), BIBLEIT_TOKEN='')
        try:
            for command in [['help'], ['help', 'read'], ['help', 'live', 'stack', 'pop'], ['version'], ['--version'], ['-v'],
                            ['completion', 'bash'], ['completion', 'zsh'], ['completion', 'fish'],
                            ['config', 'get', 'identity'], ['config', 'get', 'transport']]:
                text = run(command, ' '.join(command[:3]), fmt='table', network=False)
                check(' '.join(command[:2]) + ' nonempty output', bool(text.strip()))
            run(['config', 'set', 'format', 'json'], 'config set <key> <value>', network=False, test_env=isolated)
            run(['config', 'unset', 'format'], 'config unset <key>', network=False, test_env=isolated)
            check('config mutation tests retain copied connection defaults', 'transport: ssh' in isolated_preferences.read_text() and 'identity:' in isolated_preferences.read_text())
            for command in [['auth', 'login'], ['auth', 'login', 'ssh'], ['auth', 'info']]:
                run(command, ' '.join(command) + ' with copied defaults', test_env=isolated)
            run(['auth', 'logout'], 'auth logout isolated test login', network=False, test_env=isolated)
            check('logout cleared only temporary credentials', 'access_token:' not in isolated_preferences.read_text() and defaults['identity'] not in isolated_preferences.read_text())

            initial_ids, initial_library = lives(), library()
            assert initial_library, 'enabled translation required'
            slug = 'bpml' if 'bpml' in initial_library else sorted(initial_library)[0]
            for command in [['ping'], ['server', 'info'], ['server', 'help'], ['server', 'help', 'live'], ['server', 'help', 'account'],
                            ['auth', 'info'], ['auth', 'whoami'], ['account', 'info'], ['account', 'quotas'], ['account', 'tokens'],
                            ['account', 'quota', 'list'], ['account', 'token', 'list'], ['organization', 'list'],
                            ['organization', 'list', 'o:00000000-0000-0000-0000-000000000000']]:
                run(command, ' '.join(command[:2]))
            organizations = fields(run(['organization', 'list'], 'organization list application discovery'))
            application = next((x['application_id'] for x in organizations if 'application_id' in x), None)
            if application:
                run(['organization', 'application', application, 'status'], 'organization application <id> status')
            else:
                check('organization application status prerequisite', False, reason='No owned application available')
            run(['translation', 'info', slug], 'translation info <translation>')
            catalog = fields(run(['translation', 'catalog', slug], 'translation catalog <translation>'))
            book = next(x['name'] for x in catalog if x.get('book') == 31 and 'name' in x)
            source = one(run(['read', slug, '31', '1', '1'], 'read numeric verse'))
            plain = source.get('content') or source.get('text')
            assert plain
            query = re.findall(r'[^\W\d_]{4,}', plain, re.UNICODE)[0]
            for reference in [[book], [book, '1'], [book, '1', '1'], [book, '1:1']]:
                result = run(['read', slug, *reference], 'read <translation> <book> [chapter] [verse]')
                check('Scripture result contains source content', any(x.get('content') or x.get('text') for x in fields(result)))
            run(['search', slug, query], 'search <translation> <query>')
            added = next((s for s in ['web', 'kjv'] if s not in initial_library), None)
            if added:
                run(['translation', 'add', added], 'translation add <translation>')
                check('translation add domain readback', added in library())
                run(['translation', 'remove', added], 'translation remove <translation>')
                check('translation remove domain readback', added not in library())
                added = None
            else:
                check('library mutation prerequisite', False, reason='No unused test translation')

            for name in [None, marker]:
                created = one(run(['live', 'create', *([name] if name else [])], 'live create [name...]'))
                identifier = created['id']
                owned.add(identifier)
                if name is None:
                    live(identifier, 'delete', yes=True)
                    owned.remove(identifier)
                    continue
                live(identifier, 'info')
                live(identifier, 'stats')
                for action, key, expected in [('start', 'status', 'running'), ('stop', 'status', 'stopped'), ('start', 'status', 'running'),
                                               ('pause', 'paused', True), ('resume', 'paused', False)]:
                    live(identifier, action)
                    check(action + ' domain readback', one(live(identifier, 'info')).get(key) == expected)
                for key, value in [('name', marker + ' updated'), ('reference', book + ' 1:1'), ('translations', slug)]:
                    live(identifier, 'set', key, value)
                check('name update persisted', one(live(identifier, 'info')).get('name') == marker + ' updated')
                for ref in [[slug, book, '1', '1'], [book], [book, '1'], [book, '1', '1'], [book, '1:1'], [slug, '31', '1', '1'], [slug, '31', '1:1']]:
                    live(identifier, 'stack', 'push', *ref)
                count = stack(identifier)
                live(identifier, 'stack', 'pop')
                check('default pop removed one without confirmation', stack(identifier) == count - 1)
                live(identifier, 'stack', 'pop', '2')
                check('explicit pop removed two without confirmation', stack(identifier) == count - 3)
                live(identifier, 'stack', 'clear', yes=True)
                check('stack clear persisted', stack(identifier) == 0)
                live(identifier, 'stack', 'push', slug, '31', '1', '1')
                human_read = run(['read', slug, '31', '1', '1'], 'human Scripture read', fmt='table')
                human_stack = live(identifier, 'stack', 'info', fmt='table')
                for label, output in [('read', human_read), ('stack', human_stack)]:
                    check(label + ' consistent human Scripture columns', all(v in output for v in ['TRANSLATION', 'REFERENCE', 'TEXT']))
                    check(label + ' no endpoint or prompt noise', 'http endpoint' not in output.lower() and 'Type "yes"' not in output and 'api.bibleit.app' not in output)
                check('human read/stack use same reference', source['reference'] in human_read and source['reference'] in human_stack)
                # Presentation wraps text at terminal width; compare native text too.
                native_stack = fields(live(identifier, 'stack', 'info'))
                check('stack/read preserve same source text', any((x.get('content') or x.get('verse_text') or x.get('text', '')).removeprefix(source['reference'] + ' ') == plain.removeprefix(source['reference'] + ' ') for x in native_stack))
                live(identifier, 'clear', yes=True)
                check('clear alias persisted', stack(identifier) == 0)
                previous = None
                for action in ['create', 'rotate']:
                    path = scratch / ('audience-' + action)
                    output = live(identifier, 'secret', action, yes=action == 'rotate', extra=['--secret-output', str(path)])
                    secret = path.read_text().strip()
                    check('secret ' + action + ' protected delivery', path.stat().st_mode & 0o777 == 0o600 and len(secret) == 48 and secret != previous and secret not in json.dumps(output))
                    previous = secret
                live(identifier, 'secret', 'delete', yes=True)
                for fmt in ['table', 'ndjson']:
                    time.sleep(max(0, 4.0 - (time.monotonic() - last)))
                    last = time.monotonic()
                    with tempfile.TemporaryFile() as output, tempfile.TemporaryFile() as errors:
                        process = subprocess.Popen([binary, '--format', fmt, 'live', identifier, 'subscribe'], env=env, stdout=output, stderr=errors, stdin=subprocess.DEVNULL)
                        try:
                            time.sleep(2)
                            live(identifier, 'pause')
                            time.sleep(2)
                            process.send_signal(signal.SIGINT)
                            process.wait(timeout=5)
                            output.seek(0)
                            value = output.read()
                            valid = b'Watching Live changes.' in value if fmt == 'table' else len([json.loads(line) for line in value.splitlines()]) >= 2
                            check('live <id> subscribe', valid and process.returncode == 0, format=fmt)
                        finally:
                            if process.poll() is None:
                                process.kill()
                                process.wait()
                    live(identifier, 'resume')
                # Confirmation must fail before sending any destructive request.
                for command in [['live', identifier, 'delete'], ['live', identifier, 'stack', 'clear']]:
                    run(command, 'destructive action requires confirmation', network=False, error=(2, 'confirmation_required'))
                if not initial_ids and lives() == owned:
                    run(['live', 'delete', 'all'], 'live delete all only owned test Lives', yes=True)
                    owned.clear()
                else:
                    live(identifier, 'delete', yes=True)
                    owned.remove(identifier)

            cases = [(['live', 'stack', 'info'], 2, 'usage'), (['live', 'synthetic-nonexistent-id', 'info'], 5, 'not_found'),
                     (['live', 'L1', 'stack', 'pop', '0'], 2, 'usage'), (['unknown-command'], 2, 'usage'),
                     (['auth', 'ssh'], 2, 'usage'), (['live', 'L1', 'secret', 'synthetic-secret'], 2, 'usage'),
                     (['read', slug, '31', '--limit', '1'], 2, 'usage'), (['config', 'get', 'token'], 2, 'usage')]
            for command, code, error in cases:
                messages = [run(command, 'error: ' + error, fmt=fmt, network=error == 'not_found', error=(code, error)) for fmt in ['table', 'json']]
                check('human/JSON error message and exit code agree', messages[0] == messages[1])
        except Exception as error:
            check('campaign completion', False, reason=type(error).__name__ + '; private output suppressed')
        finally:
            for identifier in list(owned):
                try:
                    live(identifier, 'delete', yes=True)
                    owned.remove(identifier)
                except Exception:
                    pass
            if added and initial_library is not None and added not in initial_library:
                try:
                    run(['translation', 'remove', added], 'test translation cleanup')
                except Exception:
                    pass
            if initial_ids is not None:
                check('original Live scope restored; test Lives removed', lives() == initial_ids)
            if initial_library is not None:
                check('original library restored', library() == initial_library)

    unchanged = all((hashlib.sha256(p.read_bytes()).hexdigest() if p.exists() else None) == before[str(p)] for p in protected)
    check('real preferences and credentials unchanged', unchanged)
    counts = dict(collections.Counter(x['status'] for x in rows))
    report = dict(environment='production SSH using saved single-file YAML defaults', checks=rows, counts=counts,
                  transport_or_identity_flags_passed=False, remote_preferences_overridden=False,
                  auth_and_config_write_tests='isolated copies of actual preferences; empty temporary credentials',
                  remaining_test_lives=len(owned), normal_preferences_and_credentials_unchanged=unchanged,
                  private_content_credentials_and_resource_ids_recorded=False)
    args.report.parent.mkdir(parents=True, exist_ok=True)
    args.report.write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(dict(counts=counts, remaining_test_lives=len(owned), preferences_and_credentials_unchanged=unchanged)))
    return 1 if counts.get('failed') or owned else 0


if __name__ == '__main__':
    raise SystemExit(main())
