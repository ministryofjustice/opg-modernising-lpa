package donor

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCertificateProviderStoreReusable(t *testing.T) {
	certificateProviders := []donordata.CertificateProvider{{UID: actoruid.New()}, {UID: actoruid.New()}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		CertificateProviders(ctx).
		Return(certificateProviders, nil)

	service := &CertificateProviderService{reuseStore: reuseStore}
	result, err := service.Reusable(ctx)

	assert.Nil(t, err)
	assert.Equal(t, certificateProviders, result)
}

func TestCertificateProviderStoreReusableWhenNotFoundError(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		CertificateProviders(ctx).
		Return(nil, dynamo.NotFoundError{})

	service := &CertificateProviderService{reuseStore: reuseStore}
	result, err := service.Reusable(ctx)

	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestCertificateProviderStoreReusableWhenError(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		CertificateProviders(ctx).
		Return(nil, expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore}
	_, err := service.Reusable(ctx)

	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderServicePut(t *testing.T) {
	certificateProviderUID := actoruid.New()

	testcases := map[string]struct {
		certificateProvider donordata.CertificateProvider
		updated             donordata.CertificateProvider
		taskState           task.State
	}{
		"without uid": {
			updated:   donordata.CertificateProvider{UID: testUID},
			taskState: task.StateInProgress,
		},
		"with uid": {
			certificateProvider: donordata.CertificateProvider{UID: certificateProviderUID},
			updated:             donordata.CertificateProvider{UID: certificateProviderUID},
			taskState:           task.StateInProgress,
		},
		"completed professional": {
			certificateProvider: donordata.CertificateProvider{
				UID:          certificateProviderUID,
				Address:      place.Address{Line1: "1"},
				CarryOutBy:   lpadata.ChannelOnline,
				Relationship: lpadata.Professionally,
			},
			updated: donordata.CertificateProvider{
				UID:          certificateProviderUID,
				Address:      place.Address{Line1: "1"},
				CarryOutBy:   lpadata.ChannelOnline,
				Relationship: lpadata.Professionally,
			},
			taskState: task.StateCompleted,
		},
		"completed personally": {
			certificateProvider: donordata.CertificateProvider{
				UID:                certificateProviderUID,
				Address:            place.Address{Line1: "1"},
				CarryOutBy:         lpadata.ChannelOnline,
				Relationship:       lpadata.Personally,
				RelationshipLength: donordata.GreaterThanEqualToTwoYears,
			},
			updated: donordata.CertificateProvider{
				UID:                certificateProviderUID,
				Address:            place.Address{Line1: "1"},
				CarryOutBy:         lpadata.ChannelOnline,
				Relationship:       lpadata.Personally,
				RelationshipLength: donordata.GreaterThanEqualToTwoYears,
			},
			taskState: task.StateCompleted,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutCertificateProvider(ctx, tc.updated).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, &donordata.Provided{
					CertificateProvider: tc.updated,
					Tasks: donordata.Tasks{
						CertificateProvider: tc.taskState,
					},
				}).
				Return(nil)

			service := &CertificateProviderService{reuseStore: reuseStore, donorStore: donorStore, newUID: testUIDFn}
			err := service.Put(ctx, &donordata.Provided{CertificateProvider: tc.certificateProvider})

			assert.Nil(t, err)
		})
	}
}

func TestCertificateProviderServicePutWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCertificateProvider(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderServicePutWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCertificateProvider(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore, donorStore: donorStore, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderServiceDelete(t *testing.T) {
	certificateProvider := donordata.CertificateProvider{UID: actoruid.New()}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(ctx, certificateProvider).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(ctx, certificateProvider.UID).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(ctx, []scheduleddata.Action{scheduleddata.ActionRemindCertificateProviderToComplete}, "lpa-uid").
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{LpaID: "lpa-id", LpaUID: "lpa-uid"}).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(ctx, "lpa-uid").
		Return(&certificateproviderdata.Provided{
			SK: dynamo.CertificateProviderKey("cp-sub"),
		}, nil)
	certificateProviderStore.EXPECT().
		Delete(mock.MatchedBy(func(ctx context.Context) bool {
			session, _ := appcontext.SessionFromContext(ctx)

			return assert.Equal(t, &appcontext.Session{
				SessionID: "cp-sub",
				LpaID:     "lpa-id",
			}, session)
		})).
		Return(nil)

	service := &CertificateProviderService{
		reuseStore:               reuseStore,
		donorStore:               donorStore,
		scheduledStore:           scheduledStore,
		certificateProviderStore: certificateProviderStore,
		accessCodeStore:          accessCodeStore,
	}
	err := service.Delete(ctx, &donordata.Provided{
		LpaID:               "lpa-id",
		LpaUID:              "lpa-uid",
		CertificateProvider: certificateProvider,
		Tasks: donordata.Tasks{
			CertificateProvider: task.StateCompleted,
		},
	})

	assert.Nil(t, err)
}

func TestCertificateProviderServiceDeleteWhenNotAccessedYet(t *testing.T) {
	certificateProvider := donordata.CertificateProvider{UID: actoruid.New()}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(ctx, certificateProvider).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(ctx, certificateProvider.UID).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(ctx, []scheduleddata.Action{scheduleddata.ActionRemindCertificateProviderToComplete}, "lpa-uid").
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{LpaID: "lpa-id", LpaUID: "lpa-uid"}).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(ctx, "lpa-uid").
		Return(nil, dynamo.NotFoundError{})

	service := &CertificateProviderService{
		reuseStore:               reuseStore,
		donorStore:               donorStore,
		scheduledStore:           scheduledStore,
		certificateProviderStore: certificateProviderStore,
		accessCodeStore:          accessCodeStore,
	}
	err := service.Delete(ctx, &donordata.Provided{
		LpaID:               "lpa-id",
		LpaUID:              "lpa-uid",
		CertificateProvider: certificateProvider,
		Tasks: donordata.Tasks{
			CertificateProvider: task.StateCompleted,
		},
	})

	assert.Nil(t, err)
}

func TestCertificateProviderServiceDeleteWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore}
	err := service.Delete(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderServiceDeleteWhenAccessCodeStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(mock.Anything, mock.Anything).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore, accessCodeStore: accessCodeStore}
	err := service.Delete(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderServiceDeleteWhenScheduledStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(mock.Anything, mock.Anything).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(mock.Anything, mock.Anything).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore, accessCodeStore: accessCodeStore, scheduledStore: scheduledStore}
	err := service.Delete(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderServiceDeleteWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(mock.Anything, mock.Anything).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(mock.Anything, mock.Anything).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore, accessCodeStore: accessCodeStore, donorStore: donorStore, scheduledStore: scheduledStore}
	err := service.Delete(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderServiceDeleteWhenCertificateProviderStoreRetrievalErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(mock.Anything, mock.Anything).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(mock.Anything, mock.Anything).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore, accessCodeStore: accessCodeStore, donorStore: donorStore, scheduledStore: scheduledStore, certificateProviderStore: certificateProviderStore}
	err := service.Delete(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderServiceDeleteWhenCertificateProviderStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(mock.Anything, mock.Anything).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(mock.Anything, mock.Anything).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(&certificateproviderdata.Provided{
			SK: dynamo.CertificateProviderKey("cp-sub"),
		}, nil)
	certificateProviderStore.EXPECT().
		Delete(mock.Anything).
		Return(expectedError)

	service := &CertificateProviderService{reuseStore: reuseStore, accessCodeStore: accessCodeStore, donorStore: donorStore, scheduledStore: scheduledStore, certificateProviderStore: certificateProviderStore}
	err := service.Delete(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}
