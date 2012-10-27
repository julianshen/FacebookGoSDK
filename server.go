package main

import (
	"./facebook"
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
	log.Fatal(http.ListenAndServe(":3000", nil))
}
