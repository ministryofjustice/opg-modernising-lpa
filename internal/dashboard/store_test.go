package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func (m *mockDynamoClient) ExpectAllBySK(ctx, sk, data interface{}, err error) {
	m.
		On("AllBySK", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllByKeys(ctx context.Context, keys []dynamo.Keys, data []map[string]types.AttributeValue, err error) {
	m.EXPECT().
		AllByKeys(ctx, keys).
		Return(data, err)
}

func TestDashboardStoreGetAll(t *testing.T) {
	sessionID := "an-id"
	aTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	lpa0 := &lpadata.Lpa{LpaID: "0", LpaUID: "M-0", UpdatedAt: aTime}
	lpa0Donor := &donordata.Provided{
		PK:        dynamo.LpaKey("0"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey(sessionID)),
		LpaID:     "0",
		LpaUID:    "M-0",
		UpdatedAt: aTime,
	}
	lpa123 := &lpadata.Lpa{LpaID: "123", LpaUID: "M-123", UpdatedAt: aTime}
	lpa123Donor := &donordata.Provided{
		PK:        dynamo.LpaKey("123"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey(sessionID)),
		LpaID:     "123",
		LpaUID:    "M-123",
		UpdatedAt: aTime,
	}
	lpa456 := &lpadata.Lpa{LpaID: "456", LpaUID: "M-456"}
	lpa456Donor := &donordata.Provided{
		PK:     dynamo.LpaKey("456"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")),
		LpaID:  "456",
		LpaUID: "M-456",
	}
	lpa456CertificateProvider := &certificateproviderdata.Provided{
		PK:    dynamo.LpaKey("456"),
		SK:    dynamo.CertificateProviderKey(sessionID),
		LpaID: "456",
		Tasks: certificateproviderdata.Tasks{ConfirmYourDetails: task.StateCompleted},
	}
	lpa789 := &lpadata.Lpa{LpaID: "789", LpaUID: "M-789"}
	lpa789Donor := &donordata.Provided{
		PK:     dynamo.LpaKey("789"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("different-id")),
		LpaID:  "789",
		LpaUID: "M-789",
	}
	lpa789Attorney := &attorneydata.Provided{
		PK:    dynamo.LpaKey("789"),
		SK:    dynamo.AttorneyKey(sessionID),
		LpaID: "789",
		Tasks: attorneydata.Tasks{ConfirmYourDetails: task.StateInProgress},
	}
	actorUID, _ := actoruid.Parse("789-paper-attorney-uid")
	lpa789PaperAttorney := lpadata.Attorney{
		UID:      actorUID,
		Channel:  lpadata.ChannelPaper,
		SignedAt: &aTime,
	}
	lpa789.Attorneys = lpadata.Attorneys{Attorneys: []lpadata.Attorney{lpa789PaperAttorney}}
	lpaNoUIDDonor := &donordata.Provided{
		PK:        dynamo.LpaKey("0"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey(sessionID)),
		LpaID:     "999",
		UpdatedAt: aTime,
	}
	lpaCertified := &lpadata.Lpa{LpaID: "signed-by-cp", LpaUID: "M-signed-by-cp"}
	lpaCertifiedDonor := &donordata.Provided{
		PK:     dynamo.LpaKey("signed-by-cp"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")),
		LpaID:  "signed-by-cp",
		LpaUID: "M-signed-by-cp",
	}
	lpaCertifiedCertificateProvider := &certificateproviderdata.Provided{
		PK:       dynamo.LpaKey("signed-by-cp"),
		SK:       dynamo.CertificateProviderKey(sessionID),
		LpaID:    "signed-by-cp",
		SignedAt: time.Now(),
		Tasks:    certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
	}
	lpaReferenced := &lpadata.Lpa{LpaID: "referenced", LpaUID: "X"}
	lpaReferencedLink := map[string]any{
		"PK":           dynamo.LpaKey("referenced"),
		"SK":           dynamo.DonorKey(sessionID),
		"ReferencedSK": dynamo.OrganisationKey("org-id"),
	}
	lpaReferencedDonor := &donordata.Provided{
		PK:     dynamo.LpaKey("referenced"),
		SK:     dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
		LpaID:  "referenced",
		LpaUID: "X",
	}
	lpaVouched := &lpadata.Lpa{LpaID: "vouched", LpaUID: "V"}
	lpaVouchedDonor := &donordata.Provided{
		PK:     dynamo.LpaKey("vouched"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("vouched-id")),
		LpaID:  "vouched",
		LpaUID: "V",
	}
	lpaVouchedVoucher := &voucherdata.Provided{
		PK:    dynamo.LpaKey("vouched"),
		SK:    dynamo.VoucherKey(sessionID),
		LpaID: "vouched",
		Tasks: voucherdata.Tasks{ConfirmYourName: task.StateCompleted},
	}
	lpaVouchCompleted := &lpadata.Lpa{LpaID: "vouch-completed", LpaUID: "V"}
	lpaVouchCompletedDonor := &donordata.Provided{
		PK:     dynamo.LpaKey("vouch-completed"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("vouch-completed-id")),
		LpaID:  "vouch-completed",
		LpaUID: "V",
	}
	lpaVouchCompletedVoucher := &voucherdata.Provided{
		PK:    dynamo.LpaKey("vouch-completed"),
		SK:    dynamo.VoucherKey(sessionID),
		LpaID: "vouch-completed",
		Tasks: voucherdata.Tasks{ConfirmYourName: task.StateCompleted, SignTheDeclaration: task.StateCompleted},
	}

	testCases := map[string][]map[string]types.AttributeValue{
		"details returned after lpas": {
			makeAttributeValueMap(lpa123Donor),
			makeAttributeValueMap(lpa456Donor),
			makeAttributeValueMap(lpa456CertificateProvider),
			makeAttributeValueMap(lpa789Donor),
			makeAttributeValueMap(lpa789Attorney),
			makeAttributeValueMap(lpa0Donor),
			makeAttributeValueMap(lpaCertifiedDonor),
			makeAttributeValueMap(lpaCertifiedCertificateProvider),
			makeAttributeValueMap(lpaReferencedLink),
			makeAttributeValueMap(lpaVouchedDonor),
			makeAttributeValueMap(lpaVouchedVoucher),
			makeAttributeValueMap(lpaVouchCompletedDonor),
			makeAttributeValueMap(lpaVouchCompletedVoucher),
		},
		"details returned before lpas": {
			makeAttributeValueMap(lpaNoUIDDonor),
			makeAttributeValueMap(lpa456CertificateProvider),
			makeAttributeValueMap(lpa789Attorney),
			makeAttributeValueMap(lpaCertifiedCertificateProvider),
			makeAttributeValueMap(lpa123Donor),
			makeAttributeValueMap(lpa456Donor),
			makeAttributeValueMap(lpa789Donor),
			makeAttributeValueMap(lpa0Donor),
			makeAttributeValueMap(lpaCertifiedDonor),
			makeAttributeValueMap(lpaReferencedLink),
			makeAttributeValueMap(lpaVouchedVoucher),
			makeAttributeValueMap(lpaVouchedDonor),
			makeAttributeValueMap(lpaVouchCompletedVoucher),
			makeAttributeValueMap(lpaVouchCompletedDonor),
		},
	}

	for name, attributeValues := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: sessionID})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
				[]dashboarddata.LpaLink{
					{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("456"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")), ActorType: actor.TypeCertificateProvider},
					{PK: dynamo.LpaKey("789"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("different-id")), ActorType: actor.TypeAttorney},
					{PK: dynamo.LpaKey("0"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("999"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")), ActorType: actor.TypeCertificateProvider},
					{PK: dynamo.LpaKey("referenced"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("vouched"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeVoucher},
					{PK: dynamo.LpaKey("vouch-completed"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeVoucher},
				}, nil)
			dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
				{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("456"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id"))},
				{PK: dynamo.LpaKey("456"), SK: dynamo.CertificateProviderKey("an-id")},
				{PK: dynamo.LpaKey("789"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("different-id"))},
				{PK: dynamo.LpaKey("789"), SK: dynamo.AttorneyKey("an-id")},
				{PK: dynamo.LpaKey("0"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("999"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id"))},
				{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.CertificateProviderKey("an-id")},
				{PK: dynamo.LpaKey("referenced"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("vouched"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("vouched"), SK: dynamo.VoucherKey("an-id")},
				{PK: dynamo.LpaKey("vouch-completed"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("vouch-completed"), SK: dynamo.VoucherKey("an-id")},
			}, attributeValues, nil)
			dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
				{PK: dynamo.LpaKey("referenced"), SK: dynamo.OrganisationKey("org-id")},
			}, []map[string]types.AttributeValue{
				makeAttributeValueMap(lpaReferencedDonor),
			}, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				ResolveList(ctx, []*donordata.Provided{lpa123Donor, lpa456Donor, lpa789Donor, lpa0Donor, lpaCertifiedDonor, lpaVouchedDonor, lpaVouchCompletedDonor, lpaReferencedDonor}).
				Return([]*lpadata.Lpa{lpa123, lpa456, lpa789, lpa0, lpaCertified, lpaVouched, lpaVouchCompleted, lpaReferenced}, nil)

			dashboardStore := &Store{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}

			results, err := dashboardStore.GetAll(ctx)
			assert.Nil(t, err)

			assert.Equal(t, dashboarddata.Results{
				Donor:               []dashboarddata.Actor{{Lpa: lpa123, Donor: lpa123Donor}, {Lpa: lpa0, Donor: lpa0Donor}, {Lpa: lpaReferenced}},
				CertificateProvider: []dashboarddata.Actor{{Lpa: lpa456, CertificateProvider: lpa456CertificateProvider}},
				Attorney:            []dashboarddata.Actor{{Lpa: lpa789, Attorney: lpa789Attorney, LpaAttorney: &lpa789PaperAttorney}},
				Voucher:             []dashboarddata.Actor{{Lpa: lpaVouched, Voucher: lpaVouchedVoucher}},
			}, results)
		})
	}
}

func TestDashboardStoreGetAllSubmittedForAttorneys(t *testing.T) {
	sessionID := "an-id"

	lpaSubmitted := &lpadata.Lpa{LpaID: "submitted", LpaUID: "M", Submitted: true}
	lpaSubmittedDonor := &donordata.Provided{
		PK:     dynamo.LpaKey("submitted"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")),
		LpaID:  "submitted",
		LpaUID: "M",
	}
	lpaSubmittedAttorney := &attorneydata.Provided{
		PK:    dynamo.LpaKey("submitted"),
		SK:    dynamo.AttorneyKey(sessionID),
		LpaID: "submitted",
	}
	lpaSubmittedReplacement := &lpadata.Lpa{LpaID: "submitted-replacement", LpaUID: "M", Submitted: true}
	lpaSubmittedReplacementDonor := &donordata.Provided{
		PK:     dynamo.LpaKey("submitted-replacement"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")),
		LpaID:  "submitted-replacement",
		LpaUID: "M",
	}
	lpaSubmittedReplacementAttorney := &attorneydata.Provided{
		PK:            dynamo.LpaKey("submitted-replacement"),
		SK:            dynamo.AttorneyKey(sessionID),
		LpaID:         "submitted-replacement",
		IsReplacement: true,
	}

	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: sessionID})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]dashboarddata.LpaLink{
			{PK: dynamo.LpaKey("submitted"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")), ActorType: actor.TypeAttorney},
			{PK: dynamo.LpaKey("submitted-replacement"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")), ActorType: actor.TypeAttorney},
		}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("submitted"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id"))},
		{PK: dynamo.LpaKey("submitted"), SK: dynamo.AttorneyKey("an-id")},
		{PK: dynamo.LpaKey("submitted-replacement"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id"))},
		{PK: dynamo.LpaKey("submitted-replacement"), SK: dynamo.AttorneyKey("an-id")},
	}, []map[string]types.AttributeValue{
		makeAttributeValueMap(lpaSubmittedDonor),
		makeAttributeValueMap(lpaSubmittedAttorney),
		makeAttributeValueMap(lpaSubmittedReplacementDonor),
		makeAttributeValueMap(lpaSubmittedReplacementAttorney),
	}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		ResolveList(ctx, []*donordata.Provided{lpaSubmittedDonor, lpaSubmittedReplacementDonor}).
		Return([]*lpadata.Lpa{lpaSubmitted, lpaSubmittedReplacement}, nil)

	dashboardStore := &Store{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}

	results, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)

	assert.Equal(t, dashboarddata.Results{
		Attorney: []dashboarddata.Actor{
			{Lpa: lpaSubmitted, Attorney: lpaSubmittedAttorney},
		},
	}, results)
}

func makeAttributeValueMap(i any) map[string]types.AttributeValue {
	result, _ := attributevalue.MarshalMap(i)
	return result
}

func TestDashboardStoreGetAllWhenResolveErrors(t *testing.T) {
	sessionID := "an-id"
	aTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	donor := &donordata.Provided{LpaID: "0", LpaUID: "M", UpdatedAt: aTime, SK: dynamo.LpaOwnerKey(dynamo.DonorKey(sessionID)), PK: dynamo.LpaKey("0")}

	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: sessionID})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]dashboarddata.LpaLink{
			{PK: dynamo.LpaKey("0"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
		}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("0"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
	}, []map[string]types.AttributeValue{makeAttributeValueMap(donor)}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().ResolveList(ctx, mock.Anything).Return(nil, expectedError)

	dashboardStore := &Store{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}

	_, err := dashboardStore.GetAll(ctx)
	if !assert.Equal(t, expectedError, err) {
		t.Log(err.Error())
	}
}

func TestDashboardStoreGetAllWhenNone(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]map[string]any{}, nil)

	dashboardStore := &Store{dynamoClient: dynamoClient}

	results, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, dashboarddata.Results{}, results)
}

func TestDashboardStoreGetAllWhenNoneWithUID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]dashboarddata.LpaLink{
			{PK: dynamo.LpaKey("a"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("b")), ActorType: actor.TypeDonor},
		}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{{PK: dynamo.LpaKey("a"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("b"))}},
		[]map[string]types.AttributeValue{
			makeAttributeValueMap(&donordata.Provided{
				PK: dynamo.LpaKey("a"),
				SK: dynamo.LpaOwnerKey(dynamo.DonorKey("b")),
			}),
		}, nil)

	dashboardStore := &Store{dynamoClient: dynamoClient}

	results, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, dashboarddata.Results{}, results)
}

func TestDashboardStoreGetAllWhenAllForActorErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]dashboarddata.LpaLink{}, expectedError)

	dashboardStore := &Store{dynamoClient: dynamoClient}

	_, err := dashboardStore.GetAll(ctx)
	assert.Equal(t, err, expectedError)
}

func TestDashboardStoreGetAllWhenAllByKeysErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]dashboarddata.LpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor}}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
	}, nil, expectedError)

	dashboardStore := &Store{dynamoClient: dynamoClient}

	_, err := dashboardStore.GetAll(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDashboardStoreGetAllWhenReferenceGetErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]dashboarddata.LpaLink{{
			PK:        dynamo.LpaKey("123"),
			SK:        dynamo.SubKey("an-id"),
			DonorKey:  dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
			ActorType: actor.TypeDonor,
		}}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
	}, []map[string]types.AttributeValue{
		makeAttributeValueMap(map[string]any{
			"PK":           dynamo.LpaKey("123"),
			"SK":           dynamo.DonorKey("an-id"),
			"ReferencedSK": dynamo.OrganisationKey("org-id"),
		}),
	}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("123"), SK: dynamo.OrganisationKey("org-id")},
	}, nil, expectedError)

	dashboardStore := &Store{dynamoClient: dynamoClient}

	_, err := dashboardStore.GetAll(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDashboardStoreSubExists(t *testing.T) {
	testCases := map[string]struct {
		links          []dashboarddata.LpaLink
		keys           []dynamo.Keys
		attributes     []map[string]types.AttributeValue
		lpas           []*lpadata.Lpa
		expectedExists bool
		actorType      actor.Type
	}{
		"lpas exist - correct actor": {
			links: []dashboarddata.LpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor}},
			keys:  []dynamo.Keys{{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))}},
			attributes: []map[string]types.AttributeValue{
				makeAttributeValueMap(&donordata.Provided{
					PK:     dynamo.LpaKey("123"),
					SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("abc")),
					LpaID:  "123",
					LpaUID: "M-0",
				}),
			},
			lpas: []*lpadata.Lpa{{
				LpaID: "123",
			}},
			expectedExists: true,
			actorType:      actor.TypeDonor,
		},
		"lpas exist - incorrect actor": {
			links: []dashboarddata.LpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor}},
			keys:  []dynamo.Keys{{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))}},
			attributes: []map[string]types.AttributeValue{
				makeAttributeValueMap(&donordata.Provided{
					PK:     dynamo.LpaKey("123"),
					SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("abc")),
					LpaID:  "123",
					LpaUID: "M-0",
				}),
			},
			lpas: []*lpadata.Lpa{{
				LpaID: "123",
			}},
			expectedExists: false,
			actorType:      actor.TypeAttorney,
		},
		"certificate provider should not see": {
			links: []dashboarddata.LpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeCertificateProvider}},
			keys: []dynamo.Keys{
				{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("a-sub-id")},
			},
			attributes: []map[string]types.AttributeValue{
				makeAttributeValueMap(&donordata.Provided{
					PK:     dynamo.LpaKey("123"),
					SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("abc")),
					LpaID:  "123",
					LpaUID: "M-0",
				}),
				makeAttributeValueMap(&certificateproviderdata.Provided{
					PK:       dynamo.LpaKey("123"),
					SK:       dynamo.CertificateProviderKey("abc"),
					LpaID:    "123",
					SignedAt: time.Now(),
					Tasks: certificateproviderdata.Tasks{
						ConfirmYourIdentity: task.IdentityStateCompleted,
					},
				}),
			},
			lpas: []*lpadata.Lpa{{
				LpaID: "123",
			}},
			expectedExists: false,
			actorType:      actor.TypeCertificateProvider,
		},
		"voucher should not see": {
			links: []dashboarddata.LpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeVoucher}},
			keys: []dynamo.Keys{
				{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("123"), SK: dynamo.VoucherKey("a-sub-id")},
			},
			attributes: []map[string]types.AttributeValue{
				makeAttributeValueMap(&donordata.Provided{
					PK:     dynamo.LpaKey("123"),
					SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("abc")),
					LpaID:  "123",
					LpaUID: "M-0",
				}),
				makeAttributeValueMap(&voucherdata.Provided{
					PK:       dynamo.LpaKey("123"),
					SK:       dynamo.VoucherKey("abc"),
					LpaID:    "123",
					SignedAt: time.Now(),
					Tasks: voucherdata.Tasks{
						SignTheDeclaration: task.StateCompleted,
					},
				}),
			},
			lpas: []*lpadata.Lpa{{
				LpaID: "123",
			}},
			expectedExists: false,
			actorType:      actor.TypeVoucher,
		},
		"lpas do not exist": {
			links: []dashboarddata.LpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor}},
			keys:  []dynamo.Keys{{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))}},
			attributes: []map[string]types.AttributeValue{
				makeAttributeValueMap(&donordata.Provided{
					PK:     dynamo.LpaKey("123"),
					SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("abc")),
					LpaID:  "123",
					LpaUID: "M-0",
				}),
			},
			lpas:           []*lpadata.Lpa{},
			expectedExists: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.ExpectAllBySK(context.Background(), dynamo.SubKey("a-sub-id"),
				tc.links, nil)
			dynamoClient.EXPECT().
				AllByKeys(context.Background(), tc.keys).
				Return(tc.attributes, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				ResolveList(context.Background(), mock.Anything).
				Return(tc.lpas, nil)

			dashboardStore := &Store{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}
			exists, err := dashboardStore.SubExistsForActorType(context.Background(), "a-sub-id", tc.actorType)

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedExists, exists)
		})
	}
}

func TestDashboardStoreSubExistsWhenDynamoError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(context.Background(), dynamo.SubKey("a-sub-id"),
		[]dashboarddata.LpaLink{}, expectedError)

	dashboardStore := &Store{dynamoClient: dynamoClient}
	exists, err := dashboardStore.SubExistsForActorType(context.Background(), "a-sub-id", actor.TypeDonor)

	assert.Equal(t, expectedError, err)
	assert.False(t, exists)
}

func TestLpaLinkUserSub(t *testing.T) {
	assert.Equal(t, "a-sub", dashboarddata.LpaLink{SK: dynamo.SubKey("a-sub")}.UserSub())
	assert.Equal(t, "", dashboarddata.LpaLink{}.UserSub())
}
