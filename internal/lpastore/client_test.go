package lpastore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/pact-foundation/pact-go/dsl"
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

func TestClientSendLpa(t *testing.T) {
	ctx := context.Background()

	donor := &actor.DonorProvidedDetails{
		LpaUID: "M-0000-1111-2222",
		Donor: actor.Donor{
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
			Attorneys: []actor.Attorney{{
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
		ReplacementAttorneys: actor.Attorneys{
			Attorneys: []actor.Attorney{{
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
	}

	expectedBody := `{"donor":{"firstNames":"John Johnson","surname":"Smith","dateOfBirth":"2000-01-02","email":"john@example.com","address":{"line1":"line-1","line2":"line-2","line3":"line-3","town":"town","postcode":"F1 1FF","country":"GB"},"otherNamesKnownBy":"JJ"},"attorneys":[{"firstNames":"Adam","surname":"Attorney","dateOfBirth":"1999-01-02","email":"adam@example.com","address":{"line1":"a-line-1","line2":"a-line-2","line3":"a-line-3","town":"a-town","postcode":"A1 1FF","country":"GB"},"status":"active"},{"firstNames":"Alice","surname":"Attorney","dateOfBirth":"1998-01-02","email":"alice@example.com","address":{"line1":"aa-line-1","line2":"aa-line-2","line3":"aa-line-3","town":"aa-town","postcode":"A1 1AF","country":"GB"},"status":"active"},{"firstNames":"Richard","surname":"Attorney","dateOfBirth":"1999-11-12","email":"richard@example.com","address":{"line1":"r-line-1","line2":"r-line-2","line3":"r-line-3","town":"r-town","postcode":"R1 1FF","country":"GB"},"status":"replacement"},{"firstNames":"Rachel","surname":"Attorney","dateOfBirth":"1998-11-12","email":"rachel@example.com","address":{"line1":"rr-line-1","line2":"rr-line-2","line3":"rr-line-3","town":"rr-town","postcode":"R1 1RF","country":"GB"},"status":"replacement"}]}`

	var body []byte
	doer := newMockDoer(t)
	doer.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			if body == nil {
				body, _ = io.ReadAll(req.Body)
			}

			return assert.Equal(t, ctx, req.Context()) &&
				assert.Equal(t, http.MethodPut, req.Method) &&
				assert.Equal(t, "http://base/lpas/M-0000-1111-2222", req.URL.String()) &&
				assert.JSONEq(t, expectedBody, string(body))
		})).
		Return(&http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil)

	client := New("http://base", doer)
	err := client.SendLpa(ctx, donor)

	assert.Nil(t, err)
}

func TestClientSendLpaWhenNewRequestError(t *testing.T) {
	client := New("http://base", nil)
	err := client.SendLpa(nil, &actor.DonorProvidedDetails{})

	assert.NotNil(t, err)
}

func TestClientSendLpaWhenDoerError(t *testing.T) {
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.
		On("Do", mock.Anything).
		Return(nil, expectedError)

	client := New("http://base", doer)
	err := client.SendLpa(ctx, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestClientSendLpaWhenStatusCodeIsNotCreated(t *testing.T) {
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.
		On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("hey"))}, nil)

	client := New("http://base", doer)
	err := client.SendLpa(ctx, &actor.DonorProvidedDetails{})

	assert.Equal(t, responseError{name: "expected 201 response but got 400", body: "hey"}, err)
}

func TestClientServiceContract(t *testing.T) {
	pact := &dsl.Pact{
		Consumer:          "modernising-lpa",
		Provider:          "data-lpa-store",
		Host:              "localhost",
		PactFileWriteMode: "merge",
		LogDir:            "../../logs",
		PactDir:           "../../pacts",
	}
	defer pact.Teardown()

	uid := strings.ToUpper("M-" + random.String(4) + "-" + random.String(4) + "-" + random.String(4))

	pact.
		AddInteraction().
		Given("The lpa store is available").
		UponReceiving("A request for a new case").
		WithRequest(dsl.Request{
			Method: http.MethodPut,
			Path:   dsl.String("/lpas/" + uid),
			Headers: dsl.MapMatcher{
				"Content-Type":  dsl.String("application/json"),
				"Authorization": dsl.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date, Signature=98fe2cb1c34c6de900d291351991ba8aa948ca05b7bff969d781edce9b75ee20", "AWS4-HMAC-SHA256 Credential=.*\\/.*\\/.*\\/execute-api\\/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date, Signature=.*"),
				"X-Amz-Date":    dsl.String("20000102T000000Z"),
			},
			Body: dsl.Like(map[string]any{
				"donor": dsl.Like(map[string]any{
					"firstNames":  dsl.String("John Johnson"),
					"surname":     dsl.String("Smith"),
					"dateOfBirth": dsl.Regex("2000-01-02", "\\d{4}-\\d{2}-\\d{2}"),
					"email":       dsl.String("john@example.com"),
					"address": dsl.Like(map[string]any{
						"line1":    dsl.String("line-1"),
						"line2":    dsl.String("line-2"),
						"line3":    dsl.String("line-3"),
						"town":     dsl.String("town"),
						"postcode": dsl.String("F1 1FF"),
						"country":  dsl.String("GB"),
					}),
					"otherNamesKnownBy": dsl.String("JJ"),
				}),
				"attorneys": dsl.EachLike(map[string]any{
					"firstNames":  dsl.String("Adam"),
					"surname":     dsl.String("Attorney"),
					"dateOfBirth": dsl.Regex("1999-01-02", "\\d{4}-\\d{2}-\\d{2}"),
					"email":       dsl.String("adam@example.com"),
					"address": dsl.Like(map[string]any{
						"line1":    dsl.String("a-line-1"),
						"line2":    dsl.String("a-line-2"),
						"line3":    dsl.String("a-line-3"),
						"town":     dsl.String("a-town"),
						"postcode": dsl.String("A1 1FF"),
						"country":  dsl.String("GB"),
					}),
					"status": dsl.Regex("active", "active|replacement"),
				}, 1),
			}),
		}).
		WillRespondWith(dsl.Response{
			Status:  http.StatusCreated,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body:    `{}`,
		})

	pact.Verify(func() error {
		baseURL := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }

		cfg := aws.Config{
			Region:      "eu-west-1",
			Credentials: &mockCredentialsProvider{},
		}

		client := &Client{
			baseURL: baseURL,
			doer:    lambda.New(cfg, v4.NewSigner(), http.DefaultClient, now),
		}

		address := place.Address{
			Line1:      "line-1",
			Line2:      "line-2",
			Line3:      "line-3",
			TownOrCity: "town",
			Postcode:   "F1 1FF",
			Country:    "GB",
		}

		err := client.SendLpa(context.Background(), &actor.DonorProvidedDetails{
			LpaUID: uid,
			Donor: actor.Donor{
				FirstNames:  "John Johnson",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "1", "2"),
				Email:       "john@example.com",
				Address:     address,
				OtherNames:  "JJ",
			},
			Attorneys: actor.Attorneys{
				Attorneys: []actor.Attorney{{
					FirstNames:  "Alice",
					LastName:    "Attorney",
					DateOfBirth: date.New("1998", "1", "2"),
					Email:       "alice@example.com",
					Address:     address,
				}},
			},
			ReplacementAttorneys: actor.Attorneys{
				Attorneys: []actor.Attorney{{
					FirstNames:  "Richard",
					LastName:    "Attorney",
					DateOfBirth: date.New("1999", "11", "12"),
					Email:       "richard@example.com",
					Address:     address,
				}},
			},
		})

		assert.Nil(t, err)
		return err
	})
}
