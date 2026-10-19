#!/usr/bin/env python3
"""Verify the deployed contract and read-only CLI behavior; never run mutations."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import urllib.request


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *_args):
        return None


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--legacy-binary', type=Path, required=True)
    parser.add_argument('--contract', type=Path, required=True)
    parser.add_argument('--report', type=Path, required=True)
    args = parser.parse_args()
    expected = json.loads(args.contract.read_text())
    with urllib.request.build_opener(NoRedirect()).open('https://api.bibleit.app/v1/openapi.json', timeout=20) as response:
        raw = response.read(4 * 1024 * 1024)
    actual = json.loads(raw)
    if actual != expected:
        raise RuntimeError('deployed_contract_does_not_match_finalized_contract')
    operations = sum(method in {'get', 'post', 'put', 'patch', 'delete', 'head', 'options'}
                     for item in actual['paths'].values() for method in item)
    if operations != 95:
        raise RuntimeError('unexpected_production_operation_count')
    evidence = []
    with tempfile.TemporaryDirectory(prefix='bibleit-production-readonly-') as directory:
        env = dict(os.environ, BIBLEIT_PREFERENCES=str(Path(directory) / 'config.yaml'))
        # Read existing credential selection; no auth/config mutation is invoked.
        def invoke(binary, label, command, name, expected_error=None):
            prefix = tuple(command[:2])
            allowed = command[0] in ('ping', 'read', 'search') or prefix in {
                ('server', 'info'), ('server', 'help'), ('auth', 'info'),
                ('account', 'info'), ('account', 'quotas'), ('account', 'tokens'),
                ('organization', 'list'), ('translation', 'list'),
                ('translation', 'info'), ('translation', 'catalog'), ('live', 'list')}
            allowed = allowed or (command[0] == 'live' and
                                  ((len(command) == 3 and command[2] in ('info', 'stats')) or
                                   (len(command) == 4 and command[2:] == ['stack', 'info'])))
            if not allowed:
                raise RuntimeError('mutating_or_unknown_production_command_rejected')
            result = subprocess.run([str(binary.resolve()), '--transport', 'http', '--format', 'json', *command],
                                    env=env, text=True, capture_output=True, timeout=30)
            if result.returncode and expected_error is None:
                raise RuntimeError(f'{label}: {name} exited {result.returncode}; private output suppressed')
            try:
                envelope = json.loads(result.stderr if expected_error is not None else result.stdout)
            except ValueError:
                raise RuntimeError(f'{label}: {name} did not emit JSON') from None
            if expected_error is not None:
                if (not result.returncode or result.stdout or envelope.get('ok') is not False
                        or envelope.get('error', {}).get('code') != expected_error):
                    raise RuntimeError(f'{label}: {name} did not return the expected safe error')
                evidence.append({'client': label, 'command': name, 'exit_code': result.returncode,
                                 'error_code': expected_error, 'status': 'passed'})
                return None
            if envelope.get('schema_version') != 2 or envelope.get('ok') is not True:
                raise RuntimeError(f'{label}: {name} unexpected output envelope')
            evidence.append({'client': label, 'command': name, 'exit_code': 0, 'output_schema': 2, 'status': 'passed'})
            return envelope['data']

        legacy_verse = None
        for label, binary in [('legacy-v0.1.0', args.legacy_binary), ('candidate', args.binary)]:
            common = [('ping',), ('server', 'info'), ('auth', 'info'), ('account', 'info')]
            if label == 'candidate':
                common.append(('server', 'help'))
            for command in common:
                invoke(binary, label, command, ' '.join(command))
            library = invoke(binary, label, ['translation', 'list'], 'translation list')
            rows = library.get('data', [])
            slug = next((row['slug'] for row in rows if row.get('slug', '').lower() == 'bpml'), None)
            if not slug and rows:
                slug = rows[0]['slug']
            if not slug:
                raise RuntimeError(f'{label}: enabled_translation_required_for_readonly_read_checks')
            invoke(binary, label, ['translation', 'info', slug], 'translation info <translation>')
            invoke(binary, label, ['translation', 'catalog', slug], 'translation catalog <translation>')
            scripture = invoke(binary, label, ['read', slug, '31', '1', '1'], 'read <translation> <book> <chapter> <verse>')
            verses = scripture.get('verses', [])
            if not verses or verses[0].get('book_id') != 31 or not verses[0].get('text'):
                raise RuntimeError(f'{label}: scripture_coordinates_or_content_missing')
            source = {key: verses[0].get(key) for key in ('translation', 'book_id', 'chapter', 'verse', 'text')}
            if label == 'legacy-v0.1.0':
                legacy_verse = source
            elif source != legacy_verse:
                raise RuntimeError('old_new_clients_disagree_on_source_verse')
            words = re.findall(r'[^\W\d_]{4,}', verses[0]['text'], flags=re.UNICODE)
            if words:
                invoke(binary, label, ['search', slug, words[0]], 'search <translation> <query>')
            lives = invoke(binary, label, ['live', 'list'], 'live list')
            if label == 'candidate':
                complete_book = None
                for reference in [('31',), ('31', '1'), ('31', '1:1'), ('19',)]:
                    result = invoke(binary, label, ['read', slug, *reference], 'read <translation> <book> [chapter] [verse]')
                    if reference == ('31',):
                        complete_book = result
                first_page = invoke(binary, label, ['read', slug, '31', '--limit', '5'], 'read <translation> <book> --limit 5')
                if first_page.get('verses') != complete_book['verses'][:5] or not first_page.get('next_cursor'):
                    raise RuntimeError('first_page_does_not_match_complete_book')
                page_verses = first_page['verses'][:]
                cursor = first_page['next_cursor']
                seen = set()
                while cursor:
                    if cursor in seen or len(seen) > 1000:
                        raise RuntimeError('unsafe_production_book_cursor_traversal')
                    seen.add(cursor)
                    page = invoke(binary, label, ['read', slug, '31', '--limit', '3', '--cursor', cursor],
                                  'read <translation> <book> --limit 3 --cursor <opaque-cursor>')
                    if len(page.get('verses', [])) > 3 or 'next_cursor' not in page:
                        raise RuntimeError('page_limit_or_cursor_contract_failed')
                    page_verses.extend(page.get('verses', []))
                    cursor = page['next_cursor']
                if page_verses != complete_book['verses']:
                    raise RuntimeError('production_pagination_changed_source_verses')
                invoke(binary, label, ['read', slug, '31', '--cursor', 'synthetic-invalid-cursor'],
                       'read <translation> <book> --cursor <invalid-cursor>', expected_error='invalid_cursor')
                invoke(binary, label, ['read', slug, '19', '--cursor', first_page['next_cursor']],
                       'read <translation> <different-book> --cursor <opaque-cursor>', expected_error='cursor_mismatch')
                for command in [('account', 'quotas'), ('account', 'tokens'), ('organization', 'list')]:
                    invoke(binary, label, command, ' '.join(command))
                if lives.get('data'):
                    live_id = lives['data'][0]['id']
                    for action in [('info',), ('stats',), ('stack', 'info')]:
                        invoke(binary, label, ['live', live_id, *action], 'live <id> ' + ' '.join(action))
        report = {'environment': 'production; read-only HTTP requests using existing credential selection',
                  'contract_sha256': hashlib.sha256(raw).hexdigest(), 'contract_matches_finalized_sdk': True,
                  'canonical_operations': operations, 'checks': evidence, 'passed': len(evidence),
                  'mutations_invoked': False, 'credentials_or_user_content_recorded': False,
                  'homebrew_modified': False}
        args.report.parent.mkdir(parents=True, exist_ok=True)
        args.report.write_text(json.dumps(report, indent=2) + '\n')
        print(f'Production read-only verification passed: {len(evidence)} CLI cases, contract has {operations} operations, old/new clients')


if __name__ == '__main__':
    main()
