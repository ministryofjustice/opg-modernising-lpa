package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotify(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyData{
			App:  appData,
			Form: &choosePeopleToNotifyForm{},
		}).
		Return(nil)

	err := ChoosePeopleToNotify(template.Func, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetChoosePeopleToNotifyWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChoosePeopleToNotifyFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{
				validPersonToNotify,
			},
		}, nil)

	template := &mockTemplate{}

	err := ChoosePeopleToNotify(template.Func, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChoosePeopleToNotifyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyData{
			App:  appData,
			Form: &choosePeopleToNotifyForm{},
		}).
		Return(expectedError)

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
			expectedUrl: "/lpa/lpa-id" + Paths.ChoosePeopleToNotifySummary,
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
			expectedUrl: "/lpa/lpa-id" + Paths.ChoosePeopleToNotifySummary,
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
					PeopleToNotify: tc.addedPeople,
				}, nil)

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
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"email":       {"johnny@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
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

	err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("/lpa/lpa-id%s?id=123", Paths.ChoosePeopleToNotifyAddress), resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyPersonExists(t *testing.T) {
	form := url.Values{
		"first-names": {"Johnny"},
		"last-name":   {"Dear"},
		"email":       {"johnny.d@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

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
		On("Get", r.Context()).
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{existingPerson},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			PeopleToNotify: []PersonToNotify{updatedPerson},
			Tasks:          Tasks{PeopleToNotify: TaskInProgress},
		}).
		Return(nil)

	err := ChoosePeopleToNotify(nil, lpaStore, mockRandom)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChoosePeopleToNotifyAddress+"?id=123", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyFromAnotherPage(t *testing.T) {
	testcases := map[string]struct {
		requestUrl      string
		expectedNextUrl string
	}{
		"with from value": {
			"/?from=/test&id=123",
			"/lpa/lpa-id/test",
		},
		"without from value": {
			"/?from=&id=123",
			"/lpa/lpa-id" + Paths.ChoosePeopleToNotifyAddress + "?id=123",
		},
		"missing from key": {
			"/?id=123",
			"/lpa/lpa-id" + Paths.ChoosePeopleToNotifyAddress + "?id=123",
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			form := url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"email":       {"johnny@example.com"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
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
				On("Put", r.Context(), &Lpa{
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
				return assert.Equal(t, validation.With("first-names", "enterTheirFirstNames"), data.Errors)
			},
		},
		"last name missing": {
			form: url.Values{
				"first-names": {"Johnny"},
				"email":       {"name@example.com"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Equal(t, validation.With("last-name", "enterTheirLastName"), data.Errors)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{}, nil)

			template := &mockTemplate{}
			template.
				On("Func", w, mock.MatchedBy(func(data *choosePeopleToNotifyData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := ChoosePeopleToNotify(template.Func, lpaStore, mockRandom)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, lpaStore, template)
		})
	}
}

func TestPostChoosePeopleToNotifyWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"email":       {"john@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

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
		errors validation.List
	}{
		"valid": {
			form: &choosePeopleToNotifyForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@example.com",
			},
		},
		"max length": {
			form: &choosePeopleToNotifyForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
				Email:      "person@example.com",
			},
		},
		"missing all": {
			form: &choosePeopleToNotifyForm{},
			errors: validation.
				With("first-names", "enterTheirFirstNames").
				With("last-name", "enterTheirLastName").
				With("email", "enterTheirEmail"),
		},
		"too long": {
			form: &choosePeopleToNotifyForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Email:      "person@example.com",
			},
			errors: validation.
				With("first-names", "firstNamesTooLong").
				With("last-name", "lastNameTooLong"),
		},
		"invalid email": {
			form: &choosePeopleToNotifyForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@",
			},
			errors: validation.With("email", "theirEmailIncorrectFormat"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
