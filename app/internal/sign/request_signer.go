package sign

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

//go:generate mockery --testonly --inpackage --name v4Signer --structname mockV4Signer
type v4Signer interface {
	SignHTTP(ctx context.Context, credentials aws.Credentials, r *http.Request, payloadHash string, service string, region string, signingTime time.Time, optFns ...func(options *v4.SignerOptions)) error
}

type RequestSigner struct {
	v4Signer    v4Signer
	credentials aws.Credentials
	now         func() time.Time
	awsRegion   string
}

func NewRequestSigner(signer v4Signer, credentials aws.Credentials, now func() time.Time, awsRegion string) *RequestSigner {
	return &RequestSigner{
		v4Signer:    signer,
		credentials: credentials,
		now:         now,
		awsRegion:   awsRegion,
	}
}

func (rs *RequestSigner) Sign(ctx context.Context, req *http.Request, serviceName string) error {
	reqBody := []byte("")

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		reqBody = body
	}

	hash := sha256.New()
	hash.Write(reqBody)
	encodedBody := hex.EncodeToString(hash.Sum(nil))

	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	err := rs.v4Signer.SignHTTP(ctx, rs.credentials, req, encodedBody, serviceName, rs.awsRegion, rs.now())
	if err != nil {
		return err
	}

	return nil
}
