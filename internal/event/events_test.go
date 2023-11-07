package event

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func TestEventSchema(t *testing.T) {
	testcases := map[string]any{
		"application-updated": ApplicationUpdated{
			UID:       "M-0000-0000-0000",
			Type:      "hw",
			CreatedAt: time.Now(),
			Donor: ApplicationUpdatedDonor{
				FirstNames:  "syz",
				LastName:    "syz",
				DateOfBirth: date.New("2000", "01", "01"),
				Address: place.Address{
					Line1:      "line1",
					Line2:      "line2",
					Line3:      "line3",
					TownOrCity: "townOrCity",
					Postcode:   "F1 1FF",
					Country:    "GB",
				},
			},
		},
		"reduced-fee-requested": ReducedFeeRequested{
			UID:              "M-0000-0000-0000",
			RequestType:      "NoFee",
			Evidence:         []string{"key"},
			EvidenceDelivery: "upload",
		},
	}

	dir, _ := os.Getwd()

	for name, event := range testcases {
		t.Run(name, func(t *testing.T) {
			schemaLoader := gojsonschema.NewReferenceLoader("file:///" + dir + "/testdata/" + name + ".json")
			documentLoader := gojsonschema.NewGoLoader(event)

			result, err := gojsonschema.Validate(schemaLoader, documentLoader)
			assert.Nil(t, err)

			if !assert.True(t, result.Valid()) {
				lines := []string{"The document is not valid:"}
				for _, desc := range result.Errors() {
					lines = append(lines, "- "+desc.String())
				}

				t.Log(strings.Join(lines, "\n"))
			}
		})
	}
}
