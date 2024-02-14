package lpastore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

type mockCredentialsProvider struct{}

func (m *mockCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     "abc",
		SecretAccessKey: "",
	}, nil
}

func (m *mockCredentialsProvider) IsExpired() bool {
	return false
}

func ActorUID() matchers.Matcher {
	return matchers.Regex("urn:opg:poas:makeregister:users:123e4567-e89b-12d3-a456-426655440000", "urn:opg:poas:makeregister:users:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
}

func AwsAuthorization() matchers.Matcher {
	return matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=6bf29d8faab8da2f4b2df0bf24705d9d3b8774fdeb482d4a64f50bb32e178df8", "(AWS4-HMAC-SHA256 Credential=|SignedHeaders=|Signature=).*")
}

func TestResponseError(t *testing.T) {
	err := responseError{name: "name", body: 5}
	assert.Equal(t, "name", err.Error())
	assert.Equal(t, "name", err.Title())
	assert.Equal(t, 5, err.Data())
}

func TestClientSendLpa(t *testing.T) {
	var donorUID actoruid.UID
	json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users:uid"`), &donorUID)
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
"donor":{"uid":"` + donorUID.PrefixedString() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"","line3":"","town":"town","postcode":"","country":"GB"},"otherNamesKnownBy":"JJ"},
"attorneys":[{"uid":"` + attorneyUID.PrefixedString() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"","line3":"","town":"a-town","postcode":"","country":"GB"},"status":"active"}],
"certificateProvider":{"uid":"` + certificateProviderUID.PrefixedString() + `","firstNames":"Carol","lastName":"Cert","address":{"line1":"c-line-1","line2":"","line3":"","town":"c-town","postcode":"","country":"GB"},"channel":"paper"},
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
"donor":{"uid":"` + donorUID.PrefixedString() + `","firstNames":"John Johnson","lastName":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"line-2","line3":"line-3","town":"town","postcode":"F1 1FF","country":"GB"},"otherNamesKnownBy":"JJ"},
"attorneys":[
{"uid":"` + attorneyUID.PrefixedString() + `","firstNames":"Adam","lastName":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active"},
{"uid":"` + attorney2UID.PrefixedString() + `","firstNames":"Alice","lastName":"Attorney","dateOfBirth":"1998-01-02","email":"alice@example.com","address":{"line1":"aa-line-1","line2":"aa-line-2","line3":"aa-line-3","town":"aa-town","postcode":"A1 1AF","country":"GB"},"status":"active"},
{"uid":"` + replacementAttorneyUID.PrefixedString() + `","firstNames":"Richard","lastName":"Attorney","dateOfBirth":"1999-11-12","email":"richard@example.com","address":{"line1":"r-line-1","line2":"r-line-2","line3":"r-line-3","town":"r-town","postcode":"R1 1FF","country":"GB"},"status":"replacement"},
{"uid":"` + replacementAttorney2UID.PrefixedString() + `","firstNames":"Rachel","lastName":"Attorney","dateOfBirth":"1998-11-12","email":"rachel@example.com","address":{"line1":"rr-line-1","line2":"rr-line-2","line3":"rr-line-3","town":"rr-town","postcode":"R1 1RF","country":"GB"},"status":"replacement"}
],
"trustCorporations":[
{"uid":"` + trustCorporationUID.PrefixedString() + `","name":"Trusty","companyNumber":"55555","email":"trusty@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active"},
{"uid":"` + replacementTrustCorporationUID.PrefixedString() + `","name":"UnTrusty","companyNumber":"65555","email":"untrusty@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"replacement"}
],
"certificateProvider":{"uid":"` + certificateProviderUID.PrefixedString() + `","firstNames":"Carol","lastName":"Cert","email":"carol@example.com","address":{"line1":"c-line-1","line2":"c-line-2","line3":"c-line-3","town":"c-town","postcode":"C1 1FF","country":"GB"},"channel":"online"},
"peopleToNotify":[{"uid":"` + personToNotifyUID.PrefixedString() + `","firstNames":"Peter","lastName":"Notify","address":{"line1":"p-line-1","line2":"p-line-2","line3":"p-line-3","town":"p-town","postcode":"P1 1FF","country":"GB"}}],
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
						assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOnVpZCIsImlhdCI6OTQ2NzgyMjQ1fQ.4uB6hoY67WD6cmx3V-AG4R3s2SzP9gRbiWyEFgqzPJo", req.Header.Get("X-Jwt-Authorization")) &&
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

func TestClientSendCertificateProvider(t *testing.T) {
	var uid actoruid.UID
	json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users:uid"`), &uid)

	certificateProvider := &actor.CertificateProviderProvidedDetails{
		UID: uid,
		HomeAddress: place.Address{
			Line1:      "line-1",
			Line2:      "line-2",
			Line3:      "line-3",
			TownOrCity: "town",
			Postcode:   "postcode",
			Country:    "GB",
		},
		Certificate: actor.Certificate{
			Agreed: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
		},
		ContactLanguagePreference: localize.Cy,
	}
	json := `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[{"key":"/certificateProvider/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"},{"key":"/certificateProvider/contactLanguagePreference","old":null,"new":"cy"},{"key":"/certificateProvider/address/line1","old":null,"new":"line-1"},{"key":"/certificateProvider/address/line2","old":null,"new":"line-2"},{"key":"/certificateProvider/address/line3","old":null,"new":"line-3"},{"key":"/certificateProvider/address/town","old":null,"new":"town"},{"key":"/certificateProvider/address/postcode","old":null,"new":"postcode"},{"key":"/certificateProvider/address/country","old":null,"new":"GB"}]}`

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
				assert.Equal(t, http.MethodPost, req.Method) &&
				assert.Equal(t, "http://base/lpas/lpa-uid/updates", req.URL.String()) &&
				assert.Equal(t, "application/json", req.Header.Get("Content-Type")) &&
				assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOnVpZCIsImlhdCI6OTQ2NzgyMjQ1fQ.4uB6hoY67WD6cmx3V-AG4R3s2SzP9gRbiWyEFgqzPJo", req.Header.Get("X-Jwt-Authorization")) &&
				assert.JSONEq(t, json, string(body))
		})).
		Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

	client := New("http://base", secretsClient, doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendCertificateProvider(ctx, "lpa-uid", certificateProvider)

	assert.Nil(t, err)
}

func TestClientSendAttorney(t *testing.T) {
	var uid1, uid2 actoruid.UID
	json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users:uid1"`), &uid1)
	json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users:uid2"`), &uid2)

	testcases := map[string]struct {
		attorney *actor.AttorneyProvidedDetails
		donor    *actor.DonorProvidedDetails
		json     string
	}{
		"attorney": {
			attorney: &actor.AttorneyProvidedDetails{
				UID:                       uid2,
				Mobile:                    "07777",
				Confirmed:                 time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				ContactLanguagePreference: localize.Cy,
			},
			donor: &actor.DonorProvidedDetails{
				LpaUID: "lpa-uid",
				Attorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{
						{UID: uid1}, {UID: uid2},
					},
				},
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/attorneys/1/mobile","old":null,"new":"07777"},{"key":"/attorneys/1/contactLanguagePreference","old":null,"new":"cy"},{"key":"/attorneys/1/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
		},
		"replacement attorney": {
			attorney: &actor.AttorneyProvidedDetails{
				UID:                       uid2,
				IsReplacement:             true,
				Mobile:                    "07777",
				Confirmed:                 time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				ContactLanguagePreference: localize.Cy,
			},
			donor: &actor.DonorProvidedDetails{
				LpaUID: "lpa-uid",
				Attorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{
						{UID: uid1}, {UID: uid2},
					},
				},
				ReplacementAttorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{
						{UID: uid1}, {UID: uid2},
					},
				},
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/attorneys/3/mobile","old":null,"new":"07777"},{"key":"/attorneys/3/contactLanguagePreference","old":null,"new":"cy"},{"key":"/attorneys/3/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
		},
		"trust corporation": {
			attorney: &actor.AttorneyProvidedDetails{
				UID:                uid2,
				IsTrustCorporation: true,
				Mobile:             "07777",
				AuthorisedSignatories: [2]actor.TrustCorporationSignatory{{
					FirstNames:        "John",
					LastName:          "Signer",
					ProfessionalTitle: "Director",
					Confirmed:         time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				}, {
					FirstNames:        "Dave",
					LastName:          "Signer",
					ProfessionalTitle: "Assistant to the Director",
					Confirmed:         time.Date(2000, time.January, 2, 3, 4, 5, 7, time.UTC),
				}},
				ContactLanguagePreference: localize.En,
			},
			donor: &actor.DonorProvidedDetails{
				LpaUID: "lpa-uid",
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/trustCorporations/0/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/0/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/0/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/0/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/0/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/0/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"},{"key":"/trustCorporations/0/signatories/1/firstNames","old":null,"new":"Dave"},{"key":"/trustCorporations/0/signatories/1/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/0/signatories/1/professionalTitle","old":null,"new":"Assistant to the Director"},{"key":"/trustCorporations/0/signatories/1/signedAt","old":null,"new":"2000-01-02T03:04:05.000000007Z"}]}`,
		},
		"replacement trust corporation": {
			attorney: &actor.AttorneyProvidedDetails{
				UID:                uid2,
				IsTrustCorporation: true,
				IsReplacement:      true,
				Mobile:             "07777",
				AuthorisedSignatories: [2]actor.TrustCorporationSignatory{{
					FirstNames:        "John",
					LastName:          "Signer",
					ProfessionalTitle: "Director",
					Confirmed:         time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				}},
				ContactLanguagePreference: localize.En,
			},
			donor: &actor.DonorProvidedDetails{
				LpaUID: "lpa-uid",
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/trustCorporations/0/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/0/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/0/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/0/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/0/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/0/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
		},
		"replacement trust corporation when also attorney trust corporation": {
			attorney: &actor.AttorneyProvidedDetails{
				UID:                uid2,
				IsTrustCorporation: true,
				IsReplacement:      true,
				Mobile:             "07777",
				AuthorisedSignatories: [2]actor.TrustCorporationSignatory{{
					FirstNames:        "John",
					LastName:          "Signer",
					ProfessionalTitle: "Director",
					Confirmed:         time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				}},
				ContactLanguagePreference: localize.En,
			},
			donor: &actor.DonorProvidedDetails{
				LpaUID:    "lpa-uid",
				Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/trustCorporations/1/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/1/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/1/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/1/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/1/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/1/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
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
						assert.Equal(t, http.MethodPost, req.Method) &&
						assert.Equal(t, "http://base/lpas/lpa-uid/updates", req.URL.String()) &&
						assert.Equal(t, "application/json", req.Header.Get("Content-Type")) &&
						assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOnVpZDIiLCJpYXQiOjk0Njc4MjI0NX0.ZHegwOTVV4PfWMO6k2hrmhM_KYgN0NAPghDrXjS38Do", req.Header.Get("X-Jwt-Authorization")) &&
						assert.JSONEq(t, tc.json, string(body))
				})).
				Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

			client := New("http://base", secretsClient, doer)
			client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
			err := client.SendAttorney(ctx, tc.donor, tc.attorney)

			assert.Nil(t, err)
		})
	}
}

func TestClientServiceContract(t *testing.T) {
	now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }

	cfg := aws.Config{
		Region:      "eu-west-1",
		Credentials: &mockCredentialsProvider{},
	}

	address := place.Address{
		Line1:      "line-1",
		Line2:      "line-2",
		Line3:      "line-3",
		TownOrCity: "town",
		Postcode:   "F1 1FF",
		Country:    "GB",
	}

	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "modernising-lpa",
		Provider: "data-lpa-store",
		LogDir:   "../../logs",
		PactDir:  "../../pacts",
	})
	assert.Nil(t, err)

	t.Run("SendLpa", func(t *testing.T) {
		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 does not exist").
			UponReceiving("A request to create a new case").
			WithRequest(http.MethodPut, "/lpas/M-0000-1111-2222", func(b *consumer.V2RequestBuilder) {
				b.
					Header("Content-Type", matchers.String("application/json")).
					Header("Authorization", AwsAuthorization()).
					Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"lpaType":                       matchers.Regex("personal-welfare", "personal-welfare|property-and-affairs"),
						"lifeSustainingTreatmentOption": matchers.Regex("option-a", "option-a|option-b"),
						"donor": matchers.Like(map[string]any{
							"uid":         ActorUID(),
							"firstNames":  matchers.String("John Johnson"),
							"lastName":    matchers.String("Smith"),
							"dateOfBirth": matchers.Regex("2000-01-02", "\\d{4}-\\d{2}-\\d{2}"),
							"email":       matchers.String("john@example.com"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("line-1"),
								"line2":    matchers.String("line-2"),
								"line3":    matchers.String("line-3"),
								"town":     matchers.String("town"),
								"postcode": matchers.String("F1 1FF"),
								"country":  matchers.String("GB"),
							}),
						}),
						"attorneys": matchers.EachLike(map[string]any{
							"uid":         ActorUID(),
							"firstNames":  matchers.String("Adam"),
							"lastName":    matchers.String("Attorney"),
							"dateOfBirth": matchers.Regex("1999-01-02", "\\d{4}-\\d{2}-\\d{2}"),
							"email":       matchers.String("adam@example.com"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("a-line-1"),
								"line2":    matchers.String("a-line-2"),
								"line3":    matchers.String("a-line-3"),
								"town":     matchers.String("a-town"),
								"postcode": matchers.String("A1 1FF"),
								"country":  matchers.String("GB"),
							}),
							"status": matchers.Regex("active", "active|replacement"),
						}, 1),
						"certificateProvider": matchers.Like(map[string]any{
							"uid":        ActorUID(),
							"firstNames": matchers.String("Charles"),
							"lastName":   matchers.String("Certificate"),
							"email":      matchers.String("charles@example.com"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("a-line-1"),
								"line2":    matchers.String("a-line-2"),
								"line3":    matchers.String("a-line-3"),
								"town":     matchers.String("a-town"),
								"postcode": matchers.String("A1 1FF"),
								"country":  matchers.String("GB"),
							}),
							"channel": matchers.Regex("online", "online|post"),
						}),
						"restrictionsAndConditions": matchers.String("hmm"),
						"signedAt":                  matchers.Regex("2000-01-02T12:13:14.00000Z", `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?Z`),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				b.Header("Content-Type", matchers.String("application/json"))
				b.JSONBody(matchers.Map{})
			})

		assert.Nil(t, mockProvider.ExecuteTest(t, func(config consumer.MockServerConfig) error {
			baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(mock.Anything, mock.Anything).
				Return("secret", nil)

			client := &Client{
				baseURL:       baseURL,
				secretsClient: secretsClient,
				doer:          lambda.New(cfg, v4.NewSigner(), http.DefaultClient, now),
				now:           now,
			}

			err := client.SendLpa(context.Background(), &actor.DonorProvidedDetails{
				LpaUID:                        "M-0000-1111-2222",
				Type:                          actor.LpaTypePersonalWelfare,
				LifeSustainingTreatmentOption: actor.LifeSustainingTreatmentOptionA,
				Donor: actor.Donor{
					UID:         actoruid.New(),
					FirstNames:  "John Johnson",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "1", "2"),
					Email:       "john@example.com",
					Address:     address,
				},
				Attorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{{
						UID:         actoruid.New(),
						FirstNames:  "Alice",
						LastName:    "Attorney",
						DateOfBirth: date.New("1998", "1", "2"),
						Email:       "alice@example.com",
						Address:     address,
					}},
				},
				ReplacementAttorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{{
						UID:         actoruid.New(),
						FirstNames:  "Richard",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "11", "12"),
						Email:       "richard@example.com",
						Address:     address,
					}},
				},
				CertificateProvider: actor.CertificateProvider{
					UID:        actoruid.New(),
					FirstNames: "Charles",
					LastName:   "Certificate",
					Email:      "charles@example.com",
					Address:    address,
					CarryOutBy: actor.Online,
				},
				Restrictions: "hmm",
				SignedAt:     time.Date(2000, time.January, 2, 12, 13, 14, 0, time.UTC),
			})

			assert.Nil(t, err)
			return nil
		}))
	})

	t.Run("SendLpa when already exists", func(t *testing.T) {
		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 exists").
			UponReceiving("A request to create a case with existing UID").
			WithRequest(http.MethodPut, "/lpas/M-0000-1111-2222", func(b *consumer.V2RequestBuilder) {
				b.
					Header("Content-Type", matchers.String("application/json")).
					Header("Authorization", AwsAuthorization()).
					Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"lpaType": matchers.Regex("personal-welfare", "personal-welfare|property-and-affairs"),
						"donor": matchers.Like(map[string]any{
							"uid":         ActorUID(),
							"firstNames":  matchers.String("John Johnson"),
							"lastName":    matchers.String("Smith"),
							"dateOfBirth": matchers.Regex("2000-01-02", "\\d{4}-\\d{2}-\\d{2}"),
							"email":       matchers.String("john@example.com"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("line-1"),
								"line2":    matchers.String("line-2"),
								"line3":    matchers.String("line-3"),
								"town":     matchers.String("town"),
								"postcode": matchers.String("F1 1FF"),
								"country":  matchers.String("GB"),
							}),
							"otherNamesKnownBy": matchers.String("JJ"),
						}),
						"attorneys": matchers.EachLike(map[string]any{
							"uid":         ActorUID(),
							"firstNames":  matchers.String("Adam"),
							"lastName":    matchers.String("Attorney"),
							"dateOfBirth": matchers.Regex("1999-01-02", "\\d{4}-\\d{2}-\\d{2}"),
							"email":       matchers.String("adam@example.com"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("a-line-1"),
								"line2":    matchers.String("a-line-2"),
								"line3":    matchers.String("a-line-3"),
								"town":     matchers.String("a-town"),
								"postcode": matchers.String("A1 1FF"),
								"country":  matchers.String("GB"),
							}),
							"status": matchers.Regex("active", "active|replacement"),
						}, 1),
						"certificateProvider": matchers.Like(map[string]any{
							"uid":        ActorUID(),
							"firstNames": matchers.String("Charles"),
							"lastName":   matchers.String("Certificate"),
							"email":      matchers.String("charles@example.com"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("a-line-1"),
								"line2":    matchers.String("a-line-2"),
								"line3":    matchers.String("a-line-3"),
								"town":     matchers.String("a-town"),
								"postcode": matchers.String("A1 1FF"),
								"country":  matchers.String("GB"),
							}),
							"channel": matchers.Regex("online", "online|post"),
						}),
						"peopleToNotify": matchers.EachLike(map[string]any{
							"uid":        ActorUID(),
							"firstNames": matchers.String("Peter"),
							"lastName":   matchers.String("Person"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("a-line-1"),
								"line2":    matchers.String("a-line-2"),
								"line3":    matchers.String("a-line-3"),
								"town":     matchers.String("a-town"),
								"postcode": matchers.String("A1 1FF"),
								"country":  matchers.String("GB"),
							}),
						}, 0),
						"restrictionsAndConditions": matchers.String("hmm"),
						"signedAt":                  matchers.Regex("2000-01-02T12:13:14.00000Z", `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?Z`),
					})
			}).
			WillRespondWith(http.StatusBadRequest, func(b *consumer.V2ResponseBuilder) {
				b.Header("Content-Type", matchers.String("application/json"))
				b.JSONBody(matchers.Map{
					"code":   matchers.String("INVALID_REQUEST"),
					"detail": matchers.String("LPA with UID already exists"),
				})
			})

		assert.Nil(t, mockProvider.ExecuteTest(t, func(config consumer.MockServerConfig) error {
			baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(mock.Anything, mock.Anything).
				Return("secret", nil)

			client := &Client{
				baseURL:       baseURL,
				secretsClient: secretsClient,
				doer:          lambda.New(cfg, v4.NewSigner(), http.DefaultClient, now),
				now:           now,
			}

			err := client.SendLpa(context.Background(), &actor.DonorProvidedDetails{
				LpaUID: "M-0000-1111-2222",
				Type:   actor.LpaTypePersonalWelfare,
				Donor: actor.Donor{
					UID:         actoruid.New(),
					FirstNames:  "John Johnson",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "1", "2"),
					Email:       "john@example.com",
					Address:     address,
					OtherNames:  "JJ",
				},
				Attorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{{
						UID:         actoruid.New(),
						FirstNames:  "Alice",
						LastName:    "Attorney",
						DateOfBirth: date.New("1998", "1", "2"),
						Email:       "alice@example.com",
						Address:     address,
					}},
				},
				ReplacementAttorneys: actor.Attorneys{
					Attorneys: []actor.Attorney{{
						UID:         actoruid.New(),
						FirstNames:  "Richard",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "11", "12"),
						Email:       "richard@example.com",
						Address:     address,
					}},
				},
				CertificateProvider: actor.CertificateProvider{
					UID:        actoruid.New(),
					FirstNames: "Charles",
					LastName:   "Certificate",
					Email:      "charles@example.com",
					Address:    address,
					CarryOutBy: actor.Online,
				},
				PeopleToNotify: actor.PeopleToNotify{{
					UID:        actoruid.New(),
					FirstNames: "Peter",
					LastName:   "Person",
					Address:    address,
				}},
				Restrictions: "hmm",
				SignedAt:     time.Date(2000, time.January, 2, 12, 13, 14, 0, time.UTC),
			})

			assert.Equal(t, responseError{name: "expected 201 response but got 400", body: `{"code":"INVALID_REQUEST","detail":"LPA with UID already exists"}`}, err)
			return nil
		}))
	})

	t.Run("SendAttorney", func(t *testing.T) {
		uid := actoruid.New()

		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 exists").
			UponReceiving("A request to send the attorney data").
			WithRequest(http.MethodPost, "/lpas/M-0000-1111-2222/updates", func(b *consumer.V2RequestBuilder) {
				b.
					Header("Content-Type", matchers.String("application/json")).
					Header("Authorization", AwsAuthorization()).
					Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"type": matchers.Like("ATTORNEY_SIGN"),
						"changes": matchers.Like([]map[string]any{{
							"key": matchers.Like("/attorneys/0/mobile"),
							"old": matchers.Like(nil),
							"new": matchers.Like("07777777"),
						}, {
							"key": matchers.Like("/attorneys/0/contactLanguagePreference"),
							"old": matchers.Like(nil),
							"new": matchers.Like("cy"),
						}, {
							"key": matchers.Like("/attorneys/0/signedAt"),
							"old": matchers.Like(nil),
							"new": matchers.Like("2020-01-01T12:13:14Z"),
						}}),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				b.Header("Content-Type", matchers.String("application/json"))
				b.JSONBody(matchers.Map{})
			})

		assert.Nil(t, mockProvider.ExecuteTest(t, func(config consumer.MockServerConfig) error {
			baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(mock.Anything, mock.Anything).
				Return("secret", nil)

			client := &Client{
				baseURL:       baseURL,
				secretsClient: secretsClient,
				doer:          lambda.New(cfg, v4.NewSigner(), http.DefaultClient, now),
				now:           now,
			}

			err := client.SendAttorney(context.Background(),
				&actor.DonorProvidedDetails{
					LpaUID: "M-0000-1111-2222",
					Attorneys: actor.Attorneys{
						Attorneys: []actor.Attorney{{UID: uid}},
					},
				},
				&actor.AttorneyProvidedDetails{
					UID:                       uid,
					Mobile:                    "07777777",
					Confirmed:                 time.Date(2020, time.January, 1, 12, 13, 14, 0, time.UTC),
					ContactLanguagePreference: localize.Cy,
				})
			assert.Nil(t, err)
			return nil
		}))
	})

	t.Run("SendCertificateProvider", func(t *testing.T) {
		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 exists").
			UponReceiving("A request to send the certificate provider data").
			WithRequest(http.MethodPost, "/lpas/M-0000-1111-2222/updates", func(b *consumer.V2RequestBuilder) {
				b.
					Header("Content-Type", matchers.String("application/json")).
					Header("Authorization", AwsAuthorization()).
					Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"type": matchers.Like("CERTIFICATE_PROVIDER_SIGN"),
						"changes": matchers.Like([]map[string]any{{
							"key": matchers.Like("/certificateProvider/contactLanguagePreference"),
							"old": matchers.Like(nil),
							"new": matchers.Like("cy"),
						}, {
							"key": matchers.Like("/certificateProvider/signedAt"),
							"old": matchers.Like(nil),
							"new": matchers.Like("2020-01-01T12:13:14Z"),
						}}),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				b.Header("Content-Type", matchers.String("application/json"))
				b.JSONBody(matchers.Map{})
			})

		assert.Nil(t, mockProvider.ExecuteTest(t, func(config consumer.MockServerConfig) error {
			baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(mock.Anything, mock.Anything).
				Return("secret", nil)

			client := &Client{
				baseURL:       baseURL,
				secretsClient: secretsClient,
				doer:          lambda.New(cfg, v4.NewSigner(), http.DefaultClient, now),
				now:           now,
			}

			err := client.SendCertificateProvider(context.Background(), "M-0000-1111-2222",
				&actor.CertificateProviderProvidedDetails{
					Certificate: actor.Certificate{
						Agreed: time.Date(2020, time.January, 1, 12, 13, 14, 0, time.UTC),
					},
					ContactLanguagePreference: localize.Cy,
				})
			assert.Nil(t, err)
			return nil
		}))
	})

	t.Run("SendCertificateProvider when professional", func(t *testing.T) {
		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 exists").
			UponReceiving("A request to send the certificate provider data for a professional").
			WithRequest(http.MethodPost, "/lpas/M-0000-1111-2222/updates", func(b *consumer.V2RequestBuilder) {
				b.
					Header("Content-Type", matchers.String("application/json")).
					Header("Authorization", AwsAuthorization()).
					Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"type": matchers.Like("CERTIFICATE_PROVIDER_SIGN"),
						"changes": matchers.Like([]map[string]any{{
							"key": matchers.Like("/certificateProvider/contactLanguagePreference"),
							"old": matchers.Like(nil),
							"new": matchers.Like("cy"),
						}, {
							"key": matchers.Like("/certificateProvider/signedAt"),
							"old": matchers.Like(nil),
							"new": matchers.Like("2020-01-01T12:13:14Z"),
						}, {
							"key": matchers.Like("/certificateProvider/address/line1"),
							"old": matchers.Like(nil),
							"new": matchers.Like("123 Fake Street"),
						}, {
							"key": matchers.Like("/certificateProvider/address/town"),
							"old": matchers.Like(nil),
							"new": matchers.Like("Faketon"),
						}, {
							"key": matchers.Like("/certificateProvider/address/country"),
							"old": matchers.Like(nil),
							"new": matchers.Like("GB"),
						}}),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				b.Header("Content-Type", matchers.String("application/json"))
				b.JSONBody(matchers.Map{})
			})

		assert.Nil(t, mockProvider.ExecuteTest(t, func(config consumer.MockServerConfig) error {
			baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(mock.Anything, mock.Anything).
				Return("secret", nil)

			client := &Client{
				baseURL:       baseURL,
				secretsClient: secretsClient,
				doer:          lambda.New(cfg, v4.NewSigner(), http.DefaultClient, now),
				now:           now,
			}

			return client.SendCertificateProvider(context.Background(), "M-0000-1111-2222",
				&actor.CertificateProviderProvidedDetails{
					Certificate: actor.Certificate{
						Agreed: time.Date(2020, time.January, 1, 12, 13, 14, 0, time.UTC),
					},
					ContactLanguagePreference: localize.Cy,
					HomeAddress: place.Address{
						Line1:      "123 Fake Street",
						TownOrCity: "Faketon",
						Country:    "GB",
					},
				})
		}))
	})

	t.Run("sendUpdate", func(t *testing.T) {
		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 exists").
			UponReceiving("A request to update the lpa").
			WithRequest(http.MethodPost, "/lpas/M-0000-1111-2222/updates", func(b *consumer.V2RequestBuilder) {
				b.
					Header("Content-Type", matchers.String("application/json")).
					Header("Authorization", AwsAuthorization()).
					Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"type": matchers.Like("A_TYPE"),
						"changes": matchers.EachLike(map[string]any{
							"key": matchers.Like("/a/key"),
							"old": matchers.Like("old"),
							"new": matchers.Like("new"),
						}, 1),
					})
			}).
			WillRespondWith(http.StatusBadRequest, func(b *consumer.V2ResponseBuilder) {
				b.Header("Content-Type", matchers.String("application/json"))
				b.JSONBody(matchers.Map{
					"code":   matchers.String("INVALID_REQUEST"),
					"detail": matchers.String("Invalid request"),
				})
			})

		err := mockProvider.ExecuteTest(t, func(config consumer.MockServerConfig) error {
			baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(mock.Anything, mock.Anything).
				Return("secret", nil)

			client := &Client{
				baseURL:       baseURL,
				secretsClient: secretsClient,
				doer:          lambda.New(cfg, v4.NewSigner(), http.DefaultClient, now),
				now:           now,
			}

			return client.sendUpdate(context.Background(), "M-0000-1111-2222", actoruid.New(), updateRequest{
				Type: "A_TYPE",
				Changes: []updateRequestChange{
					{Key: "/a/key", Old: "old", New: "new"},
				},
			})
		})

		assert.Equal(t, responseError{name: "expected 201 response but got 400", body: `{"code":"INVALID_REQUEST","detail":"Invalid request"}`}, err)
	})
}

func TestCheckHealth(t *testing.T) {
	var endpointCalled string
	var requestMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		rBody, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(rBody))

		endpointCalled = r.URL.String()
		requestMethod = r.Method

		w.Write([]byte(`{"status":"OK"}`))
	}))

	client := New(server.URL, nil, server.Client())

	err := client.CheckHealth(context.Background())

	assert.Equal(t, http.MethodGet, requestMethod)
	assert.Equal(t, "/health-check", endpointCalled)
	assert.Nil(t, err)
}

func TestCheckHealthOnNewRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	client := New(server.URL+"`invalid-url-format", nil, server.Client())
	err := client.CheckHealth(context.Background())
	assert.NotNil(t, err)
}

func TestCheckHealthOnDoRequestError(t *testing.T) {
	httpClient := newMockDoer(t)
	httpClient.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	client := New("/", nil, httpClient)
	err := client.CheckHealth(context.Background())
	assert.Equal(t, expectedError, err)
}

func TestCheckHealthWhenNotOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	client := New(server.URL, nil, server.Client())
	err := client.CheckHealth(context.Background())
	assert.NotNil(t, err)
}
