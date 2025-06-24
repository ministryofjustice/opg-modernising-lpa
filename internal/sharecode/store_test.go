package sharecode

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	expectedError = errors.New("err")
	testNow       = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)
	testNowFn     = func() time.Time { return testNow }
)

func (m *mockDynamoClient) ExpectOneByPK(ctx, pk, data interface{}, err error) {
	m.
		On("OneByPK", ctx, pk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		}).
		Once()
}

func (m *mockDynamoClient) ExpectOneBySK(ctx, sk, data interface{}, err error) {
	m.
		On("OneBySK", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func TestShareCodeStoreGet(t *testing.T) {
	hashedCode := sharecodedata.HashedFromString("123")

	testcases := map[string]struct {
		t  actor.Type
		pk dynamo.ShareKeyType
	}{
		"attorney": {
			t:  actor.TypeAttorney,
			pk: dynamo.ShareKey(dynamo.AttorneyShareKey(hashedCode.String())),
		},
		"replacement attorney": {
			t:  actor.TypeReplacementAttorney,
			pk: dynamo.ShareKey(dynamo.AttorneyShareKey(hashedCode.String())),
		},
		"certificate provider": {
			t:  actor.TypeCertificateProvider,
			pk: dynamo.ShareKey(dynamo.CertificateProviderShareKey(hashedCode.String())),
		},
		"voucher": {
			t:  actor.TypeVoucher,
			pk: dynamo.ShareKey(dynamo.VoucherShareKey(hashedCode.String())),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data := sharecodedata.Link{LpaKey: "lpa-id"}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneByPK(ctx, tc.pk,
					data, nil)

			shareCodeStore := &Store{dynamoClient: dynamoClient}

			result, err := shareCodeStore.Get(ctx, tc.t, hashedCode)
			assert.Nil(t, err)
			assert.Equal(t, data, result)
		})
	}
}

func TestShareCodeStoreGetWhenLinked(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPK(ctx, mock.Anything,
			sharecodedata.Link{LpaLinkedAt: time.Now()}, nil)

	shareCodeStore := &Store{dynamoClient: dynamoClient}

	result, err := shareCodeStore.Get(ctx, actor.TypeDonor, sharecodedata.HashedFromString("123"))
	assert.Equal(t, dynamo.NotFoundError{}, err)
	assert.Equal(t, sharecodedata.Link{}, result)
}

func TestShareCodeStoreGetForBadActorType(t *testing.T) {
	ctx := context.Background()
	shareCodeStore := &Store{}

	_, err := shareCodeStore.Get(ctx, actor.TypeIndependentWitness, sharecodedata.HashedFromString("123"))
	assert.NotNil(t, err)
}

func TestShareCodeStoreGetOnError(t *testing.T) {
	ctx := context.Background()
	data := sharecodedata.Link{LpaKey: "lpa-id"}
	_, hashedCode := sharecodedata.Generate()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPK(ctx, dynamo.ShareKey(dynamo.AttorneyShareKey(hashedCode.String())),
			data, expectedError)

	shareCodeStore := &Store{dynamoClient: dynamoClient}

	_, err := shareCodeStore.Get(ctx, actor.TypeAttorney, hashedCode)
	assert.Equal(t, expectedError, err)
}

func TestShareCodeStorePut(t *testing.T) {
	_, hashedCode := sharecodedata.Generate()

	testcases := map[string]struct {
		actor actor.Type
		pk    dynamo.ShareKeyType
		sk    dynamo.ShareSortKeyType
	}{
		"attorney": {
			actor: actor.TypeAttorney,
			pk:    dynamo.ShareKey(dynamo.AttorneyShareKey(hashedCode.String())),
			sk:    dynamo.ShareSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"replacement attorney": {
			actor: actor.TypeReplacementAttorney,
			pk:    dynamo.ShareKey(dynamo.AttorneyShareKey(hashedCode.String())),
			sk:    dynamo.ShareSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"certificate provider": {
			actor: actor.TypeCertificateProvider,
			pk:    dynamo.ShareKey(dynamo.CertificateProviderShareKey(hashedCode.String())),
			sk:    dynamo.ShareSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"voucher": {
			actor: actor.TypeVoucher,
			pk:    dynamo.ShareKey(dynamo.VoucherShareKey(hashedCode.String())),
			sk:    dynamo.ShareSortKey(dynamo.VoucherShareSortKey("lpa-id")),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data := sharecodedata.Link{PK: tc.pk, SK: tc.sk, LpaKey: "lpa-id"}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				CreateOnly(ctx, data).
				Return(nil)

			shareCodeStore := &Store{dynamoClient: dynamoClient}

			err := shareCodeStore.Put(ctx, tc.actor, hashedCode, data)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeStorePutForBadActorType(t *testing.T) {
	ctx := context.Background()
	shareCodeStore := &Store{}

	err := shareCodeStore.Put(ctx, actor.TypePersonToNotify, sharecodedata.HashedFromString("123"), sharecodedata.Link{})
	assert.NotNil(t, err)
}

func TestShareCodeStorePutOnError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		CreateOnly(ctx, mock.Anything).
		Return(expectedError)

	shareCodeStore := &Store{dynamoClient: dynamoClient}

	err := shareCodeStore.Put(ctx, actor.TypeAttorney, sharecodedata.HashedFromString("123"), sharecodedata.Link{LpaKey: "123"})
	assert.Equal(t, expectedError, err)
}

func TestNewShareCodeStore(t *testing.T) {
	client := newMockDynamoClient(t)
	store := NewStore(client)

	assert.Equal(t, client, store.dynamoClient)
	assert.NotNil(t, store.now)
}

func TestShareCodeStoreGetDonor(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{
		OrganisationID: "org-id",
		LpaID:          "lpa-id",
	})
	data := sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id")}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.DonorInviteKey(dynamo.OrganisationKey("org-id"), dynamo.LpaKey("lpa-id")),
			data, expectedError)

	shareCodeStore := &Store{dynamoClient: dynamoClient}

	result, err := shareCodeStore.GetDonor(ctx)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, data, result)
}

func TestShareCodeStoreGetDonorWithSessionMissing(t *testing.T) {
	ctx := context.Background()
	shareCodeStore := &Store{}

	_, err := shareCodeStore.GetDonor(ctx)
	assert.NotNil(t, err)
}

func TestShareCodeStorePutDonor(t *testing.T) {
	ctx := context.Background()
	hashedCode := sharecodedata.HashedFromString("123")

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		CreateOnly(ctx, sharecodedata.Link{
			PK:          dynamo.ShareKey(dynamo.DonorShareKey(hashedCode.String())),
			SK:          dynamo.ShareSortKey(dynamo.DonorInviteKey(dynamo.OrganisationKey("org-id"), dynamo.LpaKey("lpa-id"))),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
			LpaKey:      dynamo.LpaKey("lpa-id"),
			CreatedAt:   testNow,
		}).
		Return(nil)

	shareCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := shareCodeStore.PutDonor(ctx, hashedCode, sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")), LpaKey: dynamo.LpaKey("lpa-id")})
	assert.Nil(t, err)
}

func TestShareCodeStorePutDonorWhenDonor(t *testing.T) {
	ctx := context.Background()

	shareCodeStore := &Store{}

	err := shareCodeStore.PutDonor(ctx, sharecodedata.HashedFromString("123"), sharecodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("org-id")), LpaKey: dynamo.LpaKey("lpa-id")})
	assert.Error(t, err)
}

func TestShareCodeStoreDelete(t *testing.T) {
	ctx := context.Background()
	pk := dynamo.ShareKey(dynamo.AttorneyShareKey("a-pk"))
	sk := dynamo.ShareSortKey(dynamo.MetadataKey("a-sk"))

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, pk, sk).
		Return(nil)

	shareCodeStore := &Store{dynamoClient: dynamoClient}

	err := shareCodeStore.Delete(ctx, sharecodedata.Link{LpaKey: "123", PK: pk, SK: sk})
	assert.Nil(t, err)
}

func TestShareCodeStoreDeleteOnError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, mock.Anything, mock.Anything).
		Return(expectedError)

	shareCodeStore := &Store{dynamoClient: dynamoClient}

	err := shareCodeStore.Delete(ctx, sharecodedata.Link{})
	assert.Equal(t, expectedError, err)
}

func TestShareCodeKey(t *testing.T) {
	testcases := map[actor.Type]dynamo.PK{
		actor.TypeDonor:                       dynamo.ShareKey(dynamo.DonorShareKey("8de0b3c47f112c59745f717a626932264c422a7563954872e237b223af4ad643")),
		actor.TypeAttorney:                    dynamo.ShareKey(dynamo.AttorneyShareKey("8de0b3c47f112c59745f717a626932264c422a7563954872e237b223af4ad643")),
		actor.TypeReplacementAttorney:         dynamo.ShareKey(dynamo.AttorneyShareKey("8de0b3c47f112c59745f717a626932264c422a7563954872e237b223af4ad643")),
		actor.TypeTrustCorporation:            dynamo.ShareKey(dynamo.AttorneyShareKey("8de0b3c47f112c59745f717a626932264c422a7563954872e237b223af4ad643")),
		actor.TypeReplacementTrustCorporation: dynamo.ShareKey(dynamo.AttorneyShareKey("8de0b3c47f112c59745f717a626932264c422a7563954872e237b223af4ad643")),
		actor.TypeCertificateProvider:         dynamo.ShareKey(dynamo.CertificateProviderShareKey("8de0b3c47f112c59745f717a626932264c422a7563954872e237b223af4ad643")),
		actor.TypeVoucher:                     dynamo.ShareKey(dynamo.VoucherShareKey("8de0b3c47f112c59745f717a626932264c422a7563954872e237b223af4ad643")),
	}

	for actorType, prefix := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			pk, err := shareCodeKey(actorType, sharecodedata.HashedFromString("S"))
			assert.Nil(t, err)
			assert.Equal(t, prefix, pk)
		})
	}
}

func TestShareCodeKeyWhenUnknownType(t *testing.T) {
	_, err := shareCodeKey(actor.TypeAuthorisedSignatory, sharecodedata.HashedFromString("S"))
	assert.NotNil(t, err)
}
