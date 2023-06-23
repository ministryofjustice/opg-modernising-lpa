package app

import (
	"context"
	"errors"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"golang.org/x/exp/slices"
)

type sub struct {
	PK        string
	SK        string
	DonorKey  string
	ActorType actor.Type
}

type dashboardStore struct {
	dataStore DataStore
}

func (s *dashboardStore) GetAll(ctx context.Context) (donor, attorney, certificateProvider []*page.Lpa, err error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if data.SessionID == "" {
		return nil, nil, nil, errors.New("donorStore.GetAll requires SessionID")
	}

	var keys []sub
	if err := s.dataStore.GetAllByGsi(ctx, "ActorIndex", "#SUB#"+data.SessionID, &keys); err != nil {
		return nil, nil, nil, err
	}

	searchKeys := make([]dynamo.Key, len(keys))
	keyMap := map[string]actor.Type{}
	for i, key := range keys {
		searchKeys[i] = dynamo.Key{PK: key.PK, SK: key.DonorKey}

		_, id, _ := strings.Cut(key.PK, "#")
		keyMap[id] = key.ActorType
	}

	if len(searchKeys) == 0 {
		return nil, nil, nil, nil
	}

	var items []*page.Lpa
	if err := s.dataStore.GetAllByKeys(ctx, searchKeys, &items); err != nil {
		return nil, nil, nil, err
	}

	for _, item := range items {
		switch keyMap[item.ID] {
		case actor.TypeDonor:
			donor = append(donor, item)
		case actor.TypeAttorney:
			attorney = append(attorney, item)
		case actor.TypeCertificateProvider:
			certificateProvider = append(certificateProvider, item)
		}
	}

	byUpdatedAt := func(a, b *page.Lpa) bool {
		return a.UpdatedAt.After(b.UpdatedAt)
	}

	slices.SortFunc(donor, byUpdatedAt)
	slices.SortFunc(attorney, byUpdatedAt)
	slices.SortFunc(certificateProvider, byUpdatedAt)

	return donor, attorney, certificateProvider, nil
}
