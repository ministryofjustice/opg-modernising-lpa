package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
)

var expectedError = errors.New("err")

func TestNew(t *testing.T) {
	bundle := &localize.Bundle{}

	client, err := New(nil, true, "http://base", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", http.DefaultClient, newMockEventClient(t), bundle)

	assert.Nil(t, err)
	assert.Equal(t, "http://base", client.baseURL)
	assert.Equal(t, "f33517ff-2a88-4f6e-b855-c550268ce08a", client.issuer)
	assert.Equal(t, []byte("740e5834-3a29-46b4-9a6f-16142fde533a"), client.secretKey)
	assert.Equal(t, newMockEventClient(t), client.eventClient)
	assert.Equal(t, bundle, client.bundle)
}

func TestNewWithInvalidApiKey(t *testing.T) {
	_, err := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f", http.DefaultClient, nil, nil)

	assert.NotNil(t, err)
}

func TestEmailGreeting(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json", "testdata/cy.json")
	client := &Client{bundle: bundle}

	testcases := map[string]struct {
		lpa      *lpadata.Lpa
		expected string
	}{
		"english donor": {
			lpa: &lpadata.Lpa{
				Donor: lpadata.Donor{
					ContactLanguagePreference: localize.En,
					FirstNames:                "John",
					LastName:                  "Smith",
				},
			},
			expected: "Hi John Smith",
		},
		"welsh donor": {
			lpa: &lpadata.Lpa{
				Donor: lpadata.Donor{
					ContactLanguagePreference: localize.Cy,
					FirstNames:                "John",
					LastName:                  "Smith",
				},
			},
			expected: "Hy John Smith",
		},
		"english donor with correspondent": {
			lpa: &lpadata.Lpa{
				LpaUID: "M-FAKE-1111",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					ContactLanguagePreference: localize.En,
					FirstNames:                "John",
					LastName:                  "Smith",
				},
				Correspondent: lpadata.Correspondent{
					FirstNames: "Dave",
					LastName:   "David",
				},
			},
			expected: "Hello Dave David for John Smith’s Personal welfare LPA (M-FAKE-1111)",
		},
		"welsh donor with correspondent": {
			lpa: &lpadata.Lpa{
				LpaUID: "M-FAKE-1111",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor: lpadata.Donor{
					ContactLanguagePreference: localize.Cy,
					FirstNames:                "John",
					LastName:                  "Smith",
				},
				Correspondent: lpadata.Correspondent{
					FirstNames: "Dave",
					LastName:   "David",
				},
			},
			expected: "Hyllw Dave David fwr John Smith Property and affairs LPA (M-FAKE-1111)",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			greeting := client.EmailGreeting(tc.lpa)

			assert.Equal(t, tc.expected, greeting)
		})
	}
}

type testEmail struct {
	A string
}

func (e testEmail) emailID(bool, localize.Lang) string { return "template-id" }

func TestSendEmail(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	innerCtx, _ := otel.GetTracerProvider().Tracer("mlpab").Start(ctx, "")

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			if req.Method != http.MethodPost {
				return false
			}

			var v map[string]any
			json.Unmarshal(readBody(req).Bytes(), &v)

			return assert.Equal(innerCtx, req.Context()) &&
				assert.Equal("me@example.com", v["email_address"].(string)) &&
				assert.Equal("template-id", v["template_id"].(string)) &&
				assert.Equal(map[string]any{"A": "value"}, v["personalisation"].(map[string]any))
		})).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil).
		Once()

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	err := client.SendEmail(ctx, localize.En, "me@example.com", testEmail{A: "value"})
	assert.Nil(err)
}

func TestSendEmailWhenError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	innerCtx, _ := otel.GetTracerProvider().Tracer("mlpab").Start(ctx, "")

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(innerCtx, "email send failed", slog.String("to", "me@example.com"))

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil).
		Once()

	client, _ := New(logger, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)

	err := client.SendEmail(ctx, localize.En, "me@example.com", testEmail{})
	assert.Equal(`error sending message: This happened: Plus this`, err.Error())
}

func TestSendActorEmail(t *testing.T) {
	testcases := map[string]string{
		"not previously created":           `{"notifications":[]}`,
		"previously created before window": `{"notifications":[{"status":"sending","created_at":"2020-01-02T02:53:06Z"}]}`,
		"previously failed":                `{"notifications":[{"status":"temporary-failure","created_at":"2020-01-02T02:57:06Z"}]}`,
	}

	for name, responseBody := range testcases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()
			innerCtx, _ := otel.GetTracerProvider().Tracer("mlpab").Start(ctx, "")

			doer := newMockDoer(t)
			doer.EXPECT().
				Do(mock.MatchedBy(func(req *http.Request) bool {
					if req.Method != http.MethodGet {
						return false
					}

					return assert.Equal(innerCtx, req.Context()) &&
						assert.Equal("/v2/notifications?reference=7mHebbumP4dq7lwL0a0GKXrf4Y6AzVKyY6PPfyG+4Kk", req.URL.String()) &&
						assert.Equal("", readBody(req).String())
				})).
				Return(&http.Response{
					Body: io.NopCloser(strings.NewReader(responseBody)),
				}, nil).
				Once()
			doer.EXPECT().
				Do(mock.MatchedBy(func(req *http.Request) bool {
					if req.Method != http.MethodPost {
						return false
					}

					var v map[string]any
					json.Unmarshal(readBody(req).Bytes(), &v)

					return assert.Equal("me@example.com", v["email_address"].(string)) &&
						assert.Equal("template-id", v["template_id"].(string)) &&
						assert.Equal(map[string]any{"A": "value"}, v["personalisation"].(map[string]any)) &&
						assert.Equal("7mHebbumP4dq7lwL0a0GKXrf4Y6AzVKyY6PPfyG+4Kk", v["reference"].(string))
				})).
				Return(&http.Response{
					Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
				}, nil).
				Once()

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendNotificationSent(innerCtx, event.NotificationSent{UID: "lpa-uid", NotificationID: "xyz"}).
				Return(nil)

			client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, eventClient, nil)
			client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

			err := client.SendActorEmail(ctx, localize.En, "me@example.com", "lpa-uid", testEmail{A: "value"})
			assert.Nil(err)
		})
	}
}

func TestSendActorEmailWhenToSimulated(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	for _, email := range simulatedEmails {
		doer := newMockDoer(t)
		doer.EXPECT().
			Do(mock.Anything).
			Return(&http.Response{
				Body: io.NopCloser(strings.NewReader(`{"notifications":[]}`)),
			}, nil).
			Once()
		doer.EXPECT().
			Do(mock.Anything).
			Return(&http.Response{
				Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
			}, nil).
			Once()

		client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
		client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

		err := client.SendActorEmail(ctx, localize.En, email, "lpa-uid", testEmail{A: "value"})
		assert.Nil(err)
	}
}

func TestSendActorEmailWhenAlreadyRecentlyCreated(t *testing.T) {
	testcases := map[string]string{
		"previously created":       `{"notifications":[{"status":"delivered","created_at":"2020-01-02T02:54:06Z"}]}`,
		"previously created mixed": `{"notifications":[{"status":"sending","created_at":"2020-01-02T02:53:06Z"}, {"status":"delivered","created_at":"2020-01-02T02:54:06Z"}, {"status":"permanent-failure","created_at":"2020-01-02T02:56:06Z"}]}`,
	}

	for name, responseBody := range testcases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()
			innerCtx, _ := otel.GetTracerProvider().Tracer("mlpab").Start(ctx, "")

			doer := newMockDoer(t)
			doer.EXPECT().
				Do(mock.MatchedBy(func(req *http.Request) bool {
					if req.Method != http.MethodGet {
						return false
					}

					return assert.Equal(innerCtx, req.Context()) &&
						assert.Equal("/v2/notifications?reference=7mHebbumP4dq7lwL0a0GKXrf4Y6AzVKyY6PPfyG+4Kk", req.URL.String()) &&
						assert.Equal("", readBody(req).String())
				})).
				Return(&http.Response{
					Body: io.NopCloser(strings.NewReader(responseBody)),
				}, nil).
				Once()

			client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
			client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

			err := client.SendActorEmail(ctx, localize.En, "me@example.com", "lpa-uid", testEmail{A: "value"})
			assert.Nil(err)
		})
	}
}

func TestSendActorEmailWhenReferenceExistsError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)

	err := client.SendActorEmail(context.Background(), localize.En, "me@example.com", "lpa-uid", testEmail{A: "value"})
	assert.Equal(t, expectedError, err)
}

func TestSendActorEmailWhenError(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(mock.Anything, "email send failed", slog.String("to", "me@example.com"))

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"notifications":[]}`)),
		}, nil).
		Once()
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil).
		Once()

	client, _ := New(logger, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)

	err := client.SendActorEmail(context.Background(), localize.En, "me@example.com", "lpa-uid", testEmail{})
	assert.Equal(t, "error sending message: This happened: Plus this", err.Error())
}

func TestSendActorEmailWhenEventError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"notifications":[]}`)),
		}, nil).
		Once()
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil).
		Once()

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendNotificationSent(mock.Anything, mock.Anything).
		Return(expectedError)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, eventClient, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	err := client.SendActorEmail(context.Background(), localize.En, "me@example.com", "lpa-uid", testEmail{A: "value"})
	assert.Equal(t, expectedError, err)
}

func TestNewRequest(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	doer := newMockDoer(t)

	client, _ := New(nil, true, "http://base", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, err := client.newRequest(ctx, http.MethodPost, "/an/url", map[string]string{"some": "json"})

	assert.Nil(err)
	assert.Equal(http.MethodPost, req.Method)
	assert.Equal("http://base/an/url", req.URL.String())
	assert.Equal("application/json", req.Header.Get("Content-Type"))
	assert.Equal("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJmMzM1MTdmZi0yYTg4LTRmNmUtYjg1NS1jNTUwMjY4Y2UwOGEiLCJpYXQiOjE1Nzc5MzQyNDV9.V0iR-Foo_twZdWttAxy4koJoSYJzyZHMr-tJIBwZj8k", req.Header.Get("Authorization"))
}

func TestNewRequestWhenNewRequestError(t *testing.T) {
	assert := assert.New(t)
	doer := newMockDoer(t)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
	client.now = func() time.Time { return time.Now().Add(-time.Minute) }

	_, err := client.newRequest(nil, http.MethodPost, "/an/url", map[string]string{"some": "json"})

	assert.Equal(errors.New("net/http: nil Context"), err)
}

func TestDo(t *testing.T) {
	jsonString := `{"id": "123", "status_code": 400}`

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(jsonString)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(jsonString)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.newRequest(context.Background(), http.MethodPost, "/an/url", &jsonBody)

	response, err := client.do(req)

	assert.Nil(t, err)
	assert.Equal(t, response.ID, "123")
	assert.Equal(t, response.StatusCode, 400)
}

func TestDoWhenContainsErrorList(t *testing.T) {
	jsonString := `{"id": "123", "status_code": 400, "errors": [{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(jsonString)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(jsonString)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.newRequest(context.Background(), http.MethodPost, "/an/url", &jsonBody)

	response, err := client.do(req)

	assert.Equal(t, errorsList{
		errorItem{
			Error:   "SomeError",
			Message: "This happened",
		},
		errorItem{
			Error:   "AndError",
			Message: "Plus this",
		},
	}, err)
	assert.Equal(t, "123", response.ID)
	assert.Equal(t, 400, response.StatusCode)
	assert.Equal(t, errorsList{
		errorItem{
			Error:   "SomeError",
			Message: "This happened",
		},
		errorItem{
			Error:   "AndError",
			Message: "Plus this",
		},
	}, response.Errors)
}

func TestDoRequestWhenRequestError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{Body: io.NopCloser(strings.NewReader(`{"id": "123"}`))}, expectedError)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`{"id": "123"}`)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.newRequest(context.Background(), http.MethodPost, "/an/url", &jsonBody)

	resp, err := client.do(req)

	assert.Equal(t, expectedError, err)
	assert.Equal(t, response{}, resp)
}

func TestDoRequestWhenJsonDecodeFails(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`not json`)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`not json`)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.newRequest(context.Background(), http.MethodPost, "/an/url", &jsonBody)

	resp, err := client.do(req)

	assert.IsType(t, &json.SyntaxError{}, err)
	assert.Equal(t, response{}, resp)
}

type testSMS struct {
	A string
}

func (e testSMS) smsID(bool, localize.Lang) string { return "template-id" }

func TestSendActorSMS(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	innerCtx, _ := otel.GetTracerProvider().Tracer("mlpab").Start(ctx, "")

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			var v map[string]any
			json.Unmarshal(readBody(req).Bytes(), &v)

			return assert.Equal("+447535111111", v["phone_number"].(string)) &&
				assert.Equal("template-id", v["template_id"].(string)) &&
				assert.Equal(map[string]any{"A": "value"}, v["personalisation"].(map[string]any))

		})).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendNotificationSent(innerCtx, event.NotificationSent{UID: "lpa-uid", NotificationID: "xyz"}).
		Return(nil)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, eventClient, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	err := client.SendActorSMS(ctx, localize.En, "+447535111111", "lpa-uid", testSMS{A: "value"})
	assert.Nil(err)
}

func TestSendActorSMSWhenToSimulated(t *testing.T) {
	for _, phone := range simulatedPhones {
		doer := newMockDoer(t)
		doer.EXPECT().
			Do(mock.Anything).
			Return(&http.Response{
				Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
			}, nil)

		client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)
		client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

		err := client.SendActorSMS(context.Background(), localize.En, phone, "lpa-uid", testSMS{A: "value"})
		assert.Nil(t, err)
	}
}

func TestSendActorSMSWhenError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil, nil)

	err := client.SendActorSMS(context.Background(), localize.En, "+447535111111", "lpa-uid", testSMS{})
	assert.Equal(t, "error sending message: This happened: Plus this", err.Error())
}

func TestSendActorSMSWhenEventError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`))}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendNotificationSent(mock.Anything, mock.Anything).
		Return(expectedError)

	client, _ := New(nil, true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, eventClient, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	err := client.SendActorSMS(context.Background(), localize.En, "+447535111111", "lpa-uid", testSMS{A: "value"})
	assert.Equal(t, expectedError, err)
}

func readBody(req *http.Request) *bytes.Buffer {
	var buf bytes.Buffer
	io.Copy(&buf, req.Body)
	req.Body = io.NopCloser(&buf)
	return &buf
}
