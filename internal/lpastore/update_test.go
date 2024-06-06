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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestClientSendRegister(t *testing.T) {
	json := `{"type":"REGISTER","changes":null}`

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
				assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjAwMDAwMDAwLTAwMDAtNDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsImlhdCI6OTQ2NzgyMjQ1fQ.V7MxjZw7-K8ehujYn4e0gef7s23r2UDlTbyzQtpTKvo", req.Header.Get("X-Jwt-Authorization")) &&
				assert.JSONEq(t, json, string(body))
		})).
		Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

	client := New("http://base", secretsClient, doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendRegister(ctx, "lpa-uid")

	assert.Nil(t, err)
}

func TestClientSendPerfect(t *testing.T) {
	json := `{"type":"PERFECT","changes":null}`

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
				assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjAwMDAwMDAwLTAwMDAtNDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsImlhdCI6OTQ2NzgyMjQ1fQ.V7MxjZw7-K8ehujYn4e0gef7s23r2UDlTbyzQtpTKvo", req.Header.Get("X-Jwt-Authorization")) &&
				assert.JSONEq(t, json, string(body))
		})).
		Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

	client := New("http://base", secretsClient, doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendPerfect(ctx, "lpa-uid")

	assert.Nil(t, err)
}

func TestClientSendCertificateProvider(t *testing.T) {
	uid, _ := actoruid.Parse("399ce2f7-f3bd-4feb-9207-699ff4d99cbf")

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
		Email:                     "b@example.com",
	}

	lpa := &Lpa{
		LpaUID: "lpa-uid",
		CertificateProvider: CertificateProvider{
			Channel: actor.ChannelOnline,
			Email:   "a@example.com",
		},
	}

	json := `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[{"key":"/certificateProvider/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"},{"key":"/certificateProvider/contactLanguagePreference","old":null,"new":"cy"},{"key":"/certificateProvider/address/line1","old":"","new":"line-1"},{"key":"/certificateProvider/address/line2","old":"","new":"line-2"},{"key":"/certificateProvider/address/line3","old":"","new":"line-3"},{"key":"/certificateProvider/address/town","old":"","new":"town"},{"key":"/certificateProvider/address/postcode","old":"","new":"postcode"},{"key":"/certificateProvider/address/country","old":"","new":"GB"},{"key":"/certificateProvider/email","old":"a@example.com","new":"b@example.com"}]}`

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
				assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjM5OWNlMmY3LWYzYmQtNGZlYi05MjA3LTY5OWZmNGQ5OWNiZiIsImlhdCI6OTQ2NzgyMjQ1fQ.-ZIBR-5fuznCkemcj-tbCro8VB9Li2Ieqd0sZJeooIY", req.Header.Get("X-Jwt-Authorization")) &&
				assert.JSONEq(t, json, string(body))
		})).
		Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

	client := New("http://base", secretsClient, doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendCertificateProvider(ctx, certificateProvider, lpa)

	assert.Nil(t, err)
}

func TestClientSendAttorney(t *testing.T) {
	uid1, _ := actoruid.Parse("f887edc1-bc69-413f-9e5d-b7bcc5fa1c72")
	uid2, _ := actoruid.Parse("846360af-304f-466b-bda1-df7bc47bbad6")

	testcases := map[string]struct {
		attorney *actor.AttorneyProvidedDetails
		donor    *Lpa
		json     string
	}{
		"attorney": {
			attorney: &actor.AttorneyProvidedDetails{
				UID:                       uid2,
				Mobile:                    "07777",
				Confirmed:                 time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				ContactLanguagePreference: localize.Cy,
				Email:                     "b@example.com",
			},
			donor: &Lpa{
				LpaUID: "lpa-uid",
				Attorneys: Attorneys{
					Attorneys: []Attorney{
						{UID: uid1}, {UID: uid2, Email: "a@example.com", Channel: actor.ChannelPaper},
					},
				},
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/attorneys/1/mobile","old":null,"new":"07777"},{"key":"/attorneys/1/contactLanguagePreference","old":null,"new":"cy"},{"key":"/attorneys/1/email","old":"a@example.com","new":"b@example.com"},{"key":"/attorneys/1/channel","old":"paper","new":"online"},{"key":"/attorneys/1/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
		},
		"replacement attorney": {
			attorney: &actor.AttorneyProvidedDetails{
				UID:                       uid2,
				IsReplacement:             true,
				Mobile:                    "07777",
				Confirmed:                 time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				ContactLanguagePreference: localize.Cy,
				Email:                     "b@example.com",
			},
			donor: &Lpa{
				LpaUID: "lpa-uid",
				Attorneys: Attorneys{
					Attorneys: []Attorney{
						{UID: uid1}, {UID: uid2},
					},
				},
				ReplacementAttorneys: Attorneys{
					Attorneys: []Attorney{
						{UID: uid1}, {UID: uid2, Email: "a@example.com", Channel: actor.ChannelPaper},
					},
				},
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/attorneys/3/mobile","old":null,"new":"07777"},{"key":"/attorneys/3/contactLanguagePreference","old":null,"new":"cy"},{"key":"/attorneys/3/email","old":"a@example.com","new":"b@example.com"},{"key":"/attorneys/3/channel","old":"paper","new":"online"},{"key":"/attorneys/3/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
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
				Email:                     "a@example.com",
			},
			donor: &Lpa{
				LpaUID: "lpa-uid",
				Attorneys: Attorneys{
					TrustCorporation: TrustCorporation{Channel: actor.ChannelPaper},
				},
			},
			json: `{"type":"TRUST_CORPORATION_SIGN","changes":[{"key":"/trustCorporations/0/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/0/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/0/email","old":"","new":"a@example.com"},{"key":"/trustCorporations/0/channel","old":"paper","new":"online"},{"key":"/trustCorporations/0/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/0/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/0/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/0/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"},{"key":"/trustCorporations/0/signatories/1/firstNames","old":null,"new":"Dave"},{"key":"/trustCorporations/0/signatories/1/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/0/signatories/1/professionalTitle","old":null,"new":"Assistant to the Director"},{"key":"/trustCorporations/0/signatories/1/signedAt","old":null,"new":"2000-01-02T03:04:05.000000007Z"}]}`,
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
			donor: &Lpa{
				LpaUID: "lpa-uid",
				Attorneys: Attorneys{
					TrustCorporation: TrustCorporation{Name: "a"},
				},
				ReplacementAttorneys: Attorneys{
					TrustCorporation: TrustCorporation{Channel: actor.ChannelPaper},
				},
			},
			json: `{"type":"TRUST_CORPORATION_SIGN","changes":[{"key":"/trustCorporations/1/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/1/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/1/channel","old":"paper","new":"online"},{"key":"/trustCorporations/1/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/1/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/1/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/1/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
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
			donor: &Lpa{
				LpaUID:    "lpa-uid",
				Attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
			},
			json: `{"type":"TRUST_CORPORATION_SIGN","changes":[{"key":"/trustCorporations/1/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/1/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/1/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/1/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/1/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/1/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
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
						assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjg0NjM2MGFmLTMwNGYtNDY2Yi1iZGExLWRmN2JjNDdiYmFkNiIsImlhdCI6OTQ2NzgyMjQ1fQ.InMOckjMWg_lTNayS-YvMnqcuTjEWalDVF_gCn2QeSg", req.Header.Get("X-Jwt-Authorization")) &&
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

func TestClientSendCertificateProviderOptOut(t *testing.T) {
	json := `{"type":"CERTIFICATE_PROVIDER_OPT_OUT","changes":null}`

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
				assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjAwMDAwMDAwLTAwMDAtNDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsImlhdCI6OTQ2NzgyMjQ1fQ.V7MxjZw7-K8ehujYn4e0gef7s23r2UDlTbyzQtpTKvo", req.Header.Get("X-Jwt-Authorization")) &&
				assert.JSONEq(t, json, string(body))
		})).
		Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

	client := New("http://base", secretsClient, doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendCertificateProviderOptOut(ctx, "lpa-uid", actoruid.Service)

	assert.Nil(t, err)
}
