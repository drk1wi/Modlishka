/**

    "Modlishka" Reverse Proxy.

    Copyright 2018 (C) Piotr Duszyński piotr[at]duszynski.eu. All rights reserved.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

    You should have received a copy of the Modlishka License along with this program.

**/

package plugin

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/log"
)

// Paste your CA certificate and key in the following format
// Ref: https://github.com/drk1wi/Modlishka/wiki/Quickstart-tutorial

const CA_CERT = `-----BEGIN CERTIFICATE-----
MIIDRDCCAiwCCQC/MUuFNX64sjANBgkqhkiG9w0BAQsFADBkMQswCQYDVQQGEwJw
bDELMAkGA1UECAwCcGwxCzAJBgNVBAcMAnBsMQswCQYDVQQKDAJwbDELMAkGA1UE
CwwCcGwxCzAJBgNVBAMMAnBsMRQwEgYJKoZIhvcNAQkBFgVwbEBwbDAeFw0xODEy
MjQwOTAzNDhaFw0yMTEwMTMwOTAzNDhaMGQxCzAJBgNVBAYTAnBsMQswCQYDVQQI
DAJwbDELMAkGA1UEBwwCcGwxCzAJBgNVBAoMAnBsMQswCQYDVQQLDAJwbDELMAkG
A1UEAwwCcGwxFDASBgkqhkiG9w0BCQEWBXBsQHBsMIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAvv+AkWp9ek8jxWqs810X1pYhUZrpUxswfWVFEwmxhMXe
5w+0Dd5sQPksBUzO9OS58bD1srW+CTa+gL+yUC3xb2lA+5bu4B9kfKO/X5wPx4hE
Ek1+1VIUWds5G7Mxndd5p2SGtt2sMnbKIRYgbYpxlHOpb2kLk2kF1S77SeQUSztJ
NZGJ9DGGlHErSaLvJE03s/YXerRK4BthubC7YluaZLN1PUNmMaXpsk+7GSCJFCTi
7g6PVCc9FGUZVlPtnXMYu+7ehRpdjiyiLRk7UTT0VTAQxVN7aWZxU6mBGs9mvQa1
P911Gx0ApqDPpE35i80QqHw7rhQE1ah58by+hrwNXQIDAQABMA0GCSqGSIb3DQEB
CwUAA4IBAQCQINE4Yqk59BvLSCl+kf78Wp8zEeFBiRxKBC5n6/BnW2ehDvogyLkw
MiBHifw3d9YsYEVejfw6aw+eLVGCb/fk26Yb1rlLiMqpcy02p4F2XRZjXaidJrhT
ngLpDX71HyMkgfwK0Nl7XBlT3LrTS8ASugR4Pr4xZVb/ApmIEo5BwEthDiRhwKnS
3GhJxjTwkMd01Rrr6bGVaTtUg3BZbOtrEQqsEac4zX8luG/lG4tu2SSBhbT5CdWm
oU1HwMCmJWKexXSShADTjcULtlMOL56P36y/fbu5xikEdfURAcIC/+bTqDw+twsN
sQX5u89N+VCZf3xg5wSxB0shI6WSQX0A
-----END CERTIFICATE-----`

const CA_CERT_KEY = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvv+AkWp9ek8jxWqs810X1pYhUZrpUxswfWVFEwmxhMXe5w+0
Dd5sQPksBUzO9OS58bD1srW+CTa+gL+yUC3xb2lA+5bu4B9kfKO/X5wPx4hEEk1+
1VIUWds5G7Mxndd5p2SGtt2sMnbKIRYgbYpxlHOpb2kLk2kF1S77SeQUSztJNZGJ
9DGGlHErSaLvJE03s/YXerRK4BthubC7YluaZLN1PUNmMaXpsk+7GSCJFCTi7g6P
VCc9FGUZVlPtnXMYu+7ehRpdjiyiLRk7UTT0VTAQxVN7aWZxU6mBGs9mvQa1P911
Gx0ApqDPpE35i80QqHw7rhQE1ah58by+hrwNXQIDAQABAoIBAHl833SffczMhf5O
ilAKCN2vhOX5WSxJgSBhx/wCEG5ZxhzG+kpQDh+N3phIcOOIkVXQr5fKzmPh9G7E
gFFLx+SL2I+vQ6Y/kZVOOq5AADF01Yemz2Q030kAjmS2KHsz0efNo3qxqZQ3xv4g
quPCSdiZcU6yTHCpPtKJHcG0V8w5gVGY9LIwIfa8VYrEpEAr7HtYOcQaxanZREb6
UOmi4VyAtw5KtfI6N41R7M6vdoAPMbG7izz8HyijfgQVrSBa2k/ajt93HJsYyn5Z
TVff4wI/bStxgCsrpGHmkdKaUt5abwufO5iEnvT4vgFILYc2Aum+VNzWz8PYcu3R
pibOiGkCgYEA8Po1TSeVCL3GFhE4NPN+kAzmfCDlup08VlxeZaLbnxIsxZZXiRpE
nfFv0+teHbyqIMW51O5PQCkukYj18taG3CuPaI5EG7f5CBHTCYLmIcijYPG1AWCv
pKaixq+yoKBV7kTkAZrjc/A6cXabp3kR2I8kbMczjmyO/abDDTvRvlcCgYEAyueq
waZfFEEqkkvria7Pb9vL/GdNROfN+KHtSgTMRfQTIxAI4p/4dJ6V7QvCqFz1lTxI
OAFzvStC3sN1Cc91iUynvNhgfnkhR4BF1xss3maOx0SpRnTfh8ZlSpSzCe5HrmpZ
PuKRXw/K1MOTjGfTcGcOyUT2E1/TOdsz73Z9GWsCgYEAsWjz/uaKQVI92Jc60zkE
z5a/xVkO6JHEDyyXzqnUmMrvrVQfA/AT3lgW5WUd+DSi59fKFWqRbAIlp722nN61
kLh9LxO2LtYGAJvmy9TUCsFFDyVEBkyhY03O/wnxL3J7cRzE5C2dEQkmbcxNkkF/
EvlnbrJFUbC4oSO57C9DHqcCgYAZSdZWXE3tUFHl+eBZQJhJ8LLzukw/EkTxf/z7
BK4Q6eKYtB7nX9ivcDRvXs/b+/n/p4u4rzWllga3jNTBbEHR4uPk/XLJUH99Uddi
f9iPv2h8HWqbhWV2nptxOCc4TaJRcp+83rAPkQBOlDGHhkkr8Sw+mYGx2HeS9mb6
qWHeEQKBgQDAmzS17UOTCr0YwrGx+9+XBYd05FqDsPDpljzv5iPD/IjVFu2UUfXM
j8qnz4gumUCjOg4GLhzpT0TCXlcpvP7Ua/s/WF3MMk5OvDP1AZlp8XhQNUWe8vNb
ZMRxTzdweb9zdrQ1985ffmgzLBMI6m1QmqAFaotRgtasiFwVeJF2cA==
-----END RSA PRIVATE KEY-----`

func init() {

	s := Property{}

	s.Name = "autocert"
	s.Version = "0.1"
	s.Description = "This plugin is used to auto generate certificate for you . Really useful for testing different configuration flags against your targets. "

	s.Flags = func() {

		if *config.C.ForceHTTP == false {
			if len(*config.C.TLSCertificate) == 0 && len(*config.C.TLSKey) == 0 {

				log.Infof("Autocert plugin: Auto-generating %s domain TLS certificate", *config.C.ProxyDomain)

				CAcert := CA_CERT
				CAkey := CA_CERT_KEY

				catls, err := tls.X509KeyPair([]byte(CAcert), []byte(CAkey))
				if err != nil {
					panic(err)
				}
				ca, err := x509.ParseCertificate(catls.Certificate[0])
				if err != nil {
					panic(err)
				}

				crtSerial, err := rand.Int(rand.Reader, big.NewInt(9223372036854775807))
				if err != nil {
					panic(err)
				}

				template := &x509.Certificate{
					IsCA:                  false,
					BasicConstraintsValid: true,
					SubjectKeyId:          []byte{1, 2, 3},
					SerialNumber:          crtSerial,
					DNSNames:              []string{*config.C.ProxyDomain, "*." + *config.C.ProxyDomain},
					Subject: pkix.Name{
						Country:      []string{"Earth"},
						Organization: []string{"Mother Nature"},
						CommonName:   *config.C.ProxyDomain,
					},
					NotBefore: time.Now(),
					NotAfter:  time.Now().AddDate(1, 0, 0),
				}

				// generate private key
				privatekey, err := rsa.GenerateKey(rand.Reader, 2048)

				if err != nil {
					log.Errorf("Error generating key: %s", err)
				}
				var privateKey = &pem.Block{
					Type:  "PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(privatekey),
				}

				//dump
				buf := new(bytes.Buffer)
				pem.Encode(buf, privateKey)
				tlskeyStr := buf.String()
				config.C.TLSKey = &tlskeyStr
				log.Debugf("AutoCert plugin generated TlsKey:\n %s", *config.C.TLSKey)

				// generate self signed cert
				publickey := &privatekey.PublicKey

				// create a self-signed certificate. template = parent
				//var parent = template
				var parent = ca

				cert, err := x509.CreateCertificate(rand.Reader, template, parent, publickey, catls.PrivateKey)

				buf = new(bytes.Buffer)
				pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert})

				tlscertStr := buf.String()
				config.C.TLSCertificate = &tlscertStr
				log.Debugf("AutoCert plugin generated TlsCert:\n %s", *config.C.TLSCertificate)

				//the cert is auto-generated anyway
				*config.C.TLSPool = ""

				if err != nil {
					log.Errorf("Error creating certificate: %s", err)
				}

			}
		}

	}

	// Register all the function hooks
	s.Register()
}
