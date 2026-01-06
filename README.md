# ..Modlishka..

![License](https://img.shields.io/badge/license-Author-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux%20%7C%20BSD-lightgrey.svg)
![Build Status](https://github.com/drk1wi/Modlishka/actions/workflows/reviewdog.yml/badge.svg)
![Go Version](https://img.shields.io/badge/go-1.24%2B-00ADD8.svg)

Modlishka is an open-source penetration testing tool that acts as a man-in-the-middle proxy. It introduced a new technical approach to handling browser-based HTTP traffic flow, which allows it to transparently proxy multi-domain destination traffic, both TLS and non-TLS, over a single domain, without requiring the installation of any additional certificate on the client.

In 2019, Modlishka was the first publicly released research tool to demonstrate a novel Adversary-in-the-Middle (AitM) technique capable of bypassing many common 2FA implementations — with the goal of raising awareness and improving real-world defenses.

> **Note:** This project is intended strictly for authorized research and professional security testing.

## Use Cases

**Security Testing:**
- Ethical phishing penetration tests with transparent, automated reverse proxy and universal 2FA bypass support
- Highlight [2FA](https://blog.duszynski.eu/phishing-ng-bypassing-2fa-with-modlishka/) scheme weaknesses to drive better industry security solutions

**General:**
- Wrap legacy websites with TLS
- Confuse crawler bots and automated scanners
- Universal transparent reverse proxy for other projects

## Features

**General:**
- Point-and-click HTTP and HTTPS reverse proxying of arbitrary domains
- Full control of cross-origin TLS traffic flow without client certificate installation
- Easy configuration through command line options and JSON configuration files
- Pattern-based JavaScript payload injection
- TLS wrapping, authentication, and security headers for legacy websites
- Stateless design for easy scaling via DNS load balancer
- Extensible through modular plugins
- Automatic TLS certificate generation plugin (requires self-signed CA)
- Cross-platform: Windows, macOS, Linux, BSD

**Security:**
- Support for majority of 2FA authentication schemes out of the box
- [Client Domain Hooking](https://blog.duszynski.eu/client-domain-hooking-in-practice/) attack implementation with diagnostic plugin
- User credential harvesting with URL parameter-based context
- Web panel plugin for credential management and session impersonation (beta)
- No website templates required — automatic handling in most cases

## Demo

Modlishka in action against an example 2FA scheme (SMS-based bypass):

[![Demo](https://img.shields.io/badge/Watch-Demo-red.svg)](https://vimeo.com/308709275)

## Installation

Latest source code: [zip](https://github.com/drk1wi/modlishka/zipball/master) | [tar](https://github.com/drk1wi/modlishka/tarball/master)

**Using go install:**
```bash
go install github.com/drk1wi/Modlishka@latest
```

**Manual build:**
```bash
git clone https://github.com/drk1wi/Modlishka.git
cd Modlishka
make
```

## Usage

```
./dist/proxy -h

Usage of ./dist/proxy:

  -cert string
      base64 encoded TLS certificate
  -certKey string
      base64 encoded TLS certificate key
  -certPool string
      base64 encoded Certification Authority certificate
  -config string
      JSON configuration file. Convenient instead of using command line switches.
  -controlCreds string
      Username and password to protect the credentials page. user:pass format
  -controlURL string
      URL to view captured credentials and settings. (default "SayHello2Modlishka")
  -credParams string
      Credential regexp with matching groups. e.g.: base64(username_regex),base64(password_regex)
  -debug
      Print debug information
  -disableSecurity
      Disable proxy security features like anti-SSRF. Disable at your own risk.
  -disableDynamicSubdomains
      Translate URL domain names to be the proxy domain
  -dynamicMode
      Enable dynamic mode for 'Client Domain Hooking'
  -forceHTTP
      Strip all TLS from the traffic and proxy through HTTP only
  -forceHTTPS
      Strip all clear-text from the traffic and proxy through HTTPS only
  -allowSecureCookies
      Allow secure cookies to be set. Useful when using HTTPS and cookies have SameSite=None
  -ignoreTranslateDomains string
      Comma separated list of domains to never translate and proxy
  -jsRules string
      Comma separated list of URL patterns and JS base64 encoded payloads that will be injected
      e.g.: target.tld:base64(alert(1))
  -listeningAddress string
      Listening address (default "127.0.0.1")
  -listeningPortHTTP int
      Listening port for HTTP requests (default 80)
  -listeningPortHTTPS int
      Listening port for HTTPS requests (default 443)
  -log string
      Local file to which fetched requests will be written (appended)
  -pathHostRules string
      Comma separated list of URL path patterns and target domains
      e.g.: /path/:example.com,/path2:www.example.com
  -plugins string
      Comma separated list of enabled plugin names (default "all")
  -postOnly
      Log only HTTP POST requests
  -proxyAddress string
      Proxy that should be used (socks/https/http) e.g.: http://127.0.0.1:8080
  -proxyDomain string
      Proxy domain name that will be used e.g.: proxy.tld
  -rules string
      Comma separated list of string patterns and their replacements
      e.g.: base64(old):base64(new),base64(older):base64(newer)
  -staticLocations string
      Comma separated list of FQDNs in location headers that should be preserved
  -target string
      Target domain name e.g.: target.tld
  -targetRes string
      Comma separated list of domains that were not translated automatically
      e.g.: static.target.tld
  -terminateTriggers string
      Comma separated list of URLs from target's origin which will trigger session termination
  -terminateUrl string
      URL to which a client will be redirected after session termination
  -trackingCookie string
      Name of the HTTP cookie used to track the client (default "id")
  -trackingParam string
      Name of the HTTP parameter used to track the client (default "id")
```

## Commercial Usage

Modlishka is licensed under [this License](https://raw.githubusercontent.com/drk1wi/Modlishka/master/LICENSE).

For commercial applications, please contact the author for licensing arrangements.

## Credits

Author: Modlishka was designed and implemented by Piotr Duszyński ([@drk1wi](https://twitter.com/drk1wi)). All rights reserved.

See the list of [contributors](https://github.com/drk1wi/Modlishka/graphs/contributors) who participated in this project.

## Disclaimer

This tool is made only for educational purposes and can be used in legitimate penetration tests or research only. Author does not take any responsibility for any actions taken by its users.
