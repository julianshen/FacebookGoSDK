package facebook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

const GRAPH_END_POINT = "https://graph.facebook.com"

type FacebookAuthData struct {
	Algorithm    string
	Code         string
	Issued_at    uint32
	User_id      string
	Access_token string
	Expires      int
}

type FacebookContext struct {
	app_id string
	secret string

	auth *FacebookAuthData
}

func (f *FacebookContext) ParseSignedRequest(signed_request string) {
	encoded_data := strings.SplitN(signed_request, ".", 2)

	decoded, error := base64.URLEncoding.DecodeString(encoded_data[1] + "==")

	if error != nil {
		log.Println(error)
		return
	}

	var auth FacebookAuthData
	json.Unmarshal(decoded, &auth)
	f.auth = &auth

	hash := hmac.New(sha256.New, []byte(f.secret))
	io.WriteString(hash, encoded_data[1])
	sig := base64.URLEncoding.EncodeToString(hash.Sum(nil))

	if sig[:len(encoded_data[0])] != encoded_data[0] {
		//Invalid signature
		return
	}

	if f.auth == nil || f.auth.Code == "" {
		//No authentication data
		return
	}

	v := url.Values{}
	v.Set("client_id", f.app_id)
	v.Set("client_secret", f.secret)
	v.Set("redirect_uri", "")
	v.Set("code", f.auth.Code)

	request_url := GRAPH_END_POINT + "/oauth/access_token?" + v.Encode()
	resp, error := http.Get(request_url)

	if error != nil {
		//Error to connect to server
		log.Println(error)
		return
	}

	defer resp.Body.Close()
	body, error := ioutil.ReadAll(resp.Body)

	if error != nil {
		log.Println(error)
		return
	}

	values, _ := url.ParseQuery(string(body))
	f.auth.Access_token = values.Get("access_token")
	expires := values.Get("expires")

	f.auth.Expires, error = strconv.Atoi(expires)
	if error != nil {
		log.Println(error)
		log.Println(string(body))
		f.auth.Expires = -1
	}

	return
}

func (f *FacebookContext) AccessToken() string {
	return f.auth.Access_token
}

func (f *FacebookContext) IsLogin() bool {
	return (f.auth.Access_token == "")
}

func (f *FacebookContext) Get(path string, params *url.Values, cb func(string, error)) {
	if params == nil {
		params = &url.Values{}
	}

	if f.AccessToken() != "" {
		params.Set("access_token", f.AccessToken())
	}

	resp, error := http.Get(GRAPH_END_POINT + path + "?" + params.Encode())

	if error != nil {
		cb("", error)
		return
	}

	defer resp.Body.Close()

	body, error := ioutil.ReadAll(resp.Body)

	if error != nil {
		cb("", error)
		return
	}

	cb(string(body), nil)
	return
}

func (f *FacebookContext) Me() (user *Fuser, e error) {
	params := url.Values{}
	params.Set("fields", "id,name")
	f.Get("/me", &params, func(result string, e1 error) {
		if e1 != nil {
			e = e1
			return
		}

		user = new(Fuser)
		e = json.Unmarshal([]byte(result), user)
		return
	})
	return
}

func (f *FacebookContext) Fql(queries interface{}, cb func(string, error)) {
	var q string

	if reflect.TypeOf(queries).Kind() == reflect.Map {
		//multiquery
		qstr, _ := json.Marshal(queries)
		q = string(qstr)
	} else {
		q = reflect.ValueOf(queries).String()
	}

	params := url.Values{}
	params.Add("q", q)
	f.Get("/fql", &params, cb)

	return
}

func RealtimeHandler(verifyToken string, onDataUpdated func(w http.ResponseWriter, json string, e error)) http.Handler {
	var handler http.HandlerFunc
	handler = func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			hub_mode := r.FormValue("hub.mode")
			hub_verify_token := r.FormValue("hub.verify_token")
			hub_challenge := r.FormValue("hub.challenge")

			log.Println("hub_mode: " + hub_mode)
			log.Println("hub_challenge: " + hub_challenge)
			log.Println("hub_verify_token: " + hub_verify_token)

			if hub_mode != "" && verifyToken == hub_verify_token {
				fmt.Fprintln(w, hub_challenge)
			} else {
				fmt.Fprintln(w, "!!!error!!!")
			}
		} else if r.Method == "POST" {
			defer r.Body.Close()
			json, e := ioutil.ReadAll(r.Body)
			onDataUpdated(w, json, e)
		}
		return
	}

	return handler
}

func NewBasicContext(app_id string, secret string) *FacebookContext {
	context := new(FacebookContext)
	context.app_id = app_id
	context.secret = secret

	return context
}

func New(app_id string, secret string, r *http.Request) *FacebookContext {
	context := NewBasicContext(app_id, secret)

	cookie, error := r.Cookie("fbsr_" + app_id)

	if error == nil {
		context.ParseSignedRequest(cookie.Value)
	}

	return context
}
