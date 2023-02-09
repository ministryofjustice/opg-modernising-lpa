package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDoYouWantToNotifyPeople(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &doYouWantToNotifyPeopleData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetDoYouWantToNotifyPeopleFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			DoYouWantToNotifyPeople: "yes",
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &doYouWantToNotifyPeopleData{
			App:          appData,
			WantToNotify: "yes",
			Lpa: &Lpa{
				DoYouWantToNotifyPeople: "yes",
			},
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetDoYouWantToNotifyPeopleHowAttorneysWorkTogether(t *testing.T) {
	testCases := map[string]struct {
		howWorkTogether  string
		expectedTransKey string
	}{
		"jointly": {
			howWorkTogether:  Jointly,
			expectedTransKey: "jointlyDescription",
		},
		"jointly and severally": {
			howWorkTogether:  JointlyAndSeverally,
			expectedTransKey: "jointlyAndSeverallyDescription",
		},
		"jointly for some severally for others": {
			howWorkTogether:  JointlyForSomeSeverallyForOthers,
			expectedTransKey: "jointlyForSomeSeverallyForOthersDescription",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
					DoYouWantToNotifyPeople:   "yes",
					HowAttorneysMakeDecisions: tc.howWorkTogether,
				}, nil)

			template := &mockTemplate{}
			template.
				On("Func", w, &doYouWantToNotifyPeopleData{
					App:          appData,
					WantToNotify: "yes",
					Lpa: &Lpa{
						DoYouWantToNotifyPeople:   "yes",
						HowAttorneysMakeDecisions: tc.howWorkTogether,
					},
					HowWorkTogether: tc.expectedTransKey,
				}).
				Return(nil)

			err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			mock.AssertExpectationsForObjects(t, template, lpaStore)
		})
	}

}

func TestGetDoYouWantToNotifyPeopleFromStoreWithPeople(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			PeopleToNotify: actor.PeopleToNotify{
				{ID: "123"},
			},
		}, nil)

	template := &mockTemplate{}

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))

	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetDoYouWantToNotifyPeopleWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := DoYouWantToNotifyPeople(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetDoYouWantToNotifyPeopleWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &doYouWantToNotifyPeopleData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(expectedError)

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostDoYouWantToNotifyPeople(t *testing.T) {
	testCases := []struct {
		WantToNotify     string
		ExistingAnswer   string
		ExpectedRedirect string
		ExpectedStatus   TaskState
	}{
		{
			WantToNotify:     "yes",
			ExistingAnswer:   "no",
			ExpectedRedirect: "/lpa/lpa-id" + Paths.ChoosePeopleToNotify,
			ExpectedStatus:   TaskInProgress,
		},
		{
			WantToNotify:     "no",
			ExistingAnswer:   "yes",
			ExpectedRedirect: "/lpa/lpa-id" + Paths.CheckYourLpa,
			ExpectedStatus:   TaskCompleted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.WantToNotify, func(t *testing.T) {
			form := url.Values{
				"want-to-notify": {tc.WantToNotify},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
					DoYouWantToNotifyPeople: tc.ExistingAnswer,
					Tasks: Tasks{
						YourDetails:                TaskCompleted,
						ChooseAttorneys:            TaskCompleted,
						ChooseReplacementAttorneys: TaskCompleted,
						WhenCanTheLpaBeUsed:        TaskCompleted,
						Restrictions:               TaskCompleted,
						CertificateProvider:        TaskCompleted,
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &Lpa{
					DoYouWantToNotifyPeople: tc.WantToNotify,
					Tasks: Tasks{
						YourDetails:                TaskCompleted,
						ChooseAttorneys:            TaskCompleted,
						ChooseReplacementAttorneys: TaskCompleted,
						WhenCanTheLpaBeUsed:        TaskCompleted,
						Restrictions:               TaskCompleted,
						CertificateProvider:        TaskCompleted,
						PeopleToNotify:             tc.ExpectedStatus,
					},
				}).
				Return(nil)

			err := DoYouWantToNotifyPeople(nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirect, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostDoYouWantToNotifyPeopleWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"want-to-notify": {"yes"},
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
			DoYouWantToNotifyPeople: "yes",
			Tasks:                   Tasks{PeopleToNotify: TaskInProgress},
		}).
		Return(expectedError)

	err := DoYouWantToNotifyPeople(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostDoYouWantToNotifyPeopleWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &doYouWantToNotifyPeopleData{
			App:    appData,
			Errors: validation.With("want-to-notify", validation.SelectError{Label: "yesToNotifySomeoneAboutYourLpa"}),
			Form:   &doYouWantToNotifyPeopleForm{},
			Lpa:    &Lpa{},
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadDoYouWantToNotifyPeopleForm(t *testing.T) {
	form := url.Values{
		"want-to-notify": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

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
