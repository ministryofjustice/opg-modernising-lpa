package main

import (
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/ministryofjustice/opg-go-common/env"
)

func Hello() string {
	return "Hello, world!"
}

type PageData struct {
	WebDir      string
	ServiceName string
}

func home(w http.ResponseWriter, r *http.Request) {
	webDir := env.Get("WEB_DIR", "web")

	data := PageData{
		WebDir:      webDir,
		ServiceName: "Modernising LPA",
	}

	files := []string{
		path.Join(webDir, "/template/home.gohtml"),
		path.Join(webDir, "/template/layout/base.gohtml"),
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		log.Fatal(err)
	}

	err = ts.ExecuteTemplate(w, "base", data)

	if err != nil {
		log.Fatal(err)
	}
}

func setToken(w http.ResponseWriter, r *http.Request) {
	log.Println("Made it to set_token")
	http.Redirect(w, r, "http://localhost:5050/home", http.StatusFound)
}

func main() {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./web/static/"))

	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/home", home)
	mux.HandleFunc("/set_token", setToken)

	err := http.ListenAndServe(":5000", mux)

	if err != nil {
		log.Fatal(err)
	}
}
