package accesscode

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	Create(ctx context.Context, v interface{}) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type Store struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{dynamoClient: dynamoClient, now: time.Now}
}

func (s *Store) Get(ctx context.Context, actorType actor.Type, accessCode accesscodedata.Hashed) (accesscodedata.Link, error) {
	var data accesscodedata.Link

	pk, err := accessCodeKey(actorType, accessCode)
	if err != nil {
		return data, err
	}

	if err := s.dynamoClient.OneByPK(ctx, pk, &data); err != nil {
		return accesscodedata.Link{}, err
	}

	if data.HasExpired(s.now()) || !data.LpaLinkedAt.IsZero() {
		return accesscodedata.Link{}, dynamo.NotFoundError{}
	}

	return data, nil
}

func (s *Store) Put(ctx context.Context, actorType actor.Type, accessCode accesscodedata.Hashed, data accesscodedata.Link) error {
	pk, err := accessCodeKey(actorType, accessCode)
	if err != nil {
		return err
	}

	hasActorAccess := true
	var actorAccess accesscodedata.ActorAccess
	if err := s.dynamoClient.OneByPK(ctx, dynamo.ActorAccessKey(data.ActorUID.String()), &actorAccess); err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			hasActorAccess = false
		} else {
			return err
		}
	}

	data.UpdatedAt = s.now()
	data.PK = pk
	if actorType.IsVoucher() {
		data.SK = dynamo.ShareSortKey(dynamo.VoucherShareSortKey(data.LpaKey))
	} else {
		data.SK = dynamo.ShareSortKey(dynamo.MetadataKey(accessCode.String()))
	}

	newActorAccess := accesscodedata.ActorAccess{
		PK:           dynamo.ActorAccessKey(data.ActorUID.String()),
		SK:           dynamo.MetadataKey(data.ActorUID.String()),
		ShareKey:     data.PK,
		ShareSortKey: data.SK,
	}

	transaction := dynamo.NewTransaction().Create(data)

	if hasActorAccess {
		transaction.
			Put(newActorAccess).
			Delete(dynamo.Keys{PK: actorAccess.ShareKey, SK: actorAccess.ShareSortKey})
	} else {
		transaction.Create(newActorAccess)
	}

	return s.dynamoClient.WriteTransaction(ctx, transaction)
}

func (s *Store) PutDonor(ctx context.Context, accessCode accesscodedata.Hashed, data accesscodedata.Link) error {
	organisationKey, ok := data.LpaOwnerKey.Organisation()
	if !ok {
		return errors.New("accessCodeStore.PutDonor can only be used by organisations")
	}

	data.PK = dynamo.AccessKey(dynamo.DonorAccessKey(accessCode.String()))
	data.SK = dynamo.ShareSortKey(dynamo.DonorInviteKey(organisationKey, data.LpaKey))
	data.UpdatedAt = s.now()

	return s.dynamoClient.Create(ctx, data)
}

func (s *Store) GetDonor(ctx context.Context) (accesscodedata.Link, error) {
	var data accesscodedata.Link

	sessionData, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return data, err
	}

	sk := dynamo.DonorInviteKey(dynamo.OrganisationKey(sessionData.OrganisationID), dynamo.LpaKey(sessionData.LpaID))

	if err := s.dynamoClient.OneBySK(ctx, sk, &data); err != nil {
		return accesscodedata.Link{}, err
	}

	if data.HasExpired(s.now()) {
		return accesscodedata.Link{}, dynamo.NotFoundError{}
	}

	return data, nil
}

func (s *Store) Delete(ctx context.Context, link accesscodedata.Link) error {
	return s.dynamoClient.DeleteOne(ctx, link.PK, link.SK)
}

func accessCodeKey(actorType actor.Type, accessCode accesscodedata.Hashed) (pk dynamo.ShareKeyType, err error) {
	switch actorType {
	case actor.TypeDonor:
		return dynamo.AccessKey(dynamo.DonorAccessKey(accessCode.String())), nil
	// As attorneys and replacement attorneys access the same landing page we can't
	// differentiate between them
	case actor.TypeAttorney, actor.TypeReplacementAttorney, actor.TypeTrustCorporation, actor.TypeReplacementTrustCorporation:
		return dynamo.AccessKey(dynamo.AttorneyAccessKey(accessCode.String())), nil
	case actor.TypeCertificateProvider:
		return dynamo.AccessKey(dynamo.CertificateProviderAccessKey(accessCode.String())), nil
	case actor.TypeVoucher:
		return dynamo.AccessKey(dynamo.VoucherAccessKey(accessCode.String())), nil
	default:
		return dynamo.AccessKey(nil), fmt.Errorf("cannot have access code for actorType=%v", actorType)
	}
}
