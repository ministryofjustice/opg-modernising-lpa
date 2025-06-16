package donor

import (
	"context"
	"errors"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type CorrespondentEventClient interface {
	SendCorrespondentUpdated(ctx context.Context, e event.CorrespondentUpdated) error
}

type CorrespondentService struct {
	donorStore  PutStore
	reuseStore  ReuseStore
	eventClient CorrespondentEventClient
	newUID      func() actoruid.UID
}

func NewCorrespondentService(donorStore PutStore, reuseStore ReuseStore, eventClient CorrespondentEventClient) *CorrespondentService {
	return &CorrespondentService{
		donorStore:  donorStore,
		reuseStore:  reuseStore,
		eventClient: eventClient,
		newUID:      actoruid.New,
	}
}

func (s *CorrespondentService) Reusable(ctx context.Context) ([]donordata.Correspondent, error) {
	correspondents, err := s.reuseStore.Correspondents(ctx)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return nil, fmt.Errorf("getting reusbale correspondents: %w", err)
	}

	return correspondents, nil
}

func (s *CorrespondentService) NotWanted(ctx context.Context, provided *donordata.Provided) error {
	if provided.Correspondent.FirstNames != "" {
		if err := s.reuseStore.DeleteCorrespondent(ctx, provided.Correspondent); err != nil {
			return fmt.Errorf("deleting reusable correspondent: %w", err)
		}

		if err := s.eventClient.SendCorrespondentUpdated(ctx, event.CorrespondentUpdated{
			UID: provided.LpaUID,
		}); err != nil {
			return err
		}
	}

	provided.Correspondent = donordata.Correspondent{}
	provided.AddCorrespondent = form.No
	provided.Tasks.AddCorrespondent = task.StateCompleted

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("deleting correspondent from lpa: %w", err)
	}

	return nil
}

func (s *CorrespondentService) Put(ctx context.Context, provided *donordata.Provided) error {
	if provided.Correspondent.UID.IsZero() {
		provided.Correspondent.UID = s.newUID()
	}

	if provided.Correspondent.WantAddress.IsNo() {
		provided.Correspondent.Address = place.Address{}
	}

	completed := provided.Correspondent.WantAddress.IsNo() || provided.Correspondent.Address.Line1 != ""
	if completed {
		provided.Tasks.AddCorrespondent = task.StateCompleted

		if err := s.reuseStore.PutCorrespondent(ctx, provided.Correspondent); err != nil {
			return fmt.Errorf("set reusable correspondent: %w", err)
		}

		event := event.CorrespondentUpdated{
			UID:        provided.LpaUID,
			ActorUID:   &provided.Correspondent.UID,
			FirstNames: provided.Correspondent.FirstNames,
			LastName:   provided.Correspondent.LastName,
			Email:      provided.Correspondent.Email,
			Phone:      provided.Correspondent.Phone,
		}
		if provided.Correspondent.WantAddress.IsYes() {
			event.Address = &provided.Correspondent.Address
		}

		if err := s.eventClient.SendCorrespondentUpdated(ctx, event); err != nil {
			return err
		}
	} else {
		provided.Tasks.AddCorrespondent = task.StateInProgress
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("set correspondent on lpa: %w", err)
	}

	return nil
}

func (s *CorrespondentService) Delete(ctx context.Context, provided *donordata.Provided) error {
	if err := s.reuseStore.DeleteCorrespondent(ctx, provided.Correspondent); err != nil {
		return fmt.Errorf("deleting reusable correspondent: %w", err)
	}

	provided.AddCorrespondent = form.YesNoUnknown
	provided.Correspondent = donordata.Correspondent{}
	provided.Tasks.AddCorrespondent = task.StateNotStarted

	if err := s.eventClient.SendCorrespondentUpdated(ctx, event.CorrespondentUpdated{
		UID: provided.LpaUID,
	}); err != nil {
		return err
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("deleting correspondent from lpa: %w", err)
	}

	return nil
}
