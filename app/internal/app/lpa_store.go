package app

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"golang.org/x/exp/slices"
)

type lpaStore struct {
	dataStore DataStore
	randomInt func(int) int
}

func (s *lpaStore) Create(ctx context.Context) (*page.Lpa, error) {
	lpa := &page.Lpa{
		ID: "10" + strconv.Itoa(s.randomInt(100000)),
	}

	err := s.Put(ctx, lpa)

	return lpa, err
}

func (s *lpaStore) Clone(ctx context.Context, id string) (*page.Lpa, error) {
	data := page.SessionDataFromContext(ctx)

	var lpa page.Lpa
	if err := s.dataStore.Get(ctx, data.SessionID, id, &lpa); err != nil {
		return nil, err
	}

	lpa.ID = "10" + strconv.Itoa(s.randomInt(100000))
	err := s.Put(ctx, &lpa)

	return &lpa, err
}

func (s *lpaStore) GetAll(ctx context.Context) ([]*page.Lpa, error) {
	var lpas []*page.Lpa
	err := s.dataStore.GetAll(ctx, page.SessionDataFromContext(ctx).SessionID, &lpas)

	slices.SortFunc(lpas, func(a, b *page.Lpa) bool {
		return a.UpdatedAt.After(b.UpdatedAt)
	})

	return lpas, err
}

func (s *lpaStore) Get(ctx context.Context) (*page.Lpa, error) {
	data := page.SessionDataFromContext(ctx)
	if data.LpaID == "" {
		return nil, errors.New("lpaStore.Get requires LpaID to retrieve")
	}

	var lpa page.Lpa
	if err := s.dataStore.Get(ctx, data.SessionID, data.LpaID, &lpa); err != nil {
		return nil, err
	}

	return &lpa, nil
}

func (s *lpaStore) Put(ctx context.Context, lpa *page.Lpa) error {
	lpa.UpdatedAt = time.Now()

	return s.dataStore.Put(ctx, page.SessionDataFromContext(ctx).SessionID, lpa.ID, lpa)
}
