/**

    "Modlishka" Reverse Proxy.

    Copyright 2018 (C) Piotr Duszyński piotr[at]duszynski.eu. All rights reserved.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

    You should have received a copy of the Modlishka License along with this program.

**/

package config

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/drk1wi/Modlishka/log"
)

type Options struct {
	ProxyDomain            *string `json:"proxyDomain"`
	ListeningAddress       *string `json:"listeningAddress"`
	ListeningPortHTTP      *int    `json:"listeningPortHTTP"`
	ListeningPortHTTPS     *int    `json:"listeningPortHTTPS"`
	ProxyAddress           *string `json:"proxyAddress"`
	StaticLocations        *string `json:"staticLocations"`
	Target                 *string `json:"target"`
	TargetRes              *string `json:"targetResources"`
	TargetRules            *string `json:"rules"`
	JsRules                *string `json:"jsRules"`
	TerminateTriggers      *string `json:"terminateTriggers"`
	TerminateRedirectUrl   *string `json:"terminateRedirectUrl"`
	TrackingCookie         *string `json:"trackingCookie"`
	TrackingParam          *string `json:"trackingParam"`
	Debug                  *bool   `json:"debug"`
	ForceHTTPS             *bool   `json:"forceHTTPS"`
	ForceHTTP              *bool   `json:"forceHTTP"`
	LogPostOnly            *bool   `json:"logPostOnly"`
	DisableSecurity        *bool   `json:"disableSecurity"`
	DynamicMode            *bool   `json:"dynamicMode"`
	LogRequestFile         *string `json:"log"`
	Plugins                *string `json:"plugins"`
	AllowSecureCookies     *bool   `json:"allowSecureCookies"`
	IgnoreTranslateDomains *string `json:"ignoreTranslateDomains"`
	DisableDynamicSubdomains *bool `json:"disableDynamicSubdomains"`
	PathHostRules          *string `json:"pathHostRules"`
	*TLSConfig
}

type TLSConfig struct {
	TLSCertificate *string `json:"cert"`
	TLSKey         *string `json:"certKey"`
	TLSPool        *string `json:"certPool"`
}

var (
	C = Options{
		ProxyDomain:        flag.String("proxyDomain", "", "Proxy domain name that will be used - e.g.: proxy.tld"),
		ListeningAddress:   flag.String("listeningAddress", "127.0.0.1", "Listening address - e.g.: 0.0.0.0 "),
		ListeningPortHTTP:  flag.Int("listeningPortHTTP", 80, "Listening port for HTTP requests"),
		ListeningPortHTTPS: flag.Int("listeningPortHTTPS", 443, "Listening port for HTTPS requests"),
		Target:             flag.String("target", "", "Target  domain name  - e.g.: target.tld"),
		TargetRes: flag.String("targetRes", "",
			"Comma separated list of domains that were not translated automatically. Use this to force domain translation - e.g.: static.target.tld"),
		TerminateTriggers: flag.String("terminateTriggers", "",
			"Session termination: Comma separated list of URLs from target's origin which will trigger session termination"),
		TerminateRedirectUrl: flag.String("terminateUrl", "",
			"URL to which a client will be redirected after Session Termination rules trigger"),
		TargetRules: flag.String("rules", "",
			"Comma separated list of 'string' patterns and their replacements - e.g.: base64(old):base64(new),base64(older):base64(newer)"),
		JsRules: flag.String("jsRules", "", "Comma separated list of URL patterns and JS base64 encoded payloads that will be injected - e.g.: target.tld:base64(alert(1)),..,etc"),

		ProxyAddress:    flag.String("proxyAddress", "", "Proxy that should be used (socks/https/http) - e.g.: http://127.0.0.1:8080 "),
		StaticLocations: flag.String("staticLocations", "", "FQDNs in location headers that should be preserved."),

		TrackingCookie:  flag.String("trackingCookie", "id", "Name of the HTTP cookie used for track the client"),
		TrackingParam:   flag.String("trackingParam", "id", "Name of the HTTP parameter used to set up the HTTP cookie tracking of the client"),
		Debug:           flag.Bool("debug", false, "Print extra debug information"),
		DisableSecurity: flag.Bool("disableSecurity", false, "Disable proxy security features like anti-SSRF. 'Here be dragons' - disable at your own risk."),
		DynamicMode:     flag.Bool("dynamicMode", false, "Enable dynamic mode for 'Client Domain Hooking'"),

		ForceHTTP:  flag.Bool("forceHTTP", false, "Strip all TLS from the traffic and proxy through HTTP only"),
		ForceHTTPS: flag.Bool("forceHTTPS", false, "Strip all clear-text from the traffic and proxy through HTTPS only"),

		LogRequestFile: flag.String("log", "", "Local file to which fetched requests will be written (appended)"),

		LogPostOnly: flag.Bool("postOnly", false, "Log only HTTP POST requests"),

		Plugins:                flag.String("plugins", "all", "Comma separated list of enabled plugin names"),
		AllowSecureCookies:     flag.Bool("allowSecureCookies", false, "Allow secure cookies to be set. Useful for when you are using HTTPS and cookies have SameSite=None"),
		IgnoreTranslateDomains: flag.String("ignoreTranslateDomains", "", "Comma separated list of domains to never translate and proxy"),

		DisableDynamicSubdomains: flag.Bool("disableDynamicSubdomains", false, "Translate URL domain names to be the proxy domain"),
		PathHostRules:            flag.String("pathHostRules", "",
			"Comma separated list of URL path patterns and the target domains to send the requests to - e.g.: /path/:example.com,/path2:www.example.com"),
	}

	s = TLSConfig{
		TLSCertificate: flag.String("cert", "", "base64 encoded TLS certificate"),
		TLSKey:         flag.String("certKey", "", "base64 encoded TLS certificate key"),
		TLSPool:        flag.String("certPool", "", "base64 encoded Certification Authority certificate"),
	}

	JSONConfig = flag.String("config", "", "JSON configuration file. Convenient instead of using command line switches.")
)

func ParseConfiguration() Options {

	flag.Parse()

	// Parse JSON for config
	if len(*JSONConfig) > 0 {
		C.parseJSON(*JSONConfig)
	}

	// Process TLS configuration
	C.TLSConfig = &s

	// we can assume that if someone specified one of the following cmd line parameters then he should define all of them.
	if len(*s.TLSCertificate) > 0 || len(*s.TLSKey) > 0 || len(*s.TLSPool) > 0 {

		// Handle TLS Certificates
		if *C.ForceHTTP == false {
			if len(*C.TLSCertificate) > 0 {
				decodedCertificate, err := base64.StdEncoding.DecodeString(*C.TLSCertificate)
				if err == nil {
					*C.TLSCertificate = string(decodedCertificate)

				}
			}

			if len(*C.TLSKey) > 0 {
				decodedCertificateKey, err := base64.StdEncoding.DecodeString(*C.TLSKey)
				if err == nil {
					*C.TLSKey = string(decodedCertificateKey)
				}
			}

			if len(*C.TLSPool) > 0 {
				decodedCertificatePool, err := base64.StdEncoding.DecodeString(*C.TLSPool)
				if err == nil {
					*C.TLSPool = string(decodedCertificatePool)
				}
			}
		}

	}

	return C
}

func (c *Options) parseJSON(file string) {

	ct, err := os.Open(file)
	defer ct.Close()
	if err != nil {
		log.Fatalf("Error opening JSON configuration (%s): %s . Terminating.", file, err)
	}

	ctb, _ := io.ReadAll(ct)
	err = json.Unmarshal(ctb, &c)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON configuration (%s): %s . Terminating.", file, err)
	}

	err = json.Unmarshal(ctb, &s)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON configuration (%s): %s . Terminating.", file, err)
	}

	C.TLSConfig = &s

}

func (c *Options) VerifyConfiguration() {

	if *c.ForceHTTP == true {
		if len(*c.ProxyDomain) == 0 || len(*c.ProxyDomain) == 0 {
			log.Warningf("Missing required parameters in oder start the proxy. Terminating.")
			log.Warningf("TIP: You will need to specify at least the following parameters to serve the page over HTTP: proxyDomain and target.")
			flag.PrintDefaults()
			os.Exit(1)
		}
	} else { // default + HTTPS wrapper

		if len(*c.ProxyDomain) == 0 || len(*c.ProxyDomain) == 0 {
			log.Warningf("Missing required parameters in oder start the proxy. Terminating.")
			log.Warningf("TIP: You will need to specify at least the following parameters to serve the page over HTTP: proxyDomain and target.")
			flag.PrintDefaults()
			os.Exit(1)
		}

	}

	if *c.DynamicMode == true {
		log.Warningf("Dynamic Mode enabled: Proxy will accept and hook all incoming HTTP requests.")
	}

	if *c.ForceHTTP == true {
		log.Warningf("Force HTTP wrapper enabled: Proxy will strip all TLS traffic and handle requests over HTTP only")
	}

	if *c.ForceHTTPS == true {
		log.Warningf("Force HTTPS wrapper enabled: Proxy will strip all clear-text traffic and handle requests over HTTPS only")
	}

}
