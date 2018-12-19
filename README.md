# ..Modlishka..

Modlishka is a flexible and powerful reverse proxy, that will take your phishing campaigns to the next level with minimal effort required. Enjoy :-)

Features
--------

Some of the most important 'Modlishka' features :

-   Support for majority of 2FA authentication schemes ( by design ).
-   No website templates (just point Modlishka to the target domain - it will be handled automatically ).
-   Interception of  cross origin TLS traffic from your victims browsers' ( with per victim context ).
-   Flexible  and easily configurable phishng scenarios through configuration options.
-   Per page JavaScript injection based on patterns.
-   User credential harvesting (context based on user identifiers).
-   Easy TLS certificate generation through 'Acme.sh' wrapper plugin.
-   Can be extended with your custom ideas through plugins.
-   Written in GO :-) 


Action
------
_"A picture is worth a thousand words":_

 Modlishka in action against an example 2FA (SMS) enabled authentication scheme:

[![Watch the video](https://i.vimeocdn.com/video/747311222_200x150.web)](https://vimeo.com/307369775)



Note: google.com was chosen for this POC, due to the complexity of the whole authentication flow. 


Installation
------------

Latest source code version can be fetched  from [here](https://github.com/drk1wi/modlishka/zipball/master) 

You can also clone the repository :

    git clone https://github.com/drk1wi/modlishka/master

Compile the binary and you are ready to go:

    make
    
Execute the binary:

    ./dist/proxy -h
    Usage of ./dist/proxy:
      -acmeDNSMethod string
        	Acme.sh DNS verification method to use (default "dns_aws")
      -acmeDomain string
        	Phishing domain for which we want to grab the LetsEncrypt certificate
      -acmeOuput string
        	Output directory for the generated json config file (default "/tmp/")
      -acmePath string
        	Path to the Acme.sh executable
      -cert string
        	base64 encoded TLS certificate
      -certKey string
        	base64 encoded TLS certificate key
      -certPool string
        	base64 encoded Certification Authority certificate
      -config string
        	JSON configuration file. Convenient instead of using command line switches.
      -debug
        	Print debug information
      -disableSecurity
        	Disable security features like anti-SSRF. Disable at your own risk.
      -dumpConfig
        	Print JSON config
      -jsRules string
        	Comma separated list of URL patterns and JS base64 encoded payloads that will be injected. 
      -listeningAddress string
        	Listening address (default "127.0.0.1")
      -listeningPort string
        	Listening port (default "443")
      -log string
        	Local file to which fetched requests will be written (appended)
      -phishing string
        	Phishing domain to create - Ex.: target.co
      -plugins string
        	Comma seperated list of enabled plugin names (default "all")
      -postOnly
        	Log only HTTP POST requests
      -rules string
        	Comma separated list of 'string' patterns and their replacements. 
      -target string
        	Main target to proxy - Ex.: https://target.com
      -targetRes string
        	Comma separated list of target subdomains that need to pass through the  proxy 
      -terminateTriggers string
        	Comma separated list of URLs from target's origin which will trigger session termination
      -terminateUrl string
        	URL to redirect the client after session termination triggers
      -tls
        	Enable TLS (default true)
      -trackingCookie string
        	Name of the HTTP cookie used to track the victim (default "id")
      -trackingParam string
        	Name of the HTTP parameter used to track the victim (default "id")



Usage
-----

 Check out the [wiki](https://github.com/drk1wi/Modlishka/wiki) page for a more detailed overview of the tool usage and different functionality descriptions.

Credits
-------

Thanks go to Giuseppe Trotta ([@Giutro](https://twitter.com/giutro)) 

Disclaimer
----------
This tool is made only for educational purposes and can be only used in legitimate penetration tests. Author does not take any responsibility for any actions taken by it users.

