package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v4/signer/awsv2"
)

var indexDefinition = map[string]any{
	"settings": map[string]any{
		"index": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	},
	"mappings": map[string]any{
		"properties": map[string]any{
			"PK":               map[string]any{"type": "keyword"},
			"SK":               map[string]any{"type": "keyword"},
			"Donor.FirstNames": map[string]any{"type": "keyword"},
			"Donor.LastName":   map[string]any{"type": "keyword"},
		},
	},
}

type opensearchapiClient interface {
	Search(ctx context.Context, req *opensearchapi.SearchReq) (*opensearchapi.SearchResp, error)
	Index(ctx context.Context, req opensearchapi.IndexReq) (*opensearchapi.IndexResp, error)
}

type indicesClient interface {
	Exists(ctx context.Context, req opensearchapi.IndicesExistsReq) (*opensearch.Response, error)
	Create(ctx context.Context, req opensearchapi.IndicesCreateReq) (*opensearchapi.IndicesCreateResp, error)
}

type QueryResponse struct {
	Pagination *Pagination
	Keys       []dynamo.Keys
}

type Lpa struct {
	PK    string
	SK    string
	Donor LpaDonor
}

type LpaDonor struct {
	FirstNames string
	LastName   string
}

type QueryRequest struct {
	Page     int
	PageSize int
}

type Client struct {
	svc             opensearchapiClient
	indices         indicesClient
	endpoint        string
	indexName       string
	indexingEnabled bool
}

func NewClient(cfg aws.Config, endpoint, indexName string, indexingEnabled bool) (*Client, error) {
	signer, err := requestsigner.NewSignerWithService(cfg, "aoss")
	if err != nil {
		return nil, fmt.Errorf("search could not create signer: %w", err)
	}

	svc, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Addresses: []string{endpoint},
			Signer:    signer,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("search could not create opensearch client: %w", err)
	}

	return &Client{indices: svc.Indices, svc: svc, endpoint: endpoint, indexName: indexName, indexingEnabled: indexingEnabled}, nil
}

func (c *Client) CreateIndices(ctx context.Context) error {
	body, err := json.Marshal(indexDefinition)
	if err != nil {
		return err
	}

	if _, err := c.indices.Exists(ctx, opensearchapi.IndicesExistsReq{Indices: []string{c.indexName}}); err == nil {
		return nil
	}

	if _, err := c.indices.Create(ctx, opensearchapi.IndicesCreateReq{Index: c.indexName, Body: bytes.NewReader(body)}); err != nil {
		return fmt.Errorf("search could not create index: %w", err)
	}

	return nil
}

func (c *Client) Index(ctx context.Context, lpa Lpa) error {
	if !c.indexingEnabled {
		return nil
	}

	body, err := json.Marshal(lpa)
	if err != nil {
		return err
	}

	_, err = c.svc.Index(ctx, opensearchapi.IndexReq{
		Index:      c.indexName,
		DocumentID: strings.ReplaceAll(lpa.PK, "#", "--"),
		Body:       bytes.NewReader(body),
	})

	return err
}

func (c *Client) Query(ctx context.Context, req QueryRequest) (*QueryResponse, error) {
	sk, err := getSKFromContext(ctx)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]any{
		"query": baseQuery(sk),
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.svc.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{c.indexName},
		Body:    bytes.NewReader(body),
		Params: opensearchapi.SearchParams{
			From: aws.Int((req.Page - 1) * req.PageSize),
			Size: aws.Int(req.PageSize),
			Sort: []string{"Donor.FirstNames", "Donor.LastName"},
		},
	})
	if err != nil {
		return nil, err
	}

	var keys []dynamo.Keys
	for _, hit := range resp.Hits.Hits {
		var key dynamo.Keys
		if err := json.Unmarshal(hit.Source, &key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return &QueryResponse{
		Pagination: newPagination(resp.Hits.Total.Value, req.Page, req.PageSize),
		Keys:       keys,
	}, nil
}

type CountWithQueryReq struct {
	MustNotExist string
}

func (c *Client) CountWithQuery(ctx context.Context, req CountWithQueryReq) (int, error) {
	sk, err := getSKFromContext(ctx)
	if err != nil {
		return 0, err
	}

	queryBody := map[string]any{
		"size":             0,
		"track_total_hits": true,
	}

	query := baseQuery(sk)

	if req.MustNotExist != "" {
		query["bool"]["must_not"] = map[string]any{
			"exists": map[string]any{
				"field": req.MustNotExist,
			},
		}
	}

	queryBody["query"] = query

	body, err := json.Marshal(queryBody)
	if err != nil {
		return 0, err
	}

	resp, err := c.svc.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{c.indexName},
		Body:    bytes.NewReader(body),
	})
	if err != nil {
		return 0, err
	}

	return resp.Hits.Total.Value, err
}

func baseQuery(sk string) map[string]map[string]any {
	return map[string]map[string]any{
		"bool": {
			"must": []map[string]any{
				{
					"match": map[string]string{
						"SK": sk,
					},
				},
				{
					"prefix": map[string]string{
						"PK": dynamo.LpaKey("").PK(),
					},
				},
			},
		},
	}
}

func getSKFromContext(ctx context.Context) (string, error) {
	session, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return "", err
	}

	var sk dynamo.SK = dynamo.DonorKey(session.SessionID)
	if session.OrganisationID != "" {
		sk = dynamo.OrganisationKey(session.OrganisationID)
	}

	return sk.SK(), nil
}
