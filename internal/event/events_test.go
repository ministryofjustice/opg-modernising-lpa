package event

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func pt[T any](v T) *T {
	return &v
}

var eventTests = map[string]map[string]any{
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
	"application-deleted": {
		"valid": ApplicationDeleted{
			UID: "M-0000-0000-0000",
		},
	},
	"attorney-started": {
		"valid": AttorneyStarted{
			LpaUID:   "M-0000-0000-0000",
			ActorUID: actoruid.New(),
		},
	},
	"certificate-provider-started": {
		"valid": CertificateProviderStarted{
			UID: "M-0000-0000-0000",
		},
	},
	"payment-received": {
		"valid": PaymentReceived{
			UID:       "M-0000-0000-0000",
			PaymentID: "123",
			Amount:    5,
		},
	},
	"uid-requested": {
		"valid": UidRequested{
			LpaID:          "ffec5e7a-9cea-4e46-a99b-6c086fbf1a27",
			DonorSessionID: "blah",
			OrganisationID: "blahhh",
			Type:           "pfa",
			Donor: uid.DonorDetails{
				Name:     "hey",
				Dob:      date.Today(),
				Postcode: "W1 1AA",
			},
		},
	},
	"identity-check-mismatched": {
		"valid": IdentityCheckMismatched{
			LpaUID:   "M-1111-1111-1111",
			ActorUID: actoruid.New(),
			Provided: IdentityCheckMismatchedDetails{
				FirstNames:  "a",
				LastName:    "b",
				DateOfBirth: date.Today(),
			},
			Verified: IdentityCheckMismatchedDetails{
				FirstNames:  "a",
				LastName:    "b",
				DateOfBirth: date.Today(),
			},
		},
	},
	"correspondent-updated": {
		"remove": CorrespondentUpdated{UID: "M-1111-1111-1111"},
		"without address": CorrespondentUpdated{
			UID:        "M-1111-1111-1111",
			ActorUID:   pt(actoruid.New()),
			FirstNames: "John",
			LastName:   "Smith",
			Email:      "john@example.com",
			Phone:      "07777",
		},
		"with address": CorrespondentUpdated{
			UID:        "M-1111-1111-1111",
			ActorUID:   pt(actoruid.New()),
			FirstNames: "John",
			LastName:   "Smith",
			Email:      "john@example.com",
			Phone:      "07777",
			Address: &place.Address{
				Line1:      "line-1",
				TownOrCity: "town",
				Postcode:   "F1 1FF",
				Country:    "GB",
			},
		},
	},
	"lpa-access-granted": {
		"valid": LpaAccessGranted{
			UID:     "M-1111-2222-3333",
			LpaType: "personal-welfare",
			Actors: []LpaAccessGrantedActor{{
				ActorUID:  "9ac5cb7c-fc75-40c7-8e53-059f36dbbe3d",
				SubjectID: "urn:fdc:gov.uk:2022:XXXX-XXXXXX",
			}},
		},
	},
	"letter-requested": {
		"valid": LetterRequested{
			UID:        "M-1111-2222-3333",
			LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_ACTED",
			ActorType:  actor.TypeDonor,
			ActorUID:   actoruid.New(),
		},
	},
}

func TestEventSchema(t *testing.T) {
	dir, _ := os.Getwd()

	for eventType, tcs := range eventTests {
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

func TestEventsAllTested(t *testing.T) {
	err := filepath.WalkDir("testdata", func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		name := strings.TrimSuffix(d.Name(), ".json")
		if _, ok := eventTests[name]; !ok {
			t.Fail()
			t.Log("missing testcase for event:", name)
		}

		return nil
	})
	assert.Nil(t, err)
}
