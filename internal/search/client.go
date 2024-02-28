package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/opensearch-project/opensearch-go/v3"
	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v3/signer/awsv2"
)

const (
	indexName       = "lpas4"
	indexDefinition = `
	{
		"settings": {
			"index": {
				"number_of_shards": 1,
				"number_of_replicas": 0
			}
		},
		"mappings": {
			"properties": {
				"DonorFullNameText": { "type": "text" },
				"DonorFullName": { "type": "keyword", "copy_to": "DonorFullNameText" },
				"SK": { "type": "keyword" }
			}
		}
	}`
)

type opensearchapiClient interface {
	Search(ctx context.Context, req *opensearchapi.SearchReq) (*opensearchapi.SearchResp, error)
	Index(ctx context.Context, req opensearchapi.IndexReq) (*opensearchapi.IndexResp, error)
	Info(ctx context.Context, req *opensearchapi.InfoReq) (*opensearchapi.InfoResp, error)
}

type indicesClient interface {
	Exists(ctx context.Context, req opensearchapi.IndicesExistsReq) (*opensearch.Response, error)
	Create(ctx context.Context, req opensearchapi.IndicesCreateReq) (*opensearchapi.IndicesCreateResp, error)
}

type QueryResponse struct {
	Pagination *Pagination
	Keys       []dynamo.Key
}

type Lpa struct {
	DonorFullName string
	PK            string
	SK            string
}

type QueryRequest struct {
	Page     int
	PageSize int
}

type Client struct {
	svc      opensearchapiClient
	indices  indicesClient
	endpoint string
}

func NewClient(cfg aws.Config, endpoint string) (*Client, error) {
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

	return &Client{indices: svc.Indices, svc: svc, endpoint: endpoint}, nil
}

func (c *Client) CheckHealth(ctx context.Context) error {
	resp, err := http.Get(c.endpoint)
	if err != nil {
		return fmt.Errorf("search get error: %w", err)
	}
	data, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	_, err = c.svc.Info(ctx, &opensearchapi.InfoReq{})
	if err != nil {
		return fmt.Errorf("info req failed: %w: %s", err, string(data))
	}

	return nil
}

func (c *Client) CreateIndices(ctx context.Context) error {
	_, err := c.indices.Exists(ctx, opensearchapi.IndicesExistsReq{Indices: []string{indexName}})
	if err == nil {
		return nil
	}

	settings := strings.NewReader(indexDefinition)

	if _, err := c.indices.Create(ctx, opensearchapi.IndicesCreateReq{Index: indexName, Body: settings}); err != nil {
		return fmt.Errorf("search could not create index: %w", err)
	}

	return nil
}

func (c *Client) Index(ctx context.Context, lpa Lpa) error {
	body, err := json.Marshal(lpa)
	if err != nil {
		return err
	}

	_, err = c.svc.Index(ctx, opensearchapi.IndexReq{
		Index:      indexName,
		DocumentID: strings.ReplaceAll(lpa.PK, "#", "--"),
		Body:       bytes.NewReader(body),
	})

	return err
}

func (c *Client) Query(ctx context.Context, req QueryRequest) (*QueryResponse, error) {
	session, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	sk := "#DONOR#" + session.SessionID
	if session.OrganisationID != "" {
		sk = "ORGANISATION#" + session.OrganisationID
	}

	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"SK": sk,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.svc.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{indexName},
		Body:    bytes.NewReader(body),
		Params: opensearchapi.SearchParams{
			From: aws.Int((req.Page - 1) * req.PageSize),
			Size: aws.Int(req.PageSize),
			Sort: []string{"DonorFullName"},
		},
	})
	if err != nil {
		return nil, err
	}

	var keys []dynamo.Key
	for _, hit := range resp.Hits.Hits {
		var key dynamo.Key
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
