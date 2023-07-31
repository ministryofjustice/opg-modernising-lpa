package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"golang.org/x/exp/slices"
)

// An lpaLink is used to join an actor to an LPA.
type lpaLink struct {
	// PK is the same as the PK for the LPA
	PK string
	// SK is the subKey for the current user
	SK string
	// DonorKey is the donorKey for the donor
	DonorKey string
	// ActorType is the type for the current user
	ActorType actor.Type
}

type dashboardStore struct {
	dynamoClient DynamoClient
}

func (s *dashboardStore) GetAll(ctx context.Context) (donor, attorney, certificateProvider []page.LpaAndActorTasks, err error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if data.SessionID == "" {
		return nil, nil, nil, errors.New("donorStore.GetAll requires SessionID")
	}

	var keys []lpaLink
	if err := s.dynamoClient.GetAllByGsi(ctx, "ActorIndex", subKey(data.SessionID), &keys); err != nil {
		return nil, nil, nil, err
	}

	var searchKeys []dynamo.Key
	keyMap := map[string]actor.Type{}
	for _, key := range keys {
		searchKeys = append(searchKeys, dynamo.Key{PK: key.PK, SK: key.DonorKey})

		if key.ActorType == actor.TypeAttorney {
			searchKeys = append(searchKeys, dynamo.Key{PK: key.PK, SK: attorneyKey(data.SessionID)})
		}

		if key.ActorType == actor.TypeCertificateProvider {
			searchKeys = append(searchKeys, dynamo.Key{PK: key.PK, SK: certificateProviderKey(data.SessionID)})
		}

		_, id, _ := strings.Cut(key.PK, "#")
		keyMap[id] = key.ActorType
	}

	if len(searchKeys) == 0 {
		return nil, nil, nil, nil
	}

	var lpasOrProvidedDetails []interface{}
	if err := s.dynamoClient.GetAllByKeys(ctx, searchKeys, &lpasOrProvidedDetails); err != nil {
		return nil, nil, nil, err
	}

	certificateProviderMap := make(map[string]page.LpaAndActorTasks)
	attorneyMap := make(map[string]page.LpaAndActorTasks)

	for _, item := range lpasOrProvidedDetails {
		jsonData, _ := json.Marshal(item)

		var lpa *page.Lpa
		err := json.Unmarshal(jsonData, &lpa)

		if err == nil && strings.Contains(lpa.SK, donorKey("")) {
			switch keyMap[lpa.ID] {
			case actor.TypeDonor:
				donor = append(donor, page.LpaAndActorTasks{Lpa: lpa})
			case actor.TypeAttorney:
				attorneyMap[lpa.ID] = page.LpaAndActorTasks{Lpa: lpa}
			case actor.TypeCertificateProvider:
				certificateProviderMap[lpa.ID] = page.LpaAndActorTasks{Lpa: lpa}
			}
		}
	}

	for _, item := range lpasOrProvidedDetails {
		jsonData, _ := json.Marshal(item)

		var attorneyProvidedDetails *actor.AttorneyProvidedDetails
		err = json.Unmarshal(jsonData, &attorneyProvidedDetails)
		if err == nil && strings.Contains(attorneyProvidedDetails.SK, attorneyKey("")) {
			if entry, ok := attorneyMap[attorneyProvidedDetails.LpaID]; ok {
				entry.AttorneyTasks = attorneyProvidedDetails.Tasks
				attorneyMap[attorneyProvidedDetails.LpaID] = entry
				continue
			}
		}

		var certificateProviderProvidedDetails *actor.CertificateProviderProvidedDetails
		err = json.Unmarshal(jsonData, &certificateProviderProvidedDetails)
		if err == nil && strings.Contains(certificateProviderProvidedDetails.SK, certificateProviderKey("")) {
			if entry, ok := certificateProviderMap[certificateProviderProvidedDetails.LpaID]; ok {
				entry.CertificateProviderTasks = certificateProviderProvidedDetails.Tasks
				certificateProviderMap[certificateProviderProvidedDetails.LpaID] = entry
			}
		}
	}

	for _, value := range certificateProviderMap {
		certificateProvider = append(certificateProvider, value)
	}

	for _, value := range attorneyMap {
		attorney = append(attorney, value)
	}

	byUpdatedAt := func(a, b page.LpaAndActorTasks) bool {
		return a.Lpa.UpdatedAt.After(b.Lpa.UpdatedAt)
	}

	slices.SortFunc(donor, byUpdatedAt)
	slices.SortFunc(attorney, byUpdatedAt)
	slices.SortFunc(certificateProvider, byUpdatedAt)

	return donor, attorney, certificateProvider, nil
}
