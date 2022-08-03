package main

import (
	"fmt"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/fake"
	"golang.org/x/mod/sumdb/dirhash"
	"html/template"
	"log"
	"net/http"
)

func Hello() string {
	return "Hello, world!"
}

type PageData struct {
	WebDir      string
	Prefix      string
	PrefixAsset string
}

func main() {
	fmt.Println(fake.GoodBye())
	webDir := env.Get("WEB_DIR", "web")
	prefix := env.Get("PREFIX", "")

	files := []string{
		fmt.Sprintf("%s/template/home.gohtml", webDir),
		fmt.Sprintf("%s/template/layout/base.gohtml", webDir),
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		log.Fatal(err)
	}

	staticHash, err := dirhash.HashDir(webDir+"/static", webDir, dirhash.DefaultHash)

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			WebDir: webDir,
			Prefix: prefix,
		}

		err = ts.ExecuteTemplate(w, "base", data)

		if err != nil {
			log.Fatal(err)
		}
	})

	err = http.ListenAndServe(":5000", nil)

	if err != nil {
		log.Fatal(err)
	}
}
