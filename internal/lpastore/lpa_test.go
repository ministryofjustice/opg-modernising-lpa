package lpastore

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	donordata "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
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
		donor *donordata.Provided
		json  string
	}{
		"minimal": {
			donor: &donordata.Provided{
				LpaUID: "M-0000-1111-2222",
				Type:   donordata.LpaTypePropertyAndAffairs,
				Donor: donordata.Donor{
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
					OtherNames:                "JJ",
					ContactLanguagePreference: localize.Cy,
				},
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{{
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
				ReplacementAttorneys: donordata.Attorneys{},
				WhenCanTheLpaBeUsed:  donordata.CanBeUsedWhenCapacityLost,
				CertificateProvider: donordata.CertificateProvider{
					UID:        certificateProviderUID,
					FirstNames: "Carol",
					LastName:   "Cert",
					Address: place.Address{
						Line1:      "c-line-1",
						TownOrCity: "c-town",
						Country:    "GB",
					},
					CarryOutBy: donordata.ChannelPaper,
				},
				SignedAt: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
			},
			json: `{
"lpaType":"property-and-affairs",
"channel":"online",
"donor":{"uid":"` + donorUID.String() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"","line3":"","town":"town","postcode":"","country":"GB"},"otherNamesKnownBy":"JJ","contactLanguagePreference":"cy"},
"attorneys":[{"uid":"` + attorneyUID.String() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"","line3":"","town":"a-town","postcode":"","country":"GB"},"status":"active","channel":"online"}],
"certificateProvider":{"uid":"` + certificateProviderUID.String() + `","firstNames":"Carol","lastName":"Cert","address":{"line1":"c-line-1","line2":"","line3":"","town":"c-town","postcode":"","country":"GB"},"channel":"paper"},
"restrictionsAndConditions":"",
"whenTheLpaCanBeUsed":"when-capacity-lost",
"signedAt":"2000-01-02T03:04:05.000000006Z"
}`,
		},
		"everything": {
			donor: &donordata.Provided{
				LpaUID: "M-0000-1111-2222",
				Type:   donordata.LpaTypePersonalWelfare,
				Donor: donordata.Donor{
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
					OtherNames:                "JJ",
					ContactLanguagePreference: localize.En,
				},
				Attorneys: donordata.Attorneys{
					TrustCorporation: donordata.TrustCorporation{
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
					Attorneys: []donordata.Attorney{{
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
				AttorneyDecisions: donordata.AttorneyDecisions{
					How: donordata.Jointly,
				},
				ReplacementAttorneys: donordata.Attorneys{
					TrustCorporation: donordata.TrustCorporation{
						UID:           replacementTrustCorporationUID,
						Name:          "UnTrusty",
						CompanyNumber: "65555",
						Address: place.Address{
							Line1:      "a-line-1",
							Line2:      "a-line-2",
							Line3:      "a-line-3",
							TownOrCity: "a-town",
							Postcode:   "A1 1FF",
							Country:    "GB",
						},
					},
					Attorneys: []donordata.Attorney{{
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
				ReplacementAttorneyDecisions: donordata.AttorneyDecisions{
					How:     donordata.JointlyForSomeSeverallyForOthers,
					Details: "umm",
				},
				HowShouldReplacementAttorneysStepIn: donordata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
				LifeSustainingTreatmentOption:       donordata.LifeSustainingTreatmentOptionA,
				Restrictions:                        "do not do this",
				CertificateProvider: donordata.CertificateProvider{
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
					CarryOutBy: donordata.ChannelOnline,
				},
				PeopleToNotify: donordata.PeopleToNotify{{
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
				DonorIdentityUserData: identity.UserData{
					Status:      identity.StatusConfirmed,
					FirstNames:  "John Johnson",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "1", "2"),
					RetrievedAt: time.Date(2002, time.January, 2, 12, 14, 16, 9, time.UTC),
				},
				SignedAt:                                 time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				CertificateProviderNotRelatedConfirmedAt: time.Date(2001, time.February, 3, 4, 5, 6, 7, time.UTC),
			},
			json: `{
"lpaType":"personal-welfare",
"channel":"online",
"donor":{"uid":"` + donorUID.String() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"line-2","line3":"line-3","town":"town","postcode":"F1 1FF","country":"GB"},"otherNamesKnownBy":"JJ","contactLanguagePreference":"en", "identityCheck": {"checkedAt": "2002-01-02T12:14:16.000000009Z","type":"one-login"}},
"attorneys":[
{"uid":"` + attorneyUID.String() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active","channel":"online"},
{"uid":"` + attorney2UID.String() + `","firstNames":"Alice","lastName":"Attorney","dateOfBirth":"1998-01-02","address":{"line1":"aa-line-1","line2":"aa-line-2","line3":"aa-line-3","town":"aa-town","postcode":"A1 1AF","country":"GB"},"status":"active","channel":"paper"},
{"uid":"` + replacementAttorneyUID.String() + `","firstNames":"Richard","lastName":"Attorney","dateOfBirth":"1999-11-12","email":"richard@example.com","address":{"line1":"r-line-1","line2":"r-line-2","line3":"r-line-3","town":"r-town","postcode":"R1 1FF","country":"GB"},"status":"replacement","channel":"online"},
{"uid":"` + replacementAttorney2UID.String() + `","firstNames":"Rachel","lastName":"Attorney","dateOfBirth":"1998-11-12","email":"rachel@example.com","address":{"line1":"rr-line-1","line2":"rr-line-2","line3":"rr-line-3","town":"rr-town","postcode":"R1 1RF","country":"GB"},"status":"replacement","channel":"online"}
],
"trustCorporations":[
{"uid":"` + trustCorporationUID.String() + `","name":"Trusty","companyNumber":"55555","email":"trusty@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active","channel":"online"},
{"uid":"` + replacementTrustCorporationUID.String() + `","name":"UnTrusty","companyNumber":"65555","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"replacement","channel":"paper"}
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
	err := client.SendLpa(nil, &donordata.Provided{})

	assert.NotNil(t, err)
}

func TestClientSendLpaWhenSecretsClientError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("", expectedError)

	client := New("http://base", secretsClient, nil)
	err := client.SendLpa(ctx, &donordata.Provided{})

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
	err := client.SendLpa(ctx, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestClientSendLpaWhenStatusCodeIsNotOK(t *testing.T) {
	testcases := map[int]string{
		http.StatusBadRequest:          "expected 201 response but got 400",
		http.StatusInternalServerError: "expected 201 response but got 500",
	}

	for code, errorName := range testcases {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			ctx := context.Background()

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(mock.Anything, mock.Anything).
				Return("secret", nil)

			doer := newMockDoer(t)
			doer.EXPECT().
				Do(mock.Anything).
				Return(&http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader("hey"))}, nil)

			client := New("http://base", secretsClient, doer)
			err := client.SendLpa(ctx, &donordata.Provided{})

			assert.Equal(t, responseError{name: errorName, body: "hey"}, err)
		})
	}
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
				Type:   donordata.LpaTypePropertyAndAffairs,
				Donor: Donor{
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
					Channel:    donordata.ChannelOnline,
				},
				Attorneys: Attorneys{
					Attorneys: []Attorney{{
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
				ReplacementAttorneys: Attorneys{},
				WhenCanTheLpaBeUsed:  donordata.CanBeUsedWhenCapacityLost,
				CertificateProvider: CertificateProvider{
					UID:        certificateProviderUID,
					FirstNames: "Carol",
					LastName:   "Cert",
					Address: place.Address{
						Line1:      "c-line-1",
						TownOrCity: "c-town",
						Country:    "GB",
					},
					Channel: donordata.ChannelPaper,
				},
				SignedAt: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
			},
			json: `{
"uid":"M-0000-1111-2222",
"channel":"online",
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
				Type:   donordata.LpaTypePersonalWelfare,
				Donor: Donor{
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
					Channel:    donordata.ChannelOnline,
					IdentityCheck: IdentityCheck{
						CheckedAt: time.Date(2002, time.January, 2, 12, 13, 14, 1, time.UTC),
						Type:      "one-login",
					},
				},
				Attorneys: Attorneys{
					TrustCorporation: TrustCorporation{
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
						Channel: donordata.ChannelOnline,
					},
					Attorneys: []Attorney{{
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
				AttorneyDecisions: donordata.AttorneyDecisions{
					How: donordata.Jointly,
				},
				ReplacementAttorneys: Attorneys{
					TrustCorporation: TrustCorporation{
						UID:           replacementTrustCorporationUID,
						Name:          "UnTrusty",
						CompanyNumber: "65555",
						Address: place.Address{
							Line1:      "a-line-1",
							Line2:      "a-line-2",
							Line3:      "a-line-3",
							TownOrCity: "a-town",
							Postcode:   "A1 1FF",
							Country:    "GB",
						},
						Channel: donordata.ChannelPaper,
					},
					Attorneys: []Attorney{{
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
				ReplacementAttorneyDecisions: donordata.AttorneyDecisions{
					How:     donordata.JointlyForSomeSeverallyForOthers,
					Details: "umm",
				},
				HowShouldReplacementAttorneysStepIn: donordata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
				LifeSustainingTreatmentOption:       donordata.LifeSustainingTreatmentOptionA,
				Restrictions:                        "do not do this",
				CertificateProvider: CertificateProvider{
					UID:        certificateProviderUID,
					FirstNames: "Carol",
					LastName:   "Cert",
					Email:      "carol@example.com",
					Phone:      "0700009000",
					Address: place.Address{
						Line1:      "c-line-1",
						Line2:      "c-line-2",
						Line3:      "c-line-3",
						TownOrCity: "c-town",
						Postcode:   "C1 1FF",
						Country:    "GB",
					},
					Channel: donordata.ChannelOnline,
					IdentityCheck: IdentityCheck{
						CheckedAt: time.Date(2002, time.January, 1, 13, 14, 15, 16, time.UTC),
						Type:      "one-login",
					},
				},
				PeopleToNotify: donordata.PeopleToNotify{{
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
"channel":"online",
"donor":{"uid":"` + donorUID.String() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"line-2","line3":"line-3","town":"town","postcode":"F1 1FF","country":"GB"},"otherNamesKnownBy":"JJ","identityCheck":{"checkedAt":"2002-01-02T12:13:14.0000000015Z","type":"one-login"}},
"attorneys":[
{"uid":"` + attorneyUID.String() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active"},
{"uid":"` + attorney2UID.String() + `","firstNames":"Alice","lastName":"Attorney","dateOfBirth":"1998-01-02","email":"alice@example.com","address":{"line1":"aa-line-1","line2":"aa-line-2","line3":"aa-line-3","town":"aa-town","postcode":"A1 1AF","country":"GB"},"status":"active"},
{"uid":"` + replacementAttorneyUID.String() + `","firstNames":"Richard","lastName":"Attorney","dateOfBirth":"1999-11-12","email":"richard@example.com","address":{"line1":"r-line-1","line2":"r-line-2","line3":"r-line-3","town":"r-town","postcode":"R1 1FF","country":"GB"},"status":"replacement"},
{"uid":"` + replacementAttorney2UID.String() + `","firstNames":"Rachel","lastName":"Attorney","dateOfBirth":"1998-11-12","email":"rachel@example.com","address":{"line1":"rr-line-1","line2":"rr-line-2","line3":"rr-line-3","town":"rr-town","postcode":"R1 1RF","country":"GB"},"status":"replacement"}
],
"trustCorporations":[
{"uid":"` + trustCorporationUID.String() + `","name":"Trusty","companyNumber":"55555","email":"trusty@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active","channel":"online"},
{"uid":"` + replacementTrustCorporationUID.String() + `","name":"UnTrusty","companyNumber":"65555","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"replacement","channel":"paper"}
],
"certificateProvider":{"uid":"` + certificateProviderUID.String() + `","firstNames":"Carol","lastName":"Cert","email":"carol@example.com","phone":"0700009000","address":{"line1":"c-line-1","line2":"c-line-2","line3":"c-line-3","town":"c-town","postcode":"C1 1FF","country":"GB"},"channel":"online","identityCheck":{"checkedAt":"2002-01-01T13:14:15.000000016Z","type":"one-login"}},
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

func TestClientLpaWhenStatusCodeIsNotFound(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("secret", nil)

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("hey"))}, nil)

	client := New("http://base", secretsClient, doer)
	_, err := client.Lpa(ctx, "M-0000-1111-2222")

	assert.Equal(t, ErrNotFound, err)
}

func TestClientLpaWhenStatusCodeIsNotOK(t *testing.T) {
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

func TestClientLpas(t *testing.T) {
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
		lpas []*Lpa
		json string
	}{
		"minimal": {
			lpas: []*Lpa{
				{
					LpaUID: "M-0000-1111-2222",
					Type:   donordata.LpaTypePropertyAndAffairs,
					Donor: Donor{
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
						Channel:    donordata.ChannelOnline,
					},
					Attorneys: Attorneys{
						Attorneys: []Attorney{{
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
					ReplacementAttorneys: Attorneys{},
					WhenCanTheLpaBeUsed:  donordata.CanBeUsedWhenCapacityLost,
					CertificateProvider: CertificateProvider{
						UID:        certificateProviderUID,
						FirstNames: "Carol",
						LastName:   "Cert",
						Address: place.Address{
							Line1:      "c-line-1",
							TownOrCity: "c-town",
							Country:    "GB",
						},
						Channel: donordata.ChannelPaper,
					},
					SignedAt: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
			},
			json: `{"lpas":[{
"uid":"M-0000-1111-2222",
"lpaType":"property-and-affairs",
"channel":"online",
"donor":{"uid":"` + donorUID.String() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"","line3":"","town":"town","postcode":"","country":"GB"},"otherNamesKnownBy":"JJ"},
"attorneys":[{"uid":"` + attorneyUID.String() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"","line3":"","town":"a-town","postcode":"","country":"GB"},"status":"active"}],
"certificateProvider":{"uid":"` + certificateProviderUID.String() + `","firstNames":"Carol","lastName":"Cert","address":{"line1":"c-line-1","line2":"","line3":"","town":"c-town","postcode":"","country":"GB"},"channel":"paper"},
"restrictionsAndConditions":"",
"whenTheLpaCanBeUsed":"when-capacity-lost",
"signedAt":"2000-01-02T03:04:05.000000006Z"
}]}`,
		},
		"everything": {
			lpas: []*Lpa{
				{
					LpaUID: "M-0000-1111-2222",
					Type:   donordata.LpaTypePersonalWelfare,
					Donor: Donor{
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
						Channel:    donordata.ChannelOnline,
					},
					Attorneys: Attorneys{
						TrustCorporation: TrustCorporation{
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
						Attorneys: []Attorney{{
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
					AttorneyDecisions: donordata.AttorneyDecisions{
						How: donordata.Jointly,
					},
					ReplacementAttorneys: Attorneys{
						TrustCorporation: TrustCorporation{
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
						Attorneys: []Attorney{{
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
					ReplacementAttorneyDecisions: donordata.AttorneyDecisions{
						How:     donordata.JointlyForSomeSeverallyForOthers,
						Details: "umm",
					},
					HowShouldReplacementAttorneysStepIn: donordata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
					LifeSustainingTreatmentOption:       donordata.LifeSustainingTreatmentOptionA,
					Restrictions:                        "do not do this",
					CertificateProvider: CertificateProvider{
						UID:        certificateProviderUID,
						FirstNames: "Carol",
						LastName:   "Cert",
						Email:      "carol@example.com",
						Phone:      "0700009000",
						Address: place.Address{
							Line1:      "c-line-1",
							Line2:      "c-line-2",
							Line3:      "c-line-3",
							TownOrCity: "c-town",
							Postcode:   "C1 1FF",
							Country:    "GB",
						},
						Channel: donordata.ChannelOnline,
					},
					PeopleToNotify: donordata.PeopleToNotify{{
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
			},
			json: `{"lpas":[{
"uid":"M-0000-1111-2222",
"lpaType":"personal-welfare",
"channel":"online",
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
"certificateProviderNotRelatedConfirmedAt":"2001-02-03T04:05:06.000000007Z"
}]}`,
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
						assert.Equal(t, http.MethodPost, req.Method) &&
						assert.Equal(t, "http://base/lpas", req.URL.String()) &&
						assert.Equal(t, "application/json", req.Header.Get("Content-Type")) &&
						assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjAwMDAwMDAwLTAwMDAtNDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsImlhdCI6OTQ2NzgyMjQ1fQ.V7MxjZw7-K8ehujYn4e0gef7s23r2UDlTbyzQtpTKvo", req.Header.Get("X-Jwt-Authorization"))
				})).
				Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tc.json))}, nil)

			client := New("http://base", secretsClient, doer)
			client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
			lpas, err := client.Lpas(ctx, []string{"M-0000-1111-2222"})

			assert.Nil(t, err)
			assert.Equal(t, tc.lpas, lpas)
		})
	}
}

func TestClientLpasWhenNewRequestError(t *testing.T) {
	client := New("http://base", nil, nil)
	_, err := client.Lpas(nil, []string{"M-0000-1111-2222"})

	assert.NotNil(t, err)
}

func TestClientLpasWhenSecretsClientError(t *testing.T) {
	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("", expectedError)

	client := New("http://base", secretsClient, nil)
	_, err := client.Lpas(ctx, []string{"M-0000-1111-2222"})

	assert.Equal(t, expectedError, err)
}

func TestClientLpasWhenDoerError(t *testing.T) {
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
	_, err := client.Lpas(ctx, []string{"M-0000-1111-2222"})

	assert.Equal(t, expectedError, err)
}

func TestClientLpasWhenStatusCodeIsNotOK(t *testing.T) {
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
	_, err := client.Lpas(ctx, []string{"M-0000-1111-2222"})

	assert.Equal(t, responseError{name: "expected 200 response but got 400", body: "hey"}, err)
}

func TestAllAttorneysSigned(t *testing.T) {
	attorneySigned := time.Now()

	testcases := map[string]struct {
		lpa      Lpa
		expected bool
	}{
		"no attorneys": {
			expected: false,
		},
		"need attorney to sign": {
			lpa: Lpa{
				Attorneys:            Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}, {}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}}},
			},
			expected: false,
		},
		"need replacement attorney to sign": {
			lpa: Lpa{
				Attorneys:            Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{}, {SignedAt: attorneySigned}}},
			},
			expected: false,
		},
		"all attorneys signed": {
			lpa: Lpa{
				Attorneys:            Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}, {SignedAt: attorneySigned}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}}},
			},
			expected: true,
		},
		"trust corporations not signed": {
			lpa: Lpa{
				Attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
			},
			expected: false,
		},
		"trust corporations signatory not signed": {
			lpa: Lpa{
				Attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a", Signatories: []TrustCorporationSignatory{{}}}},
			},
			expected: false,
		},
		"replacement trust corporations not signed": {
			lpa: Lpa{
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a", Signatories: []TrustCorporationSignatory{{SignedAt: attorneySigned}}}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			expected: false,
		},
		"trust corporations signed": {
			lpa: Lpa{
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a", Signatories: []TrustCorporationSignatory{{SignedAt: attorneySigned}, {SignedAt: attorneySigned}}}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r", Signatories: []TrustCorporationSignatory{{SignedAt: attorneySigned}}}},
			},
			expected: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.AllAttorneysSigned())
		})
	}
}

func TestDonorFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Donor{FirstNames: "John", LastName: "Smith"}.FullName())
}

func TestAttorneyFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Attorney{FirstNames: "John", LastName: "Smith"}.FullName())
}

func TestCertificateProviderFullName(t *testing.T) {
	assert.Equal(t, "John Smith", CertificateProvider{FirstNames: "John", LastName: "Smith"}.FullName())
}

func TestAttorneysLen(t *testing.T) {
	testcases := map[string]struct {
		attorneys Attorneys
		len       int
	}{
		"trust corporation": {
			attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
			len:       1,
		},
		"attorneys": {
			attorneys: Attorneys{Attorneys: []Attorney{{}, {}}},
			len:       2,
		},
		"both": {
			attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a"}, Attorneys: []Attorney{{}, {}}},
			len:       3,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.len, tc.attorneys.Len())
		})
	}
}

func TestAttorneysGet(t *testing.T) {
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	testCases := map[string]struct {
		attorneys        Attorneys
		expectedAttorney Attorney
		uid              actoruid.UID
		expectedFound    bool
	}{
		"attorney exists": {
			attorneys:        Attorneys{Attorneys: []Attorney{{UID: uid1, FirstNames: "Bob"}, {UID: uid2}}},
			expectedAttorney: Attorney{UID: uid1, FirstNames: "Bob"},
			uid:              uid1,
			expectedFound:    true,
		},
		"attorney does not exist": {
			attorneys:        Attorneys{Attorneys: []Attorney{{UID: uid1, FirstNames: "Bob"}, {UID: uid2}}},
			expectedAttorney: Attorney{},
			uid:              actoruid.New(),
			expectedFound:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			a, found := tc.attorneys.Get(tc.uid)

			assert.Equal(t, tc.expectedFound, found)
			assert.Equal(t, tc.expectedAttorney, a)
		})
	}
}

func TestAttorneysFullNames(t *testing.T) {
	testcases := map[string]struct {
		attorneys  Attorneys
		fullNames  []string
		firstNames []string
	}{
		"empty": {},
		"attorneys": {
			attorneys: Attorneys{
				Attorneys: []Attorney{
					{FirstNames: "Bob Alan George", LastName: "Jones"},
					{FirstNames: "Samantha", LastName: "Smith"},
					{FirstNames: "Abby Helen", LastName: "Burns-Simpson"},
				},
			},
			fullNames:  []string{"Bob Alan George Jones", "Samantha Smith", "Abby Helen Burns-Simpson"},
			firstNames: []string{"Bob Alan George", "Samantha", "Abby Helen"},
		},
		"trust corporation": {
			attorneys: Attorneys{
				TrustCorporation: TrustCorporation{Name: "Corp corp"},
			},
			fullNames:  []string{"Corp corp"},
			firstNames: []string{"Corp corp"},
		},
		"both": {
			attorneys: Attorneys{
				TrustCorporation: TrustCorporation{Name: "Corp corp"},
				Attorneys: []Attorney{
					{FirstNames: "Bob", LastName: "Jones"},
				},
			},
			fullNames:  []string{"Corp corp", "Bob Jones"},
			firstNames: []string{"Corp corp", "Bob"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.fullNames, tc.attorneys.FullNames())
		})
	}
}
