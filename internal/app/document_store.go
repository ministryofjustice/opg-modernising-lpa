package app

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type documentStore struct {
	dynamoClient DynamoClient
	s3Client     S3Client
	randomUUID   func() string
}

func NewDocumentStore(dynamoClient DynamoClient, s3Client S3Client, randomUUID func() string) *documentStore {
	return &documentStore{dynamoClient: dynamoClient, s3Client: s3Client, randomUUID: randomUUID}
}

func (s *documentStore) Create(ctx context.Context, lpa *page.Lpa, filename string, data []byte) (page.Document, error) {
	key := lpa.UID + "/evidence/" + s.randomUUID()

	document := page.Document{
		PK:       lpaKey(lpa.ID),
		SK:       documentKey(key),
		Filename: filename,
		Key:      key,
	}

	if err := s.s3Client.PutObject(ctx, document.Key, data); err != nil {
		return page.Document{}, err
	}

	if err := s.dynamoClient.Create(ctx, document); err != nil {
		return page.Document{}, err
	}

	return document, nil
}

func (s *documentStore) GetAll(ctx context.Context) (page.Documents, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("documentStore.GetAll requires LpaID")
	}

	var ds []page.Document
	if err := s.dynamoClient.AllByPartialSk(ctx, lpaKey(data.LpaID), documentKey(""), &ds); err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return nil, err
	}

	return ds, nil
}

func (s *documentStore) UpdateScanResults(ctx context.Context, lpaID, s3ObjectKey string, virusDetected bool) error {
	return s.dynamoClient.Update(ctx,
		lpaKey(lpaID),
		documentKey(s3ObjectKey),
		map[string]types.AttributeValue{
			":virusDetected": &types.AttributeValueMemberBOOL{Value: virusDetected},
			":scanned":       &types.AttributeValueMemberBOOL{Value: true},
		},
		"set VirusDetected = :virusDetected, Scanned = :scanned")
}

func (s *documentStore) BatchPut(ctx context.Context, documents []page.Document) error {
	var converted []any
	for _, d := range documents {
		converted = append(converted, d)
	}

	return s.dynamoClient.BatchPut(ctx, converted)
}

func (s *documentStore) Put(ctx context.Context, document page.Document) error {
	return s.dynamoClient.Put(ctx, document)
}

func (s *documentStore) DeleteInfectedDocuments(ctx context.Context, documents page.Documents) error {
	var dynamoKeys []dynamo.Key

	for _, d := range documents {
		if d.VirusDetected {
			dynamoKeys = append(dynamoKeys, dynamo.Key{PK: d.PK, SK: d.SK})
		}
	}

	if len(dynamoKeys) == 0 {
		return nil
	}

	return s.dynamoClient.DeleteKeys(ctx, dynamoKeys)
}

func (s *documentStore) Delete(ctx context.Context, document page.Document) error {
	if err := s.s3Client.DeleteObject(ctx, document.Key); err != nil {
		return err
	}

	return s.dynamoClient.DeleteOne(ctx, document.PK, document.SK)
}

func documentKey(s3Key string) string {
	return "#DOCUMENT#" + s3Key
}
