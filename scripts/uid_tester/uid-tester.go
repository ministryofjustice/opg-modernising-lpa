package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

type RequestSigner struct {
	v4Signer    *v4.Signer
	credentials aws.Credentials
	awsRegion   string
}

func main() {
	baseUrl := flag.String("baseUrl", "https://development.lpa-uid.api.opg.service.justice.gov.uk", "Base URL of UID service (defaults to 'https://development.lpa-uid.api.opg.service.justice.gov.uk'")

	log.Println("POSTing to: " + *baseUrl)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	credentials, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	req := buildRequest(*baseUrl)

	signer := RequestSigner{v4Signer: v4.NewSigner(), credentials: credentials, awsRegion: cfg.Region}
	err = signer.Sign(context.TODO(), req)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Request headers:\n")

	for name, values := range req.Header {
		for _, value := range values {
			log.Println(name, value)
		}
	}

	log.Println("Request body:\n")
	log.Println(req.Body)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode > http.StatusCreated {
		log.Println("Response headers:\n")

		for name, values := range resp.Header {
			for _, value := range values {
				log.Println(name, value)
			}
		}

		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("error POSTing to UID service: (%d) %s", resp.StatusCode, string(body))
	}

	log.Println("Its working!")
}

func buildRequest(baseUrl string) *http.Request {
	r, err := http.NewRequest(http.MethodPost, baseUrl+"/cases", bytes.NewReader([]byte(`{"type":"pfa","source":"APPLICANT","donor":{"name":"Jamie Smith","dob":"2000-01-02","postcode":"B14 7ED"}}`)))
	if err != nil {
		log.Fatal(err)
	}

	r.Header.Add("Content-Type", "application/json")
	return r
}

func (rs *RequestSigner) Sign(ctx context.Context, req *http.Request) error {
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

	err := rs.v4Signer.SignHTTP(ctx, rs.credentials, req, encodedBody, "execute-api", rs.awsRegion, time.Now())
	if err != nil {
		return err
	}

	return nil
}
