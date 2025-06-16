package donor

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestCorrespondentServiceReusable(t *testing.T) {
	correspondents := []donordata.Correspondent{{}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Correspondents(ctx).
		Return(correspondents, nil)

	service := &CorrespondentService{reuseStore: reuseStore}
	result, err := service.Reusable(ctx)

	assert.Nil(t, err)
	assert.Equal(t, correspondents, result)
}

func TestCorrespondentServiceReusableWhenNotFound(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Correspondents(ctx).
		Return(nil, dynamo.NotFoundError{})

	service := &CorrespondentService{reuseStore: reuseStore}
	result, err := service.Reusable(ctx)

	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestCorrespondentServiceReusableWhenError(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Correspondents(ctx).
		Return(nil, expectedError)

	service := &CorrespondentService{reuseStore: reuseStore}
	_, err := service.Reusable(ctx)

	assert.ErrorIs(t, err, expectedError)
}

func TestCorrespondentServiceNotWanted(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaID:            "lpa-id",
			AddCorrespondent: form.No,
			Tasks:            donordata.Tasks{AddCorrespondent: task.StateCompleted},
		}).
		Return(nil)

	service := &CorrespondentService{donorStore: donorStore}
	err := service.NotWanted(ctx, &donordata.Provided{LpaID: "lpa-id"})

	assert.Nil(t, err)
}

func TestCorrespondentServiceNotWantedWhenPreviouslySet(t *testing.T) {
	correspondent := donordata.Correspondent{FirstNames: "A"}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCorrespondent(ctx, correspondent).
		Return(nil)

	eventClient := newMockCorrespondentEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(ctx, event.CorrespondentUpdated{UID: "lpa-uid"}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID:           "lpa-uid",
			AddCorrespondent: form.No,
			Tasks:            donordata.Tasks{AddCorrespondent: task.StateCompleted},
		}).
		Return(nil)

	service := &CorrespondentService{donorStore: donorStore, reuseStore: reuseStore, eventClient: eventClient}
	err := service.NotWanted(ctx, &donordata.Provided{
		LpaUID:           "lpa-uid",
		Correspondent:    correspondent,
		AddCorrespondent: form.Yes,
		Tasks:            donordata.Tasks{AddCorrespondent: task.StateInProgress},
	})

	assert.Nil(t, err)
}

func TestCorrespondentServiceNotWantedWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCorrespondent(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CorrespondentService{reuseStore: reuseStore}
	err := service.NotWanted(ctx, &donordata.Provided{
		LpaUID:           "lpa-uid",
		Correspondent:    donordata.Correspondent{FirstNames: "A"},
		AddCorrespondent: form.Yes,
		Tasks:            donordata.Tasks{AddCorrespondent: task.StateInProgress},
	})

	assert.ErrorIs(t, err, expectedError)
}

func TestCorrespondentServiceNotWantedWhenEventClientErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCorrespondent(mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockCorrespondentEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CorrespondentService{reuseStore: reuseStore, eventClient: eventClient}
	err := service.NotWanted(ctx, &donordata.Provided{
		LpaUID:           "lpa-uid",
		Correspondent:    donordata.Correspondent{FirstNames: "A"},
		AddCorrespondent: form.Yes,
		Tasks:            donordata.Tasks{AddCorrespondent: task.StateInProgress},
	})

	assert.ErrorIs(t, err, expectedError)
}

func TestCorrespondentServiceNotWantedWhenDonorStoreErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CorrespondentService{donorStore: donorStore}
	err := service.NotWanted(ctx, &donordata.Provided{LpaID: "lpa-id"})

	assert.ErrorIs(t, err, expectedError)
}

func TestCorrespondentServicePutWhenDoNotWantAddress(t *testing.T) {
	correspondent := donordata.Correspondent{
		UID:         testUID,
		FirstNames:  "A",
		WantAddress: form.No,
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCorrespondent(ctx, correspondent).
		Return(nil)

	eventClient := newMockCorrespondentEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(ctx, event.CorrespondentUpdated{
			UID:        "lpa-uid",
			ActorUID:   &testUID,
			FirstNames: "A",
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID:        "lpa-uid",
			Correspondent: correspondent,
			Tasks:         donordata.Tasks{AddCorrespondent: task.StateCompleted},
		}).
		Return(nil)

	service := &CorrespondentService{donorStore: donorStore, reuseStore: reuseStore, eventClient: eventClient, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{
		LpaUID: "lpa-uid",
		Correspondent: donordata.Correspondent{
			FirstNames:  "A",
			WantAddress: form.No,
			Address:     place.Address{Line1: "A"},
		},
	})

	assert.Nil(t, err)
}

func TestCorrespondentServicePutWhenWantAddressButNotCompleted(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID: "lpa-uid",
			Correspondent: donordata.Correspondent{
				UID:         testUID,
				FirstNames:  "A",
				WantAddress: form.Yes,
			},
			Tasks: donordata.Tasks{AddCorrespondent: task.StateInProgress},
		}).
		Return(nil)

	service := &CorrespondentService{donorStore: donorStore, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{
		LpaUID: "lpa-uid",
		Correspondent: donordata.Correspondent{
			FirstNames:  "A",
			WantAddress: form.Yes,
		},
	})

	assert.Nil(t, err)
}

func TestCorrespondentServicePutWhenWantAddressAndCompleted(t *testing.T) {
	correspondent := donordata.Correspondent{
		UID:         testUID,
		FirstNames:  "A",
		WantAddress: form.Yes,
		Address:     place.Address{Line1: "B"},
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCorrespondent(ctx, correspondent).
		Return(nil)

	eventClient := newMockCorrespondentEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(ctx, event.CorrespondentUpdated{
			UID:        "lpa-uid",
			ActorUID:   &testUID,
			FirstNames: "A",
			Address:    &place.Address{Line1: "B"},
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID:        "lpa-uid",
			Correspondent: correspondent,
			Tasks:         donordata.Tasks{AddCorrespondent: task.StateCompleted},
		}).
		Return(nil)

	service := &CorrespondentService{donorStore: donorStore, reuseStore: reuseStore, eventClient: eventClient, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{
		LpaUID: "lpa-uid",
		Correspondent: donordata.Correspondent{
			FirstNames:  "A",
			WantAddress: form.Yes,
			Address:     place.Address{Line1: "B"},
		},
	})

	assert.Nil(t, err)
}

func TestCorrespondentServicePutWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCorrespondent(mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockCorrespondentEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CorrespondentService{reuseStore: reuseStore, eventClient: eventClient, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{
		LpaUID: "lpa-uid",
		Correspondent: donordata.Correspondent{
			FirstNames:  "A",
			WantAddress: form.Yes,
			Address:     place.Address{Line1: "B"},
		},
	})

	assert.ErrorIs(t, err, expectedError)
}

func TestCorrespondentServicePutWhenEventClientErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCorrespondent(mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockCorrespondentEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CorrespondentService{reuseStore: reuseStore, eventClient: eventClient, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{
		LpaUID: "lpa-uid",
		Correspondent: donordata.Correspondent{
			FirstNames:  "A",
			WantAddress: form.Yes,
			Address:     place.Address{Line1: "B"},
		},
	})

	assert.ErrorIs(t, err, expectedError)
}

func TestCorrespondentServicePutWhenDonorStoreErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &CorrespondentService{donorStore: donorStore, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{
		LpaUID: "lpa-uid",
		Correspondent: donordata.Correspondent{
			FirstNames:  "A",
			WantAddress: form.Yes,
		},
	})

	assert.ErrorIs(t, err, expectedError)
}
