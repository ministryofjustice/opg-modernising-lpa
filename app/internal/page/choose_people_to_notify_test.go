package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotify(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyData{
			App:  appData,
			Form: &choosePeopleToNotifyForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChoosePeopleToNotify(template.Func, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetChoosePeopleToNotifyWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChoosePeopleToNotifyFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{
				validPersonToNotify,
			},
		}, nil)

	template := &mockTemplate{}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChoosePeopleToNotify(template.Func, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChoosePeopleToNotifyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyData{
			App:  appData,
			Form: &choosePeopleToNotifyForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChoosePeopleToNotify(template.Func, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetChoosePeopleToNotifyPeopleLimitReached(t *testing.T) {
	personToNotify := PersonToNotify{
		FirstNames: "John",
		LastName:   "Doe",
		Email:      "johnny@example.com",
		ID:         "123",
	}

	testcases := map[string]struct {
		addedPeople []PersonToNotify
		expectedUrl string
	}{
		"5 people": {
			addedPeople: []PersonToNotify{
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
			},
			expectedUrl: Paths.ChoosePeopleToNotifySummary,
		},
		"6 people": {
			addedPeople: []PersonToNotify{
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
			},
			expectedUrl: Paths.ChoosePeopleToNotifySummary,
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{
					PeopleToNotify: tc.addedPeople,
				}, nil)

			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChoosePeopleToNotifyPersonDoesNotExists(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			PeopleToNotify: []PersonToNotify{
				{
					FirstNames: "John",
					LastName:   "Doe",
					Email:      "johnny@example.com",
					ID:         "123",
				},
			},
			Tasks: Tasks{PeopleToNotify: TaskInProgress},
		}).
		Return(nil)

	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"email":       {"johnny@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("%s?id=123", appData.Paths.ChoosePeopleToNotifyAddress), resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyPersonExists(t *testing.T) {
	w := httptest.NewRecorder()

	existingPerson := PersonToNotify{
		FirstNames: "John",
		LastName:   "Doe",
		Email:      "johnny@example.com",
		ID:         "123",
	}

	updatedPerson := PersonToNotify{
		FirstNames: "Johnny",
		LastName:   "Dear",
		Email:      "johnny.d@example.com",
		ID:         "123",
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{existingPerson},
		}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			PeopleToNotify: []PersonToNotify{updatedPerson},
			Tasks:          Tasks{PeopleToNotify: TaskInProgress},
		}).
		Return(nil)

	form := url.Values{
		"first-names": {"Johnny"},
		"last-name":   {"Dear"},
		"email":       {"johnny.d@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.ChoosePeopleToNotifyAddress+"?id=123", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyFromAnotherPage(t *testing.T) {
	testcases := map[string]struct {
		requestUrl      string
		expectedNextUrl string
	}{
		"with from value": {
			"/?from=/test&id=123",
			"/test",
		},
		"without from value": {
			"/?from=&id=123",
			Paths.ChoosePeopleToNotifyAddress + "?id=123",
		},
		"missing from key": {
			"/?id=123",
			Paths.ChoosePeopleToNotifyAddress + "?id=123",
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{
					PeopleToNotify: []PersonToNotify{
						{
							FirstNames: "John",
							LastName:   "Doe",
							Email:      "johnny@example.com",
							ID:         "123",
						},
					},
				}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{
					PeopleToNotify: []PersonToNotify{
						{
							FirstNames: "John",
							LastName:   "Doe",
							Email:      "johnny@example.com",
							ID:         "123",
						},
					},
					Tasks: Tasks{PeopleToNotify: TaskInProgress},
				}).
				Return(nil)

			form := url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"email":       {"johnny@example.com"},
			}

			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChoosePeopleToNotifyWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *choosePeopleToNotifyData) bool
	}{
		"first name missing": {
			form: url.Values{
				"last-name": {"Doe"},
				"email":     {"name@example.com"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Equal(t, map[string]string{"first-names": "enterTheirFirstNames"}, data.Errors)
			},
		},
		"last name missing": {
			form: url.Values{
				"first-names": {"Johnny"},
				"email":       {"name@example.com"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Equal(t, map[string]string{"last-name": "enterTheirLastName"}, data.Errors)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{}, nil)

			template := &mockTemplate{}
			template.
				On("Func", w, mock.MatchedBy(func(data *choosePeopleToNotifyData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChoosePeopleToNotify(template.Func, lpaStore, mockRandom)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, lpaStore, template)
		})
	}
}

func TestPostChoosePeopleToNotifyWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"email":       {"john@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestReadChoosePeopleToNotifyForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names": {"  John "},
		"last-name":   {"Doe"},
		"email":       {"john@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readChoosePeopleToNotifyForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("john@example.com", result.Email)
}

func TestChoosePeopleToNotifyFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *choosePeopleToNotifyForm
		errors map[string]string
	}{
		"valid": {
			form: &choosePeopleToNotifyForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@example.com",
			},
			errors: map[string]string{},
		},
		"max length": {
			form: &choosePeopleToNotifyForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
				Email:      "person@example.com",
			},
			errors: map[string]string{},
		},
		"missing all": {
			form: &choosePeopleToNotifyForm{},
			errors: map[string]string{
				"first-names": "enterTheirFirstNames",
				"last-name":   "enterTheirLastName",
				"email":       "enterTheirEmail",
			},
		},
		"too long": {
			form: &choosePeopleToNotifyForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Email:      "person@example.com",
			},
			errors: map[string]string{
				"first-names": "firstNamesTooLong",
				"last-name":   "lastNameTooLong",
			},
		},
		"invalid email": {
			form: &choosePeopleToNotifyForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@",
			},
			errors: map[string]string{
				"email": "theirEmailIncorrectFormat",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
