# ..Modlishka..

Modlishka is a powerful and flexible HTTP reverse proxy. It implements an entirely new approach of handling HTTP traffic flow, which allows to transparently proxy multi-domain destination TLS traffic over a single domain TLS certificate in an automated manner. What does this exactly mean? In short, it simply has a lot of potential, that can be used in many interesting ways... 

From the security perspective, Modlishka can be currently used to:
-	 Help penetration testers to carry out a modern ethical phishing campaign that requires a universal 2FA “bypass” support.
-  Hijack application HTTP TLS traffic flow through the "Client Domain Hooking" attack.
-  Wrap legacy websites with TLS layer, confuse crawler bots and automated scanners, etc.
-  TBC

Modlishka was written as an attempt overcome standard reverse proxy limitations and as a personal challenge to see what is possible with sufficient motivation and a bit of extra research time. 
The achieved results appeared to be very interesting and the tool was initially released and later updated with aim to:
- Highlight currently used two factor authentication ([2FA](https://blog.duszynski.eu/phishing-ng-bypassing-2fa-with-modlishka/)) scheme weaknesses, so adequate security solutions can be created and implemented by the industry and raise user awareness.
- Provide a diagnostic tool for the "Client Domain Hooking' attack.
- Support open source projects that could benefit from a universal and transparent TLS HTTP reverse proxy.
- Raise community awareness about modern phishing techniques and strategies and support penetration testers in their ethical phishing campaigns.


Modlishka was primarily written for security related tasks. Nevertheless, it can be helpful in other, non-security related, usage scenarios.

Efficient proxying !

Features
--------

Some of the most important 'Modlishka' features :

**General:**
-   Point-and-click HTTP and HTTPS reverse proxying of an arbitrary domain.
-   Full control of "cross" origin TLS traffic flow from your users browses (through a set of new interesting techniques).
-   Easy and fast configuration through command line options and JSON configuration files.
-   Practical implementation of the "Client Domain Hooking" attack. Supported with a diagnostic plugin.
-   Pattern based JavaScript payload injection.
-   Wrapping websites with an extra "security": TLS wrapping, authentication, relevant security headers, etc. 
-   Striping websites from all encryption and security headers (back to 90's MITM style). 
-   Stateless design. Can be scaled up easily to handle an arbitrary amount of traffic  - e.g. through a DNS load balancer.
-   Can be extended easily with your ideas through modular plugins.
-   Automatic TLS certificate generation plugin for the proxy domain (requires a self-signed CA certificate)
-   Written in Go, so it works basically on all platforms and architectures: Windows, OSX, Linux, BSD supported...

**Security related:**
- "[Client Domain Hooking](https://blog.duszynski.eu/client-domain-hooking-in-practice/)" attack in form of a diagnostic module.
-  Support for majority of 2FA authentication schemes (out of the box).
-  User credential harvesting (with context based on URL parameter passed identifiers).
-  Web panel plugin with a summary of automatically collected credentials and one-click user session impersonation module (beta POC).
-  No website templates (just point Modlishka to the target domain - in most cases, it will be handled automatically without any additional manual configuration).


Proxying In Action (2FA bypass)
------
_"A picture is worth a thousand words":_

 Modlishka in action against an example two factor authentication scheme (SMS based)  :

[![Watch the video](https://i.vimeocdn.com/video/749353683.jpg)](https://vimeo.com/308709275)

[https://vimeo.com/308709275](https://vimeo.com/308709275)


Installation
------------

Latest source code version can be fetched from [here](https://github.com/drk1wi/modlishka/zipball/master) (zip) or [here](https://github.com/drk1wi/modlishka/tarball/master) (tar).

Fetch the code with _'go get'_ :

    $ go get -u github.com/drk1wi/Modlishka

Compile the binary and you are ready to go:

    $ cd $GOPATH/src/github.com/drk1wi/Modlishka/
    $ make
    
    
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
      
      -credParams string
          	Credential regexp with matching groups. e.g. : baase64(username_regex),baase64(password_regex)

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
        	Name of the HTTP cookie used to track the victim (default "id")
      
      -trackingParam string
        	Name of the HTTP parameter used to track the victim (default "id")



References
-----

 * [WIKI](https://github.com/drk1wi/Modlishka/wiki) pages:  with more details about the tool usage and configuration.
 * [FAQ](https://github.com/drk1wi/Modlishka/wiki/FAQ)

 Blog posts:
 *  ["Modlishka introduction"](https://blog.duszynski.eu/phishing-ng-bypassing-2fa-with-modlishka/) blog post.
 * "[Hijacking browser TLS traffic through Client Domain Hooking](https://blog.duszynski.eu/hijacking-browser-tls-traffic-through-client-domain-hooking/)" technical paper - in case you are interested about the approach that is used to handle the traffic.

License
-------
Author: Modlishka was designed and implemented by Piotr Duszyński ([@drk1wi](https://twitter.com/drk1wi)) (this includes the technique described in the "Client Domain Hooking" paper) . You can find the relevant license [here](https://github.com/drk1wi/Modlishka/blob/master/LICENSE). All rights reserved.

The initial version of the tool was written as part of a bigger project that was dissolved and assets were distributed accordingly. 

Credits 
-------
Kudos for helping with the final code optimization and great support go to Giuseppe Trotta ([@Giutro](https://twitter.com/giutro)). 

Disclaimer
----------
This tool is made only for educational purposes and can be used in legitimate penetration tests or research only. Author does not take any responsibility for any actions taken by its users.

