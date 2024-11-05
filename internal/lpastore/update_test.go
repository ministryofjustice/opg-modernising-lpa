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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientSendRegister(t *testing.T) {
	json := `{"type":"REGISTER","changes":null}`

	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, "secret").
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

	client := New("http://base", secretsClient, "secret", doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendRegister(ctx, "lpa-uid")

	assert.Nil(t, err)
}

func TestClientSendPerfect(t *testing.T) {
	json := `{"type":"PERFECT","changes":null}`

	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, "secret").
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

	client := New("http://base", secretsClient, "secret", doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendPerfect(ctx, "lpa-uid")

	assert.Nil(t, err)
}

func TestClientSendDonorWithdrawLPA(t *testing.T) {
	json := `{"type":"DONOR_WITHDRAW_LPA","changes":null}`

	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, "secret").
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

	client := New("http://base", secretsClient, "secret", doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendDonorWithdrawLPA(ctx, "lpa-uid")

	assert.Nil(t, err)
}

func TestClientSendCertificateProvider(t *testing.T) {
	uid, _ := actoruid.Parse("399ce2f7-f3bd-4feb-9207-699ff4d99cbf")

	certificateProvider := &certificateproviderdata.Provided{
		UID: uid,
		HomeAddress: place.Address{
			Line1:      "line-1",
			Line2:      "line-2",
			Line3:      "line-3",
			TownOrCity: "town",
			Postcode:   "postcode",
			Country:    "GB",
		},
		SignedAt:                  time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
		ContactLanguagePreference: localize.Cy,
		Email:                     "b@example.com",
	}

	lpa := &lpadata.Lpa{
		LpaUID: "lpa-uid",
		CertificateProvider: lpadata.CertificateProvider{
			Channel: lpadata.ChannelOnline,
			Email:   "a@example.com",
		},
	}

	json := `{"type":"CERTIFICATE_PROVIDER_SIGN","changes":[{"key":"/certificateProvider/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"},{"key":"/certificateProvider/contactLanguagePreference","old":null,"new":"cy"},{"key":"/certificateProvider/address/line1","old":"","new":"line-1"},{"key":"/certificateProvider/address/line2","old":"","new":"line-2"},{"key":"/certificateProvider/address/line3","old":"","new":"line-3"},{"key":"/certificateProvider/address/town","old":"","new":"town"},{"key":"/certificateProvider/address/postcode","old":"","new":"postcode"},{"key":"/certificateProvider/address/country","old":"","new":"GB"},{"key":"/certificateProvider/email","old":"a@example.com","new":"b@example.com"}]}`

	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, "secret").
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

	client := New("http://base", secretsClient, "secret", doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendCertificateProvider(ctx, certificateProvider, lpa)

	assert.Nil(t, err)
}

func TestClientSendAttorney(t *testing.T) {
	uid1, _ := actoruid.Parse("f887edc1-bc69-413f-9e5d-b7bcc5fa1c72")
	uid2, _ := actoruid.Parse("846360af-304f-466b-bda1-df7bc47bbad6")

	testcases := map[string]struct {
		attorney *attorneydata.Provided
		donor    *lpadata.Lpa
		json     string
	}{
		"attorney": {
			attorney: &attorneydata.Provided{
				UID:                       uid2,
				Phone:                     "07777",
				SignedAt:                  time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				ContactLanguagePreference: localize.Cy,
				Email:                     "b@example.com",
			},
			donor: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{
						{UID: uid1}, {UID: uid2, Email: "a@example.com", Channel: lpadata.ChannelPaper},
					},
				},
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/attorneys/1/mobile","old":null,"new":"07777"},{"key":"/attorneys/1/contactLanguagePreference","old":null,"new":"cy"},{"key":"/attorneys/1/email","old":"a@example.com","new":"b@example.com"},{"key":"/attorneys/1/channel","old":"paper","new":"online"},{"key":"/attorneys/1/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
		},
		"replacement attorney": {
			attorney: &attorneydata.Provided{
				UID:                       uid2,
				IsReplacement:             true,
				Phone:                     "07777",
				SignedAt:                  time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				ContactLanguagePreference: localize.Cy,
				Email:                     "b@example.com",
			},
			donor: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{
						{UID: uid1}, {UID: uid2},
					},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{
						{UID: uid1}, {UID: uid2, Email: "a@example.com", Channel: lpadata.ChannelPaper},
					},
				},
			},
			json: `{"type":"ATTORNEY_SIGN","changes":[{"key":"/attorneys/3/mobile","old":null,"new":"07777"},{"key":"/attorneys/3/contactLanguagePreference","old":null,"new":"cy"},{"key":"/attorneys/3/email","old":"a@example.com","new":"b@example.com"},{"key":"/attorneys/3/channel","old":"paper","new":"online"},{"key":"/attorneys/3/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
		},
		"trust corporation": {
			attorney: &attorneydata.Provided{
				UID:                uid2,
				IsTrustCorporation: true,
				Phone:              "07777",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{
					FirstNames:        "John",
					LastName:          "Signer",
					ProfessionalTitle: "Director",
					SignedAt:          time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				}, {
					FirstNames:        "Dave",
					LastName:          "Signer",
					ProfessionalTitle: "Assistant to the Director",
					SignedAt:          time.Date(2000, time.January, 2, 3, 4, 5, 7, time.UTC),
				}},
				ContactLanguagePreference: localize.En,
				Email:                     "a@example.com",
			},
			donor: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Attorneys: lpadata.Attorneys{
					TrustCorporation: lpadata.TrustCorporation{Channel: lpadata.ChannelPaper},
				},
			},
			json: `{"type":"TRUST_CORPORATION_SIGN","changes":[{"key":"/trustCorporations/0/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/0/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/0/email","old":"","new":"a@example.com"},{"key":"/trustCorporations/0/channel","old":"paper","new":"online"},{"key":"/trustCorporations/0/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/0/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/0/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/0/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"},{"key":"/trustCorporations/0/signatories/1/firstNames","old":null,"new":"Dave"},{"key":"/trustCorporations/0/signatories/1/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/0/signatories/1/professionalTitle","old":null,"new":"Assistant to the Director"},{"key":"/trustCorporations/0/signatories/1/signedAt","old":null,"new":"2000-01-02T03:04:05.000000007Z"}]}`,
		},
		"replacement trust corporation": {
			attorney: &attorneydata.Provided{
				UID:                uid2,
				IsTrustCorporation: true,
				IsReplacement:      true,
				Phone:              "07777",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{
					FirstNames:        "John",
					LastName:          "Signer",
					ProfessionalTitle: "Director",
					SignedAt:          time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				}},
				ContactLanguagePreference: localize.En,
			},
			donor: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Attorneys: lpadata.Attorneys{
					TrustCorporation: lpadata.TrustCorporation{Name: "a"},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					TrustCorporation: lpadata.TrustCorporation{Channel: lpadata.ChannelPaper},
				},
			},
			json: `{"type":"TRUST_CORPORATION_SIGN","changes":[{"key":"/trustCorporations/1/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/1/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/1/channel","old":"paper","new":"online"},{"key":"/trustCorporations/1/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/1/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/1/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/1/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
		},
		"replacement trust corporation when also attorney trust corporation": {
			attorney: &attorneydata.Provided{
				UID:                uid2,
				IsTrustCorporation: true,
				IsReplacement:      true,
				Phone:              "07777",
				AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{
					FirstNames:        "John",
					LastName:          "Signer",
					ProfessionalTitle: "Director",
					SignedAt:          time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				}},
				ContactLanguagePreference: localize.En,
			},
			donor: &lpadata.Lpa{
				LpaUID:    "lpa-uid",
				Attorneys: lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{Name: "a"}},
			},
			json: `{"type":"TRUST_CORPORATION_SIGN","changes":[{"key":"/trustCorporations/1/mobile","old":null,"new":"07777"},{"key":"/trustCorporations/1/contactLanguagePreference","old":null,"new":"en"},{"key":"/trustCorporations/1/signatories/0/firstNames","old":null,"new":"John"},{"key":"/trustCorporations/1/signatories/0/lastName","old":null,"new":"Signer"},{"key":"/trustCorporations/1/signatories/0/professionalTitle","old":null,"new":"Director"},{"key":"/trustCorporations/1/signatories/0/signedAt","old":null,"new":"2000-01-02T03:04:05.000000006Z"}]}`,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(ctx, "secret").
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

			client := New("http://base", secretsClient, "secret", doer)
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
		Secret(ctx, "secret").
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

	client := New("http://base", secretsClient, "secret", doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendCertificateProviderOptOut(ctx, "lpa-uid", actoruid.Service)

	assert.Nil(t, err)
}

func TestClientSendDonorConfirmIdentity(t *testing.T) {
	uid, _ := actoruid.Parse("15e5df3d-a053-4537-8066-2f4dd8d1dba8")

	json := `{"type":"DONOR_CONFIRM_IDENTITY","changes": [
{"key": "/donor/identityCheck/checkedAt", "new": "2024-01-02T12:13:14.000000006Z", "old": null},
{"key": "/donor/identityCheck/type", "new": "one-login", "old": null}
]}`

	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, "secret").
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
				assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjE1ZTVkZjNkLWEwNTMtNDUzNy04MDY2LTJmNGRkOGQxZGJhOCIsImlhdCI6OTQ2NzgyMjQ1fQ.6nsN_9PRaB_jXS_sni2-JBNlnWUdHK-xEgTLda-8PD8", req.Header.Get("X-Jwt-Authorization")) &&
				assert.JSONEq(t, json, string(body))
		})).
		Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

	client := New("http://base", secretsClient, "secret", doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendDonorConfirmIdentity(ctx, &donordata.Provided{
		LpaUID: "lpa-uid",
		Donor:  donordata.Donor{UID: uid},
		IdentityUserData: identity.UserData{
			CheckedAt: time.Date(2024, time.January, 2, 12, 13, 14, 6, time.UTC),
		},
	})

	assert.Nil(t, err)
}

func TestClientSendCertificateProviderConfirmIdentity(t *testing.T) {
	uid, _ := actoruid.Parse("15e5df3d-a053-4537-8066-2f4dd8d1dba8")

	json := `{"type":"CERTIFICATE_PROVIDER_CONFIRM_IDENTITY","changes": [
{"key": "/certificateProvider/identityCheck/checkedAt", "new": "2024-01-02T12:13:14.000000006Z", "old": null},
{"key": "/certificateProvider/identityCheck/type", "new": "one-login", "old": null}
]}`

	ctx := context.Background()

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, "secret").
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
				assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOjE1ZTVkZjNkLWEwNTMtNDUzNy04MDY2LTJmNGRkOGQxZGJhOCIsImlhdCI6OTQ2NzgyMjQ1fQ.6nsN_9PRaB_jXS_sni2-JBNlnWUdHK-xEgTLda-8PD8", req.Header.Get("X-Jwt-Authorization")) &&
				assert.JSONEq(t, json, string(body))
		})).
		Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

	client := New("http://base", secretsClient, "secret", doer)
	client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }
	err := client.SendCertificateProviderConfirmIdentity(ctx, "lpa-uid", &certificateproviderdata.Provided{
		UID: uid,
		IdentityUserData: identity.UserData{
			CheckedAt: time.Date(2024, time.January, 2, 12, 13, 14, 6, time.UTC),
		},
	})

	assert.Nil(t, err)
}

func TestClientSendAttorneyOptOut(t *testing.T) {
	testcases := map[actor.Type]string{
		actor.TypeAttorney:                    "ATTORNEY_OPT_OUT",
		actor.TypeReplacementAttorney:         "ATTORNEY_OPT_OUT",
		actor.TypeTrustCorporation:            "TRUST_CORPORATION_OPT_OUT",
		actor.TypeReplacementTrustCorporation: "TRUST_CORPORATION_OPT_OUT",
	}

	for actorType, updateType := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			json := `{"type":"` + updateType + `","changes":null}`

			ctx := context.Background()

			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				Secret(ctx, "secret").
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
						assert.Equal(t, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ1cm46b3BnOnBvYXM6bWFrZXJlZ2lzdGVyOnVzZXJzOmRjNDg3ZWJiLWIzOWQtNDVlZC1iYjZhLTdmOTUwZmQzNTVjOSIsImlhdCI6OTQ2NzgyMjQ1fQ.MIHlxYV520Wpx-pP2XVvYdbUGFh3CkmCjFR99XBOX9k", req.Header.Get("X-Jwt-Authorization")) &&
						assert.JSONEq(t, json, string(body))
				})).
				Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

			client := New("http://base", secretsClient, "secret", doer)
			client.now = func() time.Time { return time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC) }

			uid, _ := actoruid.Parse("dc487ebb-b39d-45ed-bb6a-7f950fd355c9")
			err := client.SendAttorneyOptOut(ctx, "lpa-uid", uid, actorType)

			assert.Nil(t, err)
		})
	}
}
