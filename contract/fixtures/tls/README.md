# Disposable TLS fixture

This public test certificate and private key authenticate only an isolated fake
`legacy.bibleit.example`. Never deploy them or install this certificate in a host
trust store. The certificate expires in October 2036; renew this fixture before then.

Real-server integration trusts it only inside an egress-disabled container, where
the fixed hostname resolves to loopback. Native process tests use a test executable
with an explicit fixture-only trust pool and loopback dialer. Production binaries
contain neither the fixture dialer nor its trust configuration.
