package dashboard

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type DynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	OneByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	AllByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	LatestForActor(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllByKeys(ctx context.Context, keys []dynamo.Keys) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllKeysByPK(ctx context.Context, pk dynamo.PK) ([]dynamo.Keys, error)
	Put(ctx context.Context, v interface{}) error
	Create(ctx context.Context, v interface{}) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	Update(ctx context.Context, pk dynamo.PK, sk dynamo.SK, values map[string]dynamodbtypes.AttributeValue, expression string) error
	BatchPut(ctx context.Context, items []interface{}) error
	OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type LpaStoreResolvingService interface {
	ResolveList(ctx context.Context, donors []*donordata.Provided) ([]*lpadata.Lpa, error)
}

type dashboardStore struct {
	dynamoClient             DynamoClient
	lpaStoreResolvingService LpaStoreResolvingService
}

func NewStore(dynamoClient DynamoClient, lpaStoreResolvingService LpaStoreResolvingService) *dashboardStore {
	return &dashboardStore{
		dynamoClient:             dynamoClient,
		lpaStoreResolvingService: lpaStoreResolvingService,
	}
}

func isLpaKey(k dynamo.Keys) bool {
	_, donorOK := k.SK.(dynamo.DonorKeyType)
	_, orgOK := k.SK.(dynamo.OrganisationKeyType)

	return donorOK || orgOK
}

func isCertificateProviderKey(k dynamo.Keys) bool {
	_, ok := k.SK.(dynamo.CertificateProviderKeyType)
	return ok
}

func isAttorneyKey(k dynamo.Keys) bool {
	_, ok := k.SK.(dynamo.AttorneyKeyType)
	return ok
}

func (s *dashboardStore) SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error) {
	var links []dashboarddata.LpaLink
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
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if data.SessionID == "" {
		return nil, nil, nil, errors.New("donorStore.GetAll requires SessionID")
	}

	var links []dashboarddata.LpaLink
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

		_, id, _ := strings.Cut(key.PK.PK(), "#")
		keyMap[id] = key.ActorType
	}

	if len(searchKeys) == 0 {
		return nil, nil, nil, nil
	}

	lpasOrProvidedDetails, err := s.dynamoClient.AllByKeys(ctx, searchKeys)
	if err != nil {
		return nil, nil, nil, err
	}

	var (
		referencedKeys []dynamo.Keys
		donorsDetails  []*donordata.Provided
	)
	for _, item := range lpasOrProvidedDetails {
		var ks dynamo.Keys
		if err = attributevalue.UnmarshalMap(item, &ks); err != nil {
			return nil, nil, nil, err
		}

		if isLpaKey(ks) {
			var donorDetails struct {
				donordata.Provided
				ReferencedSK dynamo.OrganisationKeyType
			}
			if err := attributevalue.UnmarshalMap(item, &donorDetails); err != nil {
				return nil, nil, nil, err
			}

			if donorDetails.ReferencedSK != "" {
				referencedKeys = append(referencedKeys, dynamo.Keys{PK: ks.PK, SK: donorDetails.ReferencedSK})
			} else if donorDetails.LpaUID != "" {
				donorsDetails = append(donorsDetails, &donorDetails.Provided)
			}
		}
	}

	if len(referencedKeys) > 0 {
		referencedLpas, err := s.dynamoClient.AllByKeys(ctx, referencedKeys)
		if err != nil {
			return nil, nil, nil, err
		}

		for _, item := range referencedLpas {
			donorDetails := &donordata.Provided{}
			if err := attributevalue.UnmarshalMap(item, donorDetails); err != nil {
				return nil, nil, nil, err
			}

			if donorDetails.LpaUID != "" {
				donorsDetails = append(donorsDetails, donorDetails)
			}
		}
	}

	resolvedLpas, err := s.lpaStoreResolvingService.ResolveList(ctx, donorsDetails)
	if err != nil {
		return nil, nil, nil, err
	}

	certificateProviderMap := map[string]page.LpaAndActorTasks{}
	attorneyMap := map[string]page.LpaAndActorTasks{}

	for _, lpa := range resolvedLpas {
		switch keyMap[lpa.LpaID] {
		case actor.TypeDonor:
			donor = append(donor, page.LpaAndActorTasks{Lpa: lpa})
		case actor.TypeAttorney:
			attorneyMap[lpa.LpaID] = page.LpaAndActorTasks{Lpa: lpa}
		case actor.TypeCertificateProvider:
			certificateProviderMap[lpa.LpaID] = page.LpaAndActorTasks{Lpa: lpa}
		}
	}

	for _, item := range lpasOrProvidedDetails {
		var ks dynamo.Keys
		if err = attributevalue.UnmarshalMap(item, &ks); err != nil {
			return nil, nil, nil, err
		}

		if isAttorneyKey(ks) {
			attorneyProvidedDetails := &attorneydata.Provided{}
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

		if isCertificateProviderKey(ks) {
			certificateProviderProvidedDetails := &certificateproviderdata.Provided{}
			if err := attributevalue.UnmarshalMap(item, certificateProviderProvidedDetails); err != nil {
				return nil, nil, nil, err
			}

			lpaID := certificateProviderProvidedDetails.LpaID

			if !certificateProviderProvidedDetails.SignedAt.IsZero() {
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
		return b.Lpa.UpdatedAt.Compare(a.Lpa.UpdatedAt)
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
