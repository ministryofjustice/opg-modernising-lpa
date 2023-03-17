package main

import (
	"encoding/json"
	"net/http"

	"github.com/invopop/jsonschema"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func Schema() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reflector := new(jsonschema.Reflector)

		if err := reflector.AddGoComments("github.com/ministryofjustice/opg-modernising-lpa/internal/page", "./"); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		s := reflector.Reflect(&page.Lpa{})

		schema, _ := s.MarshalJSON()

		if r.FormValue("pretty") != "" {
			schema, _ = json.MarshalIndent(s, "", "    ")
		}

		w.Write(schema)
	}
}
