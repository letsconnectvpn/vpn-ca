Simple CA intended for use with OpenVPN.

# Why

We started out using [easy-rsa](https://github.com/OpenVPN/easy-rsa) for Let's 
Connect! / eduVPN. It is a shell script wrapped around the OpenSSL command 
line. In theory this can be (very) much cross platform, but in practice it was 
not. Only recent versions fixed some problems on other platforms than Linux.

As part of these fixes they broke backwards compatibility in their 3.x 
releases, which made "in place" upgrades impossible without (manually)
migrating to their new version(s). 

This was a good moment to think about ditching easy-rsa and come up with 
something better. Using PHP's OpenSSL binding was out due to its complexity 
while still lacking basic features.

Go has a rich standard library that has all functionality required for creating
a CA, some projects were available doing exactly that as shown above. Using 
those for inspiration, and some borrowing, stripping everything we didn't need 
resulted in a tiny CA that does exactly what we need and nothing more with a
very simple CLI API. Implementing a PHP extension seemed like overkill, so 
we simply use the CLI from PHP.

# Platform Support

We tested on Linux, OpenBSD, macOS and Windows. It works everywhere!

# Build

Use the `Makefile`:

    $ make

Or manually:

    $ go build -o _bin/vpn-ca vpn-ca/main.go

# Usage

Initialize the CA (valid for 5 years) with an RSA key of 3072 bits:

    $ _bin/vpn-ca -init

Generate a server certificate, valid for 1 year:

    $ _bin/vpn-ca -server vpn.example.org

Generate a client certificate, valid for 1 year:

    $ _bin/vpn-ca -client 12345678

Generate client certificate and specify explicitly when it expires:

    $ _bin/vpn-ca -client 12345678 -not-after 2020-12-12T12:12:12+00:00

The `-not-after` flag can be used with both `-client` and `-server`.

If you want to expire a certificate at the exact same time as the CA, you can
use `-not-after CA`.

**NOTE**: if your `-not-after`, or the default of 1 year when not specified, 
extends beyond the lifetime of the CA an error will be thrown! You should 
either reduce the certificate lifetime, or generate a new CA.

There is also the `-ca-dir` option you can use if you do not want to use
the current directory from which you run the CA command to store the CA, server
and client certificates, e.g.

    $ _bin/vpn-ca -ca-dir /tmp -init
    $ _bin/vpn-ca -ca-dir /tmp -server vpn.example.org
    $ _bin/vpn-ca -ca-dir /tmp -client 12345678

Once you specify the `-ca-dir` you MUST also use it for subsequent calls.
