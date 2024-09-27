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
			App:               testAppData,
			Form:              &canYouSignYourLpaForm{},
			YesNoMaybeOptions: donordata.YesNoMaybeValues,
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
		person   donordata.Donor
		redirect donor.Path
	}{
		"can sign": {
			form: url.Values{
				"can-sign": {donordata.Yes.String()},
			},
			person: donordata.Donor{
				ThinksCanSign: donordata.Yes,
				CanSign:       form.Yes,
			},
			redirect: donor.PathYourPreferredLanguage,
		},
		"cannot sign": {
			form: url.Values{
				"can-sign": {donordata.No.String()},
			},
			person: donordata.Donor{
				ThinksCanSign: donordata.No,
			},
			redirect: donor.PathCheckYouCanSign,
		},
		"maybe can sign": {
			form: url.Values{
				"can-sign": {donordata.Maybe.String()},
			},
			person: donordata.Donor{
				ThinksCanSign: donordata.Maybe,
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
				On("Put", r.Context(), &donordata.Provided{
					LpaID: "lpa-id",
					Donor: tc.person,
				}).
				Return(nil)

			err := CanYouSignYourLpa(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCanYouSignYourLpaWhenValidationError(t *testing.T) {
	form := url.Values{
		"can-sign": {"what"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *canYouSignYourLpaData) bool {
			return assert.Equal(t, validation.With("can-sign", validation.SelectError{Label: "yesIfCanSign"}), data.Errors)
		})).
		Return(nil)

	err := CanYouSignYourLpa(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCanYouSignYourLpaWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"can-sign": {donordata.Yes.String()},
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

func TestReadCanYouSignYourLpaForm(t *testing.T) {
	f := url.Values{
		"can-sign": {donordata.Yes.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readCanYouSignYourLpaForm(r)

	assert.Equal(t, donordata.Yes, result.CanSign)
}

func TestCanYouSignYourLpaFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *canYouSignYourLpaForm
		errors validation.List
	}{
		"valid": {
			form: &canYouSignYourLpaForm{CanSign: donordata.Yes},
		},
		"invalid": {
			form:   &canYouSignYourLpaForm{},
			errors: validation.With("can-sign", validation.SelectError{Label: "yesIfCanSign"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
