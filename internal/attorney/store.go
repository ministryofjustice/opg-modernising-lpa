package attorney

import (
	"context"
	"errors"
	"fmt"
	"time"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DynamoClient interface {
	AllByKeys(ctx context.Context, keys []dynamo.Keys) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllByLpaUIDAndPartialSK(ctx context.Context, uid string, partialSK dynamo.SK) ([]dynamo.Keys, error)
	AllByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v any) error
	AllBySK(ctx context.Context, sk dynamo.SK, v any) error
	AllKeysByPK(ctx context.Context, pk dynamo.PK) ([]dynamo.Keys, error)
	BatchPut(ctx context.Context, items []any) error
	Create(ctx context.Context, v any) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	LatestForActor(ctx context.Context, sk dynamo.SK, v any) error
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v any) error
	OneByPK(ctx context.Context, pk dynamo.PK, v any) error
	OneByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v any) error
	OneBySK(ctx context.Context, sk dynamo.SK, v any) error
	OneByUID(ctx context.Context, uid string) (dynamo.Keys, error)
	Put(ctx context.Context, v any) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type Store struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{dynamoClient: dynamoClient, now: time.Now}
}

func (s *Store) Create(ctx context.Context, link accesscodedata.Link, email string) (*attorneydata.Provided, error) {
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
		UID:                link.ActorUID,
		LpaID:              data.LpaID,
		UpdatedAt:          s.now(),
		IsReplacement:      link.IsReplacementAttorney,
		IsTrustCorporation: link.IsTrustCorporation,
		Email:              email,
	}

	actorType := actor.TypeAttorney

	if link.IsTrustCorporation && link.IsReplacementAttorney {
		actorType = actor.TypeReplacementTrustCorporation
	} else if link.IsTrustCorporation {
		actorType = actor.TypeTrustCorporation
	} else if link.IsReplacementAttorney {
		actorType = actor.TypeReplacementAttorney
	}

	transaction := dynamo.NewTransaction().
		Create(attorney).
		Create(dashboarddata.LpaLink{
			PK:        dynamo.LpaKey(data.LpaID),
			SK:        dynamo.SubKey(data.SessionID),
			LpaUID:    link.LpaUID,
			UID:       link.ActorUID,
			DonorKey:  link.LpaOwnerKey,
			ActorType: actorType,
			UpdatedAt: s.now(),
		}).
		Delete(dynamo.Keys{PK: link.PK, SK: link.SK})

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

func (s *Store) All(ctx context.Context, pk dynamo.LpaKeyType) ([]*attorneydata.Provided, error) {
	var attorneys []*attorneydata.Provided
	err := s.dynamoClient.AllByPartialSK(ctx, pk, dynamo.AttorneyKey(""), &attorneys)
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
