// Package lambda provides a client for AWS Lambda.
package lambda

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

const apiGatewayServiceName = "execute-api"

type Signer interface {
	SignHTTP(context.Context, aws.Credentials, *http.Request, string, string, string, time.Time, ...func(options *v4.SignerOptions)) error
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// A Client makes HTTP requests to AWS Lambda functions
type Client struct {
	cfg    aws.Config
	signer Signer
	doer   Doer
	now    func() time.Time
}

// New creates a Client for calling AWS Lambda functions
func New(cfg aws.Config, signer Signer, doer Doer, now func() time.Time) *Client {
	return &Client{
		cfg:    cfg,
		signer: signer,
		doer:   doer,
		now:    now,
	}
}

// Do executes the HTTP request
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	hash := sha256.New()

	if req.Body != nil {
		var reqBody bytes.Buffer

		if _, err := io.Copy(hash, io.TeeReader(req.Body, &reqBody)); err != nil {
			return nil, err
		}

		req.Body = io.NopCloser(&reqBody)
	}

	encodedBody := hex.EncodeToString(hash.Sum(nil))

	credentials, err := c.cfg.Credentials.Retrieve(req.Context())
	if err != nil {
		return nil, err
	}

	if err := c.signer.SignHTTP(req.Context(), credentials, req, encodedBody, apiGatewayServiceName, c.cfg.Region, c.now().UTC()); err != nil {
		return nil, err
	}

	return c.doer.Do(req)
}
