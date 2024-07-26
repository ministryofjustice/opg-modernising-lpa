package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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

	err := CanYouSignYourLpa(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
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

	err := CanYouSignYourLpa(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCanYouSignYourLpa(t *testing.T) {
	testCases := map[string]struct {
		form     url.Values
		person   actor.Donor
		redirect page.LpaPath
	}{
		"can sign": {
			form: url.Values{
				"can-sign": {actor.Yes.String()},
			},
			person: actor.Donor{
				ThinksCanSign: actor.Yes,
				CanSign:       form.Yes,
			},
			redirect: page.Paths.YourPreferredLanguage,
		},
		"cannot sign": {
			form: url.Values{
				"can-sign": {actor.No.String()},
			},
			person: actor.Donor{
				ThinksCanSign: actor.No,
			},
			redirect: page.Paths.CheckYouCanSign,
		},
		"maybe can sign": {
			form: url.Values{
				"can-sign": {actor.Maybe.String()},
			},
			person: actor.Donor{
				ThinksCanSign: actor.Maybe,
			},
			redirect: page.Paths.CheckYouCanSign,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &actor.DonorProvidedDetails{
					LpaID: "lpa-id",
					Donor: tc.person,
				}).
				Return(nil)

			err := CanYouSignYourLpa(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
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

	err := CanYouSignYourLpa(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCanYouSignYourLpaWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"can-sign": {actor.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := CanYouSignYourLpa(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestReadCanYouSignYourLpaForm(t *testing.T) {
	f := url.Values{
		"can-sign": {actor.Yes.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readCanYouSignYourLpaForm(r)

	assert.Equal(t, actor.Yes, result.CanSign)
	assert.Nil(t, result.CanSignError)
}

func TestCanYouSignYourLpaFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *canYouSignYourLpaForm
		errors validation.List
	}{
		"valid": {
			form: &canYouSignYourLpaForm{},
		},
		"invalid": {
			form: &canYouSignYourLpaForm{
				CanSignError: expectedError,
			},
			errors: validation.With("can-sign", validation.SelectError{Label: "yesIfCanSign"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
