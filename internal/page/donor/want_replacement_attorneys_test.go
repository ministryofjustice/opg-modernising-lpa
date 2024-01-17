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

func TestGetWantReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &wantReplacementAttorneysData{
			App:     testAppData,
			Donor:   &actor.DonorProvidedDetails{},
			Form:    form.NewYesNoForm(form.YesNoUnknown),
			Options: form.YesNoValues,
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWantReplacementAttorneysWithExistingReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)

	err := WantReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{FirstNames: "this"}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetWantReplacementAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &wantReplacementAttorneysData{
			App:     testAppData,
			Donor:   &actor.DonorProvidedDetails{WantReplacementAttorneys: form.Yes},
			Form:    form.NewYesNoForm(form.Yes),
			Options: form.YesNoValues,
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{WantReplacementAttorneys: form.Yes})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWantReplacementAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := WantReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWantReplacementAttorneys(t *testing.T) {
	testCases := map[string]struct {
		yesNo                        form.YesNo
		existingReplacementAttorneys actor.Attorneys
		expectedReplacementAttorneys actor.Attorneys
		taskState                    actor.TaskState
		redirect                     page.LpaPath
	}{
		"yes": {
			yesNo:                        form.Yes,
			existingReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}},
			expectedReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}},
			taskState:                    actor.TaskInProgress,
			redirect:                     page.Paths.ChooseReplacementAttorneys,
		},
		"no": {
			yesNo: form.No,
			existingReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
				{ID: "123"},
				{ID: "345"},
			}},
			expectedReplacementAttorneys: actor.Attorneys{},
			taskState:                    actor.TaskCompleted,
			redirect:                     page.Paths.TaskList,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"yes-no": {tc.yesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:                    "lpa-id",
					WantReplacementAttorneys: tc.yesNo,
					ReplacementAttorneys:     tc.expectedReplacementAttorneys,
					Tasks:                    actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, ChooseReplacementAttorneys: tc.taskState},
				}).
				Return(nil)

			err := WantReplacementAttorneys(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID:                "lpa-id",
				ReplacementAttorneys: tc.existingReplacementAttorneys,
				Tasks:                actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostWantReplacementAttorneysWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"yes-no": {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := WantReplacementAttorneys(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostWantReplacementAttorneysWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *wantReplacementAttorneysData) bool {
			return assert.Equal(t, validation.With("yes-no", validation.SelectError{Label: "yesToAddReplacementAttorneys"}), data.Errors)
		})).
		Return(nil)

	err := WantReplacementAttorneys(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
