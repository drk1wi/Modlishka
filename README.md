# ..Modlishka..

Modlishka is a powerful and flexible HTTP reverse proxy. It implements a new approach of handling HTTP traffic flow, which allows transparent and automated proxying of multi-destination HTTP/S traffic over a single TLS certificate without additional configuration overhead. 

Modlishka was primarily written for ethical penetration tests. Nevertheless, it can be helpful in other, non-security related, usage scenarios.

From a security testing perspective, Modlishka can be used to:
-	Carry out a modern ethical phishing campaign that requires a universal 2FA “bypass” support.
-	Diagnose and exploit browser-based applications from a "[Client Domain Hooking](https://blog.duszynski.eu/hijacking-browser-tls-traffic-through-client-domain-hooking/)" attack perspective.  

General aim of this release was to:
- Highlight currently used two factor authentication ([2FA](https://blog.duszynski.eu/phishing-ng-bypassing-2fa-with-modlishka/)) scheme weaknesses, so adequate security solutions can be created and implemented soon. 
- Create a diagnostic tool for the "Client Domain Hooking' attack.
- Support projects that could benefit from a universal and automated HTTP reverse proxy.
- Raise community awareness about modern phishing techniques and strategies.
- Support penetration testers in their ethical phishing campaigns and help to reinforce the fact that serious threat can arise from modern phishing attacks.


Enjoy  :-)

Features
--------

Some of the most important 'Modlishka' features :

-   Support for majority of 2FA authentication schemes (by design).
-   Practical implementation of the "Client Domain Hooking" attack.
-   No website templates (just point Modlishka to the target domain - in most cases, it will be handled automatically without any additional manual configuration).
-   Full control of "cross" origin TLS traffic flow from your users browsers (through custom new techniques).
-   Flexible and easily configurable phishing scenarios through configuration options.
-   Pattern based JavaScript payload injection.
-   Wrapping in-secure websites with TLS and additional security headers.
-   Striping website from all encryption and security headers (back to 90's MITM style). 
-   User credential harvesting (with context based on URL parameter passed identifiers).
-   Can be extended easily with your ideas through modular plugins.
-   Stateless design. Can be scaled up easily for an arbitrary number of users - ex. through a DNS load balancer.
-   Web panel with a summary of collected credentials and user session impersonation (beta POC).
-   Written in Go. 


Proxy In Action (2FA bypass)
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
    
    
![alt text](https://raw.githubusercontent.com/drk1wi/assets/master/7d0426a133a85a46a76a424574bf5a2acf99815e.png)

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
        	Comma separated list of URL patterns and JS base64 encoded payloads that will be injected. 
      
      -listeningAddress string
        	Listening address - e.g.: 0.0.0.0  (default "127.0.0.1")
      
      -log string
        	Local file to which fetched requests will be written (appended)
      
      -phishingDomain string
        	Proxy domain name that will be used - e.g.: proxy.tld
      
      -plugins string
        	Comma seperated list of enabled plugin names (default "all")
      
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

 * [WIKI](https://github.com/drk1wi/Modlishka/wiki) pages with detailed description of the tool usage.
 * Modlishak introduction blog [post](https://blog.duszynski.eu/phishing-ng-bypassing-2fa-with-modlishka/).
 * Implemented technique described in detaild "[Client Domain Hooking](https://blog.duszynski.eu/hijacking-browser-tls-traffic-through-client-domain-hooking/)" - in case you are interested how this tool handles multiple domains over a single TLS certificate.
 * [FAQ](https://github.com/drk1wi/Modlishka/wiki/FAQ) (Frequently Asked Questions).

License
-------
Modlishka was created by Piotr Duszyński ([@drk1wi](https://twitter.com/drk1wi)). You can find the license [here](https://github.com/drk1wi/Modlishka/blob/master/LICENSE). 

Credits
-------
Thanks for helping with the code refactoring go to Giuseppe Trotta ([@Giutro](https://twitter.com/giutro)). 

Disclaimer
----------
This tool is made only for educational purposes and can be used in legitimate penetration tests, that have all required legal approvals . Author does not take any responsibility for any actions taken by its users.

