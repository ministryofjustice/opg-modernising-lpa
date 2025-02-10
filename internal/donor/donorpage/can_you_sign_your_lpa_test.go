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

func TestGetCanYouSignYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &canYouSignYourLpaData{
			App:  testAppData,
			Form: form.NewEmptySelectForm[donordata.YesNoMaybe](donordata.YesNoMaybeValues, "yesIfCanSign"),
		}).
		Return(nil)

	err := CanYouSignYourLpa(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCanYouSignYourLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := CanYouSignYourLpa(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCanYouSignYourLpa(t *testing.T) {
	testCases := map[string]struct {
		form     url.Values
		provided *donordata.Provided
		redirect donor.Path
	}{
		"can sign": {
			form: url.Values{
				form.FieldNames.Select: {donordata.Yes.String()},
			},
			provided: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					ThinksCanSign: donordata.Yes,
					CanSign:       form.Yes,
				},
			},
			redirect: donor.PathYourPreferredLanguage,
		},
		"cannot sign": {
			form: url.Values{
				form.FieldNames.Select: {donordata.No.String()},
			},
			provided: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					ThinksCanSign: donordata.No,
				},
				AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "A"},
				IndependentWitness:  donordata.IndependentWitness{FirstNames: "I"},
				Tasks:               donordata.Tasks{ChooseYourSignatory: task.StateCompleted},
			},
			redirect: donor.PathCheckYouCanSign,
		},
		"maybe can sign": {
			form: url.Values{
				form.FieldNames.Select: {donordata.Maybe.String()},
			},
			provided: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					ThinksCanSign: donordata.Maybe,
				},
				AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "A"},
				IndependentWitness:  donordata.IndependentWitness{FirstNames: "I"},
				Tasks:               donordata.Tasks{ChooseYourSignatory: task.StateCompleted},
			},
			redirect: donor.PathCheckYouCanSign,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), tc.provided).
				Return(nil)

			err := CanYouSignYourLpa(nil, donorStore)(testAppData, w, r, &donordata.Provided{
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

func TestPostCanYouSignYourLpaWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Select: {"what"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *canYouSignYourLpaData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.Select, validation.SelectError{Label: "yesIfCanSign"}), data.Errors)
		})).
		Return(nil)

	err := CanYouSignYourLpa(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCanYouSignYourLpaWhenStoreErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {donordata.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := CanYouSignYourLpa(nil, donorStore)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}
