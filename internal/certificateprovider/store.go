package certificateprovider

import (
	"context"
	"errors"
	"fmt"
	"time"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
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

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{dynamoClient: dynamoClient, now: time.Now}
}

type Store struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func (s *Store) Create(ctx context.Context, shareCode sharecodedata.Link, email string) (*certificateproviderdata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("certificateProviderStore.Create requires LpaID and SessionID")
	}

	certificateProvider := &certificateproviderdata.Provided{
		PK:        dynamo.LpaKey(data.LpaID),
		SK:        dynamo.CertificateProviderKey(data.SessionID),
		UID:       shareCode.ActorUID,
		LpaID:     data.LpaID,
		UpdatedAt: s.now(),
		Email:     email,
	}

	transaction := dynamo.NewTransaction().
		Create(dynamo.Keys{PK: certificateProvider.PK, SK: dynamo.ReservedKey(dynamo.CertificateProviderKey)}).
		Create(certificateProvider).
		Create(dashboarddata.LpaLink{
			PK:        dynamo.LpaKey(data.LpaID),
			SK:        dynamo.SubKey(data.SessionID),
			DonorKey:  shareCode.LpaOwnerKey,
			UID:       shareCode.ActorUID,
			ActorType: actor.TypeCertificateProvider,
			UpdatedAt: s.now(),
		}).
		Delete(dynamo.Keys{PK: shareCode.PK, SK: shareCode.SK})

	if err := s.dynamoClient.WriteTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	return certificateProvider, err
}

func (s *Store) GetAny(ctx context.Context) (*certificateproviderdata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("certificateProviderStore.GetAny requires LpaID")
	}

	return s.One(ctx, dynamo.LpaKey(data.LpaID))
}

func (s *Store) One(ctx context.Context, pk dynamo.LpaKeyType) (*certificateproviderdata.Provided, error) {
	var certificateProvider certificateproviderdata.Provided
	err := s.dynamoClient.OneByPartialSK(ctx, pk, dynamo.CertificateProviderKey(""), &certificateProvider)

	return &certificateProvider, err
}

func (s *Store) Get(ctx context.Context) (*certificateproviderdata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("certificateProviderStore.Get requires LpaID and SessionID")
	}

	var certificateProvider certificateproviderdata.Provided
	err = s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), dynamo.CertificateProviderKey(data.SessionID), &certificateProvider)

	return &certificateProvider, err
}

func (s *Store) Put(ctx context.Context, certificateProvider *certificateproviderdata.Provided) error {
	certificateProvider.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, certificateProvider)
}

func (s *Store) Delete(ctx context.Context) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return errors.New("certificateProviderStore.Delete requires LpaID and SessionID")
	}

	transaction := dynamo.NewTransaction().
		Delete(dynamo.Keys{PK: dynamo.LpaKey(data.LpaID), SK: dynamo.CertificateProviderKey(data.SessionID)}).
		Delete(dynamo.Keys{PK: dynamo.LpaKey(data.LpaID), SK: dynamo.SubKey(data.SessionID)}).
		Delete(dynamo.Keys{PK: dynamo.LpaKey(data.LpaID), SK: dynamo.ReservedKey(dynamo.CertificateProviderKey)})

	if err := s.dynamoClient.WriteTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("error deleting certificate provider: %w", err)
	}

	return nil
}
