Simple CA intended for use with Let's Connect! & eduVPN.

It is based on:

* [eedev/minica](https://github.com/eedev/minica)
* [jsha/minica](https://github.com/jsha/minica)

Some ideas from:

* [FiloSottile/mkcert](https://github.com/FiloSottile/mkcert)

# Why

We started out using [easy-rsa](https://github.com/OpenVPN/easy-rsa) for Let's 
Connect! / eduVPN. It is a shell script wrapped around the OpenSSL command 
line. In theory this can be (very) much cross platform, but in practise it was 
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

# Build

Use the `Makefile`:

    $ make

Or manually:

    $ go build -o _bin/vpn-ca vpn-ca/main.go

# Usage

Initialize the CA (valid for 5 years):

    $ _bin/vpn-ca -init

Generate a server certificate (expires at the exact moment the CA expires):

    $ _bin/vpn-ca -server vpn.example.org

Generate a client certificate, valid for 90 days:

    $ _bin/vpn-ca -client 12345678

Generate client certifictate and specify when it expires:

    $ _bin/vpn-ca -client 12345678 -not-after 2019-08-16T14:00:00+00:00

There is also the `-ca-dir` option you can specify if you do not want to use
the current directory from which you run the CA command, e.g.

    $ _bin/vpn-ca -ca-dir /tmp -init
