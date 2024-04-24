package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var (
	ctx           = context.Background()
	expectedError = errors.New("err")
)

type mockCredentialsProvider struct{}

func (m *mockCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     "abc",
		SecretAccessKey: "",
	}, nil
}

func (m *mockCredentialsProvider) IsExpired() bool {
	return false
}

func TestNewClient(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{}")
	}))
	defer s.Close()

	client, err := NewClient(aws.Config{
		Region:      "eu-west-1",
		Credentials: &mockCredentialsProvider{},
	}, s.URL, true)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestClientCreateIndices(t *testing.T) {
	indices := newMockIndicesClient(t)
	indices.EXPECT().
		Exists(ctx, opensearchapi.IndicesExistsReq{Indices: []string{indexName}}).
		Return(nil, expectedError)
	indices.EXPECT().
		Create(ctx, opensearchapi.IndicesCreateReq{Index: indexName, Body: strings.NewReader(indexDefinition)}).
		Return(nil, nil)

	client := &Client{indices: indices, indexingEnabled: true}
	err := client.CreateIndices(ctx)
	assert.Nil(t, err)
}

func TestClientCreateIndicesWhenCreateErrors(t *testing.T) {
	indices := newMockIndicesClient(t)
	indices.EXPECT().
		Exists(ctx, mock.Anything).
		Return(nil, expectedError)
	indices.EXPECT().
		Create(ctx, mock.Anything).
		Return(nil, expectedError)

	client := &Client{indices: indices, indexingEnabled: true}
	err := client.CreateIndices(ctx)
	assert.ErrorIs(t, err, expectedError)
}

func TestClientCreateIndicesWhenExists(t *testing.T) {
	indices := newMockIndicesClient(t)
	indices.EXPECT().
		Exists(ctx, mock.Anything).
		Return(nil, nil)

	client := &Client{indices: indices, indexingEnabled: true}
	err := client.CreateIndices(ctx)
	assert.Nil(t, err)
}

func TestClientIndex(t *testing.T) {
	svc := newMockOpensearchapiClient(t)
	svc.EXPECT().
		Index(ctx, opensearchapi.IndexReq{
			Index:      indexName,
			DocumentID: "LPA--2020",
			Body:       bytes.NewReader([]byte(`{"DonorFullName":"x y","PK":"LPA#2020","SK":"abc#123"}`)),
		}).
		Return(nil, nil)

	client := &Client{svc: svc, indexingEnabled: true}
	err := client.Index(ctx, Lpa{DonorFullName: "x y", PK: dynamo.LpaKey("2020").PK(), SK: "abc#123"})
	assert.Nil(t, err)
}

func TestClientIndexWhenNotEnabled(t *testing.T) {
	client := &Client{}
	err := client.Index(ctx, Lpa{DonorFullName: "x y", PK: dynamo.LpaKey("2020").PK(), SK: "abc#123"})
	assert.Nil(t, err)
}

func TestClientIndexWhenIndexErrors(t *testing.T) {
	svc := newMockOpensearchapiClient(t)
	svc.EXPECT().
		Index(ctx, mock.Anything).
		Return(nil, expectedError)

	client := &Client{svc: svc, indexingEnabled: true}
	err := client.Index(ctx, Lpa{DonorFullName: "x y", PK: dynamo.LpaKey("2020").PK(), SK: "abc#123"})
	assert.Equal(t, expectedError, err)
}

func TestClientQuery(t *testing.T) {
	testcases := map[string]struct {
		session *page.SessionData
		sk      dynamo.SK
		from    int
		page    int
	}{
		"donor": {
			session: &page.SessionData{SessionID: "abc"},
			sk:      dynamo.DonorKey("abc"),
			from:    0,
			page:    1,
		},
		"organisation": {
			session: &page.SessionData{SessionID: "abc", OrganisationID: "xyz"},
			sk:      dynamo.OrganisationKey("xyz"),
			from:    0,
			page:    1,
		},
		"donor paged": {
			session: &page.SessionData{SessionID: "abc"},
			sk:      dynamo.DonorKey("abc"),
			from:    40,
			page:    5,
		},
		"organisation paged": {
			session: &page.SessionData{SessionID: "abc", OrganisationID: "xyz"},
			sk:      dynamo.OrganisationKey("xyz"),
			from:    40,
			page:    5,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(ctx, tc.session)

			resp := &opensearchapi.SearchResp{}
			resp.Hits.Total.Value = 10
			resp.Hits.Hits = []opensearchapi.SearchHit{
				{Source: json.RawMessage(`{"PK":"LPA#123","SK":"DONOR#456"}`)},
				{Source: json.RawMessage(`{"PK":"LPA#456","SK":"DONOR#789"}`)},
			}

			svc := newMockOpensearchapiClient(t)
			svc.EXPECT().
				Search(ctx, &opensearchapi.SearchReq{
					Indices: []string{indexName},
					Body:    bytes.NewReader([]byte(fmt.Sprintf(`{"query":{"match":{"SK":"%s"}}}`, tc.sk.SK()))),
					Params: opensearchapi.SearchParams{
						From: aws.Int(tc.from),
						Size: aws.Int(10),
						Sort: []string{"DonorFullName"},
					},
				}).
				Return(resp, nil)

			client := &Client{svc: svc}
			result, err := client.Query(ctx, QueryRequest{Page: tc.page, PageSize: 10})
			assert.Nil(t, err)
			assert.Equal(t, result, &QueryResponse{
				Pagination: newPagination(10, tc.page, 10),
				Keys: []dynamo.Keys{
					{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")},
					{PK: dynamo.LpaKey("456"), SK: dynamo.DonorKey("789")},
				},
			})
		})
	}
}

func TestClientQueryWhenResponseInvalid(t *testing.T) {
	ctx := page.ContextWithSessionData(ctx, &page.SessionData{SessionID: "abc"})

	resp := &opensearchapi.SearchResp{}
	resp.Hits.Total.Value = 10
	resp.Hits.Hits = []opensearchapi.SearchHit{
		{Source: json.RawMessage(`{"PK":"abc#123`)},
	}

	svc := newMockOpensearchapiClient(t)
	svc.EXPECT().
		Search(ctx, mock.Anything).
		Return(resp, nil)

	client := &Client{svc: svc}
	_, err := client.Query(ctx, QueryRequest{Page: 1, PageSize: 10})
	assert.NotNil(t, err)
}

func TestClientQueryWhenSearchErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(ctx, &page.SessionData{SessionID: "1"})

	svc := newMockOpensearchapiClient(t)
	svc.EXPECT().
		Search(ctx, mock.Anything).
		Return(nil, expectedError)

	client := &Client{svc: svc}
	_, err := client.Query(ctx, QueryRequest{Page: 1, PageSize: 10})
	assert.Equal(t, expectedError, err)
}

func TestClientQueryWhenNoSession(t *testing.T) {
	client := &Client{}
	_, err := client.Query(ctx, QueryRequest{Page: 1, PageSize: 10})
	assert.ErrorIs(t, err, page.SessionMissingError{})
}

func TestClientCountWithQuery(t *testing.T) {
	testcases := map[string]struct {
		query   CountWithQueryReq
		body    []byte
		session *page.SessionData
	}{
		"no query - donor": {
			query:   CountWithQueryReq{},
			body:    []byte(`{"query":{"bool":{"must":{"match":{"SK":"DONOR#1"}}}},"size":0,"track_total_hits":true}`),
			session: &page.SessionData{SessionID: "1"},
		},
		"no query - organisation": {
			query:   CountWithQueryReq{},
			body:    []byte(`{"query":{"bool":{"must":{"match":{"SK":"ORGANISATION#1"}}}},"size":0,"track_total_hits":true}`),
			session: &page.SessionData{OrganisationID: "1"},
		},
		"MustNotExist query - donor": {
			query:   CountWithQueryReq{MustNotExist: "a-field"},
			body:    []byte(`{"query":{"bool":{"must":{"match":{"SK":"DONOR#1"}},"must_not":{"exists":{"field":"a-field"}}}},"size":0,"track_total_hits":true}`),
			session: &page.SessionData{SessionID: "1"},
		},
		"MustNotExist query - organisation": {
			query:   CountWithQueryReq{MustNotExist: "a-field"},
			body:    []byte(`{"query":{"bool":{"must":{"match":{"SK":"ORGANISATION#1"}},"must_not":{"exists":{"field":"a-field"}}}},"size":0,"track_total_hits":true}`),
			session: &page.SessionData{OrganisationID: "1"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(ctx, tc.session)
			resp := &opensearchapi.SearchResp{}
			resp.Hits.Total.Value = 1

			svc := newMockOpensearchapiClient(t)
			svc.EXPECT().
				Search(ctx, &opensearchapi.SearchReq{
					Indices: []string{indexName},
					Body:    bytes.NewReader(tc.body),
				}).
				Return(resp, nil)

			client := &Client{
				svc: svc,
			}

			count, err := client.CountWithQuery(ctx, tc.query)

			assert.Nil(t, err)
			assert.Equal(t, 1, count)
		})
	}
}

func TestClientCountWithQueryWhenNoSession(t *testing.T) {
	client := &Client{}
	_, err := client.CountWithQuery(ctx, CountWithQueryReq{})

	assert.ErrorIs(t, err, page.SessionMissingError{})
}

func TestClientCountWithQueryWhenSearchError(t *testing.T) {
	svc := newMockOpensearchapiClient(t)
	svc.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return(&opensearchapi.SearchResp{}, expectedError)

	client := &Client{
		svc: svc,
	}

	ctx := page.ContextWithSessionData(ctx, &page.SessionData{SessionID: "1"})
	_, err := client.CountWithQuery(ctx, CountWithQueryReq{})

	assert.Error(t, err)
}
