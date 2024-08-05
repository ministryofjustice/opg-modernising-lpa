package document

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

type DynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	OneByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	AllByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	LatestForActor(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllByKeys(ctx context.Context, keys []dynamo.Keys) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllKeysByPK(ctx context.Context, pk dynamo.PK) ([]dynamo.Keys, error)
	Put(ctx context.Context, v interface{}) error
	Create(ctx context.Context, v interface{}) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	Update(ctx context.Context, pk dynamo.PK, sk dynamo.SK, values map[string]dynamodbtypes.AttributeValue, expression string) error
	BatchPut(ctx context.Context, items []interface{}) error
	OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type S3Client interface {
	PutObject(context.Context, string, []byte) error
	DeleteObject(context.Context, string) error
	DeleteObjects(ctx context.Context, keys []string) error
	PutObjectTagging(context.Context, string, map[string]string) error
}

type EventClient interface {
	SendUidRequested(context.Context, event.UidRequested) error
	SendApplicationUpdated(context.Context, event.ApplicationUpdated) error
	SendPreviousApplicationLinked(context.Context, event.PreviousApplicationLinked) error
	SendReducedFeeRequested(context.Context, event.ReducedFeeRequested) error
}

type Store struct {
	dynamoClient DynamoClient
	s3Client     S3Client
	eventClient  EventClient
	randomUUID   func() string
	now          func() time.Time
}

func NewStore(dynamoClient DynamoClient, s3Client S3Client, eventClient EventClient) *Store {
	return &Store{
		dynamoClient: dynamoClient,
		s3Client:     s3Client,
		eventClient:  eventClient,
		randomUUID:   random.UuidString,
		now:          time.Now,
	}
}

func (s *Store) Create(ctx context.Context, donor *donordata.Provided, filename string, data []byte) (Document, error) {
	key := donor.LpaUID + "/evidence/" + s.randomUUID()

	doc := Document{
		PK:       dynamo.LpaKey(donor.LpaID),
		SK:       dynamo.DocumentKey(key),
		Filename: filename,
		Key:      key,
		Uploaded: s.now(),
	}

	if err := s.s3Client.PutObject(ctx, doc.Key, data); err != nil {
		return Document{}, err
	}

	if err := s.dynamoClient.Create(ctx, doc); err != nil {
		return Document{}, err
	}

	return doc, nil
}

func (s *Store) GetAll(ctx context.Context) (Documents, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("documentStore.GetAll requires LpaID")
	}

	var ds []Document
	if err := s.dynamoClient.AllByPartialSK(ctx, dynamo.LpaKey(data.LpaID), dynamo.DocumentKey(""), &ds); err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return nil, err
	}

	return ds, nil
}

func (s *Store) UpdateScanResults(ctx context.Context, lpaID, s3ObjectKey string, virusDetected bool) error {
	return s.dynamoClient.Update(ctx,
		dynamo.LpaKey(lpaID),
		dynamo.DocumentKey(s3ObjectKey),
		map[string]types.AttributeValue{
			":virusDetected": &types.AttributeValueMemberBOOL{Value: virusDetected},
			":scanned":       &types.AttributeValueMemberBOOL{Value: true},
		},
		"set VirusDetected = :virusDetected, Scanned = :scanned")
}

func (s *Store) Put(ctx context.Context, document Document) error {
	return s.dynamoClient.Put(ctx, document)
}

func (s *Store) DeleteInfectedDocuments(ctx context.Context, documents Documents) error {
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

func (s *Store) Delete(ctx context.Context, document Document) error {
	if err := s.s3Client.DeleteObject(ctx, document.Key); err != nil {
		return err
	}

	return s.dynamoClient.DeleteOne(ctx, document.PK, document.SK)
}

func (s *Store) Submit(ctx context.Context, donor *donordata.Provided, documents Documents) error {
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
