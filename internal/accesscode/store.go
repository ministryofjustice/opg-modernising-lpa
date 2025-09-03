package accesscode

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/rate"
)

type DynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v any) error
	OneByPK(ctx context.Context, pk dynamo.PK, v any) error
	OneBySK(ctx context.Context, sk dynamo.SK, v any) error
	Create(ctx context.Context, v any) error
	Put(ctx context.Context, v any) error
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

type accessLimiter struct {
	PK        dynamo.AccessLimiterKeyType
	SK        dynamo.MetadataKeyType
	Version   int
	ExpiresAt time.Time
	Limiter   *rate.Limiter
}

func (s *Store) allowed(ctx context.Context) error {
	data, err := appcontext.SessionFromContext(ctx)
	// As a compromise we count unauthenticated requests together, these are for
	// the opt-out pages. Otherwise we'd be leaving them open for abuse, or as a
	// way to determine valid combinations to use to add LPAs. I suspect they'll
	// never get enough legitimate use to ever hit the rate.
	if err != nil {
		data = &appcontext.Session{}
	}

	var v accessLimiter
	fresh := false
	if err := s.dynamoClient.OneByPK(ctx, dynamo.AccessLimiterKey(data.SessionID), &v); err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			fresh = true
			v = accessLimiter{
				PK:      dynamo.AccessLimiterKey(data.SessionID),
				SK:      dynamo.MetadataKey(data.SessionID),
				Version: 1,
				Limiter: rate.NewLimiter(s.now(), 5*time.Minute, 5, 10),
			}
		} else {
			return fmt.Errorf("retrieve rate limiter: %w", err)
		}
	}

	allowed := v.Limiter.Allow(s.now())
	v.ExpiresAt = s.now().Add(time.Hour)

	if fresh {
		if err := s.dynamoClient.Create(ctx, v); err != nil {
			return fmt.Errorf("create rate limiter: %w", err)
		}
	} else {
		if err := s.dynamoClient.Put(ctx, v); err != nil {
			return fmt.Errorf("update rate limiter: %w", err)
		}
	}

	if !allowed {
		return dynamo.ErrTooManyRequests
	}

	return nil
}

func (s *Store) Get(ctx context.Context, actorType actor.Type, accessCode accesscodedata.Hashed) (accesscodedata.Link, error) {
	pk, err := accessCodeKey(actorType, accessCode)
	if err != nil {
		return accesscodedata.Link{}, err
	}

	if err := s.allowed(ctx); err != nil {
		return accesscodedata.Link{}, err
	}

	var data accesscodedata.Link
	if err := s.dynamoClient.OneByPK(ctx, pk, &data); err != nil {
		return accesscodedata.Link{}, err
	}

	if data.HasExpired(s.now()) {
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
		data.SK = dynamo.AccessSortKey(dynamo.VoucherAccessSortKey(data.LpaKey))
	} else {
		data.SK = dynamo.AccessSortKey(dynamo.MetadataKey(accessCode.String()))
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

func (s *Store) GetDonor(ctx context.Context, accessCode accesscodedata.Hashed) (accesscodedata.DonorLink, error) {
	if err := s.allowed(ctx); err != nil {
		return accesscodedata.DonorLink{}, err
	}

	pk, err := accessCodeKey(actor.TypeDonor, accessCode)
	if err != nil {
		return accesscodedata.DonorLink{}, err
	}

	var data accesscodedata.DonorLink
	if err := s.dynamoClient.OneByPK(ctx, pk, &data); err != nil {
		return accesscodedata.DonorLink{}, err
	}

	if data.HasExpired(s.now()) || !data.LpaLinkedAt.IsZero() {
		return accesscodedata.DonorLink{}, dynamo.NotFoundError{}
	}

	return data, nil
}

func (s *Store) PutDonor(ctx context.Context, accessCode accesscodedata.Hashed, data accesscodedata.DonorLink) error {
	organisationKey, ok := data.LpaOwnerKey.Organisation()
	if !ok {
		return errors.New("accessCodeStore.PutDonor can only be used by organisations")
	}

	data.PK = dynamo.AccessKey(dynamo.DonorAccessKey(accessCode.String()))
	data.SK = dynamo.AccessSortKey(dynamo.DonorInviteKey(organisationKey, data.LpaKey))
	data.UpdatedAt = s.now()

	return s.dynamoClient.Create(ctx, data)
}

func (s *Store) GetDonorAccess(ctx context.Context) (accesscodedata.DonorLink, error) {
	sessionData, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return accesscodedata.DonorLink{}, err
	}

	sk := dynamo.DonorInviteKey(dynamo.OrganisationKey(sessionData.OrganisationID), dynamo.LpaKey(sessionData.LpaID))

	var data accesscodedata.DonorLink
	if err := s.dynamoClient.OneBySK(ctx, sk, &data); err != nil {
		return accesscodedata.DonorLink{}, err
	}

	if data.HasExpired(s.now()) {
		return accesscodedata.DonorLink{}, dynamo.NotFoundError{}
	}

	return data, nil
}

func (s *Store) Delete(ctx context.Context, link accesscodedata.Link) error {
	transaction := dynamo.NewTransaction().
		Delete(dynamo.Keys{PK: link.PK, SK: link.SK}).
		Delete(dynamo.Keys{
			PK: dynamo.ActorAccessKey(link.ActorUID.String()),
			SK: dynamo.MetadataKey(link.ActorUID.String()),
		})

	return s.dynamoClient.WriteTransaction(ctx, transaction)
}

func (s *Store) DeleteByActor(ctx context.Context, actorUID actoruid.UID) error {
	var actorAccess accesscodedata.ActorAccess
	if err := s.dynamoClient.One(ctx, dynamo.ActorAccessKey(actorUID.String()), dynamo.MetadataKey(actorUID.String()), &actorAccess); err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			return nil
		}

		return fmt.Errorf("retrieve actor access: %w", err)
	}

	transaction := dynamo.NewTransaction().
		Delete(dynamo.Keys{PK: actorAccess.ShareKey, SK: actorAccess.ShareSortKey}).
		Delete(dynamo.Keys{PK: actorAccess.PK, SK: actorAccess.SK})

	return s.dynamoClient.WriteTransaction(ctx, transaction)
}

func (s *Store) DeleteDonor(ctx context.Context, link accesscodedata.DonorLink) error {
	return s.dynamoClient.DeleteOne(ctx, link.PK, link.SK)
}

func accessCodeKey(actorType actor.Type, accessCode accesscodedata.Hashed) (pk dynamo.AccessKeyType, err error) {
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
