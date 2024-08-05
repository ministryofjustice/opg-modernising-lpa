// Mock OS API is a mock for Ordnance Survey's Places API
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
		var postcodeJson []byte

		switch postcode {
		case "INVALID":
			postcodeJson, _ = os.ReadFile("data/invalid-postcode-error.json")
		case "NE234EE":
			postcodeJson, _ = os.ReadFile("data/no-addresses-found.json")
		default:
			postcodeJson, _ = os.ReadFile("data/multiple-addresses.json")
		}

		w.Write(postcodeJson)
		// to aid debugging e2e test failures
		log.Println("OS mock response:", string(postcodeJson))
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
