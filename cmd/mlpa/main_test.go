package main

import (
	"bytes"
	"encoding/json"
	"io"
	"iter"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

var headingTagRe = regexp.MustCompile(`<\s*(/\s*h\d\s*|h\d.+?)\s*>`)

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

	for k, v := range en.Flat() {
		if strings.Contains(v, "'") {
			t.Fail()
			t.Log("lang/en.json:", k)
		}
	}

	for k, v := range cy.Flat() {
		if strings.Contains(v, "'") {
			t.Fail()
			t.Log("lang/cy.json:", k)
		}
	}
}

func TestNoJSON(t *testing.T) {
	for _, path := range []string{"lang/en.json", "lang/cy.json"} {
		list := loadTranslations("../../" + path)

		for k, v := range list.Flat() {
			if strings.Contains(v, "\\\"") {
				t.Fail()
				t.Log(path, `contains \\\":`, k)
			}

			if strings.HasPrefix(v, "\"") {
				t.Fail()
				t.Log(path, `starts with \":`, k)
			}

			if strings.HasSuffix(v, "\",") {
				t.Fail()
				t.Log(path, `ends with \",:`, k)
			}
		}
	}
}

func TestNoWelshPossessive(t *testing.T) {
	cy := loadTranslations("../../lang/cy.json")
	possessiveRe := regexp.MustCompile(`{{\s*possessive`)

	for k, v := range cy.Flat() {
		if possessiveRe.MatchString(v) {
			t.Fail()
			t.Log("lang/cy.json:", k)
		}
	}
}

func TestTranslationVariablesMustMatch(t *testing.T) {
	en := loadTranslations("../../lang/en.json")
	cy := maps.Collect(loadTranslations("../../lang/cy.json").Flat())

	for k, enTranslation := range en.Flat() {
		if strings.Contains(enTranslation, "{{") {
			cyTranslation, _ := cy[k]

			r := regexp.MustCompile(`{{[^\.]*(\.[^ }]*)[ ]*}}`)
			enMatches := r.FindAllStringSubmatch(enTranslation, -1)
			cyMatches := r.FindAllStringSubmatch(cyTranslation, -1)

			enGroups := make([]string, len(enMatches))
			for i, matches := range enMatches {
				enGroups[i] = matches[1]
			}
			slices.Sort(enGroups)

			cyGroups := make([]string, len(cyMatches))
			for i, matches := range cyMatches {
				cyGroups[i] = matches[1]
			}
			slices.Sort(cyGroups)

			if !slices.Equal(enGroups, cyGroups) {
				t.Fail()
				t.Logf("missing translation variable in %s en: %v | cy: %v", k, enGroups, cyGroups)
			}
		}
	}
}

func TestTranslationContentMustMatch(t *testing.T) {
	en := maps.Collect(loadTranslations("../../lang/en.json").Flat())
	cy := maps.Collect(loadTranslations("../../lang/cy.json").Flat())

	mustMatch := map[string]string{
		"yourLegalRightsAndResponsibilitiesContent": "yourLegalRightsAndResponsibilitiesContent:h4",
	}

	for a, b := range mustMatch {
		assert.Equal(t, stripHeadings(en[a]), stripHeadings(en[b]))
		assert.Equal(t, stripHeadings(cy[a]), stripHeadings(cy[b]))
	}
}

func TestTranslationExternalLinksMustContainRelNoopenerNoreferrer(t *testing.T) {
	en := loadTranslations("../../lang/en.json")
	cy := loadTranslations("../../lang/cy.json")

	for k, v := range en.Flat() {
		if !externalLinksCorrect(v) {
			t.Fail()
			t.Log("lang/en.json:", k)
		}
	}

	for k, v := range cy.Flat() {
		if !externalLinksCorrect(v) {
			t.Fail()
			t.Log("lang/cy.json:", k)
		}
	}
}

func loadTranslations(path string) translationData {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var v translationData
	json.Unmarshal(data, &v)

	return v
}

type translationData map[string]any

func (d translationData) Flat() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for k, v := range d {
			switch vt := v.(type) {
			case string:
				if !yield(k, vt) {
					return
				}
			case map[string]any:
				for sk, sv := range vt {
					if !yield(k+"."+sk, sv.(string)) {
						return
					}
				}
			}
		}
	}
}

func TestStripHTML(t *testing.T) {
	assert.Equal(t, "<X>Hey<X><div>link</div><input />", stripHeadings(`<h1 class="what">Hey</h1><div>link</div><input />`))
}

func stripHeadings(s string) string {
	return headingTagRe.ReplaceAllString(s, "<X>")
}

func TestExternalLinksCorrect(t *testing.T) {
	assert.True(t, externalLinksCorrect(`No links here`))
	assert.True(t, externalLinksCorrect(`<a href="https://example.com" class="govuk-link" target="_blank" rel="noreferrer noopener">`))
	assert.True(t, externalLinksCorrect(`<a href="https://example.com" class="govuk-link" rel="noreferrer noopener"></a><a href="https://example.com" class="govuk-link" target="_blank" rel="noreferrer noopener"></a>`))
	assert.True(t, externalLinksCorrect(`<a href="/ok">`))

	assert.False(t, externalLinksCorrect(`<a href="https://example.com" class="govuk-link" target="_blank" rel="noreferrer">`))
	assert.False(t, externalLinksCorrect(`<a href="https://example.com" class="govuk-link" rel="noreferrer noopener"></a><a href="https://example.com" class="govuk-link" target="_blank" rel="noopener"></a>`))
}

func externalLinksCorrect(s string) bool {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return true
	}

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			var (
				href string
				rels []string
			)

			for _, a := range n.Attr {
				switch a.Key {
				case "href":
					href = a.Val
				case "rel":
					rels = strings.Fields(a.Val)
				}
			}

			if strings.HasPrefix(href, "https://") {
				if !slices.Contains(rels, "noreferrer") || !slices.Contains(rels, "noopener") {
					return false
				}
			}
		}
	}

	return true
}
