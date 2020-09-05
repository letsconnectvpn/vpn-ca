# ChangeLog

## 3.0.1 (...)
- Support ECDSA (P-256/prime256v1) and EdDSA (Ed25519) keys using 
  `CA_KEY_ALGO=ECDSA` and `CA_KEY_ALGO=EdDSA` environment variable

## 3.0.0 (2020-08-31)
- add name to sAN (DNSNames) for server certificates as well
- make CN of CA configurable through `-name` option when generating a CA
- change CLI parameters, now an explicit `-name` is required for `-server` and
  `-client`, optional for `-init-ca`
- default CN of CA changed to "Root CA"
- the `-init` option is renamed to `-init-ca`
- remove `-ca-dir` option, use `CA_DIR` environment variable now

## 2.0.1 (2020-04-30)
- update `Makefile` to support `install`

## 2.0.0 (2020-03-16)
- make sure issued certificates never outlive the CA (#6)
- allow specifying `CA` as a value for `-not-after` to explicitly let the 
  certificate expire at the same time as the CA (#7)
- server and client certificates now expire after 1 year by default (instead of 
  together with the CA and 90 days respectively)

## 1.0.0 (2019-11-18)
- initial release
