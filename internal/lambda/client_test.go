package lambda

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	expectedError = errors.New("an error")
	credentials   = aws.Credentials{AccessKeyID: "access-key-id", SecretAccessKey: "secret-access-key"}
	region        = "eu-west-1"
	testNow       = time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
	testNowFn     = func() time.Time { return testNow }
)

type mockCredentialsProvider struct {
	willFail bool
}

func (m *mockCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	if m.willFail {
		return aws.Credentials{}, errors.New("an error")
	}

	return credentials, nil
}

func (m *mockCredentialsProvider) IsExpired() bool {
	return false
}

func createTestConfig(willFailRetrieveCreds bool) aws.Config {
	return aws.Config{
		Region: region,
		Credentials: &mockCredentialsProvider{
			willFail: willFailRetrieveCreds,
		},
	}
}

func TestNew(t *testing.T) {
	signer := v4.NewSigner()
	cfg := createTestConfig(false)

	client := New(cfg, signer, http.DefaultClient, testNowFn)

	assert.Equal(t, cfg, client.cfg)
	assert.Equal(t, http.DefaultClient, client.doer)
	assert.Equal(t, signer, client.signer)
	assert.Equal(t, testNow, client.now())
}

func TestClientDo(t *testing.T) {
	testCases := map[string]struct {
		body          io.Reader
		encodedBody   string
		signedHeaders string
		signature     string
	}{
		"empty body": {
			body:          nil,
			encodedBody:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			signedHeaders: "a-header;host;x-amz-date",
			signature:     "99f815531e473759852fb13154796d31f4cfaccc3036f91193df440adeba0588",
		},
		"with body": {
			body:          strings.NewReader(`{"some": "body data"}`),
			encodedBody:   "50c0065bb0d1e4f12d2505e2dab0219250b8202a395e6b2ab3b62e12fe4b8100",
			signedHeaders: "a-header;content-length;host;x-amz-date",
			signature:     "c9f9b78004a45e947d0fd7ea4eba56a86d3234bed2a7b69240231a1beb8150e9",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			now := func() time.Time { return testNow }

			req, _ := http.NewRequest(http.MethodPost, "/an-url", tc.body)
			req.Header.Set("a-header", "with-a-value")

			expectedResponse := &http.Response{StatusCode: http.StatusTeapot}

			signer := newMockSigner(t)
			signer.EXPECT().
				SignHTTP(req.Context(), credentials, req, tc.encodedBody, "execute-api", region, testNow).
				Return(nil)

			doer := newMockDoer(t)
			doer.EXPECT().
				Do(req).
				Return(expectedResponse, expectedError)

			client := New(createTestConfig(false), signer, doer, now)

			resp, err := client.Do(req)
			assert.Equal(t, expectedError, err)
			assert.Equal(t, expectedResponse, resp)
		})
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 1, expectedError
}

func TestClientDoWhenReadAllError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/an-url", errReader{})
	req.Header.Set("Content-Length", "100")

	now := func() time.Time { return time.Now() }
	client := &Client{
		doer:   http.DefaultClient,
		cfg:    createTestConfig(false),
		signer: v4.NewSigner(),
		now:    now,
	}

	_, err := client.Do(req)
	assert.Equal(t, expectedError, err)
}

func TestClientDoWhenRetrieveCredentialsError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/an-url", nil)
	req.Header.Set("a-header", "with-a-value")

	client := &Client{
		doer:   http.DefaultClient,
		cfg:    createTestConfig(true),
		signer: nil,
		now:    testNowFn,
	}

	_, err := client.Do(req)
	assert.Equal(t, expectedError, err)
}

func TestClientDoWhenSignerError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/an-url", nil)
	req.Header.Set("a-header", "with-a-value")

	v4Signer := newMockSigner(t)
	v4Signer.EXPECT().
		SignHTTP(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	client := &Client{
		doer:   http.DefaultClient,
		cfg:    createTestConfig(false),
		signer: v4Signer,
		now:    testNowFn,
	}

	_, err := client.Do(req)
	assert.Equal(t, expectedError, err)
}
