package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourMobile(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourMobileData{
			App:  testAppData,
			Form: &yourMobileForm{},
		}).
		Return(nil)

	err := YourMobile(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourMobileWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := YourMobile(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourMobile(t *testing.T) {
	form := url.Values{
		"mobile": {"07000000000"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Donor: donordata.Donor{
				FirstNames: "John",
				Mobile:     "07000000000",
			},
		}).
		Return(nil)

	err := YourMobile(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{
			FirstNames: "John",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathReceivingUpdatesAboutYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourMobileWhenValidationError(t *testing.T) {
	form := url.Values{
		"mobile": {"john"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *yourMobileData) bool {
			return assert.Equal(t, validation.With("mobile", validation.PhoneError{Label: "mobile", Tmpl: "errorMobile"}), data.Errors)
		})).
		Return(nil)

	err := YourMobile(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourMobileWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"mobile": {"07000000000"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := YourMobile(nil, donorStore)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestReadYourMobileForm(t *testing.T) {
	form := url.Values{"mobile": {"07000000000"}}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourMobileForm(r)

	assert.Equal(t, "07000000000", result.Mobile)
}

func TestYourMobileFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *yourMobileForm
		errors validation.List
	}{
		"valid": {
			form: &yourMobileForm{Mobile: "07000000000"},
		},
		"empty": {
			form: &yourMobileForm{},
		},
		"invalid": {
			form:   &yourMobileForm{Mobile: "john"},
			errors: validation.With("mobile", validation.PhoneError{Label: "mobile", Tmpl: "errorMobile"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
