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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type DynamoClient interface {
	AllBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllByKeys(ctx context.Context, keys []dynamo.Keys) ([]map[string]dynamodbtypes.AttributeValue, error)
}

type LpaStoreResolvingService interface {
	ResolveList(ctx context.Context, donors []*donordata.Provided) ([]*lpadata.Lpa, error)
}

type Store struct {
	dynamoClient             DynamoClient
	lpaStoreResolvingService LpaStoreResolvingService
}

func NewStore(dynamoClient DynamoClient, lpaStoreResolvingService LpaStoreResolvingService) *Store {
	return &Store{
		dynamoClient:             dynamoClient,
		lpaStoreResolvingService: lpaStoreResolvingService,
	}
}

func (s *Store) SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error) {
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

func (s *Store) GetAll(ctx context.Context) (results dashboarddata.Results, err error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return results, err
	}

	if data.SessionID == "" {
		return results, errors.New("donorStore.GetAll requires SessionID")
	}

	var links []dashboarddata.LpaLink
	if err := s.dynamoClient.AllBySK(ctx, dynamo.SubKey(data.SessionID), &links); err != nil {
		return results, err
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

		if key.ActorType == actor.TypeVoucher {
			searchKeys = append(searchKeys, dynamo.Keys{PK: key.PK, SK: dynamo.VoucherKey(data.SessionID)})
		}

		_, id, _ := strings.Cut(key.PK.PK(), "#")
		keyMap[id] = key.ActorType
	}

	if len(searchKeys) == 0 {
		return results, nil
	}

	lpasOrProvidedDetails, err := s.dynamoClient.AllByKeys(ctx, searchKeys)
	if err != nil {
		return results, err
	}

	var (
		referencedKeys []dynamo.Keys
		donorsDetails  []*donordata.Provided
	)
	for _, item := range lpasOrProvidedDetails {
		var ks dynamo.Keys
		if err = attributevalue.UnmarshalMap(item, &ks); err != nil {
			return results, err
		}

		switch ks.SK.(type) {
		case dynamo.DonorKeyType, dynamo.OrganisationKeyType:
			var donorDetails struct {
				donordata.Provided
				ReferencedSK dynamo.OrganisationKeyType
			}
			if err := attributevalue.UnmarshalMap(item, &donorDetails); err != nil {
				return results, err
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
			return results, err
		}

		for _, item := range referencedLpas {
			donorDetails := &donordata.Provided{}
			if err := attributevalue.UnmarshalMap(item, donorDetails); err != nil {
				return results, err
			}

			if donorDetails.LpaUID != "" {
				donorsDetails = append(donorsDetails, donorDetails)
			}
		}
	}

	resolvedLpas, err := s.lpaStoreResolvingService.ResolveList(ctx, donorsDetails)
	if err != nil {
		return results, err
	}

	certificateProviderMap := map[string]dashboarddata.Actor{}
	attorneyMap := map[string]dashboarddata.Actor{}
	voucherMap := map[string]dashboarddata.Actor{}

	for _, lpa := range resolvedLpas {
		switch keyMap[lpa.LpaID] {
		case actor.TypeDonor:
			results.Donor = append(results.Donor, dashboarddata.Actor{Lpa: lpa})
		case actor.TypeAttorney:
			attorneyMap[lpa.LpaID] = dashboarddata.Actor{Lpa: lpa}
		case actor.TypeCertificateProvider:
			certificateProviderMap[lpa.LpaID] = dashboarddata.Actor{Lpa: lpa}
		case actor.TypeVoucher:
			voucherMap[lpa.LpaID] = dashboarddata.Actor{Lpa: lpa}
		}
	}

	for _, item := range lpasOrProvidedDetails {
		var ks dynamo.Keys
		if err = attributevalue.UnmarshalMap(item, &ks); err != nil {
			return results, err
		}

		switch ks.SK.(type) {
		case dynamo.AttorneyKeyType:
			attorneyProvidedDetails := &attorneydata.Provided{}
			if err := attributevalue.UnmarshalMap(item, attorneyProvidedDetails); err != nil {
				return results, err
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

		case dynamo.CertificateProviderKeyType:
			certificateProviderProvidedDetails := &certificateproviderdata.Provided{}
			if err := attributevalue.UnmarshalMap(item, certificateProviderProvidedDetails); err != nil {
				return results, err
			}

			lpaID := certificateProviderProvidedDetails.LpaID

			if !certificateProviderProvidedDetails.SignedAt.IsZero() {
				delete(certificateProviderMap, lpaID)
			}

			if entry, ok := certificateProviderMap[lpaID]; ok {
				entry.CertificateProvider = certificateProviderProvidedDetails
				certificateProviderMap[lpaID] = entry
			}

		case dynamo.VoucherKeyType:
			voucherProvidedDetails := &voucherdata.Provided{}
			if err := attributevalue.UnmarshalMap(item, voucherProvidedDetails); err != nil {
				return results, err
			}

			lpaID := voucherProvidedDetails.LpaID

			if voucherProvidedDetails.Tasks.SignTheDeclaration.IsCompleted() ||
				(voucherProvidedDetails.Tasks.ConfirmYourIdentity.IsCompleted() && !voucherProvidedDetails.IdentityConfirmed()) {
				delete(voucherMap, lpaID)
			}

			if entry, ok := voucherMap[lpaID]; ok {
				entry.Voucher = voucherProvidedDetails
				voucherMap[lpaID] = entry
			}
		}
	}

	results.CertificateProvider = mapValues(certificateProviderMap)
	results.Attorney = mapValues(attorneyMap)
	results.Voucher = mapValues(voucherMap)

	byUpdatedAt := func(a, b dashboarddata.Actor) int {
		return b.Lpa.UpdatedAt.Compare(a.Lpa.UpdatedAt)
	}

	slices.SortFunc(results.Donor, byUpdatedAt)
	slices.SortFunc(results.Attorney, byUpdatedAt)
	slices.SortFunc(results.CertificateProvider, byUpdatedAt)
	slices.SortFunc(results.Voucher, byUpdatedAt)

	return results, nil
}

func mapValues[M ~map[K]V, K comparable, V any](m M) []V {
	if len(m) == 0 {
		return nil
	}

	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}
