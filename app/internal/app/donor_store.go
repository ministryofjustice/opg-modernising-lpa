package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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
		return nil, errors.New("donorStore.Create requires SessionID to persist")
	}

	pk, sk := makeLpaKeys(lpa.ID, data.SessionID)
	err = s.dataStore.Create(ctx, pk, sk, lpa)

	return lpa, err
}

func (s *donorStore) GetAll(ctx context.Context) ([]*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.GetAll requires SessionID to retrieve")
	}

	var lpas []*page.Lpa

	sk := "#DONOR#" + data.SessionID
	err = s.dataStore.GetAllByGsi(ctx, "ActorIndex", sk, &lpas)

	slices.SortFunc(lpas, func(a, b *page.Lpa) bool {
		return a.UpdatedAt.After(b.UpdatedAt)
	})

	return lpas, err
}

func (s *donorStore) Get(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("donorStore.Get requires LpaID to retrieve")
	}

	pk := "LPA#" + data.LpaID

	var lpa *page.Lpa
	if err := s.dataStore.GetOneByPartialSk(ctx, pk, "#DONOR#", &lpa); err != nil {
		return nil, err
	}

	return lpa, nil
}

func (s *donorStore) Put(ctx context.Context, lpa *page.Lpa) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("donorStore.Put requires SessionID to persist")
	}

	lpa.UpdatedAt = time.Now()

	pk, sk := makeLpaKeys(lpa.ID, data.SessionID)
	return s.dataStore.Put(ctx, pk, sk, lpa)
}

func makeLpaKeys(lpaID, sessionID string) (string, string) {
	return "LPA#" + lpaID, "#DONOR#" + sessionID
}
