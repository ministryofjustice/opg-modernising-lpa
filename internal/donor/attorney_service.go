package donor

import (
	"context"
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
)

type AttorneyService struct {
	donorStore    PutStore
	reuseStore    ReuseStore
	newUID        func() actoruid.UID
	isReplacement bool
}

func NewAttorneyService(donorStore PutStore, reuseStore ReuseStore) *AttorneyService {
	return &AttorneyService{
		donorStore: donorStore,
		reuseStore: reuseStore,
		newUID:     actoruid.New,
	}
}

func NewReplacementAttorneyService(donorStore PutStore, reuseStore ReuseStore) *AttorneyService {
	return &AttorneyService{
		donorStore:    donorStore,
		reuseStore:    reuseStore,
		newUID:        actoruid.New,
		isReplacement: true,
	}
}

func (s *AttorneyService) Reusable(ctx context.Context, provided *donordata.Provided) ([]donordata.Attorney, error) {
	attorneys, err := s.reuseStore.Attorneys(ctx, provided)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return nil, err
	}

	return attorneys, nil
}

func (s *AttorneyService) ReusableTrustCorporations(ctx context.Context) ([]donordata.TrustCorporation, error) {
	trustCorporations, err := s.reuseStore.TrustCorporations(ctx)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return nil, err
	}

	return trustCorporations, nil
}

func (s *AttorneyService) WantReplacements(ctx context.Context, provided *donordata.Provided, yesNo form.YesNo) error {
	provided.WantReplacementAttorneys = yesNo
	provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return err
	}

	return nil
}

func (s *AttorneyService) PutMany(ctx context.Context, provided *donordata.Provided, attorneys []donordata.Attorney) error {
	for _, attorney := range attorneys {
		if attorney.UID.IsZero() {
			attorney.UID = s.newUID()
		}

		if s.isReplacement {
			provided.ReplacementAttorneys.Attorneys = append(provided.ReplacementAttorneys.Attorneys, attorney)
		} else {
			provided.Attorneys.Attorneys = append(provided.Attorneys.Attorneys, attorney)
		}
	}

	provided.UpdateDecisions()
	provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
	provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

	if err := s.reuseStore.PutAttorneys(ctx, provided.Attorneys.Attorneys); err != nil {
		return err
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return err
	}

	return nil
}

func (s *AttorneyService) Put(ctx context.Context, provided *donordata.Provided, attorney donordata.Attorney) error {
	if s.isReplacement {
		provided.ReplacementAttorneys.Put(attorney)
	} else {
		provided.Attorneys.Put(attorney)
	}

	provided.UpdateDecisions()
	provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
	provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

	if err := s.reuseStore.PutAttorney(ctx, attorney); err != nil {
		return err
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return err
	}

	return nil
}

func (s *AttorneyService) PutTrustCorporation(ctx context.Context, provided *donordata.Provided, trustCorporation donordata.TrustCorporation) error {
	if trustCorporation.UID.IsZero() {
		trustCorporation.UID = s.newUID()
	}

	if s.isReplacement {
		provided.ReplacementAttorneys.TrustCorporation = trustCorporation
	} else {
		provided.Attorneys.TrustCorporation = trustCorporation
	}

	provided.UpdateDecisions()
	provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
	provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

	if err := s.reuseStore.PutTrustCorporation(ctx, provided.Attorneys.TrustCorporation); err != nil {
		return err
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return err
	}

	return nil
}

func (s *AttorneyService) Delete(ctx context.Context, provided *donordata.Provided, attorney donordata.Attorney) error {
	if s.isReplacement {
		provided.ReplacementAttorneys.Delete(attorney)
	} else {
		provided.Attorneys.Delete(attorney)
	}

	provided.UpdateDecisions()
	provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
	provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

	if err := s.reuseStore.DeleteAttorney(ctx, attorney); err != nil {
		return err
	}

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return err
	}

	return nil
}

func (s *AttorneyService) DeleteTrustCorporation(ctx context.Context, provided *donordata.Provided) error {
	if err := s.reuseStore.DeleteTrustCorporation(ctx, provided.Attorneys.TrustCorporation); err != nil {
		return err
	}

	if s.isReplacement {
		provided.ReplacementAttorneys.TrustCorporation = donordata.TrustCorporation{}
	} else {
		provided.Attorneys.TrustCorporation = donordata.TrustCorporation{}
	}

	provided.UpdateDecisions()
	provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
	provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

	if err := s.donorStore.Put(ctx, provided); err != nil {
		return err
	}

	return nil
}

func (s *AttorneyService) IsReplacement() bool {
	return s.isReplacement
}
