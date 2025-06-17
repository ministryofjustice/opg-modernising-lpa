package donor

import (
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAttorneyServiceReusable(t *testing.T) {
	provided := &donordata.Provided{LpaUID: "lpa-uid"}
	attorneys := []donordata.Attorney{{UID: actoruid.New()}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(ctx, provided).
		Return(attorneys, nil)

	service := &AttorneyService{reuseStore: reuseStore}
	result, err := service.Reusable(ctx, provided)

	assert.Nil(t, err)
	assert.Equal(t, attorneys, result)
}

func TestAttorneyServiceReusableWhenNotFound(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	service := &AttorneyService{reuseStore: reuseStore}
	result, err := service.Reusable(ctx, &donordata.Provided{})

	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestAttorneyServiceReusableWhenError(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		Attorneys(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	service := &AttorneyService{reuseStore: reuseStore}
	_, err := service.Reusable(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServiceReusableTrustCorporations(t *testing.T) {
	trustCorporations := []donordata.TrustCorporation{{UID: actoruid.New()}}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(ctx).
		Return(trustCorporations, nil)

	service := &AttorneyService{reuseStore: reuseStore}
	result, err := service.ReusableTrustCorporations(ctx)

	assert.Nil(t, err)
	assert.Equal(t, trustCorporations, result)
}

func TestAttorneyServiceReusableTrustCorporationsWhenNotFound(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	service := &AttorneyService{reuseStore: reuseStore}
	result, err := service.ReusableTrustCorporations(ctx)

	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestAttorneyServiceReusableTrustCorporationsWhenError(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		TrustCorporations(mock.Anything).
		Return(nil, expectedError)

	service := &AttorneyService{reuseStore: reuseStore}
	_, err := service.ReusableTrustCorporations(ctx)

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServiceWantReplacements(t *testing.T) {
	testcases := map[form.YesNo]task.State{
		form.Yes: task.StateInProgress,
		form.No:  task.StateCompleted,
	}

	for yesNo, taskState := range testcases {
		t.Run(yesNo.String(), func(t *testing.T) {
			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, &donordata.Provided{
					LpaUID:                   "lpa-uid",
					WantReplacementAttorneys: yesNo,
					Tasks:                    donordata.Tasks{ChooseReplacementAttorneys: taskState},
				}).
				Return(nil)

			service := &AttorneyService{donorStore: donorStore}
			err := service.WantReplacements(ctx, &donordata.Provided{LpaUID: "lpa-uid"}, yesNo)

			assert.Nil(t, err)
		})
	}
}

func TestAttorneyServiceWantReplacementsWhenDonorStoreErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{donorStore: donorStore}
	err := service.WantReplacements(ctx, &donordata.Provided{}, form.Yes)

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServicePutMany(t *testing.T) {
	attorneyUID := actoruid.New()
	attorneys := []donordata.Attorney{{UID: attorneyUID}, {UID: testUID, FirstNames: "A"}}

	testcases := map[bool]*donordata.Provided{
		false: {
			LpaUID: "lpa-uid",
			Attorneys: donordata.Attorneys{
				Attorneys: attorneys,
			},
			Tasks: donordata.Tasks{
				ChooseAttorneys: task.StateInProgress,
			},
		},
		true: {
			LpaUID: "lpa-uid",
			ReplacementAttorneys: donordata.Attorneys{
				Attorneys: attorneys,
			},
			Tasks: donordata.Tasks{
				ChooseReplacementAttorneys: task.StateInProgress,
			},
		},
	}

	for isReplacement, provided := range testcases {
		t.Run(fmt.Sprint(isReplacement), func(t *testing.T) {
			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutAttorneys(ctx, attorneys).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, provided).
				Return(nil)

			service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore, newUID: testUIDFn, isReplacement: isReplacement}
			err := service.PutMany(ctx, &donordata.Provided{LpaUID: "lpa-uid"}, []donordata.Attorney{
				{UID: attorneyUID},
				{FirstNames: "A"},
			})

			assert.Nil(t, err)
		})
	}
}

func TestAttorneyServicePutManyWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorneys(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{reuseStore: reuseStore, newUID: testUIDFn}
	err := service.PutMany(ctx, &donordata.Provided{}, []donordata.Attorney{{}})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServicePutManyWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorneys(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore, newUID: testUIDFn}
	err := service.PutMany(ctx, &donordata.Provided{}, []donordata.Attorney{{}})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServicePut(t *testing.T) {
	attorneyUID := actoruid.New()
	attorney := donordata.Attorney{UID: attorneyUID, FirstNames: "A"}

	testcases := map[bool]*donordata.Provided{
		false: {
			LpaUID: "lpa-uid",
			Attorneys: donordata.Attorneys{
				Attorneys: []donordata.Attorney{attorney},
			},
			Tasks: donordata.Tasks{
				ChooseAttorneys: task.StateInProgress,
			},
		},
		true: {
			LpaUID: "lpa-uid",
			ReplacementAttorneys: donordata.Attorneys{
				Attorneys: []donordata.Attorney{attorney},
			},
			Tasks: donordata.Tasks{
				ChooseReplacementAttorneys: task.StateInProgress,
			},
		},
	}

	for isReplacement, provided := range testcases {
		t.Run(fmt.Sprint(isReplacement), func(t *testing.T) {
			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutAttorney(ctx, attorney).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, provided).
				Return(nil)

			service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore, newUID: testUIDFn, isReplacement: isReplacement}
			err := service.Put(ctx, &donordata.Provided{LpaUID: "lpa-uid"}, attorney)

			assert.Nil(t, err)
		})
	}
}

func TestAttorneyServicePutWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{reuseStore: reuseStore, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{LpaUID: "lpa-uid"}, donordata.Attorney{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServicePutWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore, newUID: testUIDFn}
	err := service.Put(ctx, &donordata.Provided{LpaUID: "lpa-uid"}, donordata.Attorney{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServicePutTrustCorporation(t *testing.T) {
	trustCorporationUID := actoruid.New()

	testcases := map[string]struct {
		isReplacement            bool
		trustCorporation         donordata.TrustCorporation
		reusableTrustCorporation donordata.TrustCorporation
		provided                 *donordata.Provided
	}{
		"original with uid": {
			isReplacement:            false,
			trustCorporation:         donordata.TrustCorporation{UID: trustCorporationUID, Name: "A"},
			reusableTrustCorporation: donordata.TrustCorporation{UID: trustCorporationUID, Name: "A"},
			provided: &donordata.Provided{
				LpaUID: "lpa-uid",
				Attorneys: donordata.Attorneys{
					TrustCorporation: donordata.TrustCorporation{UID: trustCorporationUID, Name: "A"},
				},
				Tasks: donordata.Tasks{
					ChooseAttorneys: task.StateInProgress,
				},
			},
		},
		"original without uid": {
			isReplacement:            false,
			trustCorporation:         donordata.TrustCorporation{Name: "A"},
			reusableTrustCorporation: donordata.TrustCorporation{UID: testUID, Name: "A"},
			provided: &donordata.Provided{
				LpaUID: "lpa-uid",
				Attorneys: donordata.Attorneys{
					TrustCorporation: donordata.TrustCorporation{UID: testUID, Name: "A"},
				},
				Tasks: donordata.Tasks{
					ChooseAttorneys: task.StateInProgress,
				},
			},
		},
		"replacement with uid": {
			isReplacement:            true,
			trustCorporation:         donordata.TrustCorporation{UID: trustCorporationUID, Name: "A"},
			reusableTrustCorporation: donordata.TrustCorporation{UID: trustCorporationUID, Name: "A"},
			provided: &donordata.Provided{
				LpaUID: "lpa-uid",
				ReplacementAttorneys: donordata.Attorneys{
					TrustCorporation: donordata.TrustCorporation{UID: trustCorporationUID, Name: "A"},
				},
				Tasks: donordata.Tasks{
					ChooseReplacementAttorneys: task.StateInProgress,
				},
			},
		},
		"replacement without uid": {
			isReplacement:            true,
			trustCorporation:         donordata.TrustCorporation{Name: "A"},
			reusableTrustCorporation: donordata.TrustCorporation{UID: testUID, Name: "A"},
			provided: &donordata.Provided{
				LpaUID: "lpa-uid",
				ReplacementAttorneys: donordata.Attorneys{
					TrustCorporation: donordata.TrustCorporation{UID: testUID, Name: "A"},
				},
				Tasks: donordata.Tasks{
					ChooseReplacementAttorneys: task.StateInProgress,
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutTrustCorporation(ctx, tc.reusableTrustCorporation).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, tc.provided).
				Return(nil)

			service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore, newUID: testUIDFn, isReplacement: tc.isReplacement}
			err := service.PutTrustCorporation(ctx, &donordata.Provided{LpaUID: "lpa-uid"}, tc.trustCorporation)

			assert.Nil(t, err)
		})
	}
}

func TestAttorneyServicePutTrustCorporationWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutTrustCorporation(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{reuseStore: reuseStore, newUID: testUIDFn}
	err := service.PutTrustCorporation(ctx, &donordata.Provided{LpaUID: "lpa-uid"}, donordata.TrustCorporation{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServicePutTrustCorporationWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutTrustCorporation(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore, newUID: testUIDFn}
	err := service.PutTrustCorporation(ctx, &donordata.Provided{LpaUID: "lpa-uid"}, donordata.TrustCorporation{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServiceDelete(t *testing.T) {
	attorney := donordata.Attorney{UID: actoruid.New()}
	otherAttorney := donordata.Attorney{UID: actoruid.New()}

	testcases := map[bool]struct {
		provided *donordata.Provided
		updated  *donordata.Provided
	}{
		false: {
			provided: &donordata.Provided{
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{attorney, otherAttorney},
				},
			},
			updated: &donordata.Provided{
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{otherAttorney},
				},
				Tasks: donordata.Tasks{ChooseAttorneys: task.StateInProgress},
			},
		},
		true: {
			provided: &donordata.Provided{
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{attorney, otherAttorney},
				},
			},
			updated: &donordata.Provided{
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{otherAttorney},
				},
				Tasks: donordata.Tasks{ChooseReplacementAttorneys: task.StateInProgress},
			},
		},
	}

	for isReplacement, tc := range testcases {
		t.Run(fmt.Sprint(isReplacement), func(t *testing.T) {
			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				DeleteAttorney(ctx, attorney).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, tc.updated).
				Return(nil)

			service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore, isReplacement: isReplacement}
			err := service.Delete(ctx, tc.provided, attorney)

			assert.Nil(t, err)
		})
	}
}

func TestAttorneyServiceDeleteWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteAttorney(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{reuseStore: reuseStore}
	err := service.Delete(ctx, &donordata.Provided{}, donordata.Attorney{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServiceDeleteWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteAttorney(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore}
	err := service.Delete(ctx, &donordata.Provided{}, donordata.Attorney{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServiceDeleteTrustCorporation(t *testing.T) {
	trustCorporation := donordata.TrustCorporation{UID: actoruid.New()}
	otherTrustCorporation := donordata.TrustCorporation{UID: actoruid.New()}

	testcases := map[bool]struct {
		provided *donordata.Provided
		updated  *donordata.Provided
	}{
		false: {
			provided: &donordata.Provided{
				Attorneys: donordata.Attorneys{
					TrustCorporation: trustCorporation,
				},
				ReplacementAttorneys: donordata.Attorneys{
					TrustCorporation: otherTrustCorporation,
				},
			},
			updated: &donordata.Provided{
				ReplacementAttorneys: donordata.Attorneys{
					TrustCorporation: otherTrustCorporation,
				},
			},
		},
		true: {
			provided: &donordata.Provided{
				Attorneys: donordata.Attorneys{
					TrustCorporation: otherTrustCorporation,
				},
				ReplacementAttorneys: donordata.Attorneys{
					TrustCorporation: trustCorporation,
				},
			},
			updated: &donordata.Provided{
				Attorneys: donordata.Attorneys{
					TrustCorporation: otherTrustCorporation,
				},
			},
		},
	}

	for isReplacement, tc := range testcases {
		t.Run(fmt.Sprint(isReplacement), func(t *testing.T) {
			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				DeleteTrustCorporation(ctx, trustCorporation).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(ctx, tc.updated).
				Return(nil)

			service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore, isReplacement: isReplacement}
			err := service.DeleteTrustCorporation(ctx, tc.provided)

			assert.Nil(t, err)
		})
	}
}

func TestAttorneyServiceDeleteTrustCorporationWhenReuseStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteTrustCorporation(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{reuseStore: reuseStore}
	err := service.DeleteTrustCorporation(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServiceDeleteTrustCorporationWhenDonorStoreErrors(t *testing.T) {
	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteTrustCorporation(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	service := &AttorneyService{donorStore: donorStore, reuseStore: reuseStore}
	err := service.DeleteTrustCorporation(ctx, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestAttorneyServiceIsReplacement(t *testing.T) {
	service := &AttorneyService{isReplacement: true}
	assert.True(t, service.IsReplacement())
}
