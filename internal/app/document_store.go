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

type EventClient interface {
	SendUidRequested(context.Context, event.UidRequested) error
	SendApplicationUpdated(context.Context, event.ApplicationUpdated) error
	SendPreviousApplicationLinked(context.Context, event.PreviousApplicationLinked) error
	SendReducedFeeRequested(context.Context, event.ReducedFeeRequested) error
}

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

func (s *documentStore) Create(ctx context.Context, donor *actor.DonorProvidedDetails, filename string, data []byte) (page.Document, error) {
	key := donor.LpaUID + "/evidence/" + s.randomUUID()

	document := page.Document{
		PK:       dynamo.LpaKey(donor.LpaID),
		SK:       dynamo.DocumentKey(key),
		Filename: filename,
		Key:      key,
		Uploaded: s.now(),
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
	if err := s.dynamoClient.AllByPartialSK(ctx, dynamo.LpaKey(data.LpaID), dynamo.DocumentKey(""), &ds); err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return nil, err
	}

	return ds, nil
}

func (s *documentStore) UpdateScanResults(ctx context.Context, lpaID, s3ObjectKey string, virusDetected bool) error {
	return s.dynamoClient.Update(ctx,
		dynamo.LpaKey(lpaID),
		dynamo.DocumentKey(s3ObjectKey),
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
	var dynamoKeys []dynamo.Keys

	for _, d := range documents {
		if d.VirusDetected {
			dynamoKeys = append(dynamoKeys, dynamo.Keys{PK: d.PK, SK: d.SK})
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

func (s *documentStore) Submit(ctx context.Context, donor *actor.DonorProvidedDetails, documents page.Documents) error {
	var unsentDocuments []any
	var unsentEvidence []event.Evidence

	tags := map[string]string{
		"replicate":         "true",
		"virus-scan-status": "ok",
	}

	for _, document := range documents {
		if document.Sent.IsZero() && !document.VirusDetected {
			document.Sent = s.now()
			unsentDocuments = append(unsentDocuments, document)
			unsentEvidence = append(unsentEvidence, event.Evidence{
				Path:     document.Key,
				Filename: document.Filename,
			})

			if err := s.s3Client.PutObjectTagging(ctx, document.Key, tags); err != nil {
				return err
			}
		}
	}

	if len(unsentDocuments) > 0 {
		if err := s.eventClient.SendReducedFeeRequested(ctx, event.ReducedFeeRequested{
			UID:              donor.LpaUID,
			RequestType:      donor.FeeType.String(),
			Evidence:         unsentEvidence,
			EvidenceDelivery: donor.EvidenceDelivery.String(),
		}); err != nil {
			return err
		}

		if err := s.dynamoClient.BatchPut(ctx, unsentDocuments); err != nil {
			return err
		}
	}

	return nil
}
