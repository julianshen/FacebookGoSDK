package facebook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const END_POINT = "https://graph.facebook.com"

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

	request_url := END_POINT + "/oauth/access_token?" + v.Encode()
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

	resp, error := http.Get(END_POINT + path + "?" + params.Encode())

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
