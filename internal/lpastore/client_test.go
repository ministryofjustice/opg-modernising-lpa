package lpastore

import (
	"bytes"
	"context"
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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestResponseError(t *testing.T) {
	err := responseError{name: "name", body: 5}
	assert.Equal(t, "name", err.Error())
	assert.Equal(t, "name", err.Title())
	assert.Equal(t, 5, err.Data())
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
					// Header("Content-Type", matchers.String("application/json")).
					// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
					// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"lpaType":                       matchers.Regex("personal-welfare", "personal-welfare|property-and-affairs"),
						"channel":                       matchers.String("online"),
						"lifeSustainingTreatmentOption": matchers.Regex("option-a", "option-a|option-b"),
						"donor": matchers.Like(map[string]any{
							"uid":         matchers.UUID(),
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
							"contactLanguagePreference": matchers.String("en"),
						}),
						"attorneys": matchers.EachLike(map[string]any{
							"uid":         matchers.UUID(),
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
							"channel": matchers.Regex("online", "online|paper"),
							"status":  matchers.Regex("active", "active|replacement"),
						}, 1),
						"trustCorporations": matchers.EachLike(map[string]any{
							"uid":           matchers.UUID(),
							"name":          matchers.String("Trust us Corp."),
							"companyNumber": matchers.String("66654321"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("tc-line-1"),
								"line2":    matchers.String("tc-line-2"),
								"line3":    matchers.String("tc-line-3"),
								"town":     matchers.String("tc-town"),
								"postcode": matchers.String("TC1 1FF"),
								"country":  matchers.String("GB"),
							}),
							"channel": matchers.Regex("paper", "online|paper"),
							"status":  matchers.Regex("active", "active|replacement"),
						}, 1),
						"certificateProvider": matchers.Like(map[string]any{
							"uid":        matchers.UUID(),
							"firstNames": matchers.String("Charles"),
							"lastName":   matchers.String("Certificate"),
							"email":      matchers.String("charles@example.com"),
							"phone":      matchers.String("0700009000"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("cp-line-1"),
								"line2":    matchers.String("cp-line-2"),
								"line3":    matchers.String("cp-line-3"),
								"town":     matchers.String("cp-town"),
								"postcode": matchers.String("CP1 1FF"),
								"country":  matchers.String("GB"),
							}),
							"channel": matchers.Regex("online", "online|paper"),
						}),
						"restrictionsAndConditions": matchers.String("hmm"),
						"signedAt":                  matchers.Regex("2000-01-02T12:13:14.00000Z", `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?Z`),
						"howAttorneysMakeDecisions": matchers.Regex("jointly", "jointly|jointly-and-severally|jointly-for-some-severally-for-others"),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				// b.Header("Content-Type", matchers.String("application/json"))
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

			err := client.SendLpa(context.Background(), &donordata.Provided{
				LpaUID:                        "M-0000-1111-2222",
				Type:                          lpadata.LpaTypePersonalWelfare,
				LifeSustainingTreatmentOption: lpadata.LifeSustainingTreatmentOptionA,
				Donor: donordata.Donor{
					UID:                       actoruid.New(),
					FirstNames:                "John Johnson",
					LastName:                  "Smith",
					DateOfBirth:               date.New("2000", "1", "2"),
					Email:                     "john@example.com",
					Address:                   address,
					ContactLanguagePreference: localize.En,
				},
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{{
						UID:         actoruid.New(),
						FirstNames:  "Alice",
						LastName:    "Attorney",
						DateOfBirth: date.New("1998", "1", "2"),
						Email:       "alice@example.com",
						Address:     address,
					}},
					TrustCorporation: donordata.TrustCorporation{
						UID:           actoruid.New(),
						Name:          "Trust us Corp.",
						CompanyNumber: "66654321",
						Address:       address,
					},
				},
				AttorneyDecisions: donordata.AttorneyDecisions{How: lpadata.Jointly},
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{{
						UID:         actoruid.New(),
						FirstNames:  "Richard",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "11", "12"),
						Email:       "richard@example.com",
						Address:     address,
					}},
				},
				CertificateProvider: donordata.CertificateProvider{
					UID:        actoruid.New(),
					FirstNames: "Charles",
					LastName:   "Certificate",
					Email:      "charles@example.com",
					Mobile:     "0700009000",
					Address:    address,
					CarryOutBy: lpadata.ChannelOnline,
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
					// Header("Content-Type", matchers.String("application/json")).
					// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
					// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"lpaType": matchers.Regex("personal-welfare", "personal-welfare|property-and-affairs"),
						"channel": matchers.String("online"),
						"donor": matchers.Like(map[string]any{
							"uid":         matchers.UUID(),
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
							"otherNamesKnownBy":         matchers.String("JJ"),
							"contactLanguagePreference": matchers.String("cy"),
						}),
						"attorneys": matchers.EachLike(map[string]any{
							"uid":         matchers.UUID(),
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
							"status":  matchers.Regex("active", "active|replacement"),
							"channel": matchers.Regex("online", "online|post"),
						}, 1),
						"certificateProvider": matchers.Like(map[string]any{
							"uid":        matchers.UUID(),
							"firstNames": matchers.String("Charles"),
							"lastName":   matchers.String("Certificate"),
							"email":      matchers.String("charles@example.com"),
							"phone":      matchers.String("0700009000"),
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
							"uid":        matchers.UUID(),
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
				// b.Header("Content-Type", matchers.String("application/json"))
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

			err := client.SendLpa(context.Background(), &donordata.Provided{
				LpaUID: "M-0000-1111-2222",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: donordata.Donor{
					UID:                       actoruid.New(),
					FirstNames:                "John Johnson",
					LastName:                  "Smith",
					DateOfBirth:               date.New("2000", "1", "2"),
					Email:                     "john@example.com",
					Address:                   address,
					OtherNames:                "JJ",
					ContactLanguagePreference: localize.Cy,
				},
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{{
						UID:         actoruid.New(),
						FirstNames:  "Alice",
						LastName:    "Attorney",
						DateOfBirth: date.New("1998", "1", "2"),
						Email:       "alice@example.com",
						Address:     address,
					}},
				},
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{{
						UID:         actoruid.New(),
						FirstNames:  "Richard",
						LastName:    "Attorney",
						DateOfBirth: date.New("1999", "11", "12"),
						Email:       "richard@example.com",
						Address:     address,
					}},
				},
				CertificateProvider: donordata.CertificateProvider{
					UID:        actoruid.New(),
					FirstNames: "Charles",
					LastName:   "Certificate",
					Email:      "charles@example.com",
					Mobile:     "0700009000",
					Address:    address,
					CarryOutBy: lpadata.ChannelOnline,
				},
				PeopleToNotify: donordata.PeopleToNotify{{
					UID:        actoruid.New(),
					FirstNames: "Peter",
					LastName:   "Person",
					Address:    address,
				}},
				Restrictions: "hmm",
				SignedAt:     time.Date(2000, time.January, 2, 12, 13, 14, 0, time.UTC),
			})

			assert.Nil(t, err)
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
					// Header("Content-Type", matchers.String("application/json")).
					// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
					// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
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
						}, {
							"key": matchers.Like("/attorneys/0/email"),
							"old": matchers.Like(nil),
							"new": matchers.Like("a@example.com"),
						}, {
							"key": matchers.Like("/attorneys/0/channel"),
							"old": matchers.Like("paper"),
							"new": matchers.Like("online"),
						}}),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				// b.Header("Content-Type", matchers.String("application/json"))
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
				&lpadata.Lpa{
					LpaUID: "M-0000-1111-2222",
					Attorneys: lpadata.Attorneys{
						Attorneys: []lpadata.Attorney{{UID: uid}},
					},
				},
				&attorneydata.Provided{
					UID:                       uid,
					Phone:                     "07777777",
					SignedAt:                  time.Date(2020, time.January, 1, 12, 13, 14, 0, time.UTC),
					ContactLanguagePreference: localize.Cy,
				})
			assert.Nil(t, err)
			return nil
		}))
	})

	t.Run("SendAttorney when trust corporation", func(t *testing.T) {
		uid := actoruid.New()

		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 exists").
			UponReceiving("A request to send the trust corporation data").
			WithRequest(http.MethodPost, "/lpas/M-0000-1111-2222/updates", func(b *consumer.V2RequestBuilder) {
				b.
					// Header("Content-Type", matchers.String("application/json")).
					// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
					// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"type": matchers.Like("TRUST_CORPORATION_SIGN"),
						"changes": matchers.Like([]map[string]any{{
							"key": matchers.Like("/trustCorporations/0/mobile"),
							"old": matchers.Like(nil),
							"new": matchers.Like("07777777"),
						}, {
							"key": matchers.Like("/trustCorporations/0/contactLanguagePreference"),
							"old": matchers.Like(nil),
							"new": matchers.Like("cy"),
						}, {
							"key": matchers.Like("/trustCorporations/0/email"),
							"old": matchers.Like(""),
							"new": matchers.Like("a@example.com"),
						}, {
							"key": matchers.Like("/trustCorporations/0/channel"),
							"old": matchers.Like("paper"),
							"new": matchers.Like("online"),
						}, {
							"key": matchers.Like("/trustCorporations/0/signatories/0/firstNames"),
							"old": matchers.Like(nil),
							"new": matchers.Like("John"),
						}, {
							"key": matchers.Like("/trustCorporations/0/signatories/0/lastName"),
							"old": matchers.Like(nil),
							"new": matchers.Like("Smith"),
						}, {
							"key": matchers.Like("/trustCorporations/0/signatories/0/professionalTitle"),
							"old": matchers.Like(nil),
							"new": matchers.Like("Director"),
						}, {
							"key": matchers.Like("/trustCorporations/0/signatories/0/signedAt"),
							"old": matchers.Like(nil),
							"new": matchers.Like("2020-01-01T12:13:14Z"),
						}}),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				// b.Header("Content-Type", matchers.String("application/json"))
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
				&lpadata.Lpa{
					LpaUID: "M-0000-1111-2222",
					Attorneys: lpadata.Attorneys{
						TrustCorporation: lpadata.TrustCorporation{
							UID:           uid,
							Name:          "Trust us Corp.",
							CompanyNumber: "66654321",
							Channel:       lpadata.ChannelPaper,
						},
					},
				},
				&attorneydata.Provided{
					UID:                       uid,
					Phone:                     "07777777",
					ContactLanguagePreference: localize.Cy,
					AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{
						{
							FirstNames:        "John",
							LastName:          "Smith",
							ProfessionalTitle: "Director",
							SignedAt:          time.Date(2020, time.January, 1, 12, 13, 14, 0, time.UTC),
						},
					},
					IsTrustCorporation: true,
					Email:              "b@example.com",
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
					// Header("Content-Type", matchers.String("application/json")).
					// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
					// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
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
							"key": matchers.Like("/certificateProvider/email"),
							"old": matchers.Like(""),
							"new": matchers.Like("a@example.com"),
						}, {
							"key": matchers.Like("/certificateProvider/channel"),
							"old": matchers.Like("paper"),
							"new": matchers.Like("online"),
						}}),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				// b.Header("Content-Type", matchers.String("application/json"))
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

			err := client.SendCertificateProvider(context.Background(),
				&certificateproviderdata.Provided{
					SignedAt:                  time.Date(2020, time.January, 1, 12, 13, 14, 0, time.UTC),
					ContactLanguagePreference: localize.Cy,
					Email:                     "a@example.com",
				}, &lpadata.Lpa{CertificateProvider: lpadata.CertificateProvider{Channel: lpadata.ChannelPaper}, LpaUID: "M-0000-1111-2222"})
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
					// Header("Content-Type", matchers.String("application/json")).
					// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
					// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
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
							"old": matchers.Like("71 South Western Terrace"),
							"new": matchers.Like("123 Fake Street"),
						}, {
							"key": matchers.Like("/certificateProvider/address/town"),
							"old": matchers.Like("Milton"),
							"new": matchers.Like("Faketon"),
						}, {
							"key": matchers.Like("/certificateProvider/address/country"),
							"old": matchers.Like("AU"),
							"new": matchers.Like("GB"),
						}, {
							"key": matchers.Like("/certificateProvider/email"),
							"old": matchers.Like(""),
							"new": matchers.Like("a@example.com"),
						}, {
							"key": matchers.Like("/certificateProvider/channel"),
							"old": matchers.Like("paper"),
							"new": matchers.Like("online"),
						}}),
					})
			}).
			WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
				// b.Header("Content-Type", matchers.String("application/json"))
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

			return client.SendCertificateProvider(context.Background(),
				&certificateproviderdata.Provided{
					SignedAt:                  time.Date(2020, time.January, 1, 12, 13, 14, 0, time.UTC),
					ContactLanguagePreference: localize.Cy,
					HomeAddress: place.Address{
						Line1:      "123 Fake Street",
						TownOrCity: "Faketon",
						Country:    "GB",
					},
					Email: "a@example.com",
				}, &lpadata.Lpa{
					CertificateProvider: lpadata.CertificateProvider{
						Channel: lpadata.ChannelPaper,
						Address: place.Address{
							Line1:      "71 South Western Terrace",
							TownOrCity: "Milton",
							Country:    "AU",
						},
					},
					LpaUID: "M-0000-1111-2222",
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
					// Header("Content-Type", matchers.String("application/json")).
					// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
					// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
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
				// b.Header("Content-Type", matchers.String("application/json"))
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

	t.Run("Lpa", func(t *testing.T) {
		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 exists").
			UponReceiving("A request to get the lpa").
			WithRequest(http.MethodGet, "/lpas/M-0000-1111-2222", func(b *consumer.V2RequestBuilder) {
				// b.
				// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
				// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
				// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+"))
			}).
			WillRespondWith(http.StatusOK, func(b *consumer.V2ResponseBuilder) {
				// b.Header("Content-Type", matchers.String("application/json"))
				b.JSONBody(matchers.Map{
					"uid":     matchers.Regex("M-0000-1111-2222", "M(-[A-Z0-9]{4}){3}"),
					"status":  matchers.Regex("in-progress", "in-progress|cannot-register|statutory-waiting-period|registered"),
					"lpaType": matchers.String("personal-welfare"),
					"channel": matchers.String("online"),
					"donor": matchers.Like(map[string]any{
						"firstNames":  matchers.String("Homer"),
						"lastName":    matchers.String("Zoller"),
						"dateOfBirth": matchers.String("1960-04-06"),
						"address": matchers.Like(map[string]any{
							"line1":    matchers.String("79 Bury Rd"),
							"town":     matchers.String("Hampton Lovett"),
							"postcode": matchers.String("WR9 2PF"),
							"country":  matchers.String("GB"),
						}),
					}),
					"attorneys": matchers.EachLike(map[string]any{
						"firstNames":  matchers.String("Jake"),
						"lastName":    matchers.String("Valler"),
						"dateOfBirth": matchers.String("2001-01-17"),
						"address": matchers.Like(map[string]any{
							"line1":   matchers.String("71 South Western Terrace"),
							"town":    matchers.String("Milton"),
							"country": matchers.String("AU"),
						}),
						"status":  matchers.String("active"),
						"channel": matchers.String("paper"),
					}, 1),
					"certificateProvider": matchers.Like(map[string]any{
						"firstNames": matchers.String("Some"),
						"lastName":   matchers.String("Provider"),
						"email":      matchers.String("some@example.com"),
						"phone":      matchers.String("0700009000"),
						"address": matchers.Like(map[string]any{
							"line1":   matchers.String("71 South Western Terrace"),
							"town":    matchers.String("Milton"),
							"country": matchers.String("AU"),
						}),
						"channel": matchers.String("online"),
					}),
					"lifeSustainingTreatmentOption": matchers.String("option-a"),
					"signedAt":                      matchers.String("2000-01-02T12:13:14Z"),
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

			donor, err := client.Lpa(context.Background(), "M-0000-1111-2222")
			if err != nil {
				return err
			}

			assert.Equal(t, &lpadata.Lpa{
				LpaUID: "M-0000-1111-2222",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					FirstNames:  "Homer",
					LastName:    "Zoller",
					DateOfBirth: date.New("1960", "04", "06"),
					Address: place.Address{
						Line1:      "79 Bury Rd",
						TownOrCity: "Hampton Lovett",
						Postcode:   "WR9 2PF",
						Country:    "GB",
					},
					Channel: lpadata.ChannelOnline,
				},
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						FirstNames:  "Jake",
						LastName:    "Valler",
						DateOfBirth: date.New("2001", "01", "17"),
						Address: place.Address{
							Line1:      "71 South Western Terrace",
							TownOrCity: "Milton",
							Country:    "AU",
						},
						Channel: lpadata.ChannelPaper,
					}},
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "Some",
					LastName:   "Provider",
					Email:      "some@example.com",
					Phone:      "0700009000",
					Address: place.Address{
						Line1:      "71 South Western Terrace",
						TownOrCity: "Milton",
						Country:    "AU",
					},
					Channel: lpadata.ChannelOnline,
				},
				LifeSustainingTreatmentOption: lpadata.LifeSustainingTreatmentOptionA,
				SignedAt:                      time.Date(2000, time.January, 2, 12, 13, 14, 0, time.UTC),
			}, donor)
			return nil
		})

		assert.Nil(t, err)
	})

	t.Run("Lpas", func(t *testing.T) {
		mockProvider.
			AddInteraction().
			Given("An LPA with UID M-0000-1111-2222 exists").
			UponReceiving("A request to get multiple lpas").
			WithRequest(http.MethodPost, "/lpas", func(b *consumer.V2RequestBuilder) {
				b.JSONBody(matchers.Map{
					"uids": matchers.EachLike("M-0000-1111-2222", 1),
				})
				// b.
				// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
				// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
				// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+"))
			}).
			WillRespondWith(http.StatusOK, func(b *consumer.V2ResponseBuilder) {
				// b.Header("Content-Type", matchers.String("application/json"))
				b.JSONBody(matchers.Map{
					"lpas": matchers.EachLike(matchers.Map{
						"uid":     matchers.Regex("M-0000-1111-2222", "M(-[A-Z0-9]{4}){3}"),
						"status":  matchers.Regex("in-progress", "in-progress|cannot-register|statutory-waiting-period|registered"),
						"lpaType": matchers.String("personal-welfare"),
						"channel": matchers.String("online"),
						"donor": matchers.Like(map[string]any{
							"firstNames":  matchers.String("Homer"),
							"lastName":    matchers.String("Zoller"),
							"dateOfBirth": matchers.String("1960-04-06"),
							"address": matchers.Like(map[string]any{
								"line1":    matchers.String("79 Bury Rd"),
								"town":     matchers.String("Hampton Lovett"),
								"postcode": matchers.String("WR9 2PF"),
								"country":  matchers.String("GB"),
							}),
						}),
						"attorneys": matchers.EachLike(map[string]any{
							"firstNames":  matchers.String("Jake"),
							"lastName":    matchers.String("Valler"),
							"dateOfBirth": matchers.String("2001-01-17"),
							"address": matchers.Like(map[string]any{
								"line1":   matchers.String("71 South Western Terrace"),
								"town":    matchers.String("Milton"),
								"country": matchers.String("AU"),
							}),
							"status":  matchers.String("active"),
							"channel": matchers.String("online"),
						}, 1),
						"certificateProvider": matchers.Like(map[string]any{
							"firstNames": matchers.String("Some"),
							"lastName":   matchers.String("Provider"),
							"email":      matchers.String("some@example.com"),
							"phone":      matchers.String("0700009000"),
							"address": matchers.Like(map[string]any{
								"line1":   matchers.String("71 South Western Terrace"),
								"town":    matchers.String("Milton"),
								"country": matchers.String("AU"),
							}),
							"channel": matchers.String("online"),
						}),
						"lifeSustainingTreatmentOption": matchers.String("option-a"),
						"signedAt":                      matchers.String("2000-01-02T12:13:14Z"),
					}, 1),
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

			lpas, err := client.Lpas(context.Background(), []string{"M-0000-1111-2222"})
			if err != nil {
				return err
			}

			assert.Equal(t, []*lpadata.Lpa{{
				LpaUID: "M-0000-1111-2222",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					FirstNames:  "Homer",
					LastName:    "Zoller",
					DateOfBirth: date.New("1960", "04", "06"),
					Address: place.Address{
						Line1:      "79 Bury Rd",
						TownOrCity: "Hampton Lovett",
						Postcode:   "WR9 2PF",
						Country:    "GB",
					},
					Channel: lpadata.ChannelOnline,
				},
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						FirstNames:  "Jake",
						LastName:    "Valler",
						DateOfBirth: date.New("2001", "01", "17"),
						Address: place.Address{
							Line1:      "71 South Western Terrace",
							TownOrCity: "Milton",
							Country:    "AU",
						},
						Channel: lpadata.ChannelOnline,
					}},
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "Some",
					LastName:   "Provider",
					Email:      "some@example.com",
					Phone:      "0700009000",
					Address: place.Address{
						Line1:      "71 South Western Terrace",
						TownOrCity: "Milton",
						Country:    "AU",
					},
					Channel: lpadata.ChannelOnline,
				},
				LifeSustainingTreatmentOption: lpadata.LifeSustainingTreatmentOptionA,
				SignedAt:                      time.Date(2000, time.January, 2, 12, 13, 14, 0, time.UTC),
			}}, lpas)
			return nil
		})

		assert.Nil(t, err)
	})
}

func TestClientDo(t *testing.T) {
	expectedResponse := &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("hey"))}

	ctx := context.Background()
	req, _ := http.NewRequest(http.MethodGet, "", nil)

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(mock.Anything, mock.Anything).
		Return("secret", nil)

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(expectedResponse, expectedError)

	client := New("http://base", secretsClient, "secret", doer)
	resp, err := client.do(ctx, actoruid.New(), req)

	assert.Equal(t, expectedError, err)
	assert.Equal(t, expectedResponse, resp)
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

	client := New(server.URL, nil, "secret", server.Client())

	err := client.CheckHealth(context.Background())

	assert.Equal(t, http.MethodGet, requestMethod)
	assert.Equal(t, "/health-check", endpointCalled)
	assert.Nil(t, err)
}

func TestCheckHealthOnNewRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	client := New(server.URL+"`invalid-url-format", nil, "secret", server.Client())
	err := client.CheckHealth(context.Background())
	assert.NotNil(t, err)
}

func TestCheckHealthOnDoRequestError(t *testing.T) {
	httpClient := newMockDoer(t)
	httpClient.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	client := New("/", nil, "secret", httpClient)
	err := client.CheckHealth(context.Background())
	assert.Equal(t, expectedError, err)
}

func TestCheckHealthWhenNotOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	client := New(server.URL, nil, "secret", server.Client())
	err := client.CheckHealth(context.Background())
	assert.NotNil(t, err)
}
