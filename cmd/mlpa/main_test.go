package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAwsRegion(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/task" {
			// https://docs.aws.amazon.com/AmazonECS/latest/userguide/task-metadata-endpoint-v4-fargate.html
			io.WriteString(w, `{"Cluster": "arn:aws:ecs:us-west-2:111122223333:cluster/default", "TaskARN": "arn:aws:ecs:us-west-2:111122223333:task/default/e9028f8d5d8e4f258373e7b93ce9a3c3", "Family": "curltest", "Revision": "3", "DesiredStatus": "RUNNING", "KnownStatus": "RUNNING", "Limits": {"CPU": 0.25, "Memory": 512}, "PullStartedAt": "2020-10-08T20:47:16.053330955Z", "PullStoppedAt": "2020-10-08T20:47:19.592684631Z", "AvailabilityZone": "us-west-2a", "Containers": [{"DockerId": "e9028f8d5d8e4f258373e7b93ce9a3c3-2495160603", "Name": "curl", "DockerName": "curl", "Image": "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest", "ImageID": "sha256:25f3695bedfb454a50f12d127839a68ad3caf91e451c1da073db34c542c4d2cb", "Labels": {"com.amazonaws.ecs.cluster": "arn:aws:ecs:us-west-2:111122223333:cluster/default", "com.amazonaws.ecs.container-name": "curl", "com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-west-2:111122223333:task/default/e9028f8d5d8e4f258373e7b93ce9a3c3", "com.amazonaws.ecs.task-definition-family": "curltest", "com.amazonaws.ecs.task-definition-version": "3"}, "DesiredStatus": "RUNNING", "KnownStatus": "RUNNING", "Limits": {"CPU": 10, "Memory": 128}, "CreatedAt": "2020-10-08T20:47:20.567813946Z", "StartedAt": "2020-10-08T20:47:20.567813946Z", "Type": "NORMAL", "Networks": [{"NetworkMode": "awsvpc", "IPv4Addresses": ["192.0.2.3"], "IPv6Addresses": ["2001:dB8:10b:1a00:32bf:a372:d80f:e958"], "AttachmentIndex": 0, "MACAddress": "02:b7:20:19:72:39", "IPv4SubnetCIDRBlock": "192.0.2.0/24", "IPv6SubnetCIDRBlock": "2600:1f13:10b:1a00::/64", "DomainNameServers": ["192.0.2.2"], "DomainNameSearchList": ["us-west-2.compute.internal"], "PrivateDNSName": "ip-172-31-30-173.us-west-2.compute.internal", "SubnetGatewayIpv4Address": "192.0.2.0/24"}], "ClockDrift": {"ClockErrorBound": 0.5458234999999999, "ReferenceTimestamp": "2021-09-07T16:57:44Z", "ClockSynchronizationStatus": "SYNCHRONIZED"}, "ContainerARN": "arn:aws:ecs:us-west-2:111122223333:container/1bdcca8b-f905-4ee6-885c-4064cb70f6e6", "LogOptions": {"awslogs-create-group": "true", "awslogs-group": "/ecs/containerlogs", "awslogs-region": "us-west-2", "awslogs-stream": "ecs/curl/e9028f8d5d8e4f258373e7b93ce9a3c3"}, "LogDriver": "awslogs"}], "LaunchType": "FARGATE"}`)
		}
	}))
	defer s.Close()

	region, err := awsRegion(s.URL)
	assert.Nil(t, err)
	assert.Equal(t, "us-west-2", region)
}

func TestLanguageFilesUniqueKeys(t *testing.T) {
	for _, path := range []string{"lang/en.json", "lang/cy.json"} {
		data, err := os.ReadFile("../../" + path)
		if err != nil {
			panic(err)
		}

		keys := map[json.Token]struct{}{}
		dec := json.NewDecoder(bytes.NewReader(data))

		// skip opening {
		_, _ = dec.Token()

		for {
			tok, err := dec.Token()
			if err != nil {
				panic(err)
			}

			if tok == json.Delim('}') {
				break
			}

			// skip value
			valTok, _ := dec.Token()

			if _, found := keys[tok]; found {
				t.Fail()
				t.Log(path, "duplicate:", tok)
			}

			keys[tok] = struct{}{}

			if valTok == json.Delim('{') {
				for dec.More() {
					_, _ = dec.Token()
				}

				// skip closing }
				_, _ = dec.Token()
			}
		}
	}
}

func TestLanguageFilesMatch(t *testing.T) {
	en := loadTranslations("../../lang/en.json")
	cy := loadTranslations("../../lang/cy.json")

	for k := range en {
		if _, ok := cy[k]; !ok {
			t.Fail()
			t.Log("lang/cy.json missing:", k)
		}
	}

	for k := range cy {
		if _, ok := en[k]; !ok {
			t.Fail()
			t.Log("lang/en.json missing:", k)
		}
	}
}

func TestApostrophesAreCurly(t *testing.T) {
	en := loadTranslations("../../lang/en.json")
	cy := loadTranslations("../../lang/cy.json")

	for k, v := range en {
		if strings.Contains(v, "'") {
			t.Fail()
			t.Log("lang/en.json:", k)
		}
	}

	for k, v := range cy {
		if strings.Contains(v, "'") {
			t.Fail()
			t.Log("lang/cy.json:", k)
		}
	}
}

func TestTranslationVariablesMustMatch(t *testing.T) {
	en := loadTranslations("../../lang/en.json")
	cy := loadTranslations("../../lang/cy.json")

	for k, enTranslation := range en {
		if strings.Contains(enTranslation, "{{") {
			cyTranslation, _ := cy[k]

			r := regexp.MustCompile(`{{[^{}]*}}`)
			enMatches := r.FindAllString(enTranslation, -1)
			cyMatches := r.FindAllString(cyTranslation, -1)

			// Account for white space in var
			for enK, _ := range enMatches {
				enMatches[enK] = strings.ReplaceAll(enMatches[enK], " ", "")
			}

			for cyK, _ := range cyMatches {
				cyMatches[cyK] = strings.ReplaceAll(cyMatches[cyK], " ", "")
			}

			if !slices.Equal(enMatches, cyMatches) {
				t.Fail()
				t.Logf("missing translation variable in %s en: %v | cy: %v", k, enMatches, cyMatches)
			}
		}
	}
}

func loadTranslations(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var v map[string]string
	json.Unmarshal(data, &v)

	return v
}
