package donor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type ScheduledStore interface {
	DeleteAllActionByUID(ctx context.Context, actions []scheduleddata.Action, uid string) error
}

type AccessCodeStore interface {
	DeleteByActor(ctx context.Context, actorUID actoruid.UID) error
}

type CertificateProviderService struct {
	donorStore               PutStore
	reuseStore               ReuseStore
	scheduledStore           ScheduledStore
	accessCodeStore          AccessCodeStore
	certificateProviderStore CertificateProviderStore
	newUID                   func() actoruid.UID
}

func NewCertificateProviderService(donorStore PutStore, reuseStore ReuseStore, scheduledStore ScheduledStore, accessCodeStore AccessCodeStore, certificateProviderStore CertificateProviderStore) *CertificateProviderService {
	return &CertificateProviderService{
		donorStore:               donorStore,
		reuseStore:               reuseStore,
		scheduledStore:           scheduledStore,
		accessCodeStore:          accessCodeStore,
		certificateProviderStore: certificateProviderStore,
		newUID:                   actoruid.New,
	}
}

func (s *CertificateProviderService) Reusable(ctx context.Context) ([]donordata.CertificateProvider, error) {
	certificateProviders, err := s.reuseStore.CertificateProviders(ctx)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return nil, fmt.Errorf("getting reusbale certificate providers: %w", err)
	}

	return certificateProviders, nil
}

func (s *CertificateProviderService) Put(ctx context.Context, provided *donordata.Provided) error {
	if provided.CertificateProvider.UID.IsZero() {
		provided.CertificateProvider.UID = s.newUID()
	}

	completed := provided.CertificateProvider.Address.Line1 != "" &&
		!provided.CertificateProvider.CarryOutBy.Empty() &&
		(provided.CertificateProvider.Relationship.IsProfessionally() ||
			provided.CertificateProvider.Relationship.IsPersonally() &&
				provided.CertificateProvider.RelationshipLength.IsGreaterThanEqualToTwoYears())

	if completed {
		provided.Tasks.CertificateProvider = task.StateCompleted
	} else {
		provided.Tasks.CertificateProvider = task.StateInProgress
	}

	if err := s.reuseStore.PutCertificateProvider(ctx, provided.CertificateProvider); err != nil {
		return fmt.Errorf("set reusable certificate provider: %w", err)
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("set certificate provider on lpa: %w", err)
	}

	return nil
}

func (s *CertificateProviderService) Delete(ctx context.Context, provided *donordata.Provided) error {
	if err := s.reuseStore.DeleteCertificateProvider(ctx, provided.CertificateProvider); err != nil {
		return fmt.Errorf("delete reusable certificate provider: %w", err)
	}

	if err := s.accessCodeStore.DeleteByActor(ctx, provided.CertificateProvider.UID); err != nil {
		return fmt.Errorf("delete certificate provider access code: %w", err)
	}

	provided.CertificateProvider = donordata.CertificateProvider{}
	provided.CertificateProviderNotRelatedConfirmedAt = time.Time{}
	provided.CertificateProviderNotRelatedConfirmedHash = 0
	provided.CertificateProviderNotRelatedConfirmedHashVersion = 0
	provided.CertificateProviderInvitedAt = time.Time{}
	provided.CheckedAt = time.Time{}
	provided.Tasks.CertificateProvider = task.StateNotStarted

	if err := s.scheduledStore.DeleteAllActionByUID(ctx, []scheduleddata.Action{scheduleddata.ActionRemindCertificateProviderToComplete}, provided.LpaUID); err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return fmt.Errorf("delete scheduled actions: %w", err)
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("delete certificate provider from lpa: %w", err)
	}

	certificateProvider, err := s.certificateProviderStore.OneByUID(ctx, provided.LpaUID)
	if err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			return nil
		}

		return fmt.Errorf("get certificate provider: %w", err)
	}

	if err := s.certificateProviderStore.Delete(appcontext.ContextWithSession(ctx, &appcontext.Session{
		LpaID:     provided.LpaID,
		SessionID: certificateProvider.SK.Sub(),
	})); err != nil {
		return fmt.Errorf("delete certificate provider: %w", err)
	}

	return nil
}
