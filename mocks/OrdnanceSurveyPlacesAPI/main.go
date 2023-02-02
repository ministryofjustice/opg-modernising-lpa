package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ministryofjustice/opg-go-common/env"
)

func main() {
	port := env.Get("PORT", "8080")

	http.HandleFunc("/search/places/v1/postcode", func(w http.ResponseWriter, r *http.Request) {
		postcode := r.URL.Query().Get("postcode")
		log.Println("postcode searched:", postcode)

		if postcode == "INVALID" {
			invalidPostcodeJson, _ := os.ReadFile("testdata/invalid-postcode-error.json")
			w.Write(invalidPostcodeJson)
		} else {
			multipleAddressJson, _ := os.ReadFile("testdata/multiple-addresses.json")
			w.Write(multipleAddressJson)
		}

	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
