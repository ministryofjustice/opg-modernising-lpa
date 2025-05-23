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

func TestGetCheckYouCanSign(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYouCanSignData{
			App:    testAppData,
			Errors: nil,
			Form:   form.NewYesNoForm(form.No),
		}).
		Return(nil)

	err := CheckYouCanSign(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{CanSign: form.No},
	})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYouCanSign(t *testing.T) {
	testcases := map[form.YesNo]struct {
		provided *donordata.Provided
		redirect donor.Path
	}{
		form.Yes: {
			provided: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{CanSign: form.Yes},
			},
			redirect: donor.PathYourPreferredLanguage,
		},
		form.No: {
			provided: &donordata.Provided{
				LpaID:               "lpa-id",
				Donor:               donordata.Donor{CanSign: form.No},
				AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "A"},
				IndependentWitness:  donordata.IndependentWitness{FirstNames: "I"},
				Tasks:               donordata.Tasks{ChooseYourSignatory: task.StateCompleted},
			},
			redirect: donor.PathNeedHelpSigningConfirmation,
		},
	}

	for yesNo, tc := range testcases {
		t.Run(yesNo.String(), func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {yesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.provided).
				Return(nil)

			err := CheckYouCanSign(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:               "lpa-id",
				AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "A"},
				IndependentWitness:  donordata.IndependentWitness{FirstNames: "I"},
				Tasks:               donordata.Tasks{ChooseYourSignatory: task.StateCompleted},
			})

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYouCanSignErrorOnPutStore(t *testing.T) {
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

	err := CheckYouCanSign(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCheckYouCanSignFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesIfYouWillBeAbleToSignYourself"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *checkYouCanSignData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := CheckYouCanSign(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
