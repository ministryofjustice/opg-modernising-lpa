package donor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotify(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyData{
			App:  testAppData,
			Form: &choosePeopleToNotifyForm{},
		}).
		Return(nil)

	err := ChoosePeopleToNotify(template.Execute, lpaStore, mockUuidString)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := ChoosePeopleToNotify(nil, lpaStore, mockUuidString)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{
				{
					ID:         "123",
					Address:    testAddress,
					FirstNames: "Johnny",
					LastName:   "Jones",
					Email:      "user@example.org",
				},
			},
		}, nil)

	template := newMockTemplate(t)

	err := ChoosePeopleToNotify(template.Execute, lpaStore, mockUuidString)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
}

func TestGetChoosePeopleToNotifyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyData{
			App:  testAppData,
			Form: &choosePeopleToNotifyForm{},
		}).
		Return(expectedError)

	err := ChoosePeopleToNotify(template.Execute, lpaStore, mockUuidString)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyPeopleLimitReached(t *testing.T) {
	personToNotify := actor.PersonToNotify{
		FirstNames: "John",
		LastName:   "Doe",
		Email:      "johnny@example.com",
		ID:         "123",
	}

	testcases := map[string]struct {
		addedPeople actor.PeopleToNotify
		expectedUrl string
	}{
		"5 people": {
			addedPeople: actor.PeopleToNotify{
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
			},
			expectedUrl: "/lpa/lpa-id" + page.Paths.ChoosePeopleToNotifySummary,
		},
		"6 people": {
			addedPeople: actor.PeopleToNotify{
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
			},
			expectedUrl: "/lpa/lpa-id" + page.Paths.ChoosePeopleToNotifySummary,
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					PeopleToNotify: tc.addedPeople,
				}, nil)

			err := ChoosePeopleToNotify(nil, lpaStore, mockUuidString)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
		})
	}
}

func TestPostChoosePeopleToNotifyPersonDoesNotExists(t *testing.T) {
	testCases := map[string]struct {
		form           url.Values
		personToNotify actor.PersonToNotify
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"email":       {"johnny@example.com"},
			},
			personToNotify: actor.PersonToNotify{
				FirstNames: "John",
				LastName:   "Doe",
				Email:      "johnny@example.com",
				ID:         "123",
			},
		},
		"name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"johnny@example.com"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypePersonToNotify, actor.TypeDonor, "Jane", "Doe").String()},
			},
			personToNotify: actor.PersonToNotify{
				FirstNames: "Jane",
				LastName:   "Doe",
				Email:      "johnny@example.com",
				ID:         "123",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					Donor:          actor.Donor{FirstNames: "Jane", LastName: "Doe"},
					PeopleToNotify: actor.PeopleToNotify{tc.personToNotify},
					Tasks:          page.Tasks{PeopleToNotify: actor.TaskInProgress},
				}).
				Return(nil)

			err := ChoosePeopleToNotify(nil, lpaStore, mockUuidString)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, fmt.Sprintf("/lpa/lpa-id%s?id=123", page.Paths.ChoosePeopleToNotifyAddress), resp.Header.Get("Location"))
		})
	}
}

func TestPostChoosePeopleToNotifyPersonExists(t *testing.T) {
	form := url.Values{
		"first-names": {"Johnny"},
		"last-name":   {"Dear"},
		"email":       {"johnny.d@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{{
				FirstNames: "John",
				LastName:   "Doe",
				Email:      "johnny@example.com",
				ID:         "123",
			}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{{
				FirstNames: "Johnny",
				LastName:   "Dear",
				Email:      "johnny.d@example.com",
				ID:         "123",
			}},
			Tasks: page.Tasks{PeopleToNotify: actor.TaskInProgress},
		}).
		Return(nil)

	err := ChoosePeopleToNotify(nil, lpaStore, mockUuidString)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifyAddress+"?id=123", resp.Header.Get("Location"))
}

func TestPostChoosePeopleToNotifyWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *choosePeopleToNotifyData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
				"email":     {"name@example.com"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Nil(t, data.NameWarning) &&
					assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
		"name warning": {
			form: url.Values{
				"first-names": {"Jane"},
				"last-name":   {"Doe"},
				"email":       {"name@example.com"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypePersonToNotify, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
		"name warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"ignore-name-warning": {"errorDonorMatchesActor|aPersonToNotify|Jane|Doe"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypePersonToNotify, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.Equal(t, validation.With("email", validation.EnterError{Label: "email"}), data.Errors)
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"name@example.com"},
				"ignore-name-warning": {"errorDonorMatchesActor|aPersonToNotify|John|Doe"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypePersonToNotify, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
				}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, mock.MatchedBy(func(data *choosePeopleToNotifyData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := ChoosePeopleToNotify(template.Execute, lpaStore, mockUuidString)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ChoosePeopleToNotify(nil, lpaStore, mockUuidString)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestReadChoosePeopleToNotifyForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names": {"  John "},
		"last-name":   {"Doe"},
		"email":       {"john@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

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
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("email", validation.EnterError{Label: "email"}),
		},
		"too long": {
			form: &choosePeopleToNotifyForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Email:      "person@example.com",
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
		"invalid email": {
			form: &choosePeopleToNotifyForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "person@",
			},
			errors: validation.With("email", validation.EmailError{Label: "email"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestPersonToNotifyMatches(t *testing.T) {
	lpa := &page.Lpa{
		Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: actor.Attorneys{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		},
		ReplacementAttorneys: actor.Attorneys{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		},
		CertificateProvider: actor.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{ID: "123", FirstNames: "o", LastName: "p"},
		},
	}

	assert.Equal(t, actor.TypeNone, personToNotifyMatches(lpa, "123", "x", "y"))
	assert.Equal(t, actor.TypeDonor, personToNotifyMatches(lpa, "123", "a", "b"))
	assert.Equal(t, actor.TypeAttorney, personToNotifyMatches(lpa, "123", "C", "D"))
	assert.Equal(t, actor.TypeAttorney, personToNotifyMatches(lpa, "123", "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, personToNotifyMatches(lpa, "123", "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, personToNotifyMatches(lpa, "123", "i", "j"))
	assert.Equal(t, actor.TypeNone, personToNotifyMatches(lpa, "123", "k", "L"))
	assert.Equal(t, actor.TypePersonToNotify, personToNotifyMatches(lpa, "123", "m", "n"))
	assert.Equal(t, actor.TypeNone, personToNotifyMatches(lpa, "123", "o", "p"))
}

func TestPersonToNotifyMatchesEmptyNamesIgnored(t *testing.T) {
	lpa := &page.Lpa{
		Donor: actor.Donor{FirstNames: "", LastName: ""},
		Attorneys: actor.Attorneys{
			{FirstNames: "", LastName: ""},
		},
		ReplacementAttorneys: actor.Attorneys{
			{FirstNames: "", LastName: ""},
		},
		CertificateProvider: actor.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "", LastName: ""},
			{ID: "123", FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, personToNotifyMatches(lpa, "123", "", ""))
}
