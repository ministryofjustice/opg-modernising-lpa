package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestGetDoYouWantToNotifyPeople(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &doYouWantToNotifyPeopleData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDoYouWantToNotifyPeopleFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			DoYouWantToNotifyPeople: "yes",
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &doYouWantToNotifyPeopleData{
			App:          testAppData,
			WantToNotify: "yes",
			Lpa: &page.Lpa{
				DoYouWantToNotifyPeople: "yes",
			},
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDoYouWantToNotifyPeopleHowAttorneysWorkTogether(t *testing.T) {
	testCases := map[string]struct {
		howWorkTogether  string
		expectedTransKey string
	}{
		"jointly": {
			howWorkTogether:  page.Jointly,
			expectedTransKey: "jointlyDescription",
		},
		"jointly and severally": {
			howWorkTogether:  page.JointlyAndSeverally,
			expectedTransKey: "jointlyAndSeverallyDescription",
		},
		"jointly for some severally for others": {
			howWorkTogether:  page.JointlyForSomeSeverallyForOthers,
			expectedTransKey: "jointlyForSomeSeverallyForOthersDescription",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					DoYouWantToNotifyPeople:   "yes",
					HowAttorneysMakeDecisions: tc.howWorkTogether,
				}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &doYouWantToNotifyPeopleData{
					App:          testAppData,
					WantToNotify: "yes",
					Lpa: &page.Lpa{
						DoYouWantToNotifyPeople:   "yes",
						HowAttorneysMakeDecisions: tc.howWorkTogether,
					},
					HowWorkTogether: tc.expectedTransKey,
				}).
				Return(nil)

			err := DoYouWantToNotifyPeople(template.Execute, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDoYouWantToNotifyPeopleFromStoreWithPeople(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{
				{ID: "123"},
			},
		}, nil)

	template := newMockTemplate(t)

	err := DoYouWantToNotifyPeople(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
}

func TestGetDoYouWantToNotifyPeopleWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := DoYouWantToNotifyPeople(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDoYouWantToNotifyPeopleWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &doYouWantToNotifyPeopleData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(expectedError)

	err := DoYouWantToNotifyPeople(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDoYouWantToNotifyPeople(t *testing.T) {
	testCases := []struct {
		WantToNotify     string
		ExistingAnswer   string
		ExpectedRedirect string
		ExpectedStatus   page.TaskState
	}{
		{
			WantToNotify:     "yes",
			ExistingAnswer:   "no",
			ExpectedRedirect: "/lpa/lpa-id" + page.Paths.ChoosePeopleToNotify,
			ExpectedStatus:   page.TaskInProgress,
		},
		{
			WantToNotify:     "no",
			ExistingAnswer:   "yes",
			ExpectedRedirect: "/lpa/lpa-id" + page.Paths.CheckYourLpa,
			ExpectedStatus:   page.TaskCompleted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.WantToNotify, func(t *testing.T) {
			form := url.Values{
				"want-to-notify": {tc.WantToNotify},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					DoYouWantToNotifyPeople: tc.ExistingAnswer,
					Tasks: page.Tasks{
						YourDetails:                page.TaskCompleted,
						ChooseAttorneys:            page.TaskCompleted,
						ChooseReplacementAttorneys: page.TaskCompleted,
						WhenCanTheLpaBeUsed:        page.TaskCompleted,
						Restrictions:               page.TaskCompleted,
						CertificateProvider:        page.TaskCompleted,
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					DoYouWantToNotifyPeople: tc.WantToNotify,
					Tasks: page.Tasks{
						YourDetails:                page.TaskCompleted,
						ChooseAttorneys:            page.TaskCompleted,
						ChooseReplacementAttorneys: page.TaskCompleted,
						WhenCanTheLpaBeUsed:        page.TaskCompleted,
						Restrictions:               page.TaskCompleted,
						CertificateProvider:        page.TaskCompleted,
						PeopleToNotify:             tc.ExpectedStatus,
					},
				}).
				Return(nil)

			err := DoYouWantToNotifyPeople(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostDoYouWantToNotifyPeopleWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"want-to-notify": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			DoYouWantToNotifyPeople: "yes",
			Tasks:                   page.Tasks{PeopleToNotify: page.TaskInProgress},
		}).
		Return(expectedError)

	err := DoYouWantToNotifyPeople(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostDoYouWantToNotifyPeopleWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &doYouWantToNotifyPeopleData{
			App:    testAppData,
			Errors: validation.With("want-to-notify", validation.SelectError{Label: "yesToNotifySomeoneAboutYourLpa"}),
			Form:   &doYouWantToNotifyPeopleForm{},
			Lpa:    &page.Lpa{},
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadDoYouWantToNotifyPeopleForm(t *testing.T) {
	form := url.Values{
		"want-to-notify": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readDoYouWantToNotifyPeople(r)

	assert.Equal(t, "yes", result.WantToNotify)
}

func TestDoYouWantToNotifyPeopleFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *doYouWantToNotifyPeopleForm
		errors validation.List
	}{
		"yes": {
			form: &doYouWantToNotifyPeopleForm{
				WantToNotify: "yes",
			},
		},
		"no": {
			form: &doYouWantToNotifyPeopleForm{
				WantToNotify: "no",
			},
		},
		"missing": {
			form:   &doYouWantToNotifyPeopleForm{},
			errors: validation.With("want-to-notify", validation.SelectError{Label: "yesToNotifySomeoneAboutYourLpa"}),
		},
		"invalid": {
			form: &doYouWantToNotifyPeopleForm{
				WantToNotify: "what",
			},
			errors: validation.With("want-to-notify", validation.SelectError{Label: "yesToNotifySomeoneAboutYourLpa"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
