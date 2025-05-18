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
	"github.com/drk1wi/Modlishka/config"
	"net/http"
	"net/url"
	"strings"

	"github.com/drk1wi/Modlishka/log"
)

var (
	Plugins           []*Property
	EnabledPluginList *[]string
)

type Property struct {
	Name        string
	Description string
	Version     string
	Active      bool

	Init            func()
	Flags           func()
	HTTPRequest     func(req *http.Request, context *HTTPContext)
	HTTPResponse    func(resp *http.Response, context *HTTPContext, buffer *[]byte)
	TerminateUser   func(userID string)
	RegisterHandler func(handler *http.ServeMux)
}

type HTTPContext struct {
	Target         *url.URL          // Target URL after going through the proxy
	OriginalTarget string            // Original Host
	Origin         string            // Origin before going through the proxy
	UserID         string            // traced user identifier
	InitUserID     string            // traced user id
	JSPayload      string            // JS Payload
	IP             string            // victim's IP address
	IsTLS          bool              //TLS request
	Extra          map[string]string //Extra Plugin Data
}

// Add the given Plugin to the list of loaded plugins
func (p *Property) Register() {
	Plugins = append(Plugins, p)
}

// Enable the specified Plugin
func (p *Property) Enable() {

	log.Infof("Enabling plugin: %s v%s", p.Name, p.Version)
	p.Active = true

	// Invoke plugin'p init function
	if p.Init != nil {
		p.Init()
	}

	// Parse plugin'p parse flags function
	if p.Flags != nil {
		p.Flags()
	}

}

func SetPluginRuntimeConfig(conf config.Options) {

	var list []string

	if EnabledPluginList == nil {
		EnabledPluginList = &list
	}

	if conf.Plugins == nil {
		return
	}

	if *conf.Plugins != "" {
		for _, i := range strings.Split(*conf.Plugins, ",") {
			i = strings.Trim(i, " ")
			list = append(list, i)
		}

		EnabledPluginList = &list
	}
}

// Enable the provided plugins
func Enable(conf config.Options) {

	for _, pluginName := range *EnabledPluginList {
		found := false
		for _, p := range Plugins {
			if *conf.Plugins == "all" {
				p.Enable()
				found = true
				continue
			}

			if strings.EqualFold(p.Name, pluginName) {
				p.Enable()
				found = true
				continue
			}
		}

		if !found {
			log.Errorf("Plugin %s not found", pluginName)
		}
	}

}

// Register HTTP handlers
func RegisterHandler(handler *http.ServeMux) {
	for _, p := range Plugins {

		if p.RegisterHandler != nil && p.Active {
			p.RegisterHandler(handler)
		}

	}
}

// Execute plugin-defined HTTP Request hooks
func (context *HTTPContext) InvokeHTTPRequestHooks(req *http.Request) {
	for _, p := range Plugins {
		if p.Active && p.HTTPRequest != nil {
			p.HTTPRequest(req, context)
		}
	}

}

// Execute plugin-defined HTTP Response hooks
func (context *HTTPContext) InvokeHTTPResponseHooks(resp *http.Response, buffer *[]byte) {
	for _, p := range Plugins {
		if p.Active == true && p.HTTPResponse != nil {
			p.HTTPResponse(resp, context, buffer)
		}
	}
}

// Execute plugin-defined Terminate User hooks
func (context *HTTPContext) InvokeTerminateUserHooks(userID string) {
	for _, p := range Plugins {
		if p.Active && p.TerminateUser != nil {
			p.TerminateUser(userID)
		}
	}

}
