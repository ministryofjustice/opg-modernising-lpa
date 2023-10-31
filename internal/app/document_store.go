package app

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type DocumentStore struct {
	dynamoClient DynamoClient
	s3Client     S3Client
}

func NewDocumentStore(dynamoClient DynamoClient, s3Client S3Client) DocumentStore {
	return DocumentStore{dynamoClient: dynamoClient, s3Client: s3Client}
}

func (s DocumentStore) GetAll(ctx context.Context) (page.Documents, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return []page.Document{}, err
	}

	if data.LpaID == "" {
		return []page.Document{}, errors.New("documentStore.GetAll requires LpaID")
	}

	var ds []page.Document
	if err := s.dynamoClient.AllByPartialSk(ctx, lpaKey(data.LpaID), DocumentKey(""), &ds); err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			return []page.Document{}, nil
		}

		return []page.Document{}, err
	}

	return ds, nil
}

func (s DocumentStore) UpdateScanResults(ctx context.Context, PK, SK string, virusDetected bool) error {
	input := &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: PK},
			"SK": &types.AttributeValueMemberS{Value: SK},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":virusDetected": &types.AttributeValueMemberBOOL{Value: virusDetected},
			":scanned":       &types.AttributeValueMemberBOOL{Value: true},
		},
		UpdateExpression: aws.String("set VirusDetected = :virusDetected, Scanned = :scanned"),
	}

	return s.dynamoClient.Update(ctx, input)
}

func (s DocumentStore) Put(ctx context.Context, document page.Document, data []byte) error {
	if data != nil {
		if err := s.s3Client.PutObject(ctx, document.Key, data); err != nil {
			return err
		}
	}

	return s.dynamoClient.Put(ctx, document)
}

func (s DocumentStore) DeleteInfectedDocuments(ctx context.Context, documents []page.Document) error {
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

	if err := s.s3Client.DeleteObjects(ctx, s3Keys); err != nil {
		return err
	}

	return s.dynamoClient.DeleteKeys(ctx, dynamoKeys)
}

func (s DocumentStore) Delete(ctx context.Context, document page.Document) error {
	if err := s.s3Client.DeleteObject(ctx, document.Key); err != nil {
		return err
	}

	return s.dynamoClient.DeleteOne(ctx, dynamo.Key{
		PK: document.PK,
		SK: document.SK,
	})
}

func DocumentKey(s3Key string) string {
	return "#DOCUMENT#" + s3Key
}
