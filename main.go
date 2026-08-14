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

package main

import (
	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/core"
	"github.com/drk1wi/Modlishka/log"
	"github.com/drk1wi/Modlishka/plugin"
	"github.com/drk1wi/Modlishka/runtime"
)

type Configuration struct{ config.Options }

// Initializes the logging object

func (c *Configuration) initLogging() {
	//
	// Logger
	//
	log.WithColors = true

	if *c.Debug == true {
		log.MinLevel = log.DEBUG
	} else {
		log.MinLevel = log.INFO
	}

	logGET := true
	if *c.LogPostOnly {
		logGET = false
	}

	log.Options = log.LoggingOptions{
		GET:            logGET,
		POST:           *c.LogPostOnly,
		LogRequestPath: *c.LogRequestFile,
	}
}

func main() {

	conf := Configuration{
		config.ParseConfiguration(),
	}

	// Initialize log
	conf.initLogging()

	// Set up runtime plugin config
	plugin.SetPluginRuntimeConfig(conf.Options)

	// Initialize plugins
	plugin.Enable(conf.Options)

	//Check if we have all of the required information to start proxy'ing requests.
	conf.VerifyConfiguration()

	// Set up runtime core config
	runtime.SetCoreRuntimeConfig(conf.Options)

	// Set up runtime server config
	core.SetServerRuntimeConfig(conf.Options)

	// Set up regexp upfront
	runtime.MakeRegexes()

	// go go go
	core.RunServer()

}
