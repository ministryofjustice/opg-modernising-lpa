package app

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type LpaStoreResolvingService interface {
	Resolve(ctx context.Context, donor *actor.DonorProvidedDetails) (*lpastore.Lpa, error)
}

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
	// UpdatedAt is set to allow this data to be queried from SKUpdatedAtIndex
	UpdatedAt time.Time
}

func (l lpaLink) UserSub() string {
	if l.SK == "" {
		return ""
	}

	return strings.Split(l.SK, dynamo.SubKey(""))[1]
}

type dashboardStore struct {
	dynamoClient             DynamoClient
	lpaStoreResolvingService LpaStoreResolvingService
}

type keys struct {
	PK, SK string
}

func (k keys) isLpa() bool {
	return strings.HasPrefix(k.SK, dynamo.DonorKey("")) || strings.HasPrefix(k.SK, dynamo.OrganisationKey(""))
}

func (k keys) isCertificateProviderDetails() bool {
	return strings.HasPrefix(k.SK, dynamo.CertificateProviderKey(""))
}

func (k keys) isAttorneyDetails() bool {
	return strings.HasPrefix(k.SK, dynamo.AttorneyKey(""))
}

func (s *dashboardStore) SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error) {
	var links []lpaLink
	if err := s.dynamoClient.AllBySK(ctx, dynamo.SubKey(sub), &links); err != nil {
		return false, err
	}

	for _, link := range links {
		if link.ActorType == actorType {
			return true, nil
		}
	}

	return false, nil
}

func (s *dashboardStore) GetAll(ctx context.Context) (donor, attorney, certificateProvider []page.LpaAndActorTasks, err error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if data.SessionID == "" {
		return nil, nil, nil, errors.New("donorStore.GetAll requires SessionID")
	}

	var links []lpaLink
	if err := s.dynamoClient.AllBySK(ctx, dynamo.SubKey(data.SessionID), &links); err != nil {
		return nil, nil, nil, err
	}

	var searchKeys []dynamo.Keys
	keyMap := map[string]actor.Type{}
	for _, key := range links {
		searchKeys = append(searchKeys, dynamo.Keys{PK: key.PK, SK: key.DonorKey})

		if key.ActorType == actor.TypeAttorney {
			searchKeys = append(searchKeys, dynamo.Keys{PK: key.PK, SK: dynamo.AttorneyKey(data.SessionID)})
		}

		if key.ActorType == actor.TypeCertificateProvider {
			searchKeys = append(searchKeys, dynamo.Keys{PK: key.PK, SK: dynamo.CertificateProviderKey(data.SessionID)})
		}

		_, id, _ := strings.Cut(key.PK, "#")
		keyMap[id] = key.ActorType
	}

	if len(searchKeys) == 0 {
		return nil, nil, nil, nil
	}

	lpasOrProvidedDetails, err := s.dynamoClient.AllByKeys(ctx, searchKeys)
	if err != nil {
		return nil, nil, nil, err
	}

	certificateProviderMap := map[string]page.LpaAndActorTasks{}
	attorneyMap := map[string]page.LpaAndActorTasks{}

	for _, item := range lpasOrProvidedDetails {
		var ks keys
		if err = attributevalue.UnmarshalMap(item, &ks); err != nil {
			return nil, nil, nil, err
		}

		if ks.isLpa() {
			donorDetails := &actor.DonorProvidedDetails{}
			if err := attributevalue.UnmarshalMap(item, donorDetails); err != nil {
				return nil, nil, nil, err
			}

			if donorDetails.LpaUID == "" {
				continue
			}

			lpa, err := s.lpaStoreResolvingService.Resolve(ctx, donorDetails)
			if err != nil {
				return nil, nil, nil, err
			}

			switch keyMap[donorDetails.LpaID] {
			case actor.TypeDonor:
				donor = append(donor, page.LpaAndActorTasks{Lpa: lpa})
			case actor.TypeAttorney:
				attorneyMap[donorDetails.LpaID] = page.LpaAndActorTasks{Lpa: lpa}
			case actor.TypeCertificateProvider:
				certificateProviderMap[donorDetails.LpaID] = page.LpaAndActorTasks{Lpa: lpa}
			}
		}
	}

	for _, item := range lpasOrProvidedDetails {
		var ks keys
		if err = attributevalue.UnmarshalMap(item, &ks); err != nil {
			return nil, nil, nil, err
		}

		if ks.isAttorneyDetails() {
			attorneyProvidedDetails := &actor.AttorneyProvidedDetails{}
			if err := attributevalue.UnmarshalMap(item, attorneyProvidedDetails); err != nil {
				return nil, nil, nil, err
			}

			lpaID := attorneyProvidedDetails.LpaID

			if entry, ok := attorneyMap[lpaID]; ok {
				if attorneyProvidedDetails.IsReplacement && entry.Lpa.Submitted {
					delete(attorneyMap, lpaID)
					continue
				}

				entry.Attorney = attorneyProvidedDetails
				attorneyMap[lpaID] = entry
				continue
			}
		}

		if ks.isCertificateProviderDetails() {
			certificateProviderProvidedDetails := &actor.CertificateProviderProvidedDetails{}
			if err := attributevalue.UnmarshalMap(item, certificateProviderProvidedDetails); err != nil {
				return nil, nil, nil, err
			}

			lpaID := certificateProviderProvidedDetails.LpaID

			if certificateProviderProvidedDetails.Certificate.AgreeToStatement {
				delete(certificateProviderMap, lpaID)
			}

			if entry, ok := certificateProviderMap[lpaID]; ok {
				entry.CertificateProvider = certificateProviderProvidedDetails
				certificateProviderMap[lpaID] = entry
			}
		}
	}

	certificateProvider = mapValues(certificateProviderMap)
	attorney = mapValues(attorneyMap)

	byUpdatedAt := func(a, b page.LpaAndActorTasks) int {
		if a.Lpa.UpdatedAt.After(b.Lpa.UpdatedAt) {
			return -1
		}
		return 1
	}

	slices.SortFunc(donor, byUpdatedAt)
	slices.SortFunc(attorney, byUpdatedAt)
	slices.SortFunc(certificateProvider, byUpdatedAt)

	return donor, attorney, certificateProvider, nil
}

func mapValues[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}
