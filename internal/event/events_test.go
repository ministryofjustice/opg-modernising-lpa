package event

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func TestEventSchema(t *testing.T) {
	eventTypes := map[string]map[string]any{
		"application-updated": {
			"valid": ApplicationUpdated{
				UID:       "M-0000-0000-0000",
				Type:      "personal-welfare",
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
		},
		"reduced-fee-requested": {
			"upload": ReducedFeeRequested{
				UID:         "M-0000-0000-0000",
				RequestType: "NoFee",
				Evidence: []Evidence{
					{Path: "M-0000-0000-0000/evidence/a-uid", Filename: "a-file.pdf"},
					{Path: "M-0000-0000-0000/evidence/b-uid", Filename: "b-file.pdf"},
				},
				EvidenceDelivery: "upload",
			},
			"post": ReducedFeeRequested{
				UID:              "M-0000-0000-0000",
				RequestType:      "NoFee",
				EvidenceDelivery: "post",
			},
		},
		"notification-sent": {
			"valid": NotificationSent{
				UID:            "M-0000-0000-0000",
				NotificationID: random.UuidString(),
			},
		},
		"paper-form-requested": {
			"certificate provider": PaperFormRequested{
				UID:       "M-0000-0000-0000",
				ActorUID:  actoruid.New(),
				ActorType: "certificateProvider",
			},
			"attorney": PaperFormRequested{
				UID:       "M-0000-0000-0000",
				ActorUID:  actoruid.New(),
				ActorType: "attorney",
			},
			"replacement attorney": PaperFormRequested{
				UID:       "M-0000-0000-0000",
				ActorUID:  actoruid.New(),
				ActorType: "replacementAttorney",
			},
			"trust corporation": PaperFormRequested{
				UID:       "M-0000-0000-0000",
				ActorUID:  actoruid.New(),
				ActorType: "trustCorporation",
			},
			"replacement trust corporation": PaperFormRequested{
				UID:       "M-0000-0000-0000",
				ActorUID:  actoruid.New(),
				ActorType: "replacementTrustCorporation",
			},
		},
	}

	dir, _ := os.Getwd()

	for eventType, tcs := range eventTypes {
		for name, event := range tcs {
			t.Run(eventType+"/"+name, func(t *testing.T) {
				schemaLoader := gojsonschema.NewReferenceLoader("file:///" + dir + "/testdata/" + eventType + ".json")
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
}
