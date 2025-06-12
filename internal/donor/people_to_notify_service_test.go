package donor

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPeopleToNotifyServiceReusable(t *testing.T) {
	provided := &donordata.Provided{LpaID: "lpa-id"}
	peopleToNotify := []donordata.PersonToNotify{{UID: actoruid.New()}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PeopleToNotify(ctx, provided).
		Return(peopleToNotify, nil)

	service := &PeopleToNotifyService{reuseStore: reuseStore}
	result, err := service.Reusable(ctx, provided)

	assert.Nil(t, err)
	assert.Equal(t, peopleToNotify, result)
}

func TestPeopleToNotifyServiceReusableWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PeopleToNotify(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	service := &PeopleToNotifyService{reuseStore: reuseStore}
	_, err := service.Reusable(nil, nil)

	assert.ErrorIs(t, err, expectedError)
}

func TestPeopleToNotifyServiceReusableWhenNotFound(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PeopleToNotify(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	service := &PeopleToNotifyService{reuseStore: reuseStore}
	result, err := service.Reusable(nil, nil)

	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestPeopleToNotifyServiceWantPeopleToNotify(t *testing.T) {
	for yesNo, taskState := range map[form.YesNo]task.State{
		form.Yes: task.StateInProgress,
		form.No:  task.StateCompleted,
	} {
		t.Run(yesNo.String(), func(t *testing.T) {
			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, &donordata.Provided{
					LpaID:                   "lpa-id",
					DoYouWantToNotifyPeople: yesNo,
					Tasks: donordata.Tasks{
						PeopleToNotify: taskState,
					},
				}).
				Return(nil)

			service := &PeopleToNotifyService{donorStore: donorStore}
			err := service.WantPeopleToNotify(ctx, &donordata.Provided{LpaID: "lpa-id"}, yesNo)

			assert.Nil(t, err)
		})
	}
}

func TestPeopleToNotifyServiceWantPeopleToNotifyWhenDonorStoreErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &PeopleToNotifyService{donorStore: donorStore}
	err := service.WantPeopleToNotify(nil, &donordata.Provided{}, form.Yes)

	assert.ErrorIs(t, err, expectedError)
}

func TestPeopleToNotifyServicePutMany(t *testing.T) {
	calls := -1
	manyUID := []actoruid.UID{actoruid.New(), actoruid.New()}
	manyUIDFn := func() actoruid.UID {
		calls++
		return manyUID[calls]
	}

	existingPerson := donordata.PersonToNotify{UID: actoruid.New(), Address: place.Address{Line1: "1"}}
	otherPerson := donordata.PersonToNotify{FirstNames: "B", Address: place.Address{Line1: "1"}}

	testcases := map[string]struct {
		person    donordata.PersonToNotify
		taskState task.State
	}{
		"has address": {
			person:    donordata.PersonToNotify{FirstNames: "A", Address: place.Address{Line1: "1"}},
			taskState: task.StateCompleted,
		},
		"no address": {
			person:    donordata.PersonToNotify{FirstNames: "A"},
			taskState: task.StateInProgress,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			calls = -1
			personWithUID := tc.person
			personWithUID.UID = manyUID[0]

			otherPersonWithUID := otherPerson
			otherPersonWithUID.UID = manyUID[1]

			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutPeopleToNotify(ctx, []donordata.PersonToNotify{
					existingPerson,
					personWithUID,
					otherPersonWithUID,
				}).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, &donordata.Provided{
					LpaID: "lpa-id",
					PeopleToNotify: []donordata.PersonToNotify{
						existingPerson,
						personWithUID,
						otherPersonWithUID,
					},
				}).
				Return(nil)

			service := &PeopleToNotifyService{reuseStore: reuseStore, donorStore: donorStore, newUID: manyUIDFn}
			err := service.PutMany(ctx, &donordata.Provided{
				LpaID:          "lpa-id",
				PeopleToNotify: []donordata.PersonToNotify{existingPerson},
			}, []donordata.PersonToNotify{tc.person, otherPerson})

			assert.Nil(t, err)
		})
	}
}

func TestPeopleToNotifyServicePutManyWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutPeopleToNotify(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &PeopleToNotifyService{reuseStore: reuseStore}
	err := service.PutMany(ctx, &donordata.Provided{}, []donordata.PersonToNotify{})

	assert.ErrorIs(t, err, expectedError)
}

func TestPeopleToNotifyServicePutManyWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutPeopleToNotify(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &PeopleToNotifyService{reuseStore: reuseStore, donorStore: donorStore}
	err := service.PutMany(ctx, &donordata.Provided{}, []donordata.PersonToNotify{})

	assert.ErrorIs(t, err, expectedError)
}

func TestPeopleToNotifyServicePut(t *testing.T) {
	existingPerson := donordata.PersonToNotify{UID: actoruid.New(), Address: place.Address{Line1: "1"}}

	testcases := map[string]struct {
		person    donordata.PersonToNotify
		taskState task.State
	}{
		"has address": {
			person:    donordata.PersonToNotify{FirstNames: "A", Address: place.Address{Line1: "1"}},
			taskState: task.StateCompleted,
		},
		"no address": {
			person:    donordata.PersonToNotify{FirstNames: "A"},
			taskState: task.StateInProgress,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			personWithUID := tc.person
			personWithUID.UID = testUID

			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutPersonToNotify(ctx, personWithUID).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, &donordata.Provided{
					LpaID:                   "lpa-id",
					DoYouWantToNotifyPeople: form.Yes,
					PeopleToNotify: []donordata.PersonToNotify{
						existingPerson,
						personWithUID,
					},
					Tasks: donordata.Tasks{
						PeopleToNotify: tc.taskState,
					},
				}).
				Return(nil)

			service := &PeopleToNotifyService{reuseStore: reuseStore, donorStore: donorStore, newUID: testUIDFn}
			uid, err := service.Put(ctx, &donordata.Provided{
				LpaID:                   "lpa-id",
				DoYouWantToNotifyPeople: form.Yes,
				PeopleToNotify:          []donordata.PersonToNotify{existingPerson},
			}, tc.person)

			assert.Nil(t, err)
			assert.Equal(t, testUID, uid)
		})
	}
}

func TestPeopleToNotifyServicePutWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutPersonToNotify(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &PeopleToNotifyService{reuseStore: reuseStore, newUID: testUIDFn}
	_, err := service.Put(nil, &donordata.Provided{}, donordata.PersonToNotify{})

	assert.ErrorIs(t, err, expectedError)
}

func TestPeopleToNotifyServicePutWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutPersonToNotify(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &PeopleToNotifyService{reuseStore: reuseStore, donorStore: donorStore, newUID: testUIDFn}
	_, err := service.Put(nil, &donordata.Provided{}, donordata.PersonToNotify{})

	assert.ErrorIs(t, err, expectedError)
}

func TestPeopleToNotifyServiceDelete(t *testing.T) {
	existingPerson := donordata.PersonToNotify{UID: actoruid.New(), Address: place.Address{Line1: "1"}}
	personToDelete := donordata.PersonToNotify{UID: actoruid.New(), Address: place.Address{Line1: "1"}}

	testcases := map[string]struct {
		existing  []donordata.PersonToNotify
		updated   []donordata.PersonToNotify
		taskState task.State
	}{
		"last": {
			existing:  []donordata.PersonToNotify{personToDelete},
			updated:   []donordata.PersonToNotify{},
			taskState: task.StateInProgress,
		},
		"not last": {
			existing:  []donordata.PersonToNotify{existingPerson, personToDelete},
			updated:   []donordata.PersonToNotify{existingPerson},
			taskState: task.StateCompleted,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				DeletePersonToNotify(ctx, personToDelete).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, &donordata.Provided{
					LpaID:                   "lpa-id",
					DoYouWantToNotifyPeople: form.Yes,
					PeopleToNotify:          tc.updated,
					Tasks: donordata.Tasks{
						PeopleToNotify: tc.taskState,
					},
				}).
				Return(nil)

			service := &PeopleToNotifyService{reuseStore: reuseStore, donorStore: donorStore, newUID: testUIDFn}
			err := service.Delete(ctx, &donordata.Provided{
				LpaID:                   "lpa-id",
				DoYouWantToNotifyPeople: form.Yes,
				PeopleToNotify:          tc.existing,
			}, personToDelete)

			assert.Nil(t, err)
		})
	}
}

func TestPeopleToNotifyServiceDeleteWhenReuseStoreErrors(t *testing.T) {
	person := donordata.PersonToNotify{UID: actoruid.New()}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeletePersonToNotify(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &PeopleToNotifyService{reuseStore: reuseStore}
	err := service.Delete(ctx, &donordata.Provided{
		PeopleToNotify: []donordata.PersonToNotify{person},
	}, person)

	assert.ErrorIs(t, err, expectedError)
}

func TestPeopleToNotifyServiceDeleteWhenDonorStoreErrors(t *testing.T) {
	person := donordata.PersonToNotify{UID: actoruid.New()}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeletePersonToNotify(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &PeopleToNotifyService{reuseStore: reuseStore, donorStore: donorStore}
	err := service.Delete(ctx, &donordata.Provided{
		PeopleToNotify: []donordata.PersonToNotify{person},
	}, person)

	assert.ErrorIs(t, err, expectedError)
}
