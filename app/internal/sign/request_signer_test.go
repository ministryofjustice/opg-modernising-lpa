package sign

import (
	"errors"
	"fmt"
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

var expectedError = errors.New("an error")

func TestNewRequestSigner(t *testing.T) {
	signer := v4.NewSigner()
	creds := aws.Credentials{AccessKeyID: "abc"}
	now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }

	requestSigner := NewRequestSigner(signer, creds, now, "region-name")

	assert.Equal(t, signer, requestSigner.v4Signer)
	assert.Equal(t, creds, requestSigner.credentials)
	assert.Equal(t, now(), requestSigner.now())
	assert.Equal(t, "region-name", requestSigner.awsRegion)
}

func TestSign(t *testing.T) {
	testCases := map[string]struct {
		Body          io.Reader
		SignedHeaders string
		Signature     string
	}{
		"empty body": {
			Body:          nil,
			SignedHeaders: "a-header;host;x-amz-date",
			Signature:     "99f815531e473759852fb13154796d31f4cfaccc3036f91193df440adeba0588",
		},
		"with body": {
			Body:          strings.NewReader(`{"some": "body data"}`),
			SignedHeaders: "a-header;content-length;host;x-amz-date",
			Signature:     "c9f9b78004a45e947d0fd7ea4eba56a86d3234bed2a7b69240231a1beb8150e9",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/an-url", tc.Body)
			req.Header.Set("a-header", "with-a-value")

			now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }
			signer := &RequestSigner{
				v4Signer:    v4.NewSigner(),
				credentials: aws.Credentials{AccessKeyID: "abc"},
				now:         now,
				awsRegion:   "eu-west-1",
			}

			err := signer.Sign(req.Context(), req, "service-name")

			assert.Nil(t, err)

			assert.Equal(t, "20000102T000000Z", req.Header.Get("X-Amz-Date"))
			assert.Equal(t, fmt.Sprintf("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/service-name/aws4_request, SignedHeaders=%s, Signature=%s", tc.SignedHeaders, tc.Signature), req.Header.Get("Authorization"))
			assert.Equal(t, "with-a-value", req.Header.Get("a-header"))
		})

	}
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 1, expectedError
}

func TestSignOnReadAllError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/an-url", errReader{})
	req.Header.Set("Content-Length", "100")

	now := func() time.Time { return time.Now() }
	signer := &RequestSigner{v4Signer: v4.NewSigner(), credentials: aws.Credentials{}, now: now}

	err := signer.Sign(req.Context(), req, "")

	assert.Equal(t, expectedError, err)
}

func TestSignOnSignHttpError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/an-url", nil)
	req.Header.Set("a-header", "with-a-value")

	now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }
	v4Signer := newMockV4Signer(t)
	v4Signer.
		On("SignHTTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	signer := &RequestSigner{v4Signer: v4Signer, credentials: aws.Credentials{AccessKeyID: "abc"}, now: now}

	err := signer.Sign(req.Context(), req, "")

	assert.Equal(t, expectedError, err)
}
