package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	client, err := New(true, "http://base", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", http.DefaultClient, newMockEventClient(t))

	assert.Nil(t, err)
	assert.Equal(t, "http://base", client.baseURL)
	assert.Equal(t, "f33517ff-2a88-4f6e-b855-c550268ce08a", client.issuer)
	assert.Equal(t, []byte("740e5834-3a29-46b4-9a6f-16142fde533a"), client.secretKey)
	assert.Equal(t, newMockEventClient(t), client.eventClient)

}

func TestNewWithInvalidApiKey(t *testing.T) {
	_, err := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f", http.DefaultClient, nil)

	assert.NotNil(t, err)
}

type testEmail struct {
	A string
}

func (e testEmail) emailID(bool) string { return "template-id" }

func TestSendEmail(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			var buf bytes.Buffer
			io.Copy(&buf, req.Body)
			req.Body = io.NopCloser(&buf)

			var v map[string]any
			json.Unmarshal(buf.Bytes(), &v)

			return assert.Equal("me@example.com", v["email_address"].(string)) &&
				assert.Equal("template-id", v["template_id"].(string)) &&
				assert.Equal(map[string]any{"A": "value"}, v["personalisation"].(map[string]any))
		})).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	err := client.SendEmail(ctx, "me@example.com", testEmail{A: "value"})
	assert.Nil(err)
}

func TestSendEmailWhenError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)

	err := client.SendEmail(ctx, "me@example.com", testEmail{})
	assert.Equal(`error sending message: This happened: Plus this`, err.Error())
}

func TestSendActorEmail(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			var buf bytes.Buffer
			io.Copy(&buf, req.Body)
			req.Body = io.NopCloser(&buf)

			var v map[string]any
			json.Unmarshal(buf.Bytes(), &v)

			return assert.Equal("me@example.com", v["email_address"].(string)) &&
				assert.Equal("template-id", v["template_id"].(string)) &&
				assert.Equal(map[string]any{"A": "value"}, v["personalisation"].(map[string]any))
		})).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendNotificationSent(ctx, event.NotificationSent{UID: "lpa-uid", NotificationID: "xyz"}).
		Return(nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, eventClient)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	err := client.SendActorEmail(ctx, "me@example.com", "lpa-uid", testEmail{A: "value"})
	assert.Nil(err)
}

func TestSendActorEmailWhenError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)

	err := client.SendActorEmail(ctx, "me@example.com", "lpa-uid", testEmail{})
	assert.Equal(`error sending message: This happened: Plus this`, err.Error())
}

func TestSendActorEmailWhenEventError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			var buf bytes.Buffer
			io.Copy(&buf, req.Body)
			req.Body = io.NopCloser(&buf)

			var v map[string]any
			json.Unmarshal(buf.Bytes(), &v)

			return assert.Equal("me@example.com", v["email_address"].(string)) &&
				assert.Equal("template-id", v["template_id"].(string)) &&
				assert.Equal(map[string]any{"A": "value"}, v["personalisation"].(map[string]any))
		})).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendNotificationSent(ctx, event.NotificationSent{UID: "lpa-uid", NotificationID: "xyz"}).
		Return(errors.New("err"))

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, eventClient)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	err := client.SendActorEmail(ctx, "me@example.com", "lpa-uid", testEmail{A: "value"})
	assert.Equal(errors.New("err"), err)
}

func TestNewRequest(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	doer := newMockDoer(t)

	client, _ := New(true, "http://base", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, err := client.newRequest(ctx, "/an/url", map[string]string{"some": "json"})

	assert.Nil(err)
	assert.Equal(http.MethodPost, req.Method)
	assert.Equal("http://base/an/url", req.URL.String())
	assert.Equal("application/json", req.Header.Get("Content-Type"))
	assert.Equal("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJmMzM1MTdmZi0yYTg4LTRmNmUtYjg1NS1jNTUwMjY4Y2UwOGEiLCJpYXQiOjE1Nzc5MzQyNDV9.V0iR-Foo_twZdWttAxy4koJoSYJzyZHMr-tJIBwZj8k", req.Header.Get("Authorization"))
}

func TestNewRequestWhenNewRequestError(t *testing.T) {
	assert := assert.New(t)
	doer := newMockDoer(t)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)
	client.now = func() time.Time { return time.Now().Add(-time.Minute) }

	_, err := client.newRequest(nil, "/an/url", map[string]string{"some": "json"})

	assert.Equal(errors.New("net/http: nil Context"), err)
}

func TestDo(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	jsonString := `{"id": "123", "status_code": 400}`

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(jsonString)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(jsonString)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.newRequest(ctx, "/an/url", &jsonBody)

	response, err := client.do(req)

	assert.Nil(err)
	assert.Equal(response.ID, "123")
	assert.Equal(response.StatusCode, 400)
}

func TestDoWhenContainsErrorList(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	jsonString := `{"id": "123", "status_code": 400, "errors": [{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(jsonString)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(jsonString)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.newRequest(ctx, "/an/url", &jsonBody)

	response, err := client.do(req)

	assert.Equal(errorsList{
		errorItem{
			Error:   "SomeError",
			Message: "This happened",
		},
		errorItem{
			Error:   "AndError",
			Message: "Plus this",
		},
	}, err)
	assert.Equal("123", response.ID)
	assert.Equal(400, response.StatusCode)
	assert.Equal(errorsList{
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
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id": "123"}`)),
		}, errors.New("err"))

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`{"id": "123"}`)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.newRequest(ctx, "/an/url", &jsonBody)

	resp, err := client.do(req)

	assert.Equal(errors.New("err"), err)
	assert.Equal(response{}, resp)
}

func TestDoRequestWhenJsonDecodeFails(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`not json`)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`not json`)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.newRequest(ctx, "/an/url", &jsonBody)

	resp, err := client.do(req)

	assert.IsType(&json.SyntaxError{}, err)
	assert.Equal(response{}, resp)
}

type testSMS struct {
	A string
}

func (e testSMS) smsID(bool) string { return "template-id" }

func TestSendSMS(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			var buf bytes.Buffer
			io.Copy(&buf, req.Body)
			req.Body = io.NopCloser(&buf)

			var v map[string]any
			json.Unmarshal(buf.Bytes(), &v)

			return assert.Equal("+447535111111", v["phone_number"].(string)) &&
				assert.Equal("template-id", v["template_id"].(string)) &&
				assert.Equal(map[string]any{"A": "value"}, v["personalisation"].(map[string]any))

		})).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	id, err := client.SendSMS(ctx, "+447535111111", testSMS{A: "value"})

	assert.Nil(err)
	assert.Equal("xyz", id)
}

func TestSendSMSWhenError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer, nil)

	_, err := client.SendSMS(ctx, "+447535111111", testSMS{})
	assert.Equal(`error sending message: This happened: Plus this`, err.Error())
}
