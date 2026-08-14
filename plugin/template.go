/**

    "Modlishka" Reverse Proxy.

    Copyright (C) 2018 Piotr Duszyński piotr[at]duszynski.eu.

    SPDX-License-Identifier: GPL-3.0-or-later

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

**/

package plugin

import (
	"encoding/json"
	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/log"
	"io"
	"net/http"
	"os"
)

type ExtendedConfiguration struct {
	*config.Options
	ExtraField string `json:"ExtraField"`
}

func init() {

	s := Property{}
	s.Name = "template"
	s.Description = "This is a template plugin, that can be used as starting point for your extensions."
	s.Version = "0.1"

	//init all of the vars, print a welcome message, init your command line flags here
	s.Init = func() {}

	// process all of the cmd line flags and config file (if supplied)
	s.Flags = func() {

		if len(*config.JSONConfig) > 0 {

			var jsonConfig ExtendedConfiguration

			ct, err := os.Open(*config.JSONConfig)
			if err != nil {
				log.Errorf("Error opening JSON configuration (%s): %s", *config.JSONConfig, err)
				return
			}
			defer ct.Close()

			ctb, _ := io.ReadAll(ct)
			err = json.Unmarshal(ctb, &jsonConfig)
			if err != nil {
				log.Errorf("Error unmarshalling JSON configuration (%s): %s", *config.JSONConfig, err)
				return
			}
		}
	}

	//process HTTP request
	s.HTTPRequest = func(req *http.Request, context *HTTPContext) {}

	//process HTTP response (responses can arrive in random order)
	s.HTTPResponse = func(resp *http.Response, context *HTTPContext, buffer *[]byte) {}

	// Register your http handlers
	s.RegisterHandler = func(handler *http.ServeMux) {

	}

	// Register all the function hooks
	s.Register()

}
