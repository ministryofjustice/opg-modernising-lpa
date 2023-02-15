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

func TestGetChoosePeopleToNotifySummary(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifySummaryData{
			App:  TestAppData,
			Lpa:  &page.Lpa{},
			Form: &choosePeopleToNotifySummaryForm{},
		}).
		Return(nil)

	err := ChoosePeopleToNotifySummary(nil, template.Func, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifySummaryWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, ExpectedError)

	logger := &MockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	err := ChoosePeopleToNotifySummary(logger, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestPostChoosePeopleToNotifySummaryAddPersonToNotify(t *testing.T) {
	form := url.Values{
		"add-person-to-notify": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{{ID: "123"}}}, nil)

	err := ChoosePeopleToNotifySummary(nil, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("/lpa/lpa-id%s?addAnother=1", page.Paths.ChoosePeopleToNotify), resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifySummaryNoFurtherPeopleToNotify(t *testing.T) {
	form := url.Values{
		"add-person-to-notify": {"no"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{{ID: "123"}},
			Tasks: page.Tasks{
				YourDetails:                page.TaskCompleted,
				ChooseAttorneys:            page.TaskCompleted,
				ChooseReplacementAttorneys: page.TaskCompleted,
				WhenCanTheLpaBeUsed:        page.TaskCompleted,
				Restrictions:               page.TaskCompleted,
				CertificateProvider:        page.TaskCompleted,
				PeopleToNotify:             page.TaskCompleted,
			},
		}, nil)

	err := ChoosePeopleToNotifySummary(nil, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.CheckYourLpa, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifySummaryFormValidation(t *testing.T) {
	form := url.Values{
		"add-person-to-notify": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	validationError := validation.With("add-person-to-notify", validation.SelectError{Label: "yesToAddAnotherPersonToNotify"})

	template := &MockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *choosePeopleToNotifySummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChoosePeopleToNotifySummary(nil, template.Func, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestChoosePeopleToNotifySummaryFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *choosePeopleToNotifySummaryForm
		errors validation.List
	}{
		"yes": {
			form: &choosePeopleToNotifySummaryForm{
				AddPersonToNotify: "yes",
			},
		},
		"no": {
			form: &choosePeopleToNotifySummaryForm{
				AddPersonToNotify: "no",
			},
		},
		"missing": {
			form:   &choosePeopleToNotifySummaryForm{},
			errors: validation.With("add-person-to-notify", validation.SelectError{Label: "yesToAddAnotherPersonToNotify"}),
		},
		"invalid": {
			form: &choosePeopleToNotifySummaryForm{
				AddPersonToNotify: "what",
			},
			errors: validation.With("add-person-to-notify", validation.SelectError{Label: "yesToAddAnotherPersonToNotify"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
