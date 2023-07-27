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

func TestNewWithEmptyBaseURL(t *testing.T) {
	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", http.DefaultClient)

	assert.Equal(t, "https://api.notifications.service.gov.uk", client.baseURL)
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
	assert.Equal(t, "95f7b0a2-1c3a-4ad9-818b-b358c549c88b", production.TemplateID(SignatureCodeEmail))
	assert.Equal(t, "e39849c0-ecab-4e16-87ec-6b22afb9d535", production.TemplateID(SignatureCodeSMS))
	assert.Equal(t, "a10341e3-3bbd-4452-b52f-ebb4f51a4d73", production.TemplateID(CertificateProviderInviteEmail))
	assert.Equal(t, "453917cd-d8bb-44af-90a1-d73ae0f3fd07", production.TemplateID(CertificateProviderReturnEmail))
	assert.Equal(t, "9f8be86f-864a-4cda-a58a-5768522bd325", production.TemplateID(CertificateProviderNameChangeEmail))
	assert.Equal(t, "9aaedb70-df4a-42a8-9c28-de435cb3d453", production.TemplateID(AttorneyInviteEmail))
	assert.Equal(t, "1e0950c5-63fa-487e-8bf3-f40445412a12", production.TemplateID(AttorneyNameChangeEmail))
	assert.Equal(t, "6be11b4a-79f9-441e-8afe-adff96f7e7fc", production.TemplateID(CertificateProviderPaperMeetingPromptSMS))
	assert.Equal(t, "1c4d5b24-fc7d-45ee-be40-f1ccda96f101", production.TemplateID(ReplacementAttorneyInviteEmail))
	assert.Equal(t, "6be11b4a-79f9-441e-8afe-adff96f7e7fc", production.TemplateID(CertificateProviderPaperMeetingPromptSMS))
	assert.Equal(t, "19948d7d-a2df-4e85-930b-5d800978f41f", production.TemplateID(CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS))
	assert.Equal(t, "d363a56f-e802-4f88-bd09-80b8c9e9d650", production.TemplateID(CertificateProviderPaperLpaDetailsChangedSMS))
	assert.Equal(t, "71d21daa-11f9-4a2a-9ae2-bb5c2247bfb7", production.TemplateID(CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS))

	test, _ := New(false, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", nil)
	assert.Equal(t, "7e8564a0-2635-4f61-9155-0166ddbe5607", test.TemplateID(SignatureCodeEmail))
	assert.Equal(t, "dfa15e16-1f23-494a-bffb-a475513df6cc", test.TemplateID(SignatureCodeSMS))
	assert.Equal(t, "dd864a1a-64b4-4b4e-b810-86267ebd6476", test.TemplateID(CertificateProviderInviteEmail))
	assert.Equal(t, "dd864a1a-64b4-4b4e-b810-86267ebd6476", test.TemplateID(CertificateProviderReturnEmail))
	assert.Equal(t, "0f111ed1-5c58-47eb-a13f-931f2077523b", test.TemplateID(CertificateProviderNameChangeEmail))
	assert.Equal(t, "9be88a99-21c0-4808-8d6a-52af366e44aa", test.TemplateID(AttorneyInviteEmail))
	assert.Equal(t, "685bbdcc-71b8-48b9-b773-03941472d3b1", test.TemplateID(AttorneyNameChangeEmail))
	assert.Equal(t, "0eba4e55-c07e-4427-b4ad-b03e08dad8ca", test.TemplateID(CertificateProviderPaperMeetingPromptSMS))
	assert.Equal(t, "bf79859b-72b7-4701-bfd3-22ac6f0908c8", test.TemplateID(ReplacementAttorneyInviteEmail))
	assert.Equal(t, "0eba4e55-c07e-4427-b4ad-b03e08dad8ca", test.TemplateID(CertificateProviderPaperMeetingPromptSMS))
	assert.Equal(t, "d7513751-49ba-4276-aef5-ad67361d29c4", test.TemplateID(CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS))
	assert.Equal(t, "94477364-281a-4032-9a88-b215f969cd12", test.TemplateID(CertificateProviderPaperLpaDetailsChangedSMS))
	assert.Equal(t, "359fffa0-e1ec-444c-a886-6f046af374ab", test.TemplateID(CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS))
}

func TestRequest(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	doer := newMockDoer(t)

	var jsonBody bytes.Buffer
	jsonBody.WriteString(`{"some": "json"}`)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)
	client.now = func() time.Time { return time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC) }

	req, err := client.request(ctx, "/an/url", &jsonBody)

	assert.Nil(err)
	assert.Equal(http.MethodPost, req.Method)
	assert.Equal("https://api.notifications.service.gov.uk/an/url", req.URL.String())
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
