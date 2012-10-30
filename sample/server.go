package main

import (
	facebook "github.com/julianshen/FacebookGoSDK"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	app_id := os.ExpandEnv("$FACEBOOK_APPID")
	secret := os.ExpandEnv("$FACEBOOK_SECRET")
	http.Handle("/", http.FileServer(http.Dir("./html")))
	http.HandleFunc("/facebook.html", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("./template/facebook.tmpl")
		t.Execute(w, app_id)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		f := facebook.New(app_id, secret, r)
		f.Get("/me", nil, func(result string, e error) {
			if e != nil {
				log.Println(e)
				fmt.Fprintln(w, "error")
			} else {
				fmt.Fprintln(w, result)
			}
			return
		})
	})

	http.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		f := facebook.New(app_id, secret, r)

		me, e := f.Me()
		if e != nil {
			log.Println(e)
			fmt.Fprintln(w, "error")
		} else {
			//me is an instance of "Fuser"
			m, _ := json.Marshal(me)
			fmt.Fprintln(w, string(m))
		}
	})

	http.HandleFunc("/friends", func(w http.ResponseWriter, r *http.Request) {
		f := facebook.New(app_id, secret, r)

		f.Fql("SELECT uid2 FROM friend WHERE uid1=me()", func(result string, e error) {
			if e != nil {
				log.Println(e)
				fmt.Fprintf(w, "error")
			} else {
				fmt.Fprintf(w, result)
			}
		})
	})

	http.HandleFunc("/allfriends", func(w http.ResponseWriter, r *http.Request) {
		f := facebook.New(app_id, secret, r)
		queries := make(map[string]string)
		queries["all friends"] = "SELECT uid2 FROM friend WHERE uid1=me()"
		queries["my name"] = "SELECT name FROM user WHERE uid=me()"

		f.Fql(queries, func(result string, e error) {
			if e != nil {
				log.Println(e)
				fmt.Fprintf(w, "error")
			} else {
				fmt.Fprintf(w, result)
			}
		})
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
