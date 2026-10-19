#!/usr/bin/env python3
"""Focused changed-native qualification inside the isolated fixture namespace."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import shutil
import socket
import subprocess
import tempfile
from urllib.parse import urlparse

ROOT = Path(__file__).resolve().parents[1]


def run(command, **kwargs):
    return subprocess.run(command, check=True, text=True, **kwargs)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--server', type=Path, required=True)
    parser.add_argument('--source-metadata', type=Path, required=True)
    parser.add_argument('--report', type=Path, required=True)
    args = parser.parse_args()
    db = urlparse(os.environ.get('BIBLEIT_TEST_DATABASE_URL', ''))
    if (not Path('/.dockerenv').exists()
            or os.environ.get('BIBLEIT_FIXED_ENDPOINT_FIXTURE') != '1'
            or db.hostname != '127.0.0.1' or db.path != '/bibleit_test'
            or any(line.split()[1] == '00000000' for line in Path('/proc/net/route').read_text().splitlines()[1:])
            or {item[4][0] for item in socket.getaddrinfo('bibleit.app', 443)} != {'127.0.0.1'}):
        raise SystemExit('focused catalogue check requires the egress-disabled disposable fixture')
    server = args.server.resolve()
    metadata = json.loads(args.source_metadata.read_text())
    with tempfile.TemporaryDirectory(prefix='bibleit-catalog-contract-') as directory:
        temporary = Path(directory)
        app = temporary / 'bibleit_server'
        ebin = app / 'ebin'
        ebin.mkdir(parents=True)
        shutil.copytree(server / 'priv', app / 'priv')
        shutil.copy(server / '_build/default/lib/bibleit_server/ebin/bibleit_server.app', ebin)
        paths = [part for p in (server / '_build/default/lib').glob('*/ebin')
                 if p.parent.name != 'bibleit_server' for part in ('-pa', str(p))]
        suites = [server / 'test/bibleit_security_SUITE.erl', server / 'test/bibleit_test_database.erl']
        run(['erlc', *paths, '-o', str(ebin),
             *map(str, (server / 'src').rglob('*.erl')), *map(str, suites)], cwd=temporary)
        env = dict(os.environ, GOTOOLCHAIN='local', GOCACHE='/tmp/catalog-go-cache')
        for name, checkout in [('go', ROOT.parent / 'bibleit-go'), ('cli', ROOT)]:
            binary = temporary / (name + '.test')
            run(['go', 'test', '-c', '-o', str(binary), '.'], cwd=checkout, env=env)
            env['BIBLEIT_CATALOG_' + name.upper() + '_BIN'] = str(binary)
        private = temporary / 'data'
        private.mkdir()
        env['BIBLEIT_TEST_PRIV'] = str(private)
        evaluation = r'''
          {ok,_}=application:ensure_all_started(cowboy),
          C=bibleit_security_SUITE:init_per_suite([{priv_dir,os:getenv("BIBLEIT_TEST_PRIV")}]),
          try
            _=bibleit_security_SUITE:init_per_testcase(catalog,C),
            gen_server:stop(bibleit_rate_limit),
            application:set_env(bibleit_server,rate_limits,#{ssh_command=>#{limit=>100}}),
            {ok,Rate}=bibleit_rate_limit:start_link(),unlink(Rate),
            ok=gen_server:stop(bibleit_http_listener),
            application:set_env(bibleit_server,http,#{port=>8080,host=>{127,0,0,1}}),
            application:set_env(bibleit_server,http_public_origin,<<"https://bibleit.app">>),
            {ok,HTTP}=bibleit_http_listener:start_link(),unlink(HTTP),
            Books=[#{<<"bookid">>=>B,<<"name">>=>case B of 42-><<"Luke">>;_->iolist_to_binary(io_lib:format("Book ~B",[B])) end}||B<-lists:seq(1,66)],
            Opening=[#{<<"book">>=>B,<<"chapter">>=>1,<<"verse">>=>1,<<"text">>=><<"fixture">>}||B<-lists:seq(1,66)],
            Sparse=[#{<<"book">>=>42,<<"chapter">>=>17,<<"verse">>=>V,<<"text">>=>Text}||{V,Text}<-[{35,<<"before-gap">>},{37,<<"after-gap">>}]],
            {ok,<<"BPML">>}=bibleit_translation_fetcher:install(<<"BPML">>,Books,Opening++Sparse),
            {ok,<<"LUT1912">>}=bibleit_translation_fetcher:install(<<"LUT1912">>,Books,Opening++Sparse),
            {ok,_}=bibleit_account:add_translation(<<"security-owner">>,<<"BPML">>),
            {ok,_}=bibleit_account:add_translation(<<"security-owner">>,<<"LUT1912">>),
            {ok,CatalogBooks}=bibleit_api:translation_catalog(<<"BPML">>),
            CatalogWire=iolist_to_binary(bibleit_protocol:encode({ok,{books,<<"BPML">>,CatalogBooks}})),
            case unicode:characters_to_list(CatalogWire) of V when is_list(V)->ok;_->io:format("Catalogue wire contains invalid UTF-8~n") end,
            {ok,_,Token}=bibleit_authorization:create_token(<<"security-owner">>,<<"security-owner">>,<<"Sparse fixture">>,[],manual,erlang:system_time(second)+3600),
            Directory=os:getenv("BIBLEIT_TEST_PRIV"),SSHDir=filename:join(Directory,"ssh"),ok=file:make_dir(SSHDir),
            Private=public_key:generate_key({rsa,2048,65537}),Public={'RSAPublicKey',element(3,Private),element(4,Private)},
            Pem=public_key:pem_encode([public_key:pem_entry_encode('RSAPrivateKey',Private)]),Identity=filename:join(SSHDir,"id_rsa"),
            [begin ok=file:write_file(F,Pem),ok=file:change_mode(F,8#600) end||F<-[Identity,filename:join(SSHDir,"ssh_host_rsa_key")]],
            PublicLine=ssh_file:encode([{Public,[]}],openssh_key),
            ok=file:write_file(Identity++".pub",PublicLine),
            {ok,_}=bibleit_account:add_key(<<"security-owner">>,PublicLine),
            {ok,SSH}=bibleit_ssh_listener:init(#{port=>22,system_dir=>SSHDir}),
            try
              ClientResults=lists:map(fun({Bin,Test})->
                P=open_port({spawn_executable,os:getenv(Bin)},[binary,exit_status,use_stdio,stderr_to_stdout,
                  {args,["-test.run="++Test,"-test.v"]},
                  {env,[{"BIBLEIT_CLIENT_TEST_DISPOSABLE","1"},{"BIBLEIT_INTEGRATION_TOKEN",binary_to_list(Token)},
                        {"BIBLEIT_INTEGRATION_IDENTITY",Identity}]}]),
                Wait=fun F()->receive {P,{data,Data}}->io:put_chars(Data),F();{P,{exit_status,Status}}->Status after 45000->error(focused_client_timeout) end end,
                Wait()
              end,[{"BIBLEIT_CATALOG_GO_BIN","^TestServerSparseCatalogIntegration$"},{"BIBLEIT_CATALOG_CLI_BIN","^TestServerSparseCatalogSSH$"}]),
              {error,index_open_failed}=bibleit_translation_publication:install(Directory,<<"BPML">>,
                <<"Genesis 1:1 first\nGenesis 1:2 second\nGenesis 1:2 duplicate\n">>,<<>>),
              {ok,<<"Luke 17:37 after-gap">>}=bibleit_api:read(<<"BPML">>,42,17,37),
              io:format("Invalid native index rejected before publication; prior sparse generation preserved.~n"),
              case lists:all(fun(Status)->Status=:=0 end,ClientResults) of true->ok;false->error({focused_clients_failed,ClientResults}) end
            after bibleit_ssh_listener:terminate(normal,SSH) end
          after application:unset_env(bibleit_server,http_public_origin),
                application:unset_env(bibleit_server,rate_limits),
                bibleit_security_SUITE:end_per_suite(C) end,
          halt().
        '''
        run(['erl', '-noshell', *paths, '-pa', str(ebin), '-eval', evaluation],
            env=env, cwd=temporary, timeout=120)
    report = dict(metadata)
    report.update({
        'reviewed_at': datetime.datetime.now(datetime.timezone.utc).date().isoformat(),
        'files_sha256': {str(p.relative_to(server)): hashlib.sha256(p.read_bytes()).hexdigest()
                         for p in sorted((server / 'src').rglob('*.erl'))},
        'test_files_sha256': {str(p.relative_to(server)): hashlib.sha256(p.read_bytes()).hexdigest()
                              for p in sorted((server / 'test').rglob('*.erl'))},
        'verification': {
            'scope': 'changed native sparse read/catalogue/publication only',
            'fixed_http_endpoint': 'https://api.bibleit.app/v1',
            'fixed_ssh_endpoint': 'bibleit.app:22',
            'egress_disabled_fixture': True,
            'tls_boundary_cases': 3,
            'go_sparse_rest_cases': 3,
            'localized_catalogues': ['BPML', 'LUT1912'],
            'cli_sparse_ssh_cases': 5,
            'invalid_index_publication_cases': 1,
            'fixture': 'synthetic 66-book BPML-shaped data, Luke17:36 absent and17:37 present',
            'broad_protocol_integration_repeated': False,
        },
    })
    args.report.write_text(json.dumps(report, indent=2) + '\n')


if __name__ == '__main__':
    main()
