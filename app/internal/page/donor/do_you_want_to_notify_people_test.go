package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDoYouWantToNotifyPeople(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &doYouWantToNotifyPeopleData{
			App:     testAppData,
			Lpa:     &page.Lpa{},
			Form:    &doYouWantToNotifyPeopleForm{},
			Options: actor.YesNoValues,
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDoYouWantToNotifyPeopleFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &doYouWantToNotifyPeopleData{
			App: testAppData,
			Lpa: &page.Lpa{
				DoYouWantToNotifyPeople: actor.Yes,
			},
			Form: &doYouWantToNotifyPeopleForm{
				WantToNotify: actor.Yes,
			},
			Options: actor.YesNoValues,
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &page.Lpa{
		DoYouWantToNotifyPeople: actor.Yes,
	})
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
			howWorkTogether:  actor.Jointly,
			expectedTransKey: "jointlyDescription",
		},
		"jointly and severally": {
			howWorkTogether:  actor.JointlyAndSeverally,
			expectedTransKey: "jointlyAndSeverallyDescription",
		},
		"jointly for some severally for others": {
			howWorkTogether:  actor.JointlyForSomeSeverallyForOthers,
			expectedTransKey: "jointlyForSomeSeverallyForOthersDescription",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &doYouWantToNotifyPeopleData{
					App: testAppData,
					Lpa: &page.Lpa{
						DoYouWantToNotifyPeople: actor.Yes,
						AttorneyDecisions:       actor.AttorneyDecisions{How: tc.howWorkTogether},
					},
					Form: &doYouWantToNotifyPeopleForm{
						WantToNotify: actor.Yes,
					},
					Options:         actor.YesNoValues,
					HowWorkTogether: tc.expectedTransKey,
				}).
				Return(nil)

			err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &page.Lpa{
				DoYouWantToNotifyPeople: actor.Yes,
				AttorneyDecisions:       actor.AttorneyDecisions{How: tc.howWorkTogether},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDoYouWantToNotifyPeopleFromStoreWithPeople(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &page.Lpa{
		ID: "lpa-id",
		PeopleToNotify: actor.PeopleToNotify{
			{ID: "123"},
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetDoYouWantToNotifyPeopleWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDoYouWantToNotifyPeople(t *testing.T) {
	testCases := []struct {
		WantToNotify     actor.YesNo
		ExistingAnswer   actor.YesNo
		ExpectedRedirect page.LpaPath
		ExpectedStatus   actor.TaskState
	}{
		{
			WantToNotify:     actor.Yes,
			ExistingAnswer:   actor.No,
			ExpectedRedirect: page.Paths.ChoosePeopleToNotify,
			ExpectedStatus:   actor.TaskInProgress,
		},
		{
			WantToNotify:     actor.No,
			ExistingAnswer:   actor.Yes,
			ExpectedRedirect: page.Paths.TaskList,
			ExpectedStatus:   actor.TaskCompleted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.WantToNotify.String(), func(t *testing.T) {
			form := url.Values{
				"want-to-notify": {tc.WantToNotify.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                      "lpa-id",
					DoYouWantToNotifyPeople: tc.WantToNotify,
					Tasks: page.Tasks{
						YourDetails:                actor.TaskCompleted,
						ChooseAttorneys:            actor.TaskCompleted,
						ChooseReplacementAttorneys: actor.TaskCompleted,
						WhenCanTheLpaBeUsed:        actor.TaskCompleted,
						Restrictions:               actor.TaskCompleted,
						CertificateProvider:        actor.TaskCompleted,
						PeopleToNotify:             tc.ExpectedStatus,
					},
				}).
				Return(nil)

			err := DoYouWantToNotifyPeople(nil, donorStore)(testAppData, w, r, &page.Lpa{
				ID:                      "lpa-id",
				DoYouWantToNotifyPeople: tc.ExistingAnswer,
				Tasks: page.Tasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					WhenCanTheLpaBeUsed:        actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostDoYouWantToNotifyPeopleWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"want-to-notify": {actor.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			DoYouWantToNotifyPeople: actor.Yes,
			Tasks:                   page.Tasks{PeopleToNotify: actor.TaskInProgress},
		}).
		Return(expectedError)

	err := DoYouWantToNotifyPeople(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostDoYouWantToNotifyPeopleWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *doYouWantToNotifyPeopleData) bool {
			return assert.Equal(t, validation.With("want-to-notify", validation.SelectError{Label: "yesToNotifySomeoneAboutYourLpa"}), data.Errors)
		})).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadDoYouWantToNotifyPeopleForm(t *testing.T) {
	form := url.Values{
		"want-to-notify": {actor.Yes.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readDoYouWantToNotifyPeople(r)

	assert.Equal(t, actor.Yes, result.WantToNotify)
}

func TestDoYouWantToNotifyPeopleFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *doYouWantToNotifyPeopleForm
		errors validation.List
	}{
		"valid": {
			form: &doYouWantToNotifyPeopleForm{},
		},
		"invalid": {
			form: &doYouWantToNotifyPeopleForm{
				Error: expectedError,
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
