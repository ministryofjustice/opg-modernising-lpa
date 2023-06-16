package app

import (
	"context"
	"errors"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"golang.org/x/exp/slices"
)

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

	var keys []struct {
		PK   string
		SK   string
		Data string
	}
	if err := s.dataStore.GetAllByGsi(ctx, "ActorIndex", "#SUB#"+data.SessionID, &keys); err != nil {
		return nil, nil, nil, err
	}

	searchKeys := make([]dynamo.Key, len(keys))
	keyMap := map[string]string{}
	for i, key := range keys {
		sk, actorType, _ := strings.Cut(key.Data, "|")
		searchKeys[i] = dynamo.Key{PK: key.PK, SK: sk}
		keyMap[key.PK] = actorType
	}

	var result []struct {
		PK   string
		Data *page.Lpa
	}
	if err := s.dataStore.GetAllByKeys(ctx, searchKeys, &result); err != nil {
		return nil, nil, nil, err
	}

	for _, item := range result {
		switch keyMap[item.PK] {
		case "DONOR":
			donor = append(donor, item.Data)
		case "ATTORNEY":
			attorney = append(attorney, item.Data)
		case "CERTIFICATE_PROVIDER":
			certificateProvider = append(certificateProvider, item.Data)
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
