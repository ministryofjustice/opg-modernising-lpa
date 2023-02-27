package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowLongHaveYouKnownDonor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := actor.Person{FirstNames: "John"}
	certificateProvider := actor.CertificateProvider{DeclaredRelationshipLength: "gte-2-years"}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{You: donor, CertificateProvider: certificateProvider}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howLongHaveYouKnownDonorData{
			App:   testAppData,
			Donor: donor,
			Form: &howLongHaveYouKnownDonorForm{
				HowLong: "gte-2-years",
			},
		}).
		Return(nil)

	err := HowLongHaveYouKnownDonor(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowLongHaveYouKnownDonorWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := HowLongHaveYouKnownDonor(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowLongHaveYouKnownDonorWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := HowLongHaveYouKnownDonor(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowLongHaveYouKnownDonor(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{DeclaredRelationshipLength: "gte-2-years"},
		}).
		Return(nil)

	err := HowLongHaveYouKnownDonor(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderYourDetails, resp.Header.Get("Location"))
}

func TestPostHowLongHaveYouKnownDonorWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := HowLongHaveYouKnownDonor(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostHowLongHaveYouKnownDonorWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howLongHaveYouKnownDonorData{
			App:    testAppData,
			Form:   &howLongHaveYouKnownDonorForm{},
			Errors: validation.With("how-long", validation.SelectError{Label: "howLongYouHaveKnownDonor"}),
		}).
		Return(nil)

	err := HowLongHaveYouKnownDonor(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowLongHaveYouKnownDonorForm(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowLongHaveYouKnownDonorForm(r)

	assert.Equal(t, "gte-2-years", result.HowLong)
}

func TestHowLongHaveYouKnownDonorFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howLongHaveYouKnownDonorForm
		errors validation.List
	}{
		"gte-2-years": {
			form: &howLongHaveYouKnownDonorForm{
				HowLong: "gte-2-years",
			},
		},
		"lt-2-years": {
			form: &howLongHaveYouKnownDonorForm{
				HowLong: "lt-2-years",
			},
			errors: validation.With("how-long", validation.CustomError{Label: "mustHaveKnownDonorTwoYears"}),
		},
		"missing": {
			form:   &howLongHaveYouKnownDonorForm{},
			errors: validation.With("how-long", validation.SelectError{Label: "howLongYouHaveKnownDonor"}),
		},
		"invalid": {
			form: &howLongHaveYouKnownDonorForm{
				HowLong: "what",
			},
			errors: validation.With("how-long", validation.SelectError{Label: "howLongYouHaveKnownDonor"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
