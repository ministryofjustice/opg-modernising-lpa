package page

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var validAttorney = Attorney{
	ID:          "123",
	Address:     address,
	FirstNames:  "Joan",
	LastName:    "Jones",
	DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
}

func TestReadDate(t *testing.T) {
	date := readDate(time.Date(2020, time.March, 12, 0, 0, 0, 0, time.Local))

	assert.Equal(t, Date{Day: "12", Month: "3", Year: "2020"}, date)
}

type mockDataStore struct {
	data interface{}
	mock.Mock
}

func (m *mockDataStore) Get(ctx context.Context, id string, v interface{}) error {
	data, _ := json.Marshal(m.data)
	json.Unmarshal(data, v)
	return m.Called(ctx, id).Error(0)
}

func (m *mockDataStore) Put(ctx context.Context, id string, v interface{}) error {
	return m.Called(ctx, id, v).Error(0)
}

func TestLpaStoreGet(t *testing.T) {
	ctx := context.Background()

	dataStore := &mockDataStore{}
	dataStore.On("Get", ctx, "an-id").Return(nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	lpa, err := lpaStore.Get(ctx, "an-id")
	assert.Nil(t, err)
	assert.Equal(t, &Lpa{ID: "10100000"}, lpa)
}

func TestLpaStoreGetWhenExists(t *testing.T) {
	existingLpa := &Lpa{ID: "an-id"}
	ctx := context.Background()

	dataStore := &mockDataStore{data: existingLpa}
	dataStore.On("Get", ctx, "an-id").Return(nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	lpa, err := lpaStore.Get(ctx, "an-id")
	assert.Nil(t, err)
	assert.Equal(t, existingLpa, lpa)
}

func TestLpaStoreGetWhenDataStoreError(t *testing.T) {
	ctx := context.Background()

	dataStore := &mockDataStore{}
	dataStore.On("Get", ctx, "an-id").Return(expectedError)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	_, err := lpaStore.Get(ctx, "an-id")
	assert.Equal(t, expectedError, err)
}

func TestLpaStorePut(t *testing.T) {
	ctx := context.Background()
	lpa := &Lpa{ID: "5"}

	dataStore := &mockDataStore{}
	dataStore.On("Put", ctx, "an-id", lpa).Return(expectedError)

	lpaStore := &lpaStore{dataStore: dataStore}

	err := lpaStore.Put(ctx, "an-id", lpa)
	assert.Equal(t, expectedError, err)
}

func TestGetAttorney(t *testing.T) {
	testCases := map[string]struct {
		lpa              *Lpa
		expectedAttorney Attorney
		id               string
		expectedFound    bool
	}{
		"attorney exists": {
			lpa: &Lpa{
				Attorneys: []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			},
			expectedAttorney: Attorney{ID: "1", FirstNames: "Bob"},
			id:               "1",
			expectedFound:    true,
		},
		"attorney does not exist": {
			lpa: &Lpa{
				Attorneys: []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			},
			expectedAttorney: Attorney{},
			id:               "4",
			expectedFound:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			a, found := tc.lpa.GetAttorney(tc.id)

			assert.Equal(t, tc.expectedFound, found)
			assert.Equal(t, tc.expectedAttorney, a)
		})
	}
}

func TestPutAttorney(t *testing.T) {
	testCases := map[string]struct {
		lpa             *Lpa
		expectedLpa     *Lpa
		updatedAttorney Attorney
		expectedUpdated bool
	}{
		"attorney exists": {
			lpa: &Lpa{
				Attorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			expectedLpa: &Lpa{
				Attorneys: []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			},
			updatedAttorney: Attorney{ID: "1", FirstNames: "Bob"},
			expectedUpdated: true,
		},
		"attorney does not exist": {
			lpa: &Lpa{
				Attorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			expectedLpa: &Lpa{
				Attorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			updatedAttorney: Attorney{ID: "3", FirstNames: "Bob"},
			expectedUpdated: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deleted := tc.lpa.PutAttorney(tc.updatedAttorney)

			assert.Equal(t, tc.expectedUpdated, deleted)
			assert.Equal(t, tc.expectedLpa, tc.lpa)
		})
	}
}

func TestGetReplacementAttorney(t *testing.T) {
	testCases := map[string]struct {
		lpa              *Lpa
		expectedAttorney Attorney
		id               string
		expectedFound    bool
	}{
		"attorney exists": {
			lpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			},
			expectedAttorney: Attorney{ID: "1", FirstNames: "Bob"},
			id:               "1",
			expectedFound:    true,
		},
		"attorney does not exist": {
			lpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			},
			expectedAttorney: Attorney{},
			id:               "4",
			expectedFound:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			a, found := tc.lpa.GetReplacementAttorney(tc.id)

			assert.Equal(t, tc.expectedFound, found)
			assert.Equal(t, tc.expectedAttorney, a)
		})
	}
}

func TestPutReplacementAttorney(t *testing.T) {
	testCases := map[string]struct {
		lpa             *Lpa
		expectedLpa     *Lpa
		updatedAttorney Attorney
		expectedUpdated bool
	}{
		"attorney exists": {
			lpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			expectedLpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			},
			updatedAttorney: Attorney{ID: "1", FirstNames: "Bob"},
			expectedUpdated: true,
		},
		"attorney does not exist": {
			lpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			expectedLpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			updatedAttorney: Attorney{ID: "3", FirstNames: "Bob"},
			expectedUpdated: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deleted := tc.lpa.PutReplacementAttorney(tc.updatedAttorney)

			assert.Equal(t, tc.expectedUpdated, deleted)
			assert.Equal(t, tc.expectedLpa, tc.lpa)
		})
	}
}

func TestDeleteAttorney(t *testing.T) {
	testCases := map[string]struct {
		lpa              *Lpa
		expectedLpa      *Lpa
		attorneyToDelete Attorney
		expectedDeleted  bool
	}{
		"attorney exists": {
			lpa: &Lpa{
				Attorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			expectedLpa: &Lpa{
				Attorneys: []Attorney{{ID: "1"}},
			},
			attorneyToDelete: Attorney{ID: "2"},
			expectedDeleted:  true,
		},
		"attorney does not exist": {
			lpa: &Lpa{
				Attorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			expectedLpa: &Lpa{
				Attorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			attorneyToDelete: Attorney{ID: "3"},
			expectedDeleted:  false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deleted := tc.lpa.DeleteAttorney(tc.attorneyToDelete)

			assert.Equal(t, tc.expectedDeleted, deleted)
			assert.Equal(t, tc.expectedLpa, tc.lpa)
		})
	}
}

func TestDeleteReplacementAttorney(t *testing.T) {
	testCases := map[string]struct {
		lpa              *Lpa
		expectedLpa      *Lpa
		attorneyToDelete Attorney
		expectedDeleted  bool
	}{
		"attorney exists": {
			lpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			expectedLpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1"}},
			},
			attorneyToDelete: Attorney{ID: "2"},
			expectedDeleted:  true,
		},
		"attorney does not exist": {
			lpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			expectedLpa: &Lpa{
				ReplacementAttorneys: []Attorney{{ID: "1"}, {ID: "2"}},
			},
			attorneyToDelete: Attorney{ID: "3"},
			expectedDeleted:  false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deleted := tc.lpa.DeleteReplacementAttorney(tc.attorneyToDelete)

			assert.Equal(t, tc.expectedDeleted, deleted)
			assert.Equal(t, tc.expectedLpa, tc.lpa)
		})
	}
}

func TestAttorneysFullNames(t *testing.T) {
	l := &Lpa{
		Attorneys: []Attorney{
			{
				FirstNames: "Bob Alan George",
				LastName:   "Jones",
			},
			{
				FirstNames: "Samantha",
				LastName:   "Smith",
			},
			{
				FirstNames: "Abby Helen",
				LastName:   "Burns-Simpson",
			},
		},
	}

	assert.Equal(t, "Bob Alan George Jones, Samantha Smith and Abby Helen Burns-Simpson", l.AttorneysFullNames())
}

func TestAttorneysFirstNames(t *testing.T) {
	l := &Lpa{
		Attorneys: []Attorney{
			{
				FirstNames: "Bob Alan George",
				LastName:   "Jones",
			},
			{
				FirstNames: "Samantha",
				LastName:   "Smith",
			},
			{
				FirstNames: "Abby Helen",
				LastName:   "Burns-Simpson",
			},
		},
	}

	assert.Equal(t, "Bob Alan George, Samantha and Abby Helen", l.AttorneysFirstNames())
}

func TestConcatSentence(t *testing.T) {
	assert.Equal(t, "Bob Smith, Alice Jones, John Doe and Paul Compton", concatSentence([]string{"Bob Smith", "Alice Jones", "John Doe", "Paul Compton"}))
	assert.Equal(t, "Bob Smith, Alice Jones and John Doe", concatSentence([]string{"Bob Smith", "Alice Jones", "John Doe"}))
	assert.Equal(t, "Bob Smith and John Doe", concatSentence([]string{"Bob Smith", "John Doe"}))
	assert.Equal(t, "Bob Smith", concatSentence([]string{"Bob Smith"}))
}

func TestReplacementAttorneysTaskComplete(t *testing.T) {
	testCases := map[string]struct {
		Attorneys                                   []Attorney
		ReplacementAttorneys                        []Attorney
		WantReplacementAttorneys                    string
		HowAttorneysMakeDecisions                   string
		HowAttorneysMakeDecisionsDetails            string
		HowReplacementAttorneysMakeDecisions        string
		HowReplacementAttorneysMakeDecisionsDetails string
		HowReplacementAttorneysStepIn               string
		HowReplacementAttorneysStepInDetails        string
		ExpectedComplete                            bool
	}{
		"replacement attorneys not required": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys:     []Attorney{},
			WantReplacementAttorneys: "no",
			ExpectedComplete:         true,
		},
		"single attorney and single replacement attorney": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys: []Attorney{
				validAttorney,
			},
			WantReplacementAttorneys: "yes",
			ExpectedComplete:         true,
		},
		"single attorney and multiple replacement attorney acting jointly": {
			Attorneys: []Attorney{validAttorney},
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly",
			ExpectedComplete:                     true,
		},
		"single attorney and multiple replacement attorney acting jointly and severally": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly-and-severally",
			ExpectedComplete:                     true,
		},
		"single attorney and multiple replacement attorneys acting mixed with details": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:                    "yes",
			HowReplacementAttorneysMakeDecisions:        "mixed",
			HowReplacementAttorneysMakeDecisionsDetails: "some details",
			ExpectedComplete:                            true,
		},
		"multiple attorneys acting jointly and severally and single replacement attorney steps in when there are no attorneys left to act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ReplacementAttorneys: []Attorney{
				validAttorney,
			},
			WantReplacementAttorneys:      "yes",
			HowReplacementAttorneysStepIn: "none",
			ExpectedComplete:              true,
		},
		"multiple attorneys acting jointly and severally and single replacement attorney steps in when one attorney can no longer act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ReplacementAttorneys: []Attorney{
				validAttorney,
			},
			WantReplacementAttorneys:      "yes",
			HowReplacementAttorneysStepIn: "one",
			ExpectedComplete:              true,
		},
		"multiple attorneys acting jointly and severally and single replacement attorney steps in in some other way with details": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ReplacementAttorneys: []Attorney{
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysStepIn:        "other",
			HowReplacementAttorneysStepInDetails: "some details",
			ExpectedComplete:                     true,
		},
		"multiple attorneys acting jointly and severally and multiple replacement attorneys acting jointly steps in when there are no attorneys left to act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly",
			HowReplacementAttorneysStepIn:        "none",
			ExpectedComplete:                     true,
		},
		"multiple attorneys acting jointly and severally and multiple replacement attorney acting jointly and severally steps in when there are no attorneys left to act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly-and-severally",
			HowReplacementAttorneysStepIn:        "none",
			ExpectedComplete:                     true,
		},
		"multiple attorneys acting jointly and severally and multiple replacement attorney acting mixed with details steps in when there are no attorneys left to act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:                    "yes",
			HowReplacementAttorneysMakeDecisions:        "mixed",
			HowReplacementAttorneysMakeDecisionsDetails: "some details",
			HowReplacementAttorneysStepIn:               "none",
			ExpectedComplete:                            true,
		},
		"multiple attorneys acting jointly and severally and multiple replacement attorneys steps in when one attorney cannot act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:      "yes",
			HowReplacementAttorneysStepIn: "one",
			ExpectedComplete:              true,
		},
		"multiple attorneys acting mixed with details and single replacement attorney with blank how to step in": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions:        "mixed",
			HowAttorneysMakeDecisionsDetails: "some details",
			ReplacementAttorneys: []Attorney{
				validAttorney,
			},
			WantReplacementAttorneys:      "yes",
			HowReplacementAttorneysStepIn: "",
			ExpectedComplete:              true,
		},
		"multiple attorneys acting mixed with details and multiple replacement attorney with blank how to step in": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions:        "mixed",
			HowAttorneysMakeDecisionsDetails: "some details",
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:      "yes",
			HowReplacementAttorneysStepIn: "",
			ExpectedComplete:              true,
		},
		"multiple attorneys acting jointly and multiple replacement attorneys acting jointly and blank how to step in": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly",
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly",
			HowReplacementAttorneysStepIn:        "",
			ExpectedComplete:                     true,
		},
		"multiple attorneys acting jointly and multiple replacement attorneys acting jointly and severally and blank how to step in": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly",
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly-and-severally",
			HowReplacementAttorneysStepIn:        "",
			ExpectedComplete:                     true,
		},
		"multiple attorneys acting jointly and multiple replacement attorneys acting mixed with details and blank how to step in": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly",
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:                    "yes",
			HowReplacementAttorneysMakeDecisions:        "mixed",
			HowReplacementAttorneysMakeDecisionsDetails: "some details",
			HowReplacementAttorneysStepIn:               "",
			ExpectedComplete:                            true,
		},
		"multiple attorneys acting jointly and single replacement attorneys and blank how to step in": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly",
			ReplacementAttorneys: []Attorney{
				validAttorney,
			},
			WantReplacementAttorneys:      "yes",
			HowReplacementAttorneysStepIn: "",
			ExpectedComplete:              true,
		},
		"replacement attorneys with missing address line 1": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys: []Attorney{
				validAttorney,
				{
					ID:          "123",
					Address:     place.Address{},
					FirstNames:  "Joan",
					LastName:    "Jones",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly",
			ExpectedComplete:                     false,
		},
		"replacement attorneys with missing first name": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys: []Attorney{
				validAttorney,
				{
					ID:          "123",
					Address:     address,
					FirstNames:  "",
					LastName:    "Jones",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly",
			ExpectedComplete:                     false,
		},
		"replacement attorneys with missing last name": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys: []Attorney{
				validAttorney,
				{
					ID:          "123",
					Address:     address,
					FirstNames:  "Joan",
					LastName:    "",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly",
			ExpectedComplete:                     false,
		},
		"replacement attorneys with missing date of birth": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys: []Attorney{
				validAttorney,
				{
					ID:          "123",
					Address:     address,
					FirstNames:  "Joan",
					LastName:    "Jones",
					DateOfBirth: time.Time{},
				},
				validAttorney,
			},
			WantReplacementAttorneys:             "yes",
			HowReplacementAttorneysMakeDecisions: "jointly",
			ExpectedComplete:                     false,
		},
		"replacement attorneys acting mixed with missing how act details": {
			Attorneys: []Attorney{
				validAttorney,
			},
			ReplacementAttorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			WantReplacementAttorneys:                    "yes",
			HowReplacementAttorneysMakeDecisions:        "mixed",
			HowReplacementAttorneysMakeDecisionsDetails: "",
			ExpectedComplete:                            false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := Lpa{
				Attorneys:                                   tc.Attorneys,
				HowAttorneysMakeDecisions:                   tc.HowAttorneysMakeDecisions,
				HowAttorneysMakeDecisionsDetails:            tc.HowAttorneysMakeDecisionsDetails,
				WantReplacementAttorneys:                    tc.WantReplacementAttorneys,
				ReplacementAttorneys:                        tc.ReplacementAttorneys,
				HowReplacementAttorneysMakeDecisions:        tc.HowReplacementAttorneysMakeDecisions,
				HowReplacementAttorneysMakeDecisionsDetails: tc.HowReplacementAttorneysMakeDecisionsDetails,
				HowShouldReplacementAttorneysStepIn:         tc.HowReplacementAttorneysStepIn,
				HowShouldReplacementAttorneysStepInDetails:  tc.HowReplacementAttorneysStepInDetails,
			}

			assert.Equal(t, tc.ExpectedComplete, lpa.ReplacementAttorneysTaskComplete())
		})
	}
}

func TestAttorneysTaskComplete(t *testing.T) {
	testCases := map[string]struct {
		Attorneys                        []Attorney
		HowAttorneysMakeDecisions        string
		HowAttorneysMakeDecisionsDetails string
		ExpectedCompleteStatus           bool
	}{
		"1 attorney": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			ExpectedCompleteStatus: true,
		},
		"Multiple attorneys acting jointly": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			HowAttorneysMakeDecisions: "jointly",
			ExpectedCompleteStatus:    true,
		},
		"Multiple attorneys acting jointly and severally": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ExpectedCompleteStatus:    true,
		},
		"Multiple attorneys acting jointly for some and severally for others with details on how they act": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			HowAttorneysMakeDecisions:        "mixed",
			HowAttorneysMakeDecisionsDetails: "some details",
			ExpectedCompleteStatus:           true,
		},
		"Multiple attorneys acting jointly for some and severally for others without details on how they act": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			HowAttorneysMakeDecisions:        "mixed",
			HowAttorneysMakeDecisionsDetails: "",
			ExpectedCompleteStatus:           false,
		},
		"Multiple attorneys unexpected value for how they act": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			HowAttorneysMakeDecisions:        "totally-not-expected",
			HowAttorneysMakeDecisionsDetails: "",
			ExpectedCompleteStatus:           false,
		},
		"1 attorney - missing address": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			ExpectedCompleteStatus: false,
		},
		"1 attorney - missing last name": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					FirstNames: "",
					LastName:   "Jones",
				},
			},
			ExpectedCompleteStatus: false,
		},
		"1 attorney - missing first name": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					FirstNames: "",
					LastName:   "Jones",
				},
			},
			ExpectedCompleteStatus: false,
		},
		"Multiple attorneys - missing first name": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			HowAttorneysMakeDecisions:        "jointly",
			HowAttorneysMakeDecisionsDetails: "",
			ExpectedCompleteStatus:           false,
		},
		"Multiple attorneys - missing last name": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			HowAttorneysMakeDecisions:        "jointly",
			HowAttorneysMakeDecisionsDetails: "",
			ExpectedCompleteStatus:           false,
		},
		"Multiple attorneys - missing address": {
			Attorneys: []Attorney{
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					FirstNames: "Joan",
					LastName:   "Jones",
				},
				{
					ID:         "123",
					Address:    address,
					FirstNames: "Joan",
					LastName:   "Jones",
				},
			},
			HowAttorneysMakeDecisions:        "jointly",
			HowAttorneysMakeDecisionsDetails: "",
			ExpectedCompleteStatus:           false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := Lpa{
				Attorneys:                        tc.Attorneys,
				HowAttorneysMakeDecisions:        tc.HowAttorneysMakeDecisions,
				HowAttorneysMakeDecisionsDetails: tc.HowAttorneysMakeDecisionsDetails,
			}

			assert.Equal(t, tc.ExpectedCompleteStatus, lpa.AttorneysTaskComplete())
		})
	}
}

func TestAllAttorneysAddressesComplete(t *testing.T) {
	testCases := map[string]struct {
		Attorneys              []Attorney
		ExpectedCompleteStatus bool
	}{
		"Addresses complete": {
			Attorneys: []Attorney{
				{
					Address: address,
				},
				{
					Address: address,
				},
				{
					Address: address,
				},
			},
			ExpectedCompleteStatus: true,
		},
		"Addresses incomplete": {
			Attorneys: []Attorney{
				{
					Address: address,
				},
				{
					ID: "123",
				},
				{
					Address: address,
				},
			},
			ExpectedCompleteStatus: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedCompleteStatus, allAttorneysAddressesComplete(tc.Attorneys))
		})
	}
}

func TestAllAttorneysNamesComplete(t *testing.T) {
	testCases := map[string]struct {
		Attorneys              []Attorney
		ExpectedCompleteStatus bool
	}{
		"All names complete": {
			Attorneys: []Attorney{
				{
					FirstNames: "Joey",
					LastName:   "Jones",
				},
				{
					FirstNames: "Joey",
					LastName:   "Jones",
				},
				{
					FirstNames: "Joey",
					LastName:   "Jones",
				},
			},
			ExpectedCompleteStatus: true,
		},
		"Missing firstnames": {
			Attorneys: []Attorney{
				{
					FirstNames: "Joey",
					LastName:   "Jones",
				},
				{
					FirstNames: "",
					LastName:   "Jones",
				},
				{
					FirstNames: "Joey",
					LastName:   "Jones",
				},
			},
			ExpectedCompleteStatus: false,
		},
		"Missing last names": {
			Attorneys: []Attorney{
				{
					FirstNames: "Joey",
					LastName:   "Jones",
				},
				{
					FirstNames: "Joey",
					LastName:   "",
				},
				{
					FirstNames: "Joey",
					LastName:   "Jones",
				},
			},
			ExpectedCompleteStatus: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedCompleteStatus, allAttorneysNamesComplete(tc.Attorneys))
		})
	}
}

func TestAllAttorneysDateOfBirthComplete(t *testing.T) {
	testCases := map[string]struct {
		Attorneys              []Attorney
		ExpectedCompleteStatus bool
	}{
		"All date of birth complete": {
			Attorneys: []Attorney{
				{
					DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			ExpectedCompleteStatus: true,
		},
		"Missing date of birth": {
			Attorneys: []Attorney{
				{
					DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					DateOfBirth: time.Time{},
				},
				{
					DateOfBirth: time.Date(1990, time.January, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			ExpectedCompleteStatus: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedCompleteStatus, allAttorneysDateOfBirthComplete(tc.Attorneys))
		})
	}
}
