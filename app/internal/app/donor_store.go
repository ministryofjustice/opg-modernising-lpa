package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"golang.org/x/exp/slices"
)

type donorStore struct {
	dataStore  DataStore
	uuidString func() string
	now        func() time.Time
}

func (s *donorStore) Create(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.Create requires SessionID")
	}

	lpaID := s.uuidString()

	lpa := &page.Lpa{
		PK:        lpaKey(lpaID),
		SK:        donorKey(data.SessionID),
		ID:        lpaID,
		UpdatedAt: s.now(),
	}

	if err := s.dataStore.Create(ctx, lpa); err != nil {
		return nil, err
	}
	if err := s.dataStore.Create(ctx, lpaLink{
		PK:        lpaKey(lpaID),
		SK:        subKey(data.SessionID),
		DonorKey:  donorKey(data.SessionID),
		ActorType: actor.TypeDonor,
	}); err != nil {
		return nil, err
	}

	return lpa, err
}

func (s *donorStore) GetAll(ctx context.Context) ([]*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.GetAll requires SessionID")
	}

	var items []*page.Lpa
	err = s.dataStore.GetAllByGsi(ctx, "ActorIndex", donorKey(data.SessionID), &items)

	slices.SortFunc(items, func(a, b *page.Lpa) bool {
		return a.UpdatedAt.After(b.UpdatedAt)
	})

	return items, err
}

func (s *donorStore) GetAny(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("donorStore.Get requires LpaID")
	}

	var lpa *page.Lpa
	if err := s.dataStore.GetOneByPartialSk(ctx, lpaKey(data.LpaID), "#DONOR#", &lpa); err != nil {
		return nil, err
	}

	return lpa, nil
}

func (s *donorStore) Get(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("donorStore.Get requires LpaID and SessionID")
	}

	var lpa *page.Lpa
	err = s.dataStore.Get(ctx, lpaKey(data.LpaID), donorKey(data.SessionID), &lpa)
	return lpa, err
}

func (s *donorStore) Put(ctx context.Context, lpa *page.Lpa) error {
	lpa.UpdatedAt = time.Now()
	return s.dataStore.Put(ctx, lpa)
}

func lpaKey(s string) string {
	return "LPA#" + s
}

func donorKey(s string) string {
	return "#DONOR#" + s
}

func subKey(s string) string {
	return "#SUB#" + s
}
