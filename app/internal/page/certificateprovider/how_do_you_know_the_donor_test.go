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

func TestGetHowDoYouKnowTheDonor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howDoYouKnowTheDonorData{
			App:  testAppData,
			Form: &howDoYouKnowTheDonorForm{},
		}).
		Return(nil)

	err := HowDoYouKnowTheDonor(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowTheDonorWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := HowDoYouKnowTheDonor(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetHowDoYouKnowTheDonorFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := actor.Person{FirstNames: "John"}
	certificateProvider := actor.CertificateProvider{DeclaredRelationship: "friend"}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			You:                 donor,
			CertificateProvider: certificateProvider,
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howDoYouKnowTheDonorData{
			App:   testAppData,
			Donor: donor,
			Form:  &howDoYouKnowTheDonorForm{How: "friend"},
		}).
		Return(nil)

	err := HowDoYouKnowTheDonor(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowTheDonorWhenTemplateErrors(t *testing.T) {
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

	err := HowDoYouKnowTheDonor(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowDoYouKnowTheDonor(t *testing.T) {
	testCases := map[string]struct {
		form                url.Values
		certificateProvider actor.CertificateProvider
		taskState           page.TaskState
		redirect            string
	}{
		"personally": {
			form: url.Values{"how": {"personally"}},
			certificateProvider: actor.CertificateProvider{
				FirstNames:           "John",
				DeclaredRelationship: "personally",
			},
			redirect: page.Paths.HowLongHaveYouKnownDonor,
		},
		"professionally": {
			form: url.Values{"how": {"professionally"}},
			certificateProvider: actor.CertificateProvider{
				FirstNames:           "John",
				DeclaredRelationship: "professionally",
			},
			redirect: page.Paths.CertificateProviderYourDetails,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					CertificateProvider: actor.CertificateProvider{FirstNames: "John"},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					CertificateProvider: tc.certificateProvider,
				}).
				Return(nil)

			err := HowDoYouKnowTheDonor(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostHowDoYouKnowTheDonorWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how": {"personally"},
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

	err := HowDoYouKnowTheDonor(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostHowDoYouKnowTheDonorWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howDoYouKnowTheDonorData{
			App:    testAppData,
			Form:   &howDoYouKnowTheDonorForm{},
			Errors: validation.With("how", validation.SelectError{Label: "howYouKnowDonor"}),
		}).
		Return(nil)

	err := HowDoYouKnowTheDonor(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowDoYouKnowTheDonorForm(t *testing.T) {
	form := url.Values{
		"how": {"personally"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowDoYouKnowTheDonorForm(r)

	assert.Equal(t, "personally", result.How)
}

func TestHowDoYouKnowTheDonorFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howDoYouKnowTheDonorForm
		errors validation.List
	}{
		"personally": {
			form: &howDoYouKnowTheDonorForm{How: "personally"},
		},
		"professionally": {
			form: &howDoYouKnowTheDonorForm{How: "professionally"},
		},
		"missing": {
			form:   &howDoYouKnowTheDonorForm{},
			errors: validation.With("how", validation.SelectError{Label: "howYouKnowDonor"}),
		},
		"invalid-option": {
			form:   &howDoYouKnowTheDonorForm{How: "what"},
			errors: validation.With("how", validation.SelectError{Label: "howYouKnowDonor"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
