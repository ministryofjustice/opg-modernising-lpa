package lpastore

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestClientSendLpa(t *testing.T) {
	donorUID, _ := actoruid.Parse("6178e739-76b0-426e-b9c4-e45be426fbdf")

	trustCorporationUID := actoruid.New()
	attorneyUID := actoruid.New()
	attorney2UID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	replacementAttorney2UID := actoruid.New()
	certificateProviderUID := actoruid.New()
	personToNotifyUID := actoruid.New()

	testcases := map[string]struct {
		donor *actor.DonorProvidedDetails
		json  string
	}{
		"minimal": {
			donor: &actor.DonorProvidedDetails{
				LpaUID: "M-0000-1111-2222",
				Type:   actor.LpaTypePropertyAndAffairs,
				Donor: actor.Donor{
					UID:         donorUID,
					FirstNames:  "John Johnson",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "1", "2"),
					Email:       "john@example.com",
					Address: place.Address{
						Line1:      "line-1",
						TownOrCity: "town",
						Country:    "GB",
					},
					OtherNames: "JJ",
				},
				Attorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{{
						UID:         attorneyUID,
						FirstNames:  "Adam",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "1", "2"),
						Email:       "adam@example.com",
						Address: place.Address{
							Line1:      "a-line-1",
							TownOrCity: "a-town",
							Country:    "GB",
						},
					}},
				},
				ReplacementAttorneys: actor.Attorneys{},
				WhenCanTheLpaBeUsed:  actor.CanBeUsedWhenCapacityLost,
				CertificateProvider: actor.CertificateProvider{
					UID:        certificateProviderUID,
					FirstNames: "Carol",
					LastName:   "Cert",
					Address: place.Address{
						Line1:      "c-line-1",
						TownOrCity: "c-town",
						Country:    "GB",
					},
					CarryOutBy: actor.Paper,
				},
				SignedAt: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
			},
			json: `{
"lpaType":"property-and-affairs",
"donor":{"uid":"` + donorUID.String() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"","line3":"","town":"town","postcode":"","country":"GB"},"otherNamesKnownBy":"JJ"},
"attorneys":[{"uid":"` + attorneyUID.String() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"","line3":"","town":"a-town","postcode":"","country":"GB"},"status":"active"}],
"certificateProvider":{"uid":"` + certificateProviderUID.String() + `","firstNames":"Carol","lastName":"Cert","address":{"line1":"c-line-1","line2":"","line3":"","town":"c-town","postcode":"","country":"GB"},"channel":"paper"},
"restrictionsAndConditions":"",
"whenTheLpaCanBeUsed":"when-capacity-lost",
"signedAt":"2000-01-02T03:04:05.000000006Z"
}`,
		},
		"everything": {
			donor: &actor.DonorProvidedDetails{
				LpaUID: "M-0000-1111-2222",
				Type:   actor.LpaTypePersonalWelfare,
				Donor: actor.Donor{
					UID:         donorUID,
					FirstNames:  "John Johnson",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "1", "2"),
					Email:       "john@example.com",
					Address: place.Address{
						Line1:      "line-1",
						Line2:      "line-2",
						Line3:      "line-3",
						TownOrCity: "town",
						Postcode:   "F1 1FF",
						Country:    "GB",
					},
					OtherNames: "JJ",
				},
				Attorneys: actor.Attorneys{
					TrustCorporation: actor.TrustCorporation{
						UID:           trustCorporationUID,
						Name:          "Trusty",
						CompanyNumber: "55555",
						Email:         "trusty@example.com",
						Address: place.Address{
							Line1:      "a-line-1",
							Line2:      "a-line-2",
							Line3:      "a-line-3",
							TownOrCity: "a-town",
							Postcode:   "A1 1FF",
							Country:    "GB",
						},
					},
					Attorneys: []actor.Attorney{{
						UID:         attorneyUID,
						FirstNames:  "Adam",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "1", "2"),
						Email:       "adam@example.com",
						Address: place.Address{
							Line1:      "a-line-1",
							Line2:      "a-line-2",
							Line3:      "a-line-3",
							TownOrCity: "a-town",
							Postcode:   "A1 1FF",
							Country:    "GB",
						},
					}, {
						UID:         attorney2UID,
						FirstNames:  "Alice",
						LastName:    "Attorney",
						DateOfBirth: date.New("1998", "1", "2"),
						Email:       "alice@example.com",
						Address: place.Address{
							Line1:      "aa-line-1",
							Line2:      "aa-line-2",
							Line3:      "aa-line-3",
							TownOrCity: "aa-town",
							Postcode:   "A1 1AF",
							Country:    "GB",
						},
					}},
				},
				AttorneyDecisions: actor.AttorneyDecisions{
					How: actor.Jointly,
				},
				ReplacementAttorneys: actor.Attorneys{
					TrustCorporation: actor.TrustCorporation{
						UID:           replacementTrustCorporationUID,
						Name:          "UnTrusty",
						CompanyNumber: "65555",
						Email:         "untrusty@example.com",
						Address: place.Address{
							Line1:      "a-line-1",
							Line2:      "a-line-2",
							Line3:      "a-line-3",
							TownOrCity: "a-town",
							Postcode:   "A1 1FF",
							Country:    "GB",
						},
					},
					Attorneys: []actor.Attorney{{
						UID:         replacementAttorneyUID,
						FirstNames:  "Richard",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "11", "12"),
						Email:       "richard@example.com",
						Address: place.Address{
							Line1:      "r-line-1",
							Line2:      "r-line-2",
							Line3:      "r-line-3",
							TownOrCity: "r-town",
							Postcode:   "R1 1FF",
							Country:    "GB",
						},
					}, {
						UID:         replacementAttorney2UID,
						FirstNames:  "Rachel",
						LastName:    "Attorney",
						DateOfBirth: date.New("1998", "11", "12"),
						Email:       "rachel@example.com",
						Address: place.Address{
							Line1:      "rr-line-1",
							Line2:      "rr-line-2",
							Line3:      "rr-line-3",
							TownOrCity: "rr-town",
							Postcode:   "R1 1RF",
							Country:    "GB",
						},
					}},
				},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{
					How:     actor.JointlyForSomeSeverallyForOthers,
					Details: "umm",
				},
				HowShouldReplacementAttorneysStepIn: actor.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
				LifeSustainingTreatmentOption:       actor.LifeSustainingTreatmentOptionA,
				Restrictions:                        "do not do this",
				CertificateProvider: actor.CertificateProvider{
					UID:        certificateProviderUID,
					FirstNames: "Carol",
					LastName:   "Cert",
					Email:      "carol@example.com",
					Mobile:     "0700009000",
					Address: place.Address{
						Line1:      "c-line-1",
						Line2:      "c-line-2",
						Line3:      "c-line-3",
						TownOrCity: "c-town",
						Postcode:   "C1 1FF",
						Country:    "GB",
					},
					CarryOutBy: actor.Online,
				},
				PeopleToNotify: actor.PeopleToNotify{{
					UID:        personToNotifyUID,
					FirstNames: "Peter",
					LastName:   "Notify",
					Address: place.Address{
						Line1:      "p-line-1",
						Line2:      "p-line-2",
						Line3:      "p-line-3",
						TownOrCity: "p-town",
						Postcode:   "P1 1FF",
						Country:    "GB",
					},
				}},
				SignedAt:                                 time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				CertificateProviderNotRelatedConfirmedAt: time.Date(2001, time.February, 3, 4, 5, 6, 7, time.UTC),
			},
			json: `{
"lpaType":"personal-welfare",
"donor":{"uid":"` + donorUID.String() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"line-2","line3":"line-3","town":"town","postcode":"F1 1FF","country":"GB"},"otherNamesKnownBy":"JJ"},
"attorneys":[
{"uid":"` + attorneyUID.String() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active"},
{"uid":"` + attorney2UID.String() + `","firstNames":"Alice","lastName":"Attorney","dateOfBirth":"1998-01-02","email":"alice@example.com","address":{"line1":"aa-line-1","line2":"aa-line-2","line3":"aa-line-3","town":"aa-town","postcode":"A1 1AF","country":"GB"},"status":"active"},
{"uid":"` + replacementAttorneyUID.String() + `","firstNames":"Richard","lastName":"Attorney","dateOfBirth":"1999-11-12","email":"richard@example.com","address":{"line1":"r-line-1","line2":"r-line-2","line3":"r-line-3","town":"r-town","postcode":"R1 1FF","country":"GB"},"status":"replacement"},
{"uid":"` + replacementAttorney2UID.String() + `","firstNames":"Rachel","lastName":"Attorney","dateOfBirth":"1998-11-12","email":"rachel@example.com","address":{"line1":"rr-line-1","line2":"rr-line-2","line3":"rr-line-3","town":"rr-town","postcode":"R1 1RF","country":"GB"},"status":"replacement"}
],
"trustCorporations":[
{"uid":"` + trustCorporationUID.String() + `","name":"Trusty","companyNumber":"55555","email":"trusty@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active"},
{"uid":"` + replacementTrustCorporationUID.String() + `","name":"UnTrusty","companyNumber":"65555","email":"untrusty@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"replacement"}
],
"certificateProvider":{"uid":"` + certificateProviderUID.String() + `","firstNames":"Carol","lastName":"Cert","email":"carol@example.com","phone":"0700009000","address":{"line1":"c-line-1","line2":"c-line-2","line3":"c-line-3","town":"c-town","postcode":"C1 1FF","country":"GB"},"channel":"online"},
"peopleToNotify":[{"uid":"` + personToNotifyUID.String() + `","firstNames":"Peter","lastName":"Notify","address":{"line1":"p-line-1","line2":"p-line-2","line3":"p-line-3","town":"p-town","postcode":"P1 1FF","country":"GB"}}],
"howAttorneysMakeDecisions":"jointly",
"howReplacementAttorneysMakeDecisions":"jointly-for-some-severally-for-others",
"howReplacementAttorneysMakeDecisionsDetails":"umm",
"restrictionsAndConditions":"do not do this",
"lifeSustainingTreatmentOption":"option-a",
"signedAt":"2000-01-02T03:04:05.000000006Z",
"certificateProviderNotRelatedConfirmedAt":"2001-02-03T04:05:06.000000007Z"}
`,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(ctx, secrets.LpaStoreJwtSecretKey).
				Return("secret", nil)

			var body []byte
			doer := newMockDoer(t)
			doer.EXPECT().
				Do(mock.MatchedBy(func(req *http.Request) bool {
					if body == nil {
						body, _ = io.ReadAll(req.Body)
					}

					return assert.Equal(t, ctx, req.Context()) &&
						assert.Equal(t, http.MethodPut, req.Method) &&
						assert.Equal(t, "http://base/lpas/M-0000-1111-2222", req.URL.String()) &&
						assert.Equal(t, "application/json", req.Header.Get("Content-Type")) &&
						assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjYxNzhlNzM5LTc2YjAtNDI2ZS1iOWM0LWU0NWJlNDI2ZmJkZiIsImlhdCI6OTQ2NzgyMjQ1fQ.6dzpmF8FHNeVpAjzivyY9Cl9sD2amq4iCmgBp1vSBaY", req.Header.Get("X-Jwt-Authorization")) &&
						assert.JSONEq(t, tc.json, string(body))
				})).
				Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

			client := New("http://base", secretsClient, doer)
			client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
			err := client.SendLpa(ctx, tc.donor)

			assert.Nil(t, err)
		})
	}
}

func TestClientSendLpaWhenNewRequestError(t *testing.T) {
	client := New("http://base", nil, nil)
	err := client.SendLpa(nil, &actor.DonorProvidedDetails{})

	assert.NotNil(t, err)
}

func TestClientSendLpaWhenSecretsClientError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("", expectedError)

	client := New("http://base", secretsClient, nil)
	err := client.SendLpa(ctx, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestClientSendLpaWhenDoerError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("secret", nil)

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	client := New("http://base", secretsClient, doer)
	err := client.SendLpa(ctx, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestClientSendLpaWhenStatusCodeIsNotCreated(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("secret", nil)

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("hey"))}, nil)

	client := New("http://base", secretsClient, doer)
	err := client.SendLpa(ctx, &actor.DonorProvidedDetails{})

	assert.Equal(t, responseError{name: "expected 201 response but got 400", body: "hey"}, err)
}

func TestClientLpa(t *testing.T) {
	donorUID, _ := actoruid.Parse("6178e739-76b0-426e-b9c4-e45be426fbdf")

	trustCorporationUID := actoruid.New()
	attorneyUID := actoruid.New()
	attorney2UID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	replacementAttorney2UID := actoruid.New()
	certificateProviderUID := actoruid.New()
	personToNotifyUID := actoruid.New()

	testcases := map[string]struct {
		donor *Lpa
		json  string
	}{
		"minimal": {
			donor: &Lpa{
				LpaUID: "M-0000-1111-2222",
				Type:   actor.LpaTypePropertyAndAffairs,
				Donor: actor.Donor{
					UID:         donorUID,
					FirstNames:  "John Johnson",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "01", "02"),
					Email:       "john@example.com",
					Address: place.Address{
						Line1:      "line-1",
						TownOrCity: "town",
						Country:    "GB",
					},
					OtherNames: "JJ",
				},
				Attorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{{
						UID:         attorneyUID,
						FirstNames:  "Adam",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "01", "02"),
						Email:       "adam@example.com",
						Address: place.Address{
							Line1:      "a-line-1",
							TownOrCity: "a-town",
							Country:    "GB",
						},
					}},
				},
				ReplacementAttorneys: actor.Attorneys{},
				WhenCanTheLpaBeUsed:  actor.CanBeUsedWhenCapacityLost,
				CertificateProvider: actor.CertificateProvider{
					UID:        certificateProviderUID,
					FirstNames: "Carol",
					LastName:   "Cert",
					Address: place.Address{
						Line1:      "c-line-1",
						TownOrCity: "c-town",
						Country:    "GB",
					},
					CarryOutBy: actor.Paper,
				},
				SignedAt: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
			},
			json: `{
"uid":"M-0000-1111-2222",
"lpaType":"property-and-affairs",
"donor":{"uid":"` + donorUID.String() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"","line3":"","town":"town","postcode":"","country":"GB"},"otherNamesKnownBy":"JJ"},
"attorneys":[{"uid":"` + attorneyUID.String() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"","line3":"","town":"a-town","postcode":"","country":"GB"},"status":"active"}],
"certificateProvider":{"uid":"` + certificateProviderUID.String() + `","firstNames":"Carol","lastName":"Cert","address":{"line1":"c-line-1","line2":"","line3":"","town":"c-town","postcode":"","country":"GB"},"channel":"paper"},
"restrictionsAndConditions":"",
"whenTheLpaCanBeUsed":"when-capacity-lost",
"signedAt":"2000-01-02T03:04:05.000000006Z"
}`,
		},
		"everything": {
			donor: &Lpa{
				LpaUID: "M-0000-1111-2222",
				Type:   actor.LpaTypePersonalWelfare,
				Donor: actor.Donor{
					UID:         donorUID,
					FirstNames:  "John Johnson",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "01", "02"),
					Email:       "john@example.com",
					Address: place.Address{
						Line1:      "line-1",
						Line2:      "line-2",
						Line3:      "line-3",
						TownOrCity: "town",
						Postcode:   "F1 1FF",
						Country:    "GB",
					},
					OtherNames: "JJ",
				},
				Attorneys: actor.Attorneys{
					TrustCorporation: actor.TrustCorporation{
						UID:           trustCorporationUID,
						Name:          "Trusty",
						CompanyNumber: "55555",
						Email:         "trusty@example.com",
						Address: place.Address{
							Line1:      "a-line-1",
							Line2:      "a-line-2",
							Line3:      "a-line-3",
							TownOrCity: "a-town",
							Postcode:   "A1 1FF",
							Country:    "GB",
						},
					},
					Attorneys: []actor.Attorney{{
						UID:         attorneyUID,
						FirstNames:  "Adam",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "01", "02"),
						Email:       "adam@example.com",
						Address: place.Address{
							Line1:      "a-line-1",
							Line2:      "a-line-2",
							Line3:      "a-line-3",
							TownOrCity: "a-town",
							Postcode:   "A1 1FF",
							Country:    "GB",
						},
					}, {
						UID:         attorney2UID,
						FirstNames:  "Alice",
						LastName:    "Attorney",
						DateOfBirth: date.New("1998", "01", "02"),
						Email:       "alice@example.com",
						Address: place.Address{
							Line1:      "aa-line-1",
							Line2:      "aa-line-2",
							Line3:      "aa-line-3",
							TownOrCity: "aa-town",
							Postcode:   "A1 1AF",
							Country:    "GB",
						},
					}},
				},
				AttorneyDecisions: actor.AttorneyDecisions{
					How: actor.Jointly,
				},
				ReplacementAttorneys: actor.Attorneys{
					TrustCorporation: actor.TrustCorporation{
						UID:           replacementTrustCorporationUID,
						Name:          "UnTrusty",
						CompanyNumber: "65555",
						Email:         "untrusty@example.com",
						Address: place.Address{
							Line1:      "a-line-1",
							Line2:      "a-line-2",
							Line3:      "a-line-3",
							TownOrCity: "a-town",
							Postcode:   "A1 1FF",
							Country:    "GB",
						},
					},
					Attorneys: []actor.Attorney{{
						UID:         replacementAttorneyUID,
						FirstNames:  "Richard",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "11", "12"),
						Email:       "richard@example.com",
						Address: place.Address{
							Line1:      "r-line-1",
							Line2:      "r-line-2",
							Line3:      "r-line-3",
							TownOrCity: "r-town",
							Postcode:   "R1 1FF",
							Country:    "GB",
						},
					}, {
						UID:         replacementAttorney2UID,
						FirstNames:  "Rachel",
						LastName:    "Attorney",
						DateOfBirth: date.New("1998", "11", "12"),
						Email:       "rachel@example.com",
						Address: place.Address{
							Line1:      "rr-line-1",
							Line2:      "rr-line-2",
							Line3:      "rr-line-3",
							TownOrCity: "rr-town",
							Postcode:   "R1 1RF",
							Country:    "GB",
						},
					}},
				},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{
					How:     actor.JointlyForSomeSeverallyForOthers,
					Details: "umm",
				},
				HowShouldReplacementAttorneysStepIn: actor.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
				LifeSustainingTreatmentOption:       actor.LifeSustainingTreatmentOptionA,
				Restrictions:                        "do not do this",
				CertificateProvider: actor.CertificateProvider{
					UID:        certificateProviderUID,
					FirstNames: "Carol",
					LastName:   "Cert",
					Email:      "carol@example.com",
					Mobile:     "0700009000",
					Address: place.Address{
						Line1:      "c-line-1",
						Line2:      "c-line-2",
						Line3:      "c-line-3",
						TownOrCity: "c-town",
						Postcode:   "C1 1FF",
						Country:    "GB",
					},
					CarryOutBy: actor.Online,
				},
				PeopleToNotify: actor.PeopleToNotify{{
					UID:        personToNotifyUID,
					FirstNames: "Peter",
					LastName:   "Notify",
					Address: place.Address{
						Line1:      "p-line-1",
						Line2:      "p-line-2",
						Line3:      "p-line-3",
						TownOrCity: "p-town",
						Postcode:   "P1 1FF",
						Country:    "GB",
					},
				}},
				SignedAt:                                 time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				CertificateProviderNotRelatedConfirmedAt: time.Date(2001, time.February, 3, 4, 5, 6, 7, time.UTC),
			},
			json: `{
"uid":"M-0000-1111-2222",
"lpaType":"personal-welfare",
"donor":{"uid":"` + donorUID.String() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"line-2","line3":"line-3","town":"town","postcode":"F1 1FF","country":"GB"},"otherNamesKnownBy":"JJ"},
"attorneys":[
{"uid":"` + attorneyUID.String() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active"},
{"uid":"` + attorney2UID.String() + `","firstNames":"Alice","lastName":"Attorney","dateOfBirth":"1998-01-02","email":"alice@example.com","address":{"line1":"aa-line-1","line2":"aa-line-2","line3":"aa-line-3","town":"aa-town","postcode":"A1 1AF","country":"GB"},"status":"active"},
{"uid":"` + replacementAttorneyUID.String() + `","firstNames":"Richard","lastName":"Attorney","dateOfBirth":"1999-11-12","email":"richard@example.com","address":{"line1":"r-line-1","line2":"r-line-2","line3":"r-line-3","town":"r-town","postcode":"R1 1FF","country":"GB"},"status":"replacement"},
{"uid":"` + replacementAttorney2UID.String() + `","firstNames":"Rachel","lastName":"Attorney","dateOfBirth":"1998-11-12","email":"rachel@example.com","address":{"line1":"rr-line-1","line2":"rr-line-2","line3":"rr-line-3","town":"rr-town","postcode":"R1 1RF","country":"GB"},"status":"replacement"}
],
"trustCorporations":[
{"uid":"` + trustCorporationUID.String() + `","name":"Trusty","companyNumber":"55555","email":"trusty@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active"},
{"uid":"` + replacementTrustCorporationUID.String() + `","name":"UnTrusty","companyNumber":"65555","email":"untrusty@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"replacement"}
],
"certificateProvider":{"uid":"` + certificateProviderUID.String() + `","firstNames":"Carol","lastName":"Cert","email":"carol@example.com","phone":"0700009000","address":{"line1":"c-line-1","line2":"c-line-2","line3":"c-line-3","town":"c-town","postcode":"C1 1FF","country":"GB"},"channel":"online"},
"peopleToNotify":[{"uid":"` + personToNotifyUID.String() + `","firstNames":"Peter","lastName":"Notify","address":{"line1":"p-line-1","line2":"p-line-2","line3":"p-line-3","town":"p-town","postcode":"P1 1FF","country":"GB"}}],
"howAttorneysMakeDecisions":"jointly",
"howReplacementAttorneysMakeDecisions":"jointly-for-some-severally-for-others",
"howReplacementAttorneysMakeDecisionsDetails":"umm",
"howReplacementAttorneysStepIn":"all-can-no-longer-act",
"restrictionsAndConditions":"do not do this",
"lifeSustainingTreatmentOption":"option-a",
"signedAt":"2000-01-02T03:04:05.000000006Z",
"certificateProviderNotRelatedConfirmedAt":"2001-02-03T04:05:06.000000007Z"}
`,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(ctx, secrets.LpaStoreJwtSecretKey).
				Return("secret", nil)

			doer := newMockDoer(t)
			doer.EXPECT().
				Do(mock.MatchedBy(func(req *http.Request) bool {
					return assert.Equal(t, ctx, req.Context()) &&
						assert.Equal(t, http.MethodGet, req.Method) &&
						assert.Equal(t, "http://base/lpas/M-0000-1111-2222", req.URL.String()) &&
						assert.Equal(t, "application/json", req.Header.Get("Content-Type")) &&
						assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjAwMDAwMDAwLTAwMDAtNDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsImlhdCI6OTQ2NzgyMjQ1fQ.V7MxjZw7-K8ehujYn4e0gef7s23r2UDlTbyzQtpTKvo", req.Header.Get("X-Jwt-Authorization"))
				})).
				Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tc.json))}, nil)

			client := New("http://base", secretsClient, doer)
			client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
			donor, err := client.Lpa(ctx, "M-0000-1111-2222")

			assert.Nil(t, err)
			assert.Equal(t, tc.donor, donor)
		})
	}
}

func TestClientLpaWhenNewRequestError(t *testing.T) {
	client := New("http://base", nil, nil)
	_, err := client.Lpa(nil, "M-0000-1111-2222")

	assert.NotNil(t, err)
}

func TestClientLpaWhenSecretsClientError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("", expectedError)

	client := New("http://base", secretsClient, nil)
	_, err := client.Lpa(ctx, "M-0000-1111-2222")

	assert.Equal(t, expectedError, err)
}

func TestClientLpaWhenDoerError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("secret", nil)

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	client := New("http://base", secretsClient, doer)
	_, err := client.Lpa(ctx, "M-0000-1111-2222")

	assert.Equal(t, expectedError, err)
}

func TestClientLpaWhenStatusCodeIsNotCreated(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("secret", nil)

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("hey"))}, nil)

	client := New("http://base", secretsClient, doer)
	_, err := client.Lpa(ctx, "M-0000-1111-2222")

	assert.Equal(t, responseError{name: "expected 200 response but got 400", body: "hey"}, err)
}

func TestAllAttorneysSigned(t *testing.T) {
	lpaSignedAt := time.Now()
	otherLpaSignedAt := lpaSignedAt.Add(time.Minute)
	attorneySigned := lpaSignedAt.Add(time.Second)

	uid1 := actoruid.New()
	uid2 := actoruid.New()
	uid3 := actoruid.New()
	uid4 := actoruid.New()
	uid5 := actoruid.New()

	testcases := map[string]struct {
		lpa       *Lpa
		attorneys []*actor.AttorneyProvidedDetails
		expected  bool
	}{
		"no attorneys": {
			expected: false,
		},
		"need attorney to sign": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}, {UID: uid2}}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid3}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid4, LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
				{UID: uid3, IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"need replacement attorney to sign": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid3}, {UID: uid5}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid3, IsReplacement: true},
				{UID: uid5, IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"all attorneys signed": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}, {UID: uid2}}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid3}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid3, IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: true,
		},
		"more attorneys signed": {
			lpa: &Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}, {UID: uid2}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid4, LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
			},
			expected: true,
		},
		"waiting for attorney to re-sign": {
			lpa: &Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}, {UID: uid2}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"trust corporations not signed": {
			lpa: &Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
			},
			expected: false,
		},
		"replacement trust corporations not signed": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
				ReplacementAttorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "r"}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]actor.TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.Yes,
					AuthorisedSignatories:    [2]actor.TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
			},
			expected: false,
		},
		"trust corporations signed": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
				ReplacementAttorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "r"}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]actor.TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
				{
					IsTrustCorporation:       true,
					IsReplacement:            true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]actor.TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
			},
			expected: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.AllAttorneysSigned(tc.attorneys))
		})
	}
}
