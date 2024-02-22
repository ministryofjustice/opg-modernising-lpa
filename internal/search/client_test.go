package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func TestClientCreateIndices(t *testing.T) {
	indices := newMockIndicesClient(t)
	indices.EXPECT().
		Exists(ctx, opensearchapi.IndicesExistsReq{Indices: []string{indexName}}).
		Return(nil, expectedError)
	indices.EXPECT().
		Create(ctx, opensearchapi.IndicesCreateReq{Index: indexName, Body: strings.NewReader(indexDefinition)}).
		Return(nil, nil)

	client := &Client{indices: indices}
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

	client := &Client{indices: indices}
	err := client.CreateIndices(ctx)
	assert.ErrorIs(t, err, expectedError)
}

func TestClientCreateIndicesWhenExists(t *testing.T) {
	indices := newMockIndicesClient(t)
	indices.EXPECT().
		Exists(ctx, mock.Anything).
		Return(nil, nil)

	client := &Client{indices: indices}
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

	client := &Client{svc: svc}
	err := client.Index(ctx, Lpa{DonorFullName: "x y", PK: "LPA#2020", SK: "abc#123"})
	assert.Nil(t, err)
}

func TestClientIndexWhenIndexErrors(t *testing.T) {
	svc := newMockOpensearchapiClient(t)
	svc.EXPECT().
		Index(ctx, mock.Anything).
		Return(nil, expectedError)

	client := &Client{svc: svc}
	err := client.Index(ctx, Lpa{DonorFullName: "x y", PK: "LPA#2020", SK: "abc#123"})
	assert.Equal(t, expectedError, err)
}

func TestClientQuery(t *testing.T) {
	testcases := map[string]struct {
		session *page.SessionData
		sk      string
		from    int
		page    int
	}{
		"donor": {
			session: &page.SessionData{SessionID: "abc"},
			sk:      "#DONOR#abc",
			from:    0,
			page:    1,
		},
		"organisation": {
			session: &page.SessionData{SessionID: "abc", OrganisationID: "xyz"},
			sk:      "ORGANISATION#xyz",
			from:    0,
			page:    1,
		},
		"donor paged": {
			session: &page.SessionData{SessionID: "abc"},
			sk:      "#DONOR#abc",
			from:    40,
			page:    5,
		},
		"organisation paged": {
			session: &page.SessionData{SessionID: "abc", OrganisationID: "xyz"},
			sk:      "ORGANISATION#xyz",
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
				{Source: json.RawMessage(`{"PK":"abc#123","SK":"xyz#456"}`)},
				{Source: json.RawMessage(`{"PK":"abc#456","SK":"xyz#789"}`)},
			}

			svc := newMockOpensearchapiClient(t)
			svc.EXPECT().
				Search(ctx, &opensearchapi.SearchReq{
					Indices: []string{indexName},
					Body:    bytes.NewReader([]byte(fmt.Sprintf(`{"query":{"match":{"SK":"%s"}}}`, tc.sk))),
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
				Keys: []dynamo.Key{
					{PK: "abc#123", SK: "xyz#456"},
					{PK: "abc#456", SK: "xyz#789"},
				},
			})
		})
	}
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
