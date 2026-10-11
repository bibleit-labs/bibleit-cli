"""TLS for disposable fixtures only; never installed in the host trust store."""
from pathlib import Path
import socket
import socketserver
import ssl
import threading

HOST = 'bibleit.app'
MATERIAL = Path(__file__).resolve().parents[1] / 'contract/fixtures/tls'


def server_context():
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    context.minimum_version = ssl.TLSVersion.TLSv1_2
    context.load_cert_chain(MATERIAL / 'server.pem', MATERIAL / 'server.key')
    return context


def verify_boundary(port):
    """Prove correct name/trust succeeds and wrong name/untrusted CA fail."""
    trusted = ssl.create_default_context(cafile=str(MATERIAL / 'server.pem'))
    for context, name, should_pass in (
            (trusted, HOST, True), (trusted, 'wrong.example', False),
            (ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT), HOST, False)):
        try:
            with socket.create_connection(('127.0.0.1', port), timeout=5) as connection:
                with context.wrap_socket(connection, server_hostname=name):
                    pass
        except ssl.SSLCertVerificationError:
            if should_pass:
                raise
        else:
            if not should_pass:
                raise AssertionError('fixture accepted wrong hostname or untrusted certificate')
    print('TLS fixture: verified hostname/trust; wrong name and untrusted CA rejected', flush=True)


class Relay(socketserver.ThreadingTCPServer):
    allow_reuse_address = True
    daemon_threads = True

    def __init__(self, port, upstream):
        self.context = server_context()
        self.upstream = upstream
        super().__init__(('127.0.0.1', port), RelayHandler)

    def get_request(self):
        connection, address = super().get_request()
        connection.settimeout(20)
        try:
            return self.context.wrap_socket(connection, server_side=True), address
        except Exception:
            connection.close()
            raise


class RelayHandler(socketserver.BaseRequestHandler):
    def handle(self):
        try:
            first = self.request.recv(65536)
        except OSError:
            return
        if not first:
            return
        with socket.create_connection(('127.0.0.1', self.server.upstream), timeout=10) as upstream:
            upstream.sendall(first)
            def copy(source, target):
                try:
                    while data := source.recv(65536):
                        target.sendall(data)
                except (OSError, ssl.SSLError):
                    pass
                finally:
                    try:
                        target.shutdown(socket.SHUT_WR)
                    except OSError:
                        pass
            reader = threading.Thread(target=copy, args=(self.request, upstream), daemon=True)
            reader.start()
            copy(upstream, self.request)
            reader.join(timeout=2)
