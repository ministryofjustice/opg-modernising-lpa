package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDoYouWantToNotifyPeople(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &doYouWantToNotifyPeopleData{
			App:   testAppData,
			Donor: &actor.DonorProvidedDetails{},
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDoYouWantToNotifyPeopleFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &doYouWantToNotifyPeopleData{
			App: testAppData,
			Donor: &actor.DonorProvidedDetails{
				DoYouWantToNotifyPeople: form.Yes,
			},
			Form: form.NewYesNoForm(form.Yes),
		}).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		DoYouWantToNotifyPeople: form.Yes,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDoYouWantToNotifyPeopleHowAttorneysWorkTogether(t *testing.T) {
	testCases := map[string]struct {
		howWorkTogether  actor.AttorneysAct
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
			template.EXPECT().
				Execute(w, &doYouWantToNotifyPeopleData{
					App: testAppData,
					Donor: &actor.DonorProvidedDetails{
						DoYouWantToNotifyPeople: form.Yes,
						AttorneyDecisions:       actor.AttorneyDecisions{How: tc.howWorkTogether},
					},
					Form:            form.NewYesNoForm(form.Yes),
					HowWorkTogether: tc.expectedTransKey,
				}).
				Return(nil)

			err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
				DoYouWantToNotifyPeople: form.Yes,
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

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
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
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDoYouWantToNotifyPeople(t *testing.T) {
	testCases := []struct {
		YesNo            form.YesNo
		ExistingAnswer   form.YesNo
		ExpectedRedirect page.LpaPath
		ExpectedStatus   actor.TaskState
	}{
		{
			YesNo:            form.Yes,
			ExistingAnswer:   form.No,
			ExpectedRedirect: page.Paths.ChoosePeopleToNotify,
			ExpectedStatus:   actor.TaskInProgress,
		},
		{
			YesNo:            form.No,
			ExistingAnswer:   form.Yes,
			ExpectedRedirect: page.Paths.TaskList,
			ExpectedStatus:   actor.TaskCompleted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.YesNo.String(), func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {tc.YesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:                   "lpa-id",
					DoYouWantToNotifyPeople: tc.YesNo,
					Tasks: actor.DonorTasks{
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

			err := DoYouWantToNotifyPeople(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID:                   "lpa-id",
				DoYouWantToNotifyPeople: tc.ExistingAnswer,
				Tasks: actor.DonorTasks{
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
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			DoYouWantToNotifyPeople: form.Yes,
			Tasks:                   actor.DonorTasks{PeopleToNotify: actor.TaskInProgress},
		}).
		Return(expectedError)

	err := DoYouWantToNotifyPeople(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostDoYouWantToNotifyPeopleWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *doYouWantToNotifyPeopleData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToNotifySomeoneAboutYourLpa"}), data.Errors)
		})).
		Return(nil)

	err := DoYouWantToNotifyPeople(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
