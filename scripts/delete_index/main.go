package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <awsRegion> <opensearchEndpoint> <environmentName>")
		return
	}

	awsRegion := os.Args[1]
	opensearchEndpoint := os.Args[2]
	environmentName := os.Args[3]
	serviceName := "aoss"

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewEnvCredentials(),
	})

	if err != nil {
		fmt.Println("Error creating session ", err)
		return
	}

	signer := v4.NewSigner(sess.Config.Credentials)

	request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", opensearchEndpoint, environmentName), nil)

	_, err = signer.Sign(request, nil, serviceName, awsRegion, time.Now())

	if err != nil {
		fmt.Println("Error signing request ", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(request)

	if err != nil {
		fmt.Println("Error making request ", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Println("Error: Unexpected response status code ", resp.StatusCode)
		return
	}

	fmt.Println("Response status code: ", resp.StatusCode)
}
