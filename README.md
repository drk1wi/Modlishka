# ..Modlishka..

Modlishka is an open-source penetration testing tool that acts as a man-in-the-middle proxy. It introduced a new technical approach to handling browser-based HTTP traffic flow, which allows it to transparently proxy multi-domain destination traffic, both TLS and non-TLS, over a single domain, without requiring the installation of any additional certificate on the client. This enables multiple practical use cases.

Modlishka was the first publicly released research tool, in 2019, to demonstrate a novel Adversary-in-the-Middle (AitM) technique capable of bypassing many common 2FA implementations — with the goal of raising awareness and improving real-world defenses.

The project was created to highlight authentication flaws and is intended strictly for authorized research and professional security testing.


From a security perspective, Modlishka can currently be used to:
-	Support ethical phishing penetration tests with a transparent and automated reverse proxy component that has a universal 2FA “bypass” support.
-  Wrap legacy websites with a TLS layer, confuse crawler bots and automated scanners, etc.


Modlishka was written as an attempt to overcome standard reverse proxy limitations and as a personal challenge to see what is possible with sufficient motivation and a bit of extra research time. 
The achieved results appeared to be very interesting and the tool was initially released and later updated with an aim to:
- Highlight currently used two-factor authentication ([2FA](https://blog.duszynski.eu/phishing-ng-bypassing-2fa-with-modlishka/)) scheme weaknesses, so adequate security solutions can be created and implemented by the industry.
- Support other projects that could benefit from a universal and transparent reverse proxy.
- Raise community awareness about modern phishing techniques and strategies and support penetration testers in their daily work.


Modlishka was primarily written for security-related tasks. Nevertheless, it can be helpful in other, non-security related, usage scenarios.

Features
--------

Key features of Modlishka include:

**General:**
-   Point-and-click HTTP and HTTPS reverse proxying of an arbitrary domain/s.
-   Full control of "cross" origin TLS traffic flow from your users browsers (without the requirement of installing any additional certificate on the client).
-   Easy and fast configuration through command line options and JSON configuration files.
-   Pattern based JavaScript payload injection.
-   Wrapping websites with an extra "security": TLS wrapping, authentication, relevant security headers, etc. 
-   Stripping websites of all encryption and security headers (back to 90's MITM style). 
-   Stateless design. Can be scaled up easily to handle an arbitrary amount of traffic  - e.g. through a DNS load balancer.
-   Can be extended easily with your ideas through modular plugins.
-   Automatic test TLS certificate generation plugin for the proxy domain (requires a self-signed CA certificate)
-   Written in Go, so it works basically on all platforms and architectures: Windows, OSX, Linux, BSD supported...

**Security related:**
-  Support for majority of 2FA authentication schemes (out of the box).
-   Practical implementation of the "[Client Domain Hooking](https://blog.duszynski.eu/client-domain-hooking-in-practice/)" attack. Supported with a diagnostic plugin.
-  User credential harvesting (with context based on URL parameter passed identifiers).
-  Web panel plugin with a summary of automatically collected credentials and one-click user session impersonation module (proof-of-concept/beta).
-  No website templates (just point Modlishka to the target domain - in most cases, it will be handled automatically without any additional manual configuration).


Proxying In Action (2FA bypass)
------
_"A picture is worth a thousand words":_

Modlishka in action against an example two factor authentication scheme (SMS based bypass proof-of-concept)  :

[https://vimeo.com/308709275](https://vimeo.com/308709275)

Installation
------------

Latest source code version can be fetched from [here](https://github.com/drk1wi/modlishka/zipball/master) (zip) or [here](https://github.com/drk1wi/modlishka/tarball/master) (tar).



Fetch the code with _'go install'_ :

    $ go install github.com/drk1wi/Modlishka@latest

Compile manually:

    $ git clone https://github.com/drk1wi/Modlishka.git
    $ cd Modlishka
    $ make
    
------

![alt text](https://github.com/drk1wi/assets/raw/master/0876a672f771046e833f2242f6be5d3cf01519efdbb9dad0e1ed2d33e33fecbc.png)

    # ./dist/proxy -h
  
    
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
          Username and password to protect the credentials page.  user:pass format
          
      -controlURL string
          URL to view captured credentials and settings. (default "SayHello2Modlishka")
          
      -credParams string
          	Credential regexp with matching groups. e.g. : base64(username_regex),base64(password_regex)

      -debug
        	Print debug information
      
      -disableSecurity
        	Disable proxy security features like anti-SSRF. 'Here be dragons' - disable at your own risk.
      
      -dynamicMode
          	Enable dynamic mode for 'Client Domain Hooking'
      
      -forceHTTP
         	Strip all TLS from the traffic and proxy through HTTP only
    
      -forceHTTPS
         	Strip all clear-text from the traffic and proxy through HTTPS only
     
      -jsRules string
        	Comma separated list of URL patterns and JS base64 encoded payloads that will be injected - e.g.: target.tld:base64(alert(1)),..,etc
      
      -listeningAddress string
        	Listening address - e.g.: 0.0.0.0  (default "127.0.0.1")
      
      -log string
        	Local file to which fetched requests will be written (appended)
      
      -plugins string
        	Comma seperated list of enabled plugin names (default "all")
      
      -proxyAddress string
    	    Proxy that should be used (socks/https/http) - e.g.: http://127.0.0.1:8080 
         
      -proxyDomain string
        	Proxy domain name that will be used - e.g.: proxy.tld
      
      -postOnly
        	Log only HTTP POST requests
      
      -rules string
          	Comma separated list of 'string' patterns and their replacements - e.g.: base64(new):base64(old),base64(newer):base64(older)

      -target string
        	Target domain name  - e.g.: target.tld
         
      -targetRes string
        	Comma separated list of domains that were not translated automatically. Use this to force domain translation - e.g.: static.target.tld 
      
      -terminateTriggers string
        	Session termination: Comma separated list of URLs from target's origin which will trigger session termination
        		
      -terminateUrl string
        	URL to which a client will be redirected after Session Termination rules trigger
      
      -trackingCookie string
        	Name of the HTTP cookie used to track the client (default "id")
      
      -trackingParam string
        	Name of the HTTP parameter used to track the client (default "id")


Commercial Usage
-------
Modlishka is licensed under [this License](https://raw.githubusercontent.com/drk1wi/Modlishka/master/LICENSE). 

For commercial, legitimate applications, please contact the author for the appropriate licensing arrangements.

Credits 
-------
Author: Modlishka was designed and implemented by Piotr Duszyński ([@drk1wi](https://twitter.com/drk1wi)). All rights reserved.

See the list of [contributors](https://github.com/drk1wi/Modlishka/graphs/contributors) who participated in this project.

* sentence copied directly from another project .

Disclaimer
----------
This tool is made only for educational purposes and can be used in legitimate penetration tests or research only. Author does not take any responsibility for any actions taken by its users.
