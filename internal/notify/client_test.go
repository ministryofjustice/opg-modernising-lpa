package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	client, err := New(true, "http://base", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", http.DefaultClient)

	assert.Nil(t, err)
	assert.Equal(t, "http://base", client.baseURL)
	assert.Equal(t, "f33517ff-2a88-4f6e-b855-c550268ce08a", client.issuer)
	assert.Equal(t, []byte("740e5834-3a29-46b4-9a6f-16142fde533a"), client.secretKey)

}

func TestNewWithInvalidApiKey(t *testing.T) {
	_, err := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f", http.DefaultClient)

	assert.NotNil(t, err)
}

func TestEmail(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			var buf bytes.Buffer
			io.Copy(&buf, req.Body)
			req.Body = ioutil.NopCloser(&buf)

			var v map[string]string
			json.Unmarshal(buf.Bytes(), &v)

			return assert.Equal("me@example.com", v["email_address"]) &&
				assert.Equal("template-123", v["template_id"])
		})).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	id, err := client.Email(ctx, Email{EmailAddress: "me@example.com", TemplateID: "template-123"})
	assert.Nil(err)
	assert.Equal("xyz", id)
}

func TestEmailWhenError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.
		On("Do", mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)

	_, err := client.Email(ctx, Email{EmailAddress: "me@example.com", TemplateID: "template-123"})
	assert.Equal(`error sending message: This happened: Plus this`, err.Error())
}

func TestTemplateID(t *testing.T) {
	production, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", nil)
	assert.Equal(t, "e39849c0-ecab-4e16-87ec-6b22afb9d535", production.TemplateID(WitnessCodeSMS))
	assert.Equal(t, "", production.TemplateID(Template(200)))

	test, _ := New(false, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", nil)
	assert.Equal(t, "dfa15e16-1f23-494a-bffb-a475513df6cc", test.TemplateID(WitnessCodeSMS))
	assert.Equal(t, "", test.TemplateID(Template(200)))
}

func TestRequest(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	doer := newMockDoer(t)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`{"some": "json"}`)

	client, _ := New(true, "http://base", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, err := client.request(ctx, "/an/url", &jsonBody)

	assert.Nil(err)
	assert.Equal(http.MethodPost, req.Method)
	assert.Equal("http://base/an/url", req.URL.String())
	assert.Equal("application/json", req.Header.Get("Content-Type"))
	assert.Equal("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJmMzM1MTdmZi0yYTg4LTRmNmUtYjg1NS1jNTUwMjY4Y2UwOGEiLCJpYXQiOjE1Nzc5MzQyNDV9.V0iR-Foo_twZdWttAxy4koJoSYJzyZHMr-tJIBwZj8k", req.Header.Get("Authorization"))
}

func TestRequestWhenNewRequestError(t *testing.T) {
	assert := assert.New(t)
	doer := newMockDoer(t)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`{"some": "json"}`)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Now().Add(-time.Minute) }

	_, err := client.request(nil, "/an/url", &jsonBody)

	assert.Equal(errors.New("net/http: nil Context"), err)
}

func TestDoRequest(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	jsonString := `{"id": "123", "status_code": 400}`

	doer := newMockDoer(t)
	doer.
		On("Do", mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(jsonString)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(jsonString)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.request(ctx, "/an/url", &jsonBody)

	response, err := client.doRequest(req)

	assert.Nil(err)
	assert.Equal(response.ID, "123")
	assert.Equal(response.StatusCode, 400)
}

func TestDoRequestWhenContainsErrorList(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	jsonString := `{"id": "123", "status_code": 400, "errors": [{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`

	doer := newMockDoer(t)
	doer.
		On("Do", mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(jsonString)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(jsonString)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.request(ctx, "/an/url", &jsonBody)

	response, err := client.doRequest(req)

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
	doer.
		On("Do", mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id": "123"}`)),
		}, errors.New("err"))

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`{"id": "123"}`)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.request(ctx, "/an/url", &jsonBody)

	resp, err := client.doRequest(req)

	assert.Equal(errors.New("err"), err)
	assert.Equal(response{}, resp)
}

func TestDoRequestWhenJsonDecodeFails(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.
		On("Do", mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`not json`)),
		}, nil)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`not json`)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, _ := client.request(ctx, "/an/url", &jsonBody)

	resp, err := client.doRequest(req)

	assert.IsType(&json.SyntaxError{}, err)
	assert.Equal(response{}, resp)
}

func TestSms(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			var buf bytes.Buffer
			io.Copy(&buf, req.Body)
			req.Body = ioutil.NopCloser(&buf)

			var v map[string]string
			json.Unmarshal(buf.Bytes(), &v)

			return assert.Equal("+447535111111", v["phone_number"]) &&
				assert.Equal("template-123", v["template_id"])
		})).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"id":"xyz"}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	id, err := client.Sms(ctx, Sms{PhoneNumber: "+447535111111", TemplateID: "template-123"})

	assert.Nil(err)
	assert.Equal("xyz", id)
}

func TestSmsWhenError(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	doer := newMockDoer(t)
	doer.
		On("Do", mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)

	_, err := client.Sms(ctx, Sms{PhoneNumber: "+447535111111", TemplateID: "template-123"})
	assert.Equal(`error sending message: This happened: Plus this`, err.Error())
}
