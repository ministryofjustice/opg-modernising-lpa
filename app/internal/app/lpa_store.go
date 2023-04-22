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

func (s *lpaStore) GetAll(ctx context.Context) ([]*page.Lpa, error) {
	var lpas []*page.Lpa

	data := page.SessionDataFromContext(ctx)
	if data.SessionID == "" {
		return nil, errors.New("lpaStore.GetAll requires SessionID to retrieve")
	}

	sk := "#DONOR#" + data.SessionID
	err := s.dataStore.GetAllByGsi(ctx, "ActorIndex", sk, &lpas)

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

	pk := "LPA#" + data.LpaID

	var lpa page.Lpa
	if err := s.dataStore.GetOneByPartialSk(ctx, pk, "#DONOR#", &lpa); err != nil {
		return nil, err
	}

	return &lpa, nil
}

func (s *lpaStore) Put(ctx context.Context, lpa *page.Lpa) error {
	data := page.SessionDataFromContext(ctx)
	if data.SessionID == "" {
		return errors.New("lpaStore.Put requires SessionID to persist")
	}

	lpa.UpdatedAt = time.Now()

	pk := "LPA#" + lpa.ID
	sk := "#DONOR#" + data.SessionID
	return s.dataStore.Put(ctx, pk, sk, lpa)
}
