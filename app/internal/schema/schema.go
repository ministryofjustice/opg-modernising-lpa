// go:build exclude

package main

import (
	"encoding/json"
	"os"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"

	"github.com/invopop/jsonschema"
)

func main() {
	r := new(jsonschema.Reflector)
	r.AllowAdditionalProperties = true
	if err := r.AddGoComments("github.com/ministryofjustice/opg-modernising-lpa/internal/page", "./app/internal/page"); err != nil {
		panic(err)
	}
	s := r.Reflect(&page.Lpa{})
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile("schema.json", data, 0o644)
}
