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

func TestGetChoosePeopleToNotifySummary(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifySummaryData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChoosePeopleToNotifySummary(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifySummaryWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	logger := &mockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChoosePeopleToNotifySummary(logger, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestPostChoosePeopleToNotifySummaryAddPersonToNotify(t *testing.T) {
	testcases := map[string]struct {
		addMoreFormValue string
		expectedUrl      string
		PeopleToNotify   []PersonToNotify
	}{
		"add person to notify": {
			addMoreFormValue: "yes",
			expectedUrl:      fmt.Sprintf("%s?addAnother=1", appData.Paths.ChoosePeopleToNotify),
			PeopleToNotify:   []PersonToNotify{},
		},
		"do not add person to notify": {
			addMoreFormValue: "no",
			expectedUrl:      appData.Paths.CheckYourLpa,
			PeopleToNotify:   []PersonToNotify{{ID: "123"}},
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{PeopleToNotify: tc.PeopleToNotify}, nil)

			form := url.Values{
				"add-person-to-notify": {tc.addMoreFormValue},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChoosePeopleToNotifySummary(nil, nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostChoosePeopleToNotifySummaryFormValidation(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	validationError := map[string]string{
		"add-person-to-notify": "selectAddMorePeopleToNotify",
	}

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *choosePeopleToNotifySummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"add-person-to-notify": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifySummary(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestChoosePeopleToNotifySummaryFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *choosePeopleToNotifySummaryForm
		errors map[string]string
	}{
		"yes": {
			form: &choosePeopleToNotifySummaryForm{
				AddPersonToNotify: "yes",
			},
			errors: map[string]string{},
		},
		"no": {
			form: &choosePeopleToNotifySummaryForm{
				AddPersonToNotify: "no",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &choosePeopleToNotifySummaryForm{},
			errors: map[string]string{
				"add-person-to-notify": "selectAddMorePeopleToNotify",
			},
		},
		"invalid": {
			form: &choosePeopleToNotifySummaryForm{
				AddPersonToNotify: "what",
			},
			errors: map[string]string{
				"add-person-to-notify": "selectAddMorePeopleToNotify",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
