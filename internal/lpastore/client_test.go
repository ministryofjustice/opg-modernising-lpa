package lpastore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
							"status": matchers.Regex("active", "active|replacement"),
						}, 1),
						"certificateProvider": matchers.Like(map[string]any{
							"uid":        matchers.UUID(),
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
					// Header("Content-Type", matchers.String("application/json")).
					// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date;x-jwt-authorization, Signature=3fe9cd4a65c746d7531c3f3d9ae4479eec81886f5b6863680fcf7cf804aa4d6b", "AWS4-HMAC-SHA256 .*")).
					// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
					// Header("X-Jwt-Authorization", matchers.Regex("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvcGcucG9hcy5tYWtlcmVnaXN0ZXIiLCJzdWIiOiJ0b2RvIiwiaWF0Ijo5NDY3NzEyMDB9.teh381oIhucqUD3EhBTaaBTLFI1O2FOWGe-44Ftk0LY", "Bearer .+")).
					JSONBody(matchers.Map{
						"lpaType": matchers.Regex("personal-welfare", "personal-welfare|property-and-affairs"),
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
							"otherNamesKnownBy": matchers.String("JJ"),
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
							"status": matchers.Regex("active", "active|replacement"),
						}, 1),
						"certificateProvider": matchers.Like(map[string]any{
							"uid":        matchers.UUID(),
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
					"uid": matchers.Regex("M-0000-1111-2222", "M(-\\d{4}){3}"),
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

			assert.Equal(t, &actor.DonorProvidedDetails{LpaUID: "M-0000-1111-2222"}, donor)
			return nil
		})

		assert.Nil(t, err)
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
