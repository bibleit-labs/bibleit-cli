#!/usr/bin/env python3
"""Build fixture NIFs for the container OS, then execute the real-server harness."""
from pathlib import Path
import shutil
import subprocess
import tempfile


def run(command, **kwargs):
    return subprocess.run(command, check=True, text=True, **kwargs)


def main():
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
        run(['python3', 'scripts/check_server.py', '--integration', '--external-disposable-database',
             '--source-metadata', str(workspace / 'server-metadata.json'),
             '--baseline', '/results/server-baseline.json'], cwd=cli)


if __name__ == '__main__':
    main()
