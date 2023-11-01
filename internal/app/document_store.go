package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type PartialBatchWriteError struct {
	Written  int
	Expected int
}

func (e PartialBatchWriteError) Error() string {
	return fmt.Sprintf("Expected to write %d but %d were written", e.Expected, e.Written)
}

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
	var converted []interface{}
	for _, d := range documents {
		converted = append(converted, d)
	}

	toWrite := len(converted)
	written, err := s.dynamoClient.BatchPut(ctx, converted)

	if err != nil {
		return err
	} else if written != toWrite {
		return PartialBatchWriteError{Written: written, Expected: toWrite}
	}

	return nil
}

func (s *documentStore) Put(ctx context.Context, document page.Document) error {
	return s.dynamoClient.Put(ctx, document)
}

func (s *documentStore) DeleteInfectedDocuments(ctx context.Context, documents page.Documents) error {
	var dynamoKeys []dynamo.Key
	var s3Keys []string

	for _, d := range documents {
		if d.VirusDetected {
			dynamoKeys = append(dynamoKeys, dynamo.Key{
				PK: d.PK,
				SK: d.SK,
			})
			s3Keys = append(s3Keys, d.Key)
		}
	}

	if len(dynamoKeys) == 0 {
		return nil
	}

	if err := s.s3Client.DeleteObjects(ctx, s3Keys); err != nil {
		return err
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
