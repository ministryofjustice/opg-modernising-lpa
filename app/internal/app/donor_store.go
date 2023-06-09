package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"golang.org/x/exp/slices"
)

type donorStore struct {
	dataStore  DataStore
	uuidString func() string
	now        func() time.Time
}

func (s *donorStore) Create(ctx context.Context) (*page.Lpa, error) {
	lpa := &page.Lpa{
		ID:        s.uuidString(),
		UpdatedAt: s.now(),
	}

	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.Create requires SessionID")
	}

	pk, sk, subk := makeDonorKeys(lpa.ID, data.SessionID)
	if err := s.dataStore.Create(ctx, pk, sk, lpa); err != nil {
		return nil, err
	}
	if err := s.dataStore.Create(ctx, pk, subk, sk+"|DONOR"); err != nil {
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

	var items []struct {
		Data *page.Lpa
	}

	sk := "#DONOR#" + data.SessionID
	err = s.dataStore.GetAllByGsi(ctx, "ActorIndex", sk, &items)

	lpas := make([]*page.Lpa, len(items))
	for i, item := range items {
		lpas[i] = item.Data
	}

	slices.SortFunc(lpas, func(a, b *page.Lpa) bool {
		return a.UpdatedAt.After(b.UpdatedAt)
	})

	return lpas, err
}

func (s *donorStore) GetAny(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("donorStore.Get requires LpaID")
	}

	pk := "LPA#" + data.LpaID

	var lpa *page.Lpa
	if err := s.dataStore.GetOneByPartialSk(ctx, pk, "#DONOR#", &lpa); err != nil {
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

	pk, sk, _ := makeDonorKeys(data.LpaID, data.SessionID)

	var lpa *page.Lpa
	err = s.dataStore.Get(ctx, pk, sk, &lpa)
	return lpa, err
}

func (s *donorStore) Put(ctx context.Context, lpa *page.Lpa) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("donorStore.Put requires SessionID")
	}

	lpa.UpdatedAt = time.Now()

	pk, sk, _ := makeDonorKeys(lpa.ID, data.SessionID)
	return s.dataStore.Put(ctx, pk, sk, lpa)
}

func makeDonorKeys(lpaID, sessionID string) (string, string, string) {
	return "LPA#" + lpaID, "#DONOR#" + sessionID, "#SUB#" + sessionID
}
