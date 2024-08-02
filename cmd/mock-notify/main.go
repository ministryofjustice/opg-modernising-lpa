// Mock notify is a mock for GOV.UK's Notify service.
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/env"
)

func main() {
	port := env.Get("PORT", "8080")

	http.HandleFunc("/v2/notifications", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"notifications": []any{}})
	})

	http.HandleFunc("/v2/notifications/email", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		json.NewDecoder(r.Body).Decode(&v)
		log.Println("email:", v)
		json.NewEncoder(w).Encode(map[string]string{"id": "an-email-id"})
	})

	http.HandleFunc("/v2/notifications/sms", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		json.NewDecoder(r.Body).Decode(&v)
		log.Println("sms:", v)
		json.NewEncoder(w).Encode(map[string]string{"id": "an-sms-id"})
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
