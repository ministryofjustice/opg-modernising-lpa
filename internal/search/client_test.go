package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var (
	ctx           = context.Background()
	expectedError = errors.New("err")
	testIndexName = "index"
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
	}, s.URL, testIndexName, true)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestClientCreateIndices(t *testing.T) {
	data, _ := json.Marshal(indexDefinition)

	indices := newMockIndicesClient(t)
	indices.EXPECT().
		Exists(ctx, opensearchapi.IndicesExistsReq{Indices: []string{testIndexName}}).
		Return(nil, expectedError)
	indices.EXPECT().
		Create(ctx, opensearchapi.IndicesCreateReq{Index: testIndexName, Body: bytes.NewReader(data)}).
		Return(nil, nil)

	client := &Client{indices: indices, indexName: testIndexName, indexingEnabled: true}
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
			Index:      testIndexName,
			DocumentID: "LPA--2020",
			Body:       bytes.NewReader([]byte(`{"PK":"LPA#2020","SK":"abc#123","Donor":{"FirstNames":"x","LastName":"y"}}`)),
		}).
		Return(nil, nil)

	client := &Client{svc: svc, indexName: testIndexName, indexingEnabled: true}
	err := client.Index(ctx, Lpa{Donor: LpaDonor{FirstNames: "x", LastName: "y"}, PK: dynamo.LpaKey("2020").PK(), SK: "abc#123"})
	assert.Nil(t, err)
}

func TestClientIndexWhenNotEnabled(t *testing.T) {
	client := &Client{}
	err := client.Index(ctx, Lpa{Donor: LpaDonor{FirstNames: "x", LastName: "y"}, PK: dynamo.LpaKey("2020").PK(), SK: "abc#123"})
	assert.Nil(t, err)
}

func TestClientIndexWhenIndexErrors(t *testing.T) {
	svc := newMockOpensearchapiClient(t)
	svc.EXPECT().
		Index(ctx, mock.Anything).
		Return(nil, expectedError)

	client := &Client{svc: svc, indexingEnabled: true}
	err := client.Index(ctx, Lpa{Donor: LpaDonor{FirstNames: "x", LastName: "y"}, PK: dynamo.LpaKey("2020").PK(), SK: "abc#123"})
	assert.Equal(t, expectedError, err)
}

func TestClientQuery(t *testing.T) {
	testcases := map[string]struct {
		session *appcontext.SessionData
		sk      dynamo.SK
		from    int
		page    int
	}{
		"donor": {
			session: &appcontext.SessionData{SessionID: "abc"},
			sk:      dynamo.DonorKey("abc"),
			from:    0,
			page:    1,
		},
		"organisation": {
			session: &appcontext.SessionData{SessionID: "abc", OrganisationID: "xyz"},
			sk:      dynamo.OrganisationKey("xyz"),
			from:    0,
			page:    1,
		},
		"donor paged": {
			session: &appcontext.SessionData{SessionID: "abc"},
			sk:      dynamo.DonorKey("abc"),
			from:    40,
			page:    5,
		},
		"organisation paged": {
			session: &appcontext.SessionData{SessionID: "abc", OrganisationID: "xyz"},
			sk:      dynamo.OrganisationKey("xyz"),
			from:    40,
			page:    5,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSessionData(ctx, tc.session)

			resp := &opensearchapi.SearchResp{}
			resp.Hits.Total.Value = 10
			resp.Hits.Hits = []opensearchapi.SearchHit{
				{Source: json.RawMessage(`{"PK":"LPA#123","SK":"DONOR#456"}`)},
				{Source: json.RawMessage(`{"PK":"LPA#456","SK":"DONOR#789"}`)},
			}

			svc := newMockOpensearchapiClient(t)
			svc.EXPECT().
				Search(ctx, &opensearchapi.SearchReq{
					Indices: []string{testIndexName},
					Body:    bytes.NewReader([]byte(fmt.Sprintf(`{"query":{"bool":{"must":[{"match":{"SK":"%s"}},{"prefix":{"PK":"LPA#"}}]}}}`, tc.sk.SK()))),
					Params: opensearchapi.SearchParams{
						From: aws.Int(tc.from),
						Size: aws.Int(10),
						Sort: []string{"Donor.FirstNames", "Donor.LastName"},
					},
				}).
				Return(resp, nil)

			client := &Client{svc: svc, indexName: testIndexName}
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
	ctx := appcontext.ContextWithSessionData(ctx, &appcontext.SessionData{SessionID: "abc"})

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
	ctx := appcontext.ContextWithSessionData(ctx, &appcontext.SessionData{SessionID: "1"})

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
	assert.ErrorIs(t, err, appcontext.SessionMissingError{})
}

func TestClientCountWithQuery(t *testing.T) {
	testcases := map[string]struct {
		query   CountWithQueryReq
		body    []byte
		session *appcontext.SessionData
	}{
		"no query - donor": {
			query:   CountWithQueryReq{},
			body:    []byte(`{"query":{"bool":{"must":[{"match":{"SK":"DONOR#1"}},{"prefix":{"PK":"LPA#"}}]}},"size":0,"track_total_hits":true}`),
			session: &appcontext.SessionData{SessionID: "1"},
		},
		"no query - organisation": {
			query:   CountWithQueryReq{},
			body:    []byte(`{"query":{"bool":{"must":[{"match":{"SK":"ORGANISATION#1"}},{"prefix":{"PK":"LPA#"}}]}},"size":0,"track_total_hits":true}`),
			session: &appcontext.SessionData{OrganisationID: "1"},
		},
		"MustNotExist query - donor": {
			query:   CountWithQueryReq{MustNotExist: "a-field"},
			body:    []byte(`{"query":{"bool":{"must":[{"match":{"SK":"DONOR#1"}},{"prefix":{"PK":"LPA#"}}],"must_not":{"exists":{"field":"a-field"}}}},"size":0,"track_total_hits":true}`),
			session: &appcontext.SessionData{SessionID: "1"},
		},
		"MustNotExist query - organisation": {
			query:   CountWithQueryReq{MustNotExist: "a-field"},
			body:    []byte(`{"query":{"bool":{"must":[{"match":{"SK":"ORGANISATION#1"}},{"prefix":{"PK":"LPA#"}}],"must_not":{"exists":{"field":"a-field"}}}},"size":0,"track_total_hits":true}`),
			session: &appcontext.SessionData{OrganisationID: "1"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSessionData(ctx, tc.session)
			resp := &opensearchapi.SearchResp{}
			resp.Hits.Total.Value = 1

			svc := newMockOpensearchapiClient(t)
			svc.EXPECT().
				Search(ctx, &opensearchapi.SearchReq{
					Indices: []string{testIndexName},
					Body:    bytes.NewReader(tc.body),
				}).
				Return(resp, nil)

			client := &Client{
				svc:       svc,
				indexName: testIndexName,
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

	assert.ErrorIs(t, err, appcontext.SessionMissingError{})
}

func TestClientCountWithQueryWhenSearchError(t *testing.T) {
	svc := newMockOpensearchapiClient(t)
	svc.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return(&opensearchapi.SearchResp{}, expectedError)

	client := &Client{
		svc: svc,
	}

	ctx := appcontext.ContextWithSessionData(ctx, &appcontext.SessionData{SessionID: "1"})
	_, err := client.CountWithQuery(ctx, CountWithQueryReq{})

	assert.Error(t, err)
}
