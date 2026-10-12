#!/usr/bin/env python3
"""Exercise the generated formula in a disposable Homebrew prefix using local CLI archives."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--assets', required=True)
    parser.add_argument('--formula', help='Formula from the separate tap; defaults to the generated candidate formula')
    parser.add_argument('--report', default='build/validation/homebrew.json')
    args = parser.parse_args()
    assets = Path(args.assets).resolve()
    version = json.loads((assets / 'manifest.json').read_text())['version'][1:]
    brew = shutil.which('brew')
    if not brew:
        parser.error('Homebrew is required on a native macOS/Linux host')
    base_env = dict(os.environ, HOMEBREW_NO_AUTO_UPDATE='1', HOMEBREW_NO_ANALYTICS='1')
    repository = Path(subprocess.check_output([brew, '--repository'], env=base_env, text=True).strip())
    evidence = []
    # Homebrew's sandbox grants broad access to the OS temp directory; putting
    # its executable there defeats its inherited-sandbox write-protection check.
    validation = Path(__file__).resolve().parents[1] / 'build/validation'
    validation.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='bibleit-brew-', dir=validation) as scratch:
        scratch = Path(scratch)
        prefix = scratch / 'brew'
        subprocess.run(['git', 'clone', '--local', '--no-hardlinks', '--quiet', str(repository), str(prefix)], check=True)
        # Uninstall can request core metadata even for a dependency-free tap.
        # Seed it outside the sandbox without fetching core's full Git history.
        core = prefix / 'Library/Taps/homebrew/homebrew-core'
        core.parent.mkdir(parents=True, exist_ok=True)
        cached_core = repository / 'Library/Taps/homebrew/homebrew-core'
        source = str(cached_core) if (cached_core / '.git').exists() else 'https://github.com/Homebrew/homebrew-core'
        clone_options = ['--local', '--no-hardlinks'] if source == str(cached_core) else ['--depth=1']
        git_env = dict(base_env, GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_SYSTEM=os.devnull, GIT_TERMINAL_PROMPT='0')
        subprocess.run(['git', 'clone', *clone_options, '--quiet', source, str(core)], env=git_env, check=True, timeout=300)
        core_commit = subprocess.check_output(['git', '-C', str(core), 'rev-parse', 'HEAD'], env=git_env, text=True).strip()
        # Reuse the installed interpreter read-only. Homebrew may still fetch
        # developer gems or core metadata into the disposable prefix.
        ruby = repository / 'Library/Homebrew/vendor/portable-ruby/current'
        if ruby.exists():
            portable = prefix / 'Library/Homebrew/vendor/portable-ruby'
            portable.mkdir(parents=True, exist_ok=True)
            (portable / ruby.resolve().name).symlink_to(ruby.resolve(), target_is_directory=True)
            (portable / 'current').symlink_to(ruby.resolve().name, target_is_directory=True)
        isolated = prefix / 'bin/brew'
        env = dict(base_env, HOMEBREW_NO_INSTALL_FROM_API='1', HOMEBREW_NO_INSTALL_CLEANUP='1', HOMEBREW_NO_AUTOREMOVE='1', HOMEBREW_NO_ENV_HINTS='1', HOMEBREW_CACHE=str(scratch / 'cache'), HOMEBREW_LOGS=str(scratch / 'logs'), HOMEBREW_TEMP=str(scratch / 'temp'))
        (scratch / 'temp').mkdir()
        tap = prefix / 'Library/Taps/bibleit-labs/homebrew-local-validation'
        (tap / 'Formula').mkdir(parents=True)
        formula_path = Path(args.formula) if args.formula else assets / 'bibleit-cli.rb'
        if not args.formula and not formula_path.exists():
            formula_path = assets / 'bibleit.rb'  # Previously generated candidates.
        formula_name = formula_path.stem
        if formula_name not in ('bibleit', 'bibleit-cli'):
            raise RuntimeError(f'unsupported Bibleit formula name: {formula_name}')
        original = formula_path.read_text()
        # Only URLs change in the fixture. Install/test methods and actual hashes
        # remain the generated production formula's implementation.
        def local_url(match):
            name = match[1].rsplit('/', 1)[-1]
            archive = assets / name
            if not archive.is_file():
                raise RuntimeError(f'missing archive: {name}')
            return f'url "{archive.as_uri()}"'
        formula = re.sub(r'url "(https://github.com/bibleit-labs/bibleit-cli/releases/download/[^\"]+)"', local_url, original)
        (tap / 'Formula' / f'{formula_name}.rb').write_text(formula)
        subprocess.run(['git', 'init', '--quiet', str(tap)], check=True)
        name = f'bibleit-labs/local-validation/{formula_name}'
        def run(*command):
            result = subprocess.run([str(isolated), *command], env=env, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
            evidence.append({'command': ['brew', *command], 'exit_code': result.returncode, 'output': result.stdout})
            if result.returncode:
                raise RuntimeError(result.stdout)
            print(f'PASS: brew {" ".join(command)}', flush=True)
            return result.stdout.strip()
        actual_prefix = Path(run('--prefix'))
        if actual_prefix.resolve() != prefix.resolve():
            raise RuntimeError(f'refusing installation outside fixture prefix: {actual_prefix}')
        run('install', '--formula', '--build-from-source', name)
        run('test', name)
        binary = prefix / 'bin/bibleit'
        expected = f'bibleit CLI version {version}'
        assert subprocess.check_output([str(binary), 'version'], text=True).strip() == expected
        before = hashlib.sha256(binary.read_bytes()).hexdigest()
        for path in ('etc/bash_completion.d/bibleit', 'share/zsh/site-functions/_bibleit', 'share/fish/vendor_completions.d/bibleit.fish'):
            assert (prefix / path).is_file(), path
        def verify_notices():
            for notice in ('LICENSE', 'THIRD_PARTY_NOTICES'):
                assert (prefix / 'share' / formula_name / notice).read_bytes() == (Path(__file__).resolve().parents[1] / notice).read_bytes(), notice
        verify_notices()
        run('reinstall', '--formula', '--build-from-source', name)
        assert hashlib.sha256(binary.read_bytes()).hexdigest() == before
        verify_notices()
        run('test', name)
        # The disposable prefix has no dependent packages. Avoid Homebrew's
        # dependent checks/autoremove where supported by the installed Homebrew.
        assert {p.name for p in (prefix / 'Cellar').iterdir() if p.is_dir()} == {formula_name}
        run('uninstall', '--formula', '--ignore-dependencies', name)
        assert not binary.exists() and not binary.is_symlink()
        assert not (prefix / 'Cellar' / formula_name).exists()
    report = Path(args.report)
    report.parent.mkdir(parents=True, exist_ok=True)
    report.write_text(json.dumps({'platform': os.uname().sysname, 'architecture': os.uname().machine, 'version': version, 'local_cli_archives': True, 'isolated_prefix': True, 'redistribution_notices_verified': True, 'core_metadata_commit': core_commit, 'core_metadata_source': 'runner cache' if source == str(cached_core) else 'public shallow clone', 'checks': evidence}, indent=2) + '\n')
    print(f'PASS: Homebrew install/test/reinstall/completions/uninstall; report {report}')


if __name__ == '__main__':
    main()
