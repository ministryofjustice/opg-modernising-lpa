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
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifySummaryData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	err := ChoosePeopleToNotifySummary(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifySummaryWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	logger := &mockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	err := ChoosePeopleToNotifySummary(logger, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestPostChoosePeopleToNotifySummaryAddPersonToNotify(t *testing.T) {
	form := url.Values{
		"add-person-to-notify": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{PeopleToNotify: []PersonToNotify{{ID: "123"}}}, nil)

	err := ChoosePeopleToNotifySummary(nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("/lpa/lpa-id%s?addAnother=1", Paths.ChoosePeopleToNotify), resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifySummaryNoFurtherPeopleToNotify(t *testing.T) {
	form := url.Values{
		"add-person-to-notify": {"no"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{{ID: "123"}},
			Tasks:          Tasks{ChooseAttorneys: TaskCompleted, CertificateProvider: TaskCompleted},
		}, nil)

	err := ChoosePeopleToNotifySummary(nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.CheckYourLpa, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifySummaryFormValidation(t *testing.T) {
	form := url.Values{
		"add-person-to-notify": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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
