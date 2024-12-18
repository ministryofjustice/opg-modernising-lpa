package attorney

import (
	"context"
	"errors"
	"fmt"
	"time"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
)

type DynamoClient interface {
	AllByKeys(ctx context.Context, keys []dynamo.Keys) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllByLpaUIDAndPartialSK(ctx context.Context, uid string, partialSK dynamo.SK, v interface{}) error
	AllByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	AllBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllKeysByPK(ctx context.Context, pk dynamo.PK) ([]dynamo.Keys, error)
	BatchPut(ctx context.Context, items []interface{}) error
	Create(ctx context.Context, v interface{}) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	LatestForActor(ctx context.Context, sk dynamo.SK, v interface{}) error
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	OneByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	Put(ctx context.Context, v interface{}) error
	Update(ctx context.Context, pk dynamo.PK, sk dynamo.SK, values map[string]dynamodbtypes.AttributeValue, expression string) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type Store struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{dynamoClient: dynamoClient, now: time.Now}
}

func (s *Store) Create(ctx context.Context, shareCode sharecodedata.Link, email string) (*attorneydata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("attorneyStore.Create requires LpaID and SessionID")
	}

	attorney := &attorneydata.Provided{
		PK:                 dynamo.LpaKey(data.LpaID),
		SK:                 dynamo.AttorneyKey(data.SessionID),
		UID:                shareCode.ActorUID,
		LpaID:              data.LpaID,
		UpdatedAt:          s.now(),
		IsReplacement:      shareCode.IsReplacementAttorney,
		IsTrustCorporation: shareCode.IsTrustCorporation,
		Email:              email,
	}

	transaction := dynamo.NewTransaction().
		Create(attorney).
		Create(dashboarddata.LpaLink{
			PK:        dynamo.LpaKey(data.LpaID),
			SK:        dynamo.SubKey(data.SessionID),
			UID:       shareCode.ActorUID,
			DonorKey:  shareCode.LpaOwnerKey,
			ActorType: actor.TypeAttorney,
			UpdatedAt: s.now(),
		}).
		Delete(dynamo.Keys{PK: shareCode.PK, SK: shareCode.SK})

	if err := s.dynamoClient.WriteTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	return attorney, err
}

func (s *Store) Get(ctx context.Context) (*attorneydata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("attorneyStore.Get requires LpaID and SessionID")
	}

	var attorney attorneydata.Provided
	err = s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), dynamo.AttorneyKey(data.SessionID), &attorney)

	return &attorney, err
}

func (s *Store) All(ctx context.Context, lpaUID string) ([]*attorneydata.Provided, error) {
	var attorneys []*attorneydata.Provided
	err := s.dynamoClient.AllByLpaUIDAndPartialSK(ctx, lpaUID, dynamo.AttorneyKey(""), &attorneys)
	return attorneys, err
}

func (s *Store) Put(ctx context.Context, attorney *attorneydata.Provided) error {
	attorney.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, attorney)
}

func (s *Store) Delete(ctx context.Context) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return errors.New("attorneyStore.Delete requires LpaID and SessionID")
	}

	if err := s.dynamoClient.DeleteOne(ctx, dynamo.LpaKey(data.LpaID), dynamo.AttorneyKey(data.SessionID)); err != nil {
		return fmt.Errorf("error deleting attorney: %w", err)
	}

	return nil
}
