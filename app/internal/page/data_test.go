package page

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
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

var validPersonToNotify = PersonToNotify{
	ID:         "123",
	Address:    address,
	FirstNames: "Johnny",
	LastName:   "Jones",
	Email:      "user@example.org",
}

var mockRandom = func(int) string { return "123" }

var address = place.Address{
	Line1:      "a",
	Line2:      "b",
	Line3:      "c",
	TownOrCity: "d",
	Postcode:   "e",
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

func TestIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		expected bool
	}{
		"yoti": {
			lpa:      &Lpa{YotiUserData: identity.UserData{OK: true}},
			expected: true,
		},
		"one login": {
			lpa:      &Lpa{OneLoginUserData: identity.UserData{OK: true}},
			expected: true,
		},
		"none": {
			lpa:      &Lpa{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.IdentityConfirmed())
		})
	}
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

func TestReplacementAttorneysFullNames(t *testing.T) {
	l := &Lpa{
		ReplacementAttorneys: []Attorney{
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

	assert.Equal(t, "Bob Alan George Jones, Samantha Smith and Abby Helen Burns-Simpson", l.ReplacementAttorneysFullNames())
}

func TestReplacementAttorneysFirstNames(t *testing.T) {
	l := &Lpa{
		ReplacementAttorneys: []Attorney{
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

	assert.Equal(t, "Bob Alan George, Samantha and Abby Helen", l.ReplacementAttorneysFirstNames())
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
			HowReplacementAttorneysStepIn: AllCanNoLongerAct,
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
			HowReplacementAttorneysStepIn: OneCanNoLongerAct,
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
			HowReplacementAttorneysStepIn:        SomeOtherWay,
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
			HowReplacementAttorneysStepIn:        AllCanNoLongerAct,
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
			HowReplacementAttorneysStepIn:        AllCanNoLongerAct,
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
			HowReplacementAttorneysStepIn:               AllCanNoLongerAct,
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
			HowReplacementAttorneysStepIn: OneCanNoLongerAct,
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
				validAttorney,
			},
			ExpectedCompleteStatus: true,
		},
		"Multiple attorneys acting jointly": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly",
			ExpectedCompleteStatus:    true,
		},
		"Multiple attorneys acting jointly and severally": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly-and-severally",
			ExpectedCompleteStatus:    true,
		},
		"Multiple attorneys acting jointly for some and severally for others with details on how they act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions:        "mixed",
			HowAttorneysMakeDecisionsDetails: "some details",
			ExpectedCompleteStatus:           true,
		},
		"Multiple attorneys acting jointly for some and severally for others without details on how they act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions:        "mixed",
			HowAttorneysMakeDecisionsDetails: "",
			ExpectedCompleteStatus:           false,
		},
		"Multiple attorneys unexpected value for how they act": {
			Attorneys: []Attorney{
				validAttorney,
				validAttorney,
			},
			HowAttorneysMakeDecisions:        "totally-not-expected",
			HowAttorneysMakeDecisionsDetails: "",
			ExpectedCompleteStatus:           false,
		},
		"1 attorney - missing address": {
			Attorneys: []Attorney{
				{
					ID:          "123",
					FirstNames:  "Joan",
					LastName:    "Jones",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
			},
			ExpectedCompleteStatus: false,
		},
		"1 attorney - missing first name": {
			Attorneys: []Attorney{
				{
					ID:          "123",
					LastName:    "Jones",
					Address:     address,
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
			},
			ExpectedCompleteStatus: false,
		},
		"1 attorney - missing last name": {
			Attorneys: []Attorney{
				{
					ID:          "123",
					FirstNames:  "John",
					Address:     address,
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
			},
			ExpectedCompleteStatus: false,
		},
		"Multiple attorneys - missing first name": {
			Attorneys: []Attorney{
				validAttorney,
				{
					ID:          "123",
					Address:     address,
					LastName:    "Jones",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly",
			ExpectedCompleteStatus:    false,
		},
		"Multiple attorneys - missing last name": {
			Attorneys: []Attorney{
				validAttorney,
				{
					ID:          "123",
					Address:     address,
					FirstNames:  "Joan",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly",
			ExpectedCompleteStatus:    false,
		},
		"Multiple attorneys - missing address": {
			Attorneys: []Attorney{
				validAttorney,
				{
					ID:          "123",
					FirstNames:  "Joan",
					LastName:    "Jones",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
			HowAttorneysMakeDecisions: "jointly",
			ExpectedCompleteStatus:    false,
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

func TestAttorneysValid(t *testing.T) {
	lpa := Lpa{
		Attorneys: []Attorney{
			validAttorney,
			validAttorney,
			validAttorney,
		},
	}

	assert.Equal(t, true, lpa.AttorneysValid())
}

func TestAttorneysValidInvalid(t *testing.T) {
	testCases := map[string]struct {
		Attorneys []Attorney
	}{
		"Missing firstnames": {
			Attorneys: []Attorney{
				validAttorney,
				{
					FirstNames:  "",
					LastName:    "Jones",
					ID:          "123",
					Address:     address,
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
		},
		"Missing last names": {
			Attorneys: []Attorney{
				validAttorney,
				{
					FirstNames:  "Joey",
					LastName:    "",
					ID:          "123",
					Address:     address,
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
		},
		"Missing date of birth": {
			Attorneys: []Attorney{
				validAttorney,
				{
					FirstNames:  "Joey",
					LastName:    "Jones",
					ID:          "123",
					Address:     address,
					DateOfBirth: time.Time{},
				},
				validAttorney,
			},
		},
		"Addresses incomplete": {
			Attorneys: []Attorney{
				validAttorney,
				{
					FirstNames: "Joey",
					LastName:   "Jones",
					ID:         "123",
					Address: place.Address{
						Line1:      "",
						Line2:      "a",
						Line3:      "b",
						TownOrCity: "c",
						Postcode:   "d",
					},
					DateOfBirth: time.Time{},
				},
				validAttorney,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := Lpa{
				Attorneys: tc.Attorneys,
			}

			assert.Equal(t, false, lpa.AttorneysValid())
		})
	}
}

func TestReplacementAttorneysValid(t *testing.T) {
	lpa := Lpa{
		ReplacementAttorneys: []Attorney{
			validAttorney,
			validAttorney,
			validAttorney,
		},
	}

	assert.Equal(t, true, lpa.ReplacementAttorneysValid())
}

func TestReplacementAttorneysValidInvalid(t *testing.T) {
	testCases := map[string]struct {
		Attorneys []Attorney
	}{
		"Missing firstnames": {
			Attorneys: []Attorney{
				validAttorney,
				{
					FirstNames:  "",
					LastName:    "Jones",
					ID:          "123",
					Address:     address,
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
		},
		"Missing last names": {
			Attorneys: []Attorney{
				validAttorney,
				{
					FirstNames:  "Joey",
					LastName:    "",
					ID:          "123",
					Address:     address,
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
				validAttorney,
			},
		},
		"Missing date of birth": {
			Attorneys: []Attorney{
				validAttorney,
				{
					FirstNames:  "Joey",
					LastName:    "Jones",
					ID:          "123",
					Address:     address,
					DateOfBirth: time.Time{},
				},
				validAttorney,
			},
		},
		"Addresses incomplete": {
			Attorneys: []Attorney{
				validAttorney,
				{
					FirstNames: "Joey",
					LastName:   "Jones",
					ID:         "123",
					Address: place.Address{
						Line1:      "",
						Line2:      "a",
						Line3:      "b",
						TownOrCity: "c",
						Postcode:   "d",
					},
					DateOfBirth: time.Time{},
				},
				validAttorney,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := Lpa{
				ReplacementAttorneys: tc.Attorneys,
			}

			assert.Equal(t, false, lpa.ReplacementAttorneysValid())
		})
	}
}

func TestPeopleToNotifyValid(t *testing.T) {
	lpa := Lpa{
		PeopleToNotify: []PersonToNotify{
			validPersonToNotify,
			validPersonToNotify,
			validPersonToNotify,
		},
	}

	assert.Equal(t, true, lpa.PeopleToNotifyValid())
}

func TestPeopleToNotifyValidInvalid(t *testing.T) {
	testCases := map[string]struct {
		PeopleToNotify []PersonToNotify
	}{
		"Missing firstnames": {
			PeopleToNotify: []PersonToNotify{
				validPersonToNotify,
				{
					FirstNames: "",
					LastName:   "Jones",
					ID:         "123",
					Address:    address,
				},
				validPersonToNotify,
			},
		},
		"Missing last names": {
			PeopleToNotify: []PersonToNotify{
				validPersonToNotify,
				{
					FirstNames: "Joey",
					LastName:   "",
					ID:         "123",
					Address:    address,
				},
				validPersonToNotify,
			},
		},
		"Addresses incomplete": {
			PeopleToNotify: []PersonToNotify{
				validPersonToNotify,
				{
					FirstNames: "Joey",
					LastName:   "Jones",
					ID:         "123",
					Address: place.Address{
						Line1:      "",
						Line2:      "a",
						Line3:      "b",
						TownOrCity: "c",
						Postcode:   "d",
					},
				},
				validPersonToNotify,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := Lpa{
				PeopleToNotify: tc.PeopleToNotify,
			}

			assert.Equal(t, false, lpa.PeopleToNotifyValid())
		})
	}
}

func TestAttorneysActJointlyOrJointlyAndSeverally(t *testing.T) {
	testCases := map[string]struct {
		HowAct   string
		Expected bool
	}{
		"Jointly":                               {HowAct: Jointly, Expected: true},
		"Jointly and severally":                 {HowAct: JointlyAndSeverally, Expected: true},
		"Jointly for some severally for others": {HowAct: JointlyForSomeSeverallyForOthers, Expected: false},
		"Unexpected":                            {HowAct: "what", Expected: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{HowAttorneysMakeDecisions: tc.HowAct}
			assert.Equal(t, tc.Expected, lpa.AttorneysActJointlyOrJointlyAndSeverally())
		})
	}
}

func TestAttorneysActJointlyForSomeSeverallyForOthersWithDetails(t *testing.T) {
	testCases := map[string]struct {
		HowAct        string
		HowActDetails string
		Expected      bool
	}{
		"With how acting": {
			HowAct:        JointlyForSomeSeverallyForOthers,
			HowActDetails: "",
			Expected:      false,
		},
		"With details": {
			HowAct:        "",
			HowActDetails: "some details",
			Expected:      false,
		},
		"With both": {
			HowAct:        JointlyForSomeSeverallyForOthers,
			HowActDetails: "some details",
			Expected:      true,
		},
		"Mismatch status": {
			HowAct:        Jointly,
			HowActDetails: "some details",
			Expected:      false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{HowAttorneysMakeDecisions: tc.HowAct, HowAttorneysMakeDecisionsDetails: tc.HowActDetails}
			assert.Equal(t, tc.Expected, lpa.AttorneysActJointlyForSomeSeverallyForOthersWithDetails())
		})
	}
}

func TestReplacementAttorneysActJointlyOrJointlyAndSeverally(t *testing.T) {
	testCases := map[string]struct {
		HowAct   string
		Expected bool
	}{
		"Jointly":                               {HowAct: Jointly, Expected: true},
		"Jointly and severally":                 {HowAct: JointlyAndSeverally, Expected: true},
		"Jointly for some severally for others": {HowAct: JointlyForSomeSeverallyForOthers, Expected: false},
		"Unexpected":                            {HowAct: "what", Expected: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{HowReplacementAttorneysMakeDecisions: tc.HowAct}
			assert.Equal(t, tc.Expected, lpa.ReplacementAttorneysActJointlyOrJointlyAndSeverally())
		})
	}
}

func TestReplacementAttorneysActJointlyForSomeSeverallyForOthersWithDetails(t *testing.T) {
	testCases := map[string]struct {
		HowAct        string
		HowActDetails string
		Expected      bool
	}{
		"With how acting": {
			HowAct:        JointlyForSomeSeverallyForOthers,
			HowActDetails: "",
			Expected:      false,
		},
		"With details": {
			HowAct:        "",
			HowActDetails: "some details",
			Expected:      false,
		},
		"With both": {
			HowAct:        JointlyForSomeSeverallyForOthers,
			HowActDetails: "some details",
			Expected:      true,
		},
		"Mismatch status": {
			HowAct:        Jointly,
			HowActDetails: "some details",
			Expected:      false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{
				HowReplacementAttorneysMakeDecisions:        tc.HowAct,
				HowReplacementAttorneysMakeDecisionsDetails: tc.HowActDetails,
			}

			assert.Equal(t, tc.Expected, lpa.ReplacementAttorneysActJointlyForSomeSeverallyForOthersWithDetails())
		})
	}
}

func TestReplacementAttorneysStepInWhenOneOrAllAttorneysCannotAct(t *testing.T) {
	testCases := map[string]struct {
		WhenStepIn string
		Expected   bool
	}{
		"One attorney cannot act":  {WhenStepIn: OneCanNoLongerAct, Expected: true},
		"All attorneys cannot act": {WhenStepIn: AllCanNoLongerAct, Expected: true},
		"Some other way":           {WhenStepIn: SomeOtherWay, Expected: false},
		"Unexpected":               {WhenStepIn: "what", Expected: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{HowShouldReplacementAttorneysStepIn: tc.WhenStepIn}
			assert.Equal(t, tc.Expected, lpa.ReplacementAttorneysStepInWhenOneOrAllAttorneysCannotAct())
		})
	}
}

func TestReplacementAttorneysStepInSomeOtherWayWithDetails(t *testing.T) {
	testCases := map[string]struct {
		WhenStepIn        string
		WhenStepInDetails string
		Expected          bool
	}{
		"With when step in only": {
			WhenStepIn:        SomeOtherWay,
			WhenStepInDetails: "",
			Expected:          false,
		},
		"With details only": {
			WhenStepIn:        "",
			WhenStepInDetails: "some details",
			Expected:          false,
		},
		"With both": {
			WhenStepIn:        SomeOtherWay,
			WhenStepInDetails: "some details",
			Expected:          true,
		},
		"Mismatch status": {
			WhenStepIn:        OneCanNoLongerAct,
			WhenStepInDetails: "some details",
			Expected:          false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{
				HowShouldReplacementAttorneysStepIn:        tc.WhenStepIn,
				HowShouldReplacementAttorneysStepInDetails: tc.WhenStepInDetails,
			}

			assert.Equal(t, tc.Expected, lpa.ReplacementAttorneysStepInSomeOtherWayWithDetails())
		})
	}
}

func TestDonorFullName(t *testing.T) {
	l := &Lpa{
		You: Person{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"},
	}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", l.DonorFullName())
}

func TestCertificateProviderFullName(t *testing.T) {
	l := &Lpa{
		CertificateProvider: CertificateProvider{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"},
	}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", l.CertificateProviderFullName())
}

func TestLpaLegalTermTransKey(t *testing.T) {
	testCases := map[string]struct {
		LpaType           string
		ExpectedLegalTerm string
	}{
		"PFA": {
			LpaType:           LpaTypePropertyFinance,
			ExpectedLegalTerm: "pfaLegalTerm",
		},
		"HW": {
			LpaType:           LpaTypeHealthWelfare,
			ExpectedLegalTerm: "hwLegalTerm",
		},
		"Combined": {
			LpaType:           LpaTypeCombined,
			ExpectedLegalTerm: "combinedLegalTerm",
		},
		"unexpected": {
			LpaType:           "not-a-type",
			ExpectedLegalTerm: "",
		},
		"empty": {
			LpaType:           "",
			ExpectedLegalTerm: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := Lpa{Type: tc.LpaType}
			assert.Equal(t, tc.ExpectedLegalTerm, lpa.LpaLegalTermTransKey())
		})
	}
}

func TestWitnessCodeExpired(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		Duration string
		Expected bool
	}{
		"now": {
			Duration: "0s",
			Expected: false,
		},
		"29m59s ago": {
			Duration: "-29m59s",
			Expected: false,
		},
		"30m ago": {
			Duration: "-30m",
			Expected: true,
		},
		"30m01s ago": {
			Duration: "-30m01s",
			Expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			duration, _ := time.ParseDuration(tc.Duration)

			lpa := Lpa{WitnessCode: WitnessCode{
				Created: now.Add(duration),
			}}

			assert.Equal(t, tc.Expected, lpa.WitnessCode.HasExpired())
		})
	}
}

func TestAttorneysSigningDeadline(t *testing.T) {
	lpa := Lpa{
		Submitted: time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC),
	}

	expected := time.Date(2020, time.January, 30, 3, 4, 5, 6, time.UTC)
	assert.Equal(t, expected, lpa.AttorneysAndCpSigningDeadline())
}

func TestCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		url      string
		expected bool
	}{
		"empty path": {
			lpa:      &Lpa{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			lpa:      &Lpa{},
			url:      "/whatever",
			expected: true,
		},
		"about payment without task": {
			lpa:      &Lpa{},
			url:      Paths.AboutPayment,
			expected: false,
		},
		"about payment with task": {
			lpa:      &Lpa{Tasks: Tasks{CheckYourLpa: TaskCompleted}},
			url:      Paths.AboutPayment,
			expected: true,
		},
		"select your identity options without task": {
			lpa:      &Lpa{},
			url:      Paths.SelectYourIdentityOptions,
			expected: false,
		},
		"select your identity options with task": {
			lpa:      &Lpa{Tasks: Tasks{PayForLpa: TaskCompleted}},
			url:      Paths.SelectYourIdentityOptions,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.CanGoTo(tc.url))
		})
	}
}

func TestEntered(t *testing.T) {
	testCases := map[string]struct {
		Date     Date
		Expected bool
	}{
		"valid": {
			Date: Date{
				Day:   "1",
				Month: "2",
				Year:  "3",
			},
			Expected: true,
		},
		"missing day": {
			Date: Date{
				Month: "2",
				Year:  "3",
			},
			Expected: false,
		},
		"missing month": {
			Date: Date{
				Day:  "1",
				Year: "3",
			},
			Expected: false,
		},
		"missing year": {
			Date: Date{
				Day:   "1",
				Month: "2",
			},
			Expected: false,
		},
		"missing all": {
			Date:     Date{},
			Expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Date.Entered())
		})
	}

}
