package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAddCorrespondent(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &addCorrespondentData{
			App:   testAppData,
			Donor: &donordata.Provided{},
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := AddCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAddCorrespondentFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &addCorrespondentData{
			App:   testAppData,
			Donor: &donordata.Provided{AddCorrespondent: form.Yes},
			Form:  form.NewYesNoForm(form.Yes),
		}).
		Return(nil)

	err := AddCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{AddCorrespondent: form.Yes})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAddCorrespondentWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := AddCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAddCorrespondent(t *testing.T) {
	testCases := map[string]struct {
		yesNo                 form.YesNo
		existingCorrespondent donordata.Correspondent
		existingTaskState     task.State
		expectedCorrespondent donordata.Correspondent
		expectedTaskState     task.State
		redirect              donor.Path
	}{
		"yes was yes": {
			yesNo:                 form.Yes,
			existingCorrespondent: donordata.Correspondent{FirstNames: "John"},
			existingTaskState:     task.StateCompleted,
			expectedCorrespondent: donordata.Correspondent{FirstNames: "John"},
			expectedTaskState:     task.StateCompleted,
			redirect:              page.Paths.EnterCorrespondentDetails,
		},
		"yes was no": {
			yesNo:             form.Yes,
			existingTaskState: task.StateCompleted,
			expectedTaskState: task.StateInProgress,
			redirect:          page.Paths.EnterCorrespondentDetails,
		},
		"yes": {
			yesNo:             form.Yes,
			expectedTaskState: task.StateInProgress,
			redirect:          page.Paths.EnterCorrespondentDetails,
		},
		"no was yes": {
			yesNo:                 form.No,
			existingCorrespondent: donordata.Correspondent{FirstNames: "John"},
			existingTaskState:     task.StateCompleted,
			expectedTaskState:     task.StateCompleted,
			redirect:              page.Paths.TaskList,
		},
		"no": {
			yesNo:             form.No,
			expectedTaskState: task.StateCompleted,
			redirect:          page.Paths.TaskList,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {tc.yesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:            "lpa-id",
					AddCorrespondent: tc.yesNo,
					Correspondent:    tc.expectedCorrespondent,
					Tasks:            donordata.Tasks{AddCorrespondent: tc.expectedTaskState},
				}).
				Return(nil)

			err := AddCorrespondent(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:         "lpa-id",
				Correspondent: tc.existingCorrespondent,
				Tasks:         donordata.Tasks{AddCorrespondent: tc.existingTaskState},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostAddCorrespondentWhenStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := AddCorrespondent(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostAddCorrespondentWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *addCorrespondentData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToAddCorrespondent"}), data.Errors)
		})).
		Return(nil)

	err := AddCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
