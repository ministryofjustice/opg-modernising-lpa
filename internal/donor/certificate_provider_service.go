package donor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type ScheduledStore interface {
	DeleteAllActionByUID(ctx context.Context, actions []scheduleddata.Action, uid string) error
}

type CertificateProviderService struct {
	donorStore     PutStore
	reuseStore     ReuseStore
	scheduledStore ScheduledStore
	newUID         func() actoruid.UID
}

func NewCertificateProviderService(donorStore PutStore, reuseStore ReuseStore, scheduledStore ScheduledStore) *CertificateProviderService {
	return &CertificateProviderService{
		donorStore:     donorStore,
		reuseStore:     reuseStore,
		newUID:         actoruid.New,
		scheduledStore: scheduledStore,
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
		return fmt.Errorf("deleting reusable certificate provider: %w", err)
	}

	provided.CertificateProvider = donordata.CertificateProvider{}
	provided.CertificateProviderNotRelatedConfirmedAt = time.Time{}
	provided.CertificateProviderNotRelatedConfirmedHash = 0
	provided.CertificateProviderNotRelatedConfirmedHashVersion = 0
	provided.CertificateProviderInvitedAt = time.Time{}
	provided.CheckedAt = time.Time{}
	provided.Tasks.CertificateProvider = task.StateNotStarted

	if err := s.scheduledStore.DeleteAllActionByUID(ctx, []scheduleddata.Action{scheduleddata.ActionRemindCertificateProviderToComplete}, provided.LpaUID); err != nil {
		return fmt.Errorf("deleting scheduled actions: %w", err)
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("deleting certificate provider from lpa: %w", err)
	}

	return nil
}
