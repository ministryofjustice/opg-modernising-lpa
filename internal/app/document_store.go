package app

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type documentStore struct {
	dynamoClient DynamoClient
	s3Client     S3Client
	eventClient  EventClient
	randomUUID   func() string
	now          func() time.Time
}

func NewDocumentStore(dynamoClient DynamoClient, s3Client S3Client, eventClient EventClient, randomUUID func() string, now func() time.Time) *documentStore {
	return &documentStore{
		dynamoClient: dynamoClient,
		s3Client:     s3Client,
		eventClient:  eventClient,
		randomUUID:   randomUUID,
		now:          now,
	}
}

func (s *documentStore) Create(ctx context.Context, lpa *actor.Lpa, filename string, data []byte) (page.Document, error) {
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

func (s *documentStore) Submit(ctx context.Context, lpa *actor.Lpa, documents page.Documents) error {
	var unsentDocuments []any
	var unsentDocumentKeys []string

	tags := map[string]string{
		"replicate":         "true",
		"virus-scan-status": "ok",
	}

	for _, document := range documents {
		if document.Sent.IsZero() && !document.VirusDetected {
			document.Sent = s.now()
			unsentDocuments = append(unsentDocuments, document)
			unsentDocumentKeys = append(unsentDocumentKeys, document.Key)

			if err := s.s3Client.PutObjectTagging(ctx, document.Key, tags); err != nil {
				return err
			}
		}
	}

	if len(unsentDocuments) > 0 {
		if err := s.eventClient.SendReducedFeeRequested(ctx, event.ReducedFeeRequested{
			UID:              lpa.UID,
			RequestType:      lpa.FeeType.String(),
			Evidence:         unsentDocumentKeys,
			EvidenceDelivery: lpa.EvidenceDelivery.String(),
		}); err != nil {
			return err
		}

		if err := s.dynamoClient.BatchPut(ctx, unsentDocuments); err != nil {
			return err
		}
	}

	return nil
}

func documentKey(s3Key string) string {
	return "#DOCUMENT#" + s3Key
}
