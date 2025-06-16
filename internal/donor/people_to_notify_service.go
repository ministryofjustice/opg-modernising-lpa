package donor

import (
	"context"
	"errors"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type PeopleToNotifyService struct {
	donorStore PutStore
	reuseStore ReuseStore
	newUID     func() actoruid.UID
}

func NewPeopleToNotifyService(donorStore PutStore, reuseStore ReuseStore) *PeopleToNotifyService {
	return &PeopleToNotifyService{
		donorStore: donorStore,
		reuseStore: reuseStore,
		newUID:     actoruid.New,
	}
}

func (s *PeopleToNotifyService) Reusable(ctx context.Context, provided *donordata.Provided) ([]donordata.PersonToNotify, error) {
	peopleToNotify, err := s.reuseStore.PeopleToNotify(ctx, provided)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return nil, fmt.Errorf("getting reusable people to notify: %w", err)
	}

	return peopleToNotify, nil
}

func (s *PeopleToNotifyService) WantPeopleToNotify(ctx context.Context, provided *donordata.Provided, yesNo form.YesNo) error {
	provided.DoYouWantToNotifyPeople = yesNo
	provided.Tasks.PeopleToNotify = s.taskState(provided)

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("setting want people to notify to lpa: %w", err)
	}

	return nil
}

func (s *PeopleToNotifyService) PutMany(ctx context.Context, provided *donordata.Provided, people []donordata.PersonToNotify) error {
	for _, person := range people {
		if person.UID.IsZero() {
			person.UID = s.newUID()
		}
		provided.PeopleToNotify.Put(person)
	}
	provided.Tasks.PeopleToNotify = s.taskState(provided)

	if err := s.reuseStore.PutPeopleToNotify(ctx, provided.PeopleToNotify); err != nil {
		return fmt.Errorf("adding many reusable people to notify: %w", err)
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("adding many people to notify to lpa: %w", err)
	}

	return nil
}

func (s *PeopleToNotifyService) Put(ctx context.Context, provided *donordata.Provided, person donordata.PersonToNotify) (actoruid.UID, error) {
	if person.UID.IsZero() {
		person.UID = s.newUID()
	}
	provided.PeopleToNotify.Put(person)
	provided.Tasks.PeopleToNotify = s.taskState(provided)

	if err := s.reuseStore.PutPersonToNotify(ctx, person); err != nil {
		return actoruid.UID{}, fmt.Errorf("adding reusable person to notify: %w", err)
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return actoruid.UID{}, fmt.Errorf("adding person to notify to lpa: %w", err)
	}

	return person.UID, nil
}

func (s *PeopleToNotifyService) Delete(ctx context.Context, provided *donordata.Provided, person donordata.PersonToNotify) error {
	provided.PeopleToNotify.Delete(person)
	provided.Tasks.PeopleToNotify = s.taskState(provided)

	if err := s.reuseStore.DeletePersonToNotify(ctx, person); err != nil {
		return fmt.Errorf("removing reusable person to notify: %w", err)
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("removing person to notify from lpa: %w", err)
	}

	return nil
}

func (s *PeopleToNotifyService) taskState(provided *donordata.Provided) task.State {
	switch provided.DoYouWantToNotifyPeople {
	case form.No:
		return task.StateCompleted

	case form.Yes:
		if len(provided.PeopleToNotify) == 0 {
			return task.StateInProgress
		}

		for _, person := range provided.PeopleToNotify {
			if person.Address.Line1 == "" {
				return task.StateInProgress
			}
		}

		return task.StateCompleted

	default:
		return task.StateNotStarted
	}
}
