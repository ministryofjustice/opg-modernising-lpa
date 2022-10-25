package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDoer struct {
	mock.Mock
}

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

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

	doer := &mockDoer{}
	doer.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			var v map[string]string
			json.NewDecoder(req.Body).Decode(&v)

			return assert.Equal(http.MethodPost, req.Method) &&
				assert.Equal("https://api.notifications.service.gov.uk/v2/notifications/email", req.URL.String()) &&
				assert.Equal("application/json", req.Header.Get("Content-Type")) &&
				assert.Equal("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJmMzM1MTdmZi0yYTg4LTRmNmUtYjg1NS1jNTUwMjY4Y2UwOGEiLCJpYXQiOjE1Nzc5MzQyNDV9.V0iR-Foo_twZdWttAxy4koJoSYJzyZHMr-tJIBwZj8k", req.Header.Get("Authorization")) &&
				assert.Equal("me@example.com", v["email_address"]) &&
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

	doer := &mockDoer{}
	doer.
		On("Do", mock.Anything).
		Return(&http.Response{
			Body: io.NopCloser(strings.NewReader(`{"errors":[{"error":"SomeError","message":"This happened"}, {"error":"AndError","message":"Plus this"}]}`)),
		}, nil)

	client, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", doer)

	_, err := client.Email(ctx, Email{EmailAddress: "me@example.com", TemplateID: "template-123"})
	assert.Equal(`error sending email: This happened: Plus this`, err.Error())
}

func TestTemplateID(t *testing.T) {
	production, _ := New(true, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", nil)
	assert.Equal(t, "95f7b0a2-1c3a-4ad9-818b-b358c549c88b", production.TemplateID("MLPA Beta signature code"))

	test, _ := New(false, "", "my_client-f33517ff-2a88-4f6e-b855-c550268ce08a-740e5834-3a29-46b4-9a6f-16142fde533a", nil)
	assert.Equal(t, "7e8564a0-2635-4f61-9155-0166ddbe5607", test.TemplateID("MLPA Beta signature code"))
}
