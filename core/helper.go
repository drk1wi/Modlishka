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

package core

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"github.com/drk1wi/Modlishka/runtime"
	"net/http"
)

// GZIP content
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

// Deflate content
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
		http.Redirect(w, r, url, 302)
	} else {
		http.Redirect(w, r, "http://"+runtime.TopLevelDomain, 302)
	}
}
