package accesscode

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/rate"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	expectedError = errors.New("err")
	testNow       = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)
	testNowFn     = func() time.Time { return testNow }
)

func (c *mockDynamoClient_OneByPK_Call) SetData(data any) *mockDynamoClient_OneByPK_Call {
	return c.Run(func(_ context.Context, _ dynamo.PK, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}

func (c *mockDynamoClient_OneBySK_Call) SetData(data any) {
	c.Run(func(_ context.Context, _ dynamo.SK, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}

func (c *mockDynamoClient_One_Call) SetData(data any) {
	c.Run(func(_ context.Context, _ dynamo.PK, _ dynamo.SK, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}

func TestAccessCodeStoreGet(t *testing.T) {
	hashedCode := accesscodedata.HashedFromString("123", "Jones")

	testcases := map[string]struct {
		t  actor.Type
		pk dynamo.AccessKeyType
	}{
		"attorney": {
			t:  actor.TypeAttorney,
			pk: dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
		},
		"replacement attorney": {
			t:  actor.TypeReplacementAttorney,
			pk: dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
		},
		"certificate provider": {
			t:  actor.TypeCertificateProvider,
			pk: dynamo.AccessKey(dynamo.CertificateProviderAccessKey(hashedCode.String())),
		},
		"voucher": {
			t:  actor.TypeVoucher,
			pk: dynamo.AccessKey(dynamo.VoucherAccessKey(hashedCode.String())),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
			data := accesscodedata.Link{LpaKey: "lpa-id", ExpiresAt: testNow}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				OneByPK(ctx, dynamo.AccessLimiterKey("session-id"), mock.Anything).
				Return(nil).
				SetData(accessLimiter{
					Limiter: &rate.Limiter{TokenPer: time.Minute, MaxTokens: 5},
				}).
				Once()
			dynamoClient.EXPECT().
				Put(ctx, accessLimiter{
					Limiter:   &rate.Limiter{TokenPer: time.Minute, MaxTokens: 5, Tokens: 4, TokensAt: testNow},
					ExpiresAt: testNow.Add(time.Hour),
				}).
				Return(nil)
			dynamoClient.EXPECT().
				OneByPK(ctx, tc.pk, mock.Anything).
				Return(nil).
				SetData(data).
				Once()

			accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

			result, err := accessCodeStore.Get(ctx, tc.t, hashedCode)
			assert.Nil(t, err)
			assert.Equal(t, data, result)
		})
	}
}

func TestAccessCodeStoreGetWhenExpired(t *testing.T) {
	hashedCode := accesscodedata.HashedFromString("123", "Jones")

	testcases := map[string]struct {
		t  actor.Type
		pk dynamo.AccessKeyType
	}{
		"attorney": {
			t:  actor.TypeAttorney,
			pk: dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
		},
		"replacement attorney": {
			t:  actor.TypeReplacementAttorney,
			pk: dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
		},
		"certificate provider": {
			t:  actor.TypeCertificateProvider,
			pk: dynamo.AccessKey(dynamo.CertificateProviderAccessKey(hashedCode.String())),
		},
		"voucher": {
			t:  actor.TypeVoucher,
			pk: dynamo.AccessKey(dynamo.VoucherAccessKey(hashedCode.String())),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data := accesscodedata.Link{LpaKey: "lpa-id", UpdatedAt: testNow.AddDate(-2, 0, -1)}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				OneByPK(ctx, dynamo.AccessLimiterKey(""), mock.Anything).
				Return(dynamo.NotFoundError{}).
				Once()
			dynamoClient.EXPECT().
				Create(ctx, accessLimiter{
					PK:        dynamo.AccessLimiterKey(""),
					SK:        dynamo.MetadataKey(""),
					Version:   1,
					Limiter:   &rate.Limiter{TokenPer: 5 * time.Minute, MaxTokens: 10, Tokens: 9, TokensAt: testNow},
					ExpiresAt: testNow.Add(time.Hour),
				}).
				Return(nil)
			dynamoClient.EXPECT().
				OneByPK(ctx, tc.pk, mock.Anything).
				Return(nil).
				SetData(data).
				Once()

			accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

			_, err := accessCodeStore.Get(ctx, tc.t, hashedCode)
			assert.ErrorIs(t, err, dynamo.NotFoundError{})
		})
	}
}

func TestAccessCodeStoreGetWhenLimited(t *testing.T) {
	hashedCode := accesscodedata.HashedFromString("123", "Jones")

	testcases := map[string]struct {
		t  actor.Type
		pk dynamo.AccessKeyType
	}{
		"attorney": {
			t:  actor.TypeAttorney,
			pk: dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
		},
		"replacement attorney": {
			t:  actor.TypeReplacementAttorney,
			pk: dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
		},
		"certificate provider": {
			t:  actor.TypeCertificateProvider,
			pk: dynamo.AccessKey(dynamo.CertificateProviderAccessKey(hashedCode.String())),
		},
		"voucher": {
			t:  actor.TypeVoucher,
			pk: dynamo.AccessKey(dynamo.VoucherAccessKey(hashedCode.String())),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				OneByPK(ctx, dynamo.AccessLimiterKey("session-id"), mock.Anything).
				Return(nil).
				SetData(accessLimiter{
					Limiter: &rate.Limiter{TokenPer: time.Minute, MaxTokens: 5, Tokens: -1, TokensAt: testNow.Add(-time.Minute)},
				}).
				Once()
			dynamoClient.EXPECT().
				Put(ctx, accessLimiter{
					Limiter:   &rate.Limiter{TokenPer: time.Minute, MaxTokens: 5, Tokens: 0, TokensAt: testNow},
					ExpiresAt: testNow.Add(time.Hour),
				}).
				Return(nil)

			accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

			_, err := accessCodeStore.Get(ctx, tc.t, hashedCode)
			assert.ErrorIs(t, err, dynamo.ErrTooManyRequests)
		})
	}
}

func TestAccessCodeStoreGetForBadActorType(t *testing.T) {
	ctx := context.Background()
	accessCodeStore := &Store{}

	_, err := accessCodeStore.Get(ctx, actor.TypeIndependentWitness, accesscodedata.HashedFromString("123", "Jones"))
	assert.NotNil(t, err)
}

func TestAccessCodeStoreGetWhenGetRateLimiterError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	_, err := accessCodeStore.Get(ctx, actor.TypeAttorney, accesscodedata.HashedFromString("123", "Jones"))
	assert.ErrorIs(t, err, expectedError)
}

func TestAccessCodeStoreGetWhenPutRateLimiterError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData(accessLimiter{
			Limiter: &rate.Limiter{TokenPer: time.Minute, MaxTokens: 5},
		})
	dynamoClient.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	_, err := accessCodeStore.Get(ctx, actor.TypeAttorney, accesscodedata.HashedFromString("123", "Jones"))
	assert.ErrorIs(t, err, expectedError)
}

func TestAccessCodeStoreGetWhenCreateRateLimiterError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(dynamo.NotFoundError{})
	dynamoClient.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	_, err := accessCodeStore.Get(ctx, actor.TypeAttorney, accesscodedata.HashedFromString("123", "Jones"))
	assert.ErrorIs(t, err, expectedError)
}

func TestAccessCodeStoreGetOnError(t *testing.T) {
	ctx := context.Background()
	data := accesscodedata.Link{LpaKey: "lpa-id"}
	_, hashedCode := accesscodedata.Generate("Jones")

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(dynamo.NotFoundError{}).
		Once()
	dynamoClient.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(nil)
	dynamoClient.EXPECT().
		OneByPK(ctx, dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())), mock.Anything).
		Return(expectedError).
		SetData(data).
		Once()

	accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	_, err := accessCodeStore.Get(ctx, actor.TypeAttorney, hashedCode)
	assert.Equal(t, expectedError, err)
}

func TestAccessCodeStorePut(t *testing.T) {
	_, hashedCode := accesscodedata.Generate("Jones")

	testcases := map[string]struct {
		actor actor.Type
		pk    dynamo.AccessKeyType
		sk    dynamo.AccessSortKeyType
	}{
		"attorney": {
			actor: actor.TypeAttorney,
			pk:    dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
			sk:    dynamo.AccessSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"replacement attorney": {
			actor: actor.TypeReplacementAttorney,
			pk:    dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
			sk:    dynamo.AccessSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"certificate provider": {
			actor: actor.TypeCertificateProvider,
			pk:    dynamo.AccessKey(dynamo.CertificateProviderAccessKey(hashedCode.String())),
			sk:    dynamo.AccessSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"voucher": {
			actor: actor.TypeVoucher,
			pk:    dynamo.AccessKey(dynamo.VoucherAccessKey(hashedCode.String())),
			sk:    dynamo.AccessSortKey(dynamo.VoucherAccessSortKey("lpa-id")),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			actorUID := actoruid.New()
			data := accesscodedata.Link{
				PK:        tc.pk,
				SK:        tc.sk,
				LpaKey:    "lpa-id",
				ActorUID:  actorUID,
				UpdatedAt: testNow,
				ExpiresAt: testNow.AddDate(2, 0, 0),
			}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				OneByPK(ctx, dynamo.ActorAccessKey(actorUID.String()), mock.Anything).
				Return(dynamo.NotFoundError{})
			dynamoClient.EXPECT().
				WriteTransaction(ctx, &dynamo.Transaction{
					Creates: []any{
						data,
						accesscodedata.ActorAccess{
							PK:           dynamo.ActorAccessKey(actorUID.String()),
							SK:           dynamo.MetadataKey(actorUID.String()),
							ShareKey:     tc.pk,
							ShareSortKey: tc.sk,
						},
					},
				}).
				Return(nil)

			accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

			err := accessCodeStore.Put(ctx, tc.actor, hashedCode, data)
			assert.Nil(t, err)
		})
	}
}

func TestAccessCodeStorePutWhenHasAccessCode(t *testing.T) {
	_, hashedCode := accesscodedata.Generate("Jones")

	testcases := map[string]struct {
		actor actor.Type
		pk    dynamo.AccessKeyType
		sk    dynamo.AccessSortKeyType
	}{
		"attorney": {
			actor: actor.TypeAttorney,
			pk:    dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
			sk:    dynamo.AccessSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"replacement attorney": {
			actor: actor.TypeReplacementAttorney,
			pk:    dynamo.AccessKey(dynamo.AttorneyAccessKey(hashedCode.String())),
			sk:    dynamo.AccessSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"certificate provider": {
			actor: actor.TypeCertificateProvider,
			pk:    dynamo.AccessKey(dynamo.CertificateProviderAccessKey(hashedCode.String())),
			sk:    dynamo.AccessSortKey(dynamo.MetadataKey(hashedCode.String())),
		},
		"voucher": {
			actor: actor.TypeVoucher,
			pk:    dynamo.AccessKey(dynamo.VoucherAccessKey(hashedCode.String())),
			sk:    dynamo.AccessSortKey(dynamo.VoucherAccessSortKey("lpa-id")),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			actorUID := actoruid.New()
			data := accesscodedata.Link{
				PK:        tc.pk,
				SK:        tc.sk,
				LpaKey:    "lpa-id",
				ActorUID:  actorUID,
				UpdatedAt: testNow,
				ExpiresAt: testNow.AddDate(2, 0, 0),
			}
			actorAccess := accesscodedata.ActorAccess{
				ShareKey:     dynamo.AccessKey(dynamo.DonorAccessKey("what")),
				ShareSortKey: dynamo.AccessSortKey(dynamo.MetadataKey("what")),
			}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				OneByPK(ctx, dynamo.ActorAccessKey(actorUID.String()), mock.Anything).
				Return(nil).
				SetData(actorAccess)
			dynamoClient.EXPECT().
				WriteTransaction(ctx, &dynamo.Transaction{
					Creates: []any{data},
					Puts: []any{
						accesscodedata.ActorAccess{
							PK:           dynamo.ActorAccessKey(actorUID.String()),
							SK:           dynamo.MetadataKey(actorUID.String()),
							ShareKey:     tc.pk,
							ShareSortKey: tc.sk,
						},
					},
					Deletes: []dynamo.Keys{{PK: actorAccess.ShareKey, SK: actorAccess.ShareSortKey}},
				}).
				Return(nil)

			accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

			err := accessCodeStore.Put(ctx, tc.actor, hashedCode, data)
			assert.Nil(t, err)
		})
	}
}

func TestAccessCodeStorePutForBadActorType(t *testing.T) {
	ctx := context.Background()
	accessCodeStore := &Store{}

	err := accessCodeStore.Put(ctx, actor.TypePersonToNotify, accesscodedata.HashedFromString("123", "Jones"), accesscodedata.Link{
		ActorUID: actoruid.New(),
	})
	assert.NotNil(t, err)
}

func TestAccessCodeStoreWhenOneByPKError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := accessCodeStore.Put(ctx, actor.TypeAttorney, accesscodedata.HashedFromString("123", "Jones"), accesscodedata.Link{LpaKey: "123"})
	assert.Equal(t, expectedError, err)
}

func TestNewAccessCodeStore(t *testing.T) {
	client := newMockDynamoClient(t)
	store := NewStore(client)

	assert.Equal(t, client, store.dynamoClient)
	assert.NotNil(t, store.now)
}

func TestAccessCodeStoreGetDonorAccess(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{
		OrganisationID: "org-id",
		LpaID:          "lpa-id",
	})

	data := supporterdata.LpaLink{
		InviteSentTo: "barry@example.com",
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(ctx, dynamo.LpaKey("lpa-id"), dynamo.OrganisationLinkKey("org-id"), mock.Anything).
		Return(nil).
		SetData(data)

	accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	result, err := accessCodeStore.GetDonorAccess(ctx)
	assert.Nil(t, err)
	assert.Equal(t, data, result)
}

func TestAccessCodeStoreGetDonorAccessWithSessionMissing(t *testing.T) {
	ctx := context.Background()
	accessCodeStore := &Store{}

	_, err := accessCodeStore.GetDonorAccess(ctx)
	assert.NotNil(t, err)
}

func TestAccessCodeStoreGetDonorAccessWhenDynamoErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{
		OrganisationID: "org-id",
		LpaID:          "lpa-id",
	})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	_, err := accessCodeStore.GetDonorAccess(ctx)
	assert.Equal(t, err, expectedError)
}

func TestAccessCodeStorePutDonorAccess(t *testing.T) {
	ctx := context.Background()
	hashedCode := accesscodedata.HashedFromString("123", "Jones")
	actorUID := actoruid.New()

	accessCode := accesscodedata.Link{
		PK:          dynamo.AccessKey(dynamo.DonorAccessKey(hashedCode.String())),
		SK:          dynamo.AccessSortKey(dynamo.DonorInviteKey(dynamo.OrganisationKey("org-id"), dynamo.LpaKey("lpa-id"))),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
		LpaKey:      dynamo.LpaKey("lpa-id"),
		ActorUID:    actorUID,
		UpdatedAt:   testNow,
		ExpiresAt:   testNow.AddDate(0, 3, 0),
	}

	transaction := &dynamo.Transaction{
		Creates: []any{
			accessCode,
			supporterdata.LpaLink{
				PK:           dynamo.LpaKey("lpa-id"),
				SK:           dynamo.OrganisationLinkKey("org-id"),
				InviteKey:    dynamo.DonorInviteKey(dynamo.OrganisationKey("org-id"), dynamo.LpaKey("lpa-id")),
				InviteSentTo: "john@example.com",
				InviteSentAt: testNow,
			},
			accesscodedata.ActorAccess{
				PK:           dynamo.ActorAccessKey(actorUID.String()),
				SK:           dynamo.MetadataKey(actorUID.String()),
				ShareKey:     accessCode.PK,
				ShareSortKey: accessCode.SK,
			},
		},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPK(ctx, dynamo.ActorAccessKey(actorUID.String()), mock.Anything).
		Return(dynamo.NotFoundError{})
	dynamoClient.EXPECT().
		WriteTransaction(ctx, transaction).
		Return(nil)

	accessCodeStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := accessCodeStore.PutDonorAccess(ctx, hashedCode, accesscodedata.Link{
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
		LpaKey:      dynamo.LpaKey("lpa-id"),
		ActorUID:    actorUID,
	}, "john@example.com")
	assert.Nil(t, err)
}

func TestAccessCodeStorePutDonorAccessWhenDonor(t *testing.T) {
	ctx := context.Background()

	accessCodeStore := &Store{}

	err := accessCodeStore.PutDonorAccess(ctx, accesscodedata.HashedFromString("123", "Jones"), accesscodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("org-id")), LpaKey: dynamo.LpaKey("lpa-id")}, "john@example.com")
	assert.Error(t, err)
}

func TestAccessCodeStoreDelete(t *testing.T) {
	actorUID := actoruid.New()
	pk := dynamo.AccessKey(dynamo.AttorneyAccessKey("a-pk"))
	sk := dynamo.AccessSortKey(dynamo.MetadataKey("a-sk"))

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{
				{PK: pk, SK: sk},
				{PK: dynamo.ActorAccessKey(actorUID.String()), SK: dynamo.MetadataKey(actorUID.String())},
			},
		}).
		Return(nil)

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.Delete(ctx, accesscodedata.Link{LpaKey: "123", PK: pk, SK: sk, ActorUID: actorUID})
	assert.Nil(t, err)
}

func TestAccessCodeStoreDeleteWhenError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.Delete(ctx, accesscodedata.Link{})
	assert.Equal(t, expectedError, err)
}

func TestAccessCodeStoreDeleteDonorAccess(t *testing.T) {
	supporterLink := supporterdata.LpaLink{
		PK:        dynamo.LpaKey("lpa-id"),
		SK:        dynamo.OrganisationLinkKey("org-id"),
		InviteKey: dynamo.DonorInviteKey(dynamo.OrganisationKey("org-id"), dynamo.LpaKey("lpa-id")),
	}

	accessCode := accesscodedata.Link{
		PK:       dynamo.AccessKey(dynamo.DonorAccessKey("code")),
		SK:       dynamo.AccessSortKey(supporterLink.InviteKey),
		ActorUID: actoruid.New(),
	}

	transaction := &dynamo.Transaction{
		Deletes: []dynamo.Keys{
			{PK: supporterLink.PK, SK: supporterLink.SK},
			{PK: accessCode.PK, SK: accessCode.SK},
			{PK: dynamo.ActorAccessKey(accessCode.ActorUID.String()), SK: dynamo.MetadataKey(accessCode.ActorUID.String())},
		},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneBySK(ctx, supporterLink.InviteKey, mock.Anything).
		Return(nil).
		SetData(accessCode)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, transaction).
		Return(nil)

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.DeleteDonorAccess(ctx, supporterLink)
	assert.Nil(t, err)
}

func TestAccessCodeStoreDeleteDonorAccessWhenAccessCodeUsedOrExpired(t *testing.T) {
	supporterLink := supporterdata.LpaLink{
		PK:        dynamo.LpaKey("lpa-id"),
		SK:        dynamo.OrganisationLinkKey("org-id"),
		InviteKey: dynamo.DonorInviteKey(dynamo.OrganisationKey("org-id"), dynamo.LpaKey("lpa-id")),
	}

	transaction := &dynamo.Transaction{
		Deletes: []dynamo.Keys{
			{PK: supporterLink.PK, SK: supporterLink.SK},
		},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneBySK(ctx, supporterLink.InviteKey, mock.Anything).
		Return(dynamo.NotFoundError{})
	dynamoClient.EXPECT().
		WriteTransaction(ctx, transaction).
		Return(nil)

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.DeleteDonorAccess(ctx, supporterLink)
	assert.Nil(t, err)
}

func TestAccessCodeStoreDeleteDonorAccessWhenErrorGettingLink(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneBySK(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.DeleteDonorAccess(ctx, supporterdata.LpaLink{})
	assert.Equal(t, expectedError, err)
}

func TestAccessCodeStoreDeleteDonorAccessWhenTransactionError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneBySK(mock.Anything, mock.Anything, mock.Anything).
		Return(dynamo.NotFoundError{})
	dynamoClient.EXPECT().
		WriteTransaction(ctx, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.DeleteDonorAccess(ctx, supporterdata.LpaLink{})
	assert.Equal(t, expectedError, err)
}

func TestAccessCodeStoreDeleteByActor(t *testing.T) {
	actorUID := actoruid.New()
	actorAccess := accesscodedata.ActorAccess{
		PK:           dynamo.ActorAccessKey(actorUID.String()),
		SK:           dynamo.MetadataKey(actorUID.String()),
		ShareKey:     dynamo.AccessKey(dynamo.CertificateProviderAccessKey("blah")),
		ShareSortKey: dynamo.AccessSortKey(dynamo.MetadataKey("blah")),
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(ctx, actorAccess.PK, actorAccess.SK, mock.Anything).
		Return(nil).
		SetData(actorAccess)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{
				{PK: actorAccess.ShareKey, SK: actorAccess.ShareSortKey},
				{PK: actorAccess.PK, SK: actorAccess.SK},
			},
		}).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.DeleteByActor(ctx, actorUID)
	assert.Equal(t, expectedError, err)
}

func TestAccessCodeStoreDeleteByActorWhenNotFound(t *testing.T) {
	actorUID := actoruid.New()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(dynamo.NotFoundError{})

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.DeleteByActor(ctx, actorUID)
	assert.Nil(t, err)
}

func TestAccessCodeStoreDeleteByActorWhenError(t *testing.T) {
	actorUID := actoruid.New()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	accessCodeStore := &Store{dynamoClient: dynamoClient}

	err := accessCodeStore.DeleteByActor(ctx, actorUID)
	assert.ErrorIs(t, err, expectedError)
}

func TestAccessKey(t *testing.T) {
	testcases := map[actor.Type]dynamo.PK{
		actor.TypeDonor:                       dynamo.AccessKey(dynamo.DonorAccessKey("955e0e614eac028c6a648f09308987e34e8e7d7ca9ecb2ba42694d8c3bf6a419")),
		actor.TypeAttorney:                    dynamo.AccessKey(dynamo.AttorneyAccessKey("955e0e614eac028c6a648f09308987e34e8e7d7ca9ecb2ba42694d8c3bf6a419")),
		actor.TypeReplacementAttorney:         dynamo.AccessKey(dynamo.AttorneyAccessKey("955e0e614eac028c6a648f09308987e34e8e7d7ca9ecb2ba42694d8c3bf6a419")),
		actor.TypeTrustCorporation:            dynamo.AccessKey(dynamo.AttorneyAccessKey("955e0e614eac028c6a648f09308987e34e8e7d7ca9ecb2ba42694d8c3bf6a419")),
		actor.TypeReplacementTrustCorporation: dynamo.AccessKey(dynamo.AttorneyAccessKey("955e0e614eac028c6a648f09308987e34e8e7d7ca9ecb2ba42694d8c3bf6a419")),
		actor.TypeCertificateProvider:         dynamo.AccessKey(dynamo.CertificateProviderAccessKey("955e0e614eac028c6a648f09308987e34e8e7d7ca9ecb2ba42694d8c3bf6a419")),
		actor.TypeVoucher:                     dynamo.AccessKey(dynamo.VoucherAccessKey("955e0e614eac028c6a648f09308987e34e8e7d7ca9ecb2ba42694d8c3bf6a419")),
	}

	for actorType, prefix := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			pk, err := accessCodeKey(actorType, accesscodedata.HashedFromString("S", "Jones"))
			assert.Nil(t, err)
			assert.Equal(t, prefix, pk)
		})
	}
}

func TestAccessKeyWhenUnknownType(t *testing.T) {
	_, err := accessCodeKey(actor.TypeAuthorisedSignatory, accesscodedata.HashedFromString("S", "Jones"))
	assert.NotNil(t, err)
}
