/**

    "Modlishka" Reverse Proxy.

    Copyright 2018 (C) Piotr Duszyński piotr[at]duszynski.eu. All rights reserved.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

    You should have received a copy of the Modlishka License along with this program.

**/

package main

import (
	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/log"
	"github.com/drk1wi/Modlishka/plugin"
	"github.com/drk1wi/Modlishka/runtime"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/drk1wi/Modlishka/core"

	"golang.org/x/net/publicsuffix"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	runtime.MakeRegexes()
	//var f = false
	//config.C.Debug = &f
}
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var TestsTranslatePhishtoURL = []struct {
	input    string // input
	expected string // expected result
}{
	{"uc45bnjw1b6gyp4z7ab0.google.dev", "accounts.youtube.com"},
	{"www.google.dev", "www.google.com"},
	{"www.google.com", "www.google.com"},
	{"www.google.com.google.dev", "www.google.com.google.com"},
}

var TestsTranslateURLtoPhish = []struct {
	input    string // input
	expected string // expected result
}{
	{"assets-cdn.github.com", "assets-cdn.phish-github.dev"},
	{"accounts.youtube.com", "uc45bnjw1b6gyp4z7ab0.phish-github.dev"},
	{"github.com", "phish-github.dev"},
}

var TestsDynamicTranslateURLHost = []struct {
	input    string // input
	expected string // expected result
}{
	{"assets-cdn.github.com", "assets-cdn.github.com"},
	{"accounts.youtube.com", "accounts.youtube.com"},
	{"github.com", "github.com"},
}

func TestEncodeDecode(t *testing.T) {

	input := randSeq(20)
	encoded, err := runtime.EncodeSubdomain(input, false)
	if err != nil {
		t.Errorf("TestEncodeDecode(%s):  error %s ", "encode", err.Error())
	}

	decoded, _, _, err := runtime.DecodeSubdomain(encoded)
	if err != nil {
		t.Errorf("TestEncodeDecode(%s):  error %s ", "decode", err.Error())
	}

	if input != decoded {
		t.Errorf("TestEncodeDecode(%s): expected %s, actual %s", input, input, decoded)
	}

}

func TestRegex(t *testing.T) {

	runtime.RC4_KEY = `7afa263b3d6efb65dfde80875cf3883cdc4da6cef9b64034a5ba895317e98e39`
	runtime.ProxyDomain = "google.dev"
	runtime.TopLevelDomain = "google.com"
	runtime.MakeRegexes()

}

func TestTranslatePhishtoURL(t *testing.T) {

	runtime.RC4_KEY = `7afa263b3d6efb65dfde80875cf3883cdc4da6cef9b64034a5ba895317e98e39`
	target := "google.com"
	phishing := "google.dev"
	runtime.ProxyDomain = phishing
	runtime.TopLevelDomain = target
	runtime.MakeRegexes()

	domain, _ := publicsuffix.EffectiveTLDPlusOne(target)
	runtime.TopLevelDomain = strings.Replace(domain, "https://", "", -1)
	runtime.TopLevelDomain = strings.Replace(runtime.TopLevelDomain, "http://", "", -1)

	runtime.ProxyDomain = string(phishing)

	// core.Logger = core.InitializeLogger(*debugInfo)

	for _, tt := range TestsTranslatePhishtoURL {
		actual := runtime.PhishURLToRealURL(tt.input)
		if actual != tt.expected {
			t.Errorf("TestsTranslatePhishtoURL(%s): expected %s, actual %s", tt.input, tt.expected, actual)
		}
	}

}

func TestDynamicTranslateURLHost(t *testing.T) {

	// configure
	runtime.RC4_KEY = `7afa263b3d6efb65dfde80875cf3883cdc4da6cef9b64034a5ba895317e98e39`
	target := "github.com"
	phishing := "phish-github.dev"
	runtime.ProxyDomain = phishing
	runtime.TopLevelDomain = target
	runtime.MakeRegexes()
	domain, _ := publicsuffix.EffectiveTLDPlusOne(target)
	runtime.TopLevelDomain = strings.Replace(domain, "https://", "", -1)
	runtime.TopLevelDomain = strings.Replace(runtime.TopLevelDomain, "http://", "", -1)
	runtime.ProxyDomain = string(phishing)
	runtime.DynamicMode = true
	//

	for _, tt := range TestsDynamicTranslateURLHost {
		actual, _, _ := runtime.TranslateRequestHost(tt.input)
		if actual != tt.expected {
			t.Errorf("TestsTranslateURLtoPhish(%s): expected %s, actual %s", tt.input, tt.expected, actual)
		}
	}

}

func TestTranslateURLtoPhish(t *testing.T) {

	runtime.RC4_KEY = `7afa263b3d6efb65dfde80875cf3883cdc4da6cef9b64034a5ba895317e98e39`
	target := "github.com"
	phishing := "phish-github.dev"
	runtime.ProxyDomain = phishing
	runtime.TopLevelDomain = target
	runtime.MakeRegexes()

	domain, _ := publicsuffix.EffectiveTLDPlusOne(target)
	runtime.TopLevelDomain = strings.Replace(domain, "https://", "", -1)
	runtime.TopLevelDomain = strings.Replace(runtime.TopLevelDomain, "http://", "", -1)

	runtime.ProxyDomain = string(phishing)

	// core.Logger = core.InitializeLogger(*debugInfo)

	for _, tt := range TestsTranslateURLtoPhish {
		actual := runtime.RealURLtoPhish(tt.input)
		if actual != tt.expected {
			t.Errorf("TestsTranslateURLtoPhish(%s): expected %s, actual %s", tt.input, tt.expected, actual)
		}
	}

}

func getFieldString(v *config.Options, field string) string {
	out := reflect.ValueOf(v).Elem().FieldByName(field)
	return out.Elem().String()
}

func getFieldBool(v *config.Options, field string) bool {
	out := reflect.ValueOf(v).Elem().FieldByName(field)
	return out.Elem().Bool()
}

func TestCmdLineFlags(t *testing.T) {

	in := map[string]string{
		"ProxyDomain":          "https://google.dev",
		"ListeningAddress":     "0.0.0.0",
		"ProxyAddress":         "http://127.0.0.1:8080",
		"Target":               "google.com",
		"TargetRes":            "test.google.com,test1.google.com",
		"TerminateTriggers":    "terminate.google.dev,terminate2.google.dev",
		"TerminateRedirectUrl": "redirect.google.com",
		"TargetRules":          "eHh4:eXl5", //xxx:yyy
		"TrackingCookie":       "id",
		"TrackingParam":        "id",
		"LogRequestFile":       "logfile",
		"Plugins":              "plugin1,plugin2,plugin2",
	}

	in_bool := map[string]bool{
		"Debug":           true,
		"DisableSecurity": true,
		"LogPostOnly":     true,
		"ForceHTTP":       true,
		"ForceHTTPS":      true,
	}

	//encodedCert := base64.StdEncoding.EncodeToString([]byte(in["TLSCertificate"]))
	//encodedKey := base64.StdEncoding.EncodeToString([]byte(in["TLSKey"]))

	args := "   -proxyDomain " + in["ProxyDomain"] +
		" -proxyAddress " + in["ProxyAddress"] +
		" -target " + in["Target"] +
		" -listeningAddress " + in["ListeningAddress"] +
		" -targetRes " + in["TargetRes"] +
		" -terminateTriggers " + in["TerminateTriggers"] +
		" -terminateUrl " + in["TerminateRedirectUrl"] +
		" -rules " + in["TargetRules"] +
		" -trackingCookie " + in["TrackingCookie"] +
		" -trackingParam " + in["TrackingParam"] +
		" -log " + in["LogRequestFile"] +
		" -plugins " + in["Plugins"]

	if in_bool["Debug"] {
		args += " -debug "
	}

	if in_bool["ForceHTTP"] {
		args += " -forceHTTP "
	}

	if in_bool["ForceHTTPS"] {
		args += " -forceHTTPS "
	}

	if in_bool["DisableSecurity"] {
		args += " -disableSecurity "
	}

	if in_bool["LogPostOnly"] {
		args += " -postOnly "
	}

	arg := []string{os.Args[0]}
	for _, v := range strings.Fields(args) {
		arg = append(arg, v)

	}

	os.Args = arg

	options := config.ParseConfiguration()

	conf := Configuration{
		options,
	}

	for k, _ := range in {

		if getFieldString(&options, k) != in[k] {
			t.Errorf("TestCmdLineFlags ParseConfiguration (%s): expected %s, actual %s", k, in[k], getFieldString(&options, k))
		}
	}

	for k, _ := range in_bool {

		if getFieldBool(&options, k) != in_bool[k] {
			t.Errorf("TestCmdLineFlags ParseConfiguration (%s): expected %t, actual %t", k, in_bool[k], getFieldBool(&options, k))
		}
	}

	// Set up runtime core config
	runtime.SetCoreRuntimeConfig(conf.Options)

	if getFieldString(&options, "ProxyDomain") != runtime.ProxyDomain {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "ProxyDomain", getFieldString(&options, "ProxyDomain"), runtime.ProxyDomain)
	}

	if getFieldString(&options, "TrackingCookie") != runtime.TrackingCookie {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TrackingCookie", getFieldString(&options, "TrackingCookie"), runtime.TrackingCookie)

	}

	if getFieldString(&options, "TrackingParam") != runtime.TrackingParam {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TrackingParam", getFieldString(&options, "TrackingParam"), runtime.TrackingParam)

	}

	if "google.com" != runtime.TopLevelDomain {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TopLevelDomain", "google.com", runtime.TopLevelDomain)

	}

	if "yyy" != runtime.ReplaceStrings["xxx"] {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "ReplaceStrings", "xxx", runtime.ReplaceStrings["xxx"])

	}

	if strings.Join(runtime.TargetResources, ",") != getFieldString(&options, "TargetRes") {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TargetResources", getFieldString(&options, "TargetRes"), strings.Join(runtime.TargetResources, ","))

	}

	if strings.Join(runtime.TerminateTriggers, ",") != getFieldString(&options, "TerminateTriggers") {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TerminateTriggers", getFieldString(&options, "TerminateRedirectUrl"), strings.Join(runtime.TerminateTriggers, ","))

	}

	// Set up runtime server config
	core.SetServerRuntimeConfig(conf.Options)

	if in["TLSCertificate"] != *core.ServerRuntimeConfig.TLSCertificate {
		t.Errorf("TestCmdLineFlags SetServerRuntimeConfig (%s): expected %s, actual %s", "TLSCertificate", in["TLSCertificate"], *core.ServerRuntimeConfig.TLSCertificate)

	}

	if in["TLSKey"] != *core.ServerRuntimeConfig.TLSKey {
		t.Errorf("TestCmdLineFlags SetServerRuntimeConfig (%s): expected %s, actual %s", "TLSKey", in["TLSKey"], *core.ServerRuntimeConfig.TLSKey)

	}

	// Set up runtime plugin config
	plugin.SetPluginRuntimeConfig(conf.Options)

}

var jsonfile1 = `{
    "proxyDomain": "https://google.dev",
    "listeningAddress": "0.0.0.0",
    "target": "google.com",
    "targetResources": "test.google.com,test1.google.com",
    "targetRules": "eHh4:eXl5",
    "terminateTriggers": "terminate.google.dev,terminate2.google.dev",
    "terminateUrl": "redirect.google.com",
    "trackingCookie": "id",
    "trackingParam": "id",
    "debug": true,
    "logPostOnly": false,
    "disableSecurity": false,
    "log": "logfile",
    "plugins": "plugin1,plugin2,plugin2",
    "cert": "-----BEGIN CERTIFICATE-----\nMIIDEDCCAfigAwIBAgIEKfekOzANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQKEwdB\nY21lIENvMB4XDTE4MTIwMjIwMTc1NloXDTI0MDUwNzE5MTc1NlowPTEOMAwGA1UE\nBhMFRWFydGgxFjAUBgNVBAoTDU1vdGhlciBOYXR1cmUxEzARBgNVBAMTCmdvb2ds\nZS5kZXYwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDBzp66XCX6iPGK\n3DSy2ZcvcyDzL263U6CGHqwkFGySr8J3qrNeh4NZcnlYoAnobUlna9OCUPXFqA4/\nHjL6BuBsrLE//8gnrqP5Bga0ZYaTcq00EQuYxEpNuHBPsX0VBev/5qmJGa20Rd2O\nXajNGyK5S2eJhSOEDYY14tIVocPD9DTXsZ8TkVUxXZ8UqEaBDPp23OHL/HAFY/rd\nOybt1e9SZWC2bqsFjeoVM/xHBpuNDfhjivHI5AMNJGYvOxGtiqfOVUFNDc3zE1TC\nnBCpsesrpG4jB/6Q1yWdYogy5/7aUtM69GiXDDD4wG3l5MMxGhVFaspfKSc28IFG\nfJjMxH37AgMBAAGjQzBBMAwGA1UdEwEB/wQCMAAwDAYDVR0OBAUEAwECAzAjBgNV\nHREEHDAaggpnb29nbGUuZGV2ggwqLmdvb2dsZS5kZXYwDQYJKoZIhvcNAQELBQAD\nggEBAKSaZ04Q+Pv00PpugEi3FQtQOBz6JK/Exz8BOW6zOeY0NhfGrXjfa9rTqGdx\n0yxU1LQZhcNrdLKgIN3GGY/lYN0GKqBJFqmyy9zRxdob19Lb5HcL8ZY4fvFdrXBK\nI6D8eJhRmVY2Mr+v8fc2mDYg7q/kmgrcAtANtx3KC5QLtIWRxWn6iu+NO7FDKcsZ\nmJmHRikPR4PrhKyzuU9S5llUi7MvkHyZ+Daxj4pCvigEAPSVRepmdF96rf63fLWb\n0t0Uc01pFkyGFOZEBo/XkdOhWE4MRiYT0wFyGZLwJ9YOWRT1KwYsWedEUD+w1Elt\nUp4TXBYFCvw7HY+CQI9HKHh1GkM=\n-----END CERTIFICATE-----\n",
    "certKey": "-----BEGIN PRIVATE KEY-----\nMIIEpQIBAAKCAQEAwc6eulwl+ojxitw0stmXL3Mg8y9ut1Oghh6sJBRskq/Cd6qz\nXoeDWXJ5WKAJ6G1JZ2vTglD1xagOPx4y+gbgbKyxP//IJ66j+QYGtGWGk3KtNBEL\nmMRKTbhwT7F9FQXr/+apiRmttEXdjl2ozRsiuUtniYUjhA2GNeLSFaHDw/Q017Gf\nE5FVMV2fFKhGgQz6dtzhy/xwBWP63Tsm7dXvUmVgtm6rBY3qFTP8RwabjQ34Y4rx\nyOQDDSRmLzsRrYqnzlVBTQ3N8xNUwpwQqbHrK6RuIwf+kNclnWKIMuf+2lLTOvRo\nlwww+MBt5eTDMRoVRWrKXyknNvCBRnyYzMR9+wIDAQABAoIBAQCj6+X3DA+XWxKp\nd10fVMj5+i+JYLoNVy8zoWfJ0HiQjRY3burtbkLbeeZG3n3i1+S5E8s+ssldl6oN\nGrbVINHbOSlmTyp03dKUwtMS67gqqFj06+HaIVQTboeX8DAyguK8e9UzP8Pa8SjW\nzEME0AnLnYqCF1kVzPaSRzmX0E2rQz4ezJkMOUdjiH0OmMVLnezlrLr7w6Q8Swp3\nfyD2hd8g3ieoPLYOEVxYA8AVERxAVdli8Jm6w/Xcng7UlMnA+RP6zXJzdZx1iY8S\nNW9Yt/BlL34+3iHUt6lMUBa0SSzGxcgpBNU1/f5aAQZFGJIN7tJ1e8700jRTzvu+\ntFz31D5RAoGBAOXc3N1MiMXD4Gz0aSfmzWYEuJkvBBCmLHBNV2aMa05F4bnF0oZf\nEDLYKqqDxYqzzHuy1ySTKV1Z1P10hx+jbmZgQY6R8Uehc0TGnRnuz7AF9qDijjIY\nGiAZ4HoW3GT4l0SBZfcdb2dJSIO/PEgWn4CCN9sjSD9OwGLM5hyKxWRzAoGBANfY\nNDbj+aPg7hRbDFm4ZV1n+nwIGWq4M77/EuOPZcppfVrYl8EfCGcuoG+k8Wld2SoS\nz2N4kT2mnowSFE5OW0hRBojhOyUOPR7hLx8VoOF8Ymjl4WFsahELvQuXP+1Apq7Q\nZ0h+Gb2NkpRrgTJK8cUQf+8uIQM4SYpgAGw2dqZZAoGBANjdSoeDOJsVFXzWDwv1\nRh4VIDBt2jD3BoAhh+8ZVffwGGnTyK67q6W8qmxbjBkzTx35ed9o9CK9qSRDN2TT\nJUpzUAZ5jKEfIohltjyMQef5iFj7xlpewO8+Wrn1LZQZsWWRi6jcXYmd60tZNj9x\nEKUGtjoKjJQl8X6FgCi3iEofAoGARYgoieY27UvwZi5OdDiqrsRoNLyHM5HTWZvi\nAdyX9fS1pSZQ/K16j4K9vDlua3sIEj2tAWY9o5ahTI4mbHNhhJJVgJLN8sn7do8k\nFudoxDrFmPU0/aVnJcaaR7mZplxFVdtc6kV1FVMd/SIEpKbv64O9MtexWtAvIJx8\nhl+lKUECgYEAu9sAdc0pbzmdTeNterIScCXnclpANW1jsfCQvOv3qWqvU0uBreyd\nhVW67M9XzMzn6baZ3jLi0RxmIkxnLwkfLUTpMcmQO+1WY77MkROXDBmDQ87sBIDP\nluG0g5iz09m0QIt8nFUAZlogqgUXoMsBTtNk/jY4jpdTSzoh1kUeZIw=\n-----END PRIVATE KEY-----\n",
    "certPool": ""
}`

func TestJSONConfig(t *testing.T) {

	configFile, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatalf("%s... Terminating.\n", err.Error())
	}

	err = os.WriteFile(configFile.Name(), []byte(jsonfile1), 0644)
	if err != nil {
		log.Fatalf("%s... Terminating.\n", err.Error())
	}

	defer os.Remove(configFile.Name())

	in := map[string]string{
		"ProxyDomain":          "https://google.dev",
		"ListeningAddress":     "0.0.0.0",
		"Target":               "google.com",
		"TargetRes":            "test.google.com,test1.google.com",
		"TerminateTriggers":    "terminate.google.dev,terminate2.google.dev",
		"TerminateRedirectUrl": "redirect.google.com",
		"TargetRules":          "eHh4:eXl5", //xxx:yyy
		"TrackingCookie":       "id",
		"TrackingParam":        "id",
		"LogRequestFile":       "logfile",
		"Plugins":              "plugin1,plugin2,plugin2",
		"TLSCertificate":       "-----BEGIN CERTIFICATE-----\nMIIDEDCCAfigAwIBAgIEKfekOzANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQKEwdB\nY21lIENvMB4XDTE4MTIwMjIwMTc1NloXDTI0MDUwNzE5MTc1NlowPTEOMAwGA1UE\nBhMFRWFydGgxFjAUBgNVBAoTDU1vdGhlciBOYXR1cmUxEzARBgNVBAMTCmdvb2ds\nZS5kZXYwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDBzp66XCX6iPGK\n3DSy2ZcvcyDzL263U6CGHqwkFGySr8J3qrNeh4NZcnlYoAnobUlna9OCUPXFqA4/\nHjL6BuBsrLE//8gnrqP5Bga0ZYaTcq00EQuYxEpNuHBPsX0VBev/5qmJGa20Rd2O\nXajNGyK5S2eJhSOEDYY14tIVocPD9DTXsZ8TkVUxXZ8UqEaBDPp23OHL/HAFY/rd\nOybt1e9SZWC2bqsFjeoVM/xHBpuNDfhjivHI5AMNJGYvOxGtiqfOVUFNDc3zE1TC\nnBCpsesrpG4jB/6Q1yWdYogy5/7aUtM69GiXDDD4wG3l5MMxGhVFaspfKSc28IFG\nfJjMxH37AgMBAAGjQzBBMAwGA1UdEwEB/wQCMAAwDAYDVR0OBAUEAwECAzAjBgNV\nHREEHDAaggpnb29nbGUuZGV2ggwqLmdvb2dsZS5kZXYwDQYJKoZIhvcNAQELBQAD\nggEBAKSaZ04Q+Pv00PpugEi3FQtQOBz6JK/Exz8BOW6zOeY0NhfGrXjfa9rTqGdx\n0yxU1LQZhcNrdLKgIN3GGY/lYN0GKqBJFqmyy9zRxdob19Lb5HcL8ZY4fvFdrXBK\nI6D8eJhRmVY2Mr+v8fc2mDYg7q/kmgrcAtANtx3KC5QLtIWRxWn6iu+NO7FDKcsZ\nmJmHRikPR4PrhKyzuU9S5llUi7MvkHyZ+Daxj4pCvigEAPSVRepmdF96rf63fLWb\n0t0Uc01pFkyGFOZEBo/XkdOhWE4MRiYT0wFyGZLwJ9YOWRT1KwYsWedEUD+w1Elt\nUp4TXBYFCvw7HY+CQI9HKHh1GkM=\n-----END CERTIFICATE-----\n",
		"TLSKey":               "-----BEGIN PRIVATE KEY-----\nMIIEpQIBAAKCAQEAwc6eulwl+ojxitw0stmXL3Mg8y9ut1Oghh6sJBRskq/Cd6qz\nXoeDWXJ5WKAJ6G1JZ2vTglD1xagOPx4y+gbgbKyxP//IJ66j+QYGtGWGk3KtNBEL\nmMRKTbhwT7F9FQXr/+apiRmttEXdjl2ozRsiuUtniYUjhA2GNeLSFaHDw/Q017Gf\nE5FVMV2fFKhGgQz6dtzhy/xwBWP63Tsm7dXvUmVgtm6rBY3qFTP8RwabjQ34Y4rx\nyOQDDSRmLzsRrYqnzlVBTQ3N8xNUwpwQqbHrK6RuIwf+kNclnWKIMuf+2lLTOvRo\nlwww+MBt5eTDMRoVRWrKXyknNvCBRnyYzMR9+wIDAQABAoIBAQCj6+X3DA+XWxKp\nd10fVMj5+i+JYLoNVy8zoWfJ0HiQjRY3burtbkLbeeZG3n3i1+S5E8s+ssldl6oN\nGrbVINHbOSlmTyp03dKUwtMS67gqqFj06+HaIVQTboeX8DAyguK8e9UzP8Pa8SjW\nzEME0AnLnYqCF1kVzPaSRzmX0E2rQz4ezJkMOUdjiH0OmMVLnezlrLr7w6Q8Swp3\nfyD2hd8g3ieoPLYOEVxYA8AVERxAVdli8Jm6w/Xcng7UlMnA+RP6zXJzdZx1iY8S\nNW9Yt/BlL34+3iHUt6lMUBa0SSzGxcgpBNU1/f5aAQZFGJIN7tJ1e8700jRTzvu+\ntFz31D5RAoGBAOXc3N1MiMXD4Gz0aSfmzWYEuJkvBBCmLHBNV2aMa05F4bnF0oZf\nEDLYKqqDxYqzzHuy1ySTKV1Z1P10hx+jbmZgQY6R8Uehc0TGnRnuz7AF9qDijjIY\nGiAZ4HoW3GT4l0SBZfcdb2dJSIO/PEgWn4CCN9sjSD9OwGLM5hyKxWRzAoGBANfY\nNDbj+aPg7hRbDFm4ZV1n+nwIGWq4M77/EuOPZcppfVrYl8EfCGcuoG+k8Wld2SoS\nz2N4kT2mnowSFE5OW0hRBojhOyUOPR7hLx8VoOF8Ymjl4WFsahELvQuXP+1Apq7Q\nZ0h+Gb2NkpRrgTJK8cUQf+8uIQM4SYpgAGw2dqZZAoGBANjdSoeDOJsVFXzWDwv1\nRh4VIDBt2jD3BoAhh+8ZVffwGGnTyK67q6W8qmxbjBkzTx35ed9o9CK9qSRDN2TT\nJUpzUAZ5jKEfIohltjyMQef5iFj7xlpewO8+Wrn1LZQZsWWRi6jcXYmd60tZNj9x\nEKUGtjoKjJQl8X6FgCi3iEofAoGARYgoieY27UvwZi5OdDiqrsRoNLyHM5HTWZvi\nAdyX9fS1pSZQ/K16j4K9vDlua3sIEj2tAWY9o5ahTI4mbHNhhJJVgJLN8sn7do8k\nFudoxDrFmPU0/aVnJcaaR7mZplxFVdtc6kV1FVMd/SIEpKbv64O9MtexWtAvIJx8\nhl+lKUECgYEAu9sAdc0pbzmdTeNterIScCXnclpANW1jsfCQvOv3qWqvU0uBreyd\nhVW67M9XzMzn6baZ3jLi0RxmIkxnLwkfLUTpMcmQO+1WY77MkROXDBmDQ87sBIDP\nluG0g5iz09m0QIt8nFUAZlogqgUXoMsBTtNk/jY4jpdTSzoh1kUeZIw=\n-----END PRIVATE KEY-----\n",
	}

	in_bool := map[string]bool{
		"Debug":           true,
		"DisableSecurity": false,
		"LogPostOnly":     false,
	}

	args := "  -config " + configFile.Name()

	arg := []string{os.Args[0]}
	for _, v := range strings.Fields(args) {
		arg = append(arg, v)

	}

	os.Args = arg

	options := config.ParseConfiguration()

	conf := Configuration{
		options,
	}

	for k, _ := range in {

		if getFieldString(&options, k) != in[k] {
			t.Errorf("TestCmdLineFlags ParseConfiguration (%s): expected %s, actual %s", k, in[k], getFieldString(&options, k))
		}
	}

	for k, _ := range in_bool {

		if getFieldBool(&options, k) != in_bool[k] {
			t.Errorf("TestCmdLineFlags ParseConfiguration (%s): expected %t, actual %t", k, in_bool[k], getFieldBool(&options, k))
		}
	}

	// Set up runtime core config
	runtime.SetCoreRuntimeConfig(conf.Options)

	if getFieldString(&options, "ProxyDomain") != runtime.ProxyDomain {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "ProxyDomain", getFieldString(&options, "ProxyDomain"), runtime.ProxyDomain)
	}

	if getFieldString(&options, "TrackingCookie") != runtime.TrackingCookie {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TrackingCookie", getFieldString(&options, "TrackingCookie"), runtime.TrackingCookie)

	}

	if getFieldString(&options, "TrackingParam") != runtime.TrackingParam {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TrackingParam", getFieldString(&options, "TrackingParam"), runtime.TrackingParam)

	}

	if "google.com" != runtime.TopLevelDomain {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TopLevelDomain", "google.com", runtime.TopLevelDomain)

	}

	if "yyy" != runtime.ReplaceStrings["xxx"] {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "ReplaceStrings", "xxx", runtime.ReplaceStrings["xxx"])

	}

	if strings.Join(runtime.TargetResources, ",") != getFieldString(&options, "TargetRes") {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TargetResources", getFieldString(&options, "TargetRes"), strings.Join(runtime.TargetResources, ","))

	}

	if strings.Join(runtime.TerminateTriggers, ",") != getFieldString(&options, "TerminateTriggers") {
		t.Errorf("TestCmdLineFlags SetCoreRuntimeConfig (%s): expected %s, actual %s", "TerminateTriggers", getFieldString(&options, "TerminateRedirectUrl"), strings.Join(runtime.TerminateTriggers, ","))

	}

	// Set up runtime server config
	core.SetServerRuntimeConfig(conf.Options)

	if in["TLSCertificate"] != *core.ServerRuntimeConfig.TLSCertificate {
		t.Errorf("TestCmdLineFlags SetServerRuntimeConfig (%s): expected %s, actual %s", "TLSCertificate", in["TLSCertificate"], *core.ServerRuntimeConfig.TLSCertificate)

	}

	if in["TLSKey"] != *core.ServerRuntimeConfig.TLSKey {
		t.Errorf("TestCmdLineFlags SetServerRuntimeConfig (%s): expected %s, actual %s", "TLSKey", in["TLSKey"], *core.ServerRuntimeConfig.TLSKey)

	}

	// Set up runtime plugin config
	plugin.SetPluginRuntimeConfig(conf.Options)

}
