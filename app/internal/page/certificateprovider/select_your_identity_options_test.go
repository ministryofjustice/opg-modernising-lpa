package certificateprovider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSelectYourIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &selectYourIdentityOptionsData{
			App:  testAppData,
			Page: 2,
			Form: &selectYourIdentityOptionsForm{},
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Execute, 2, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSelectYourIdentityOptionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := SelectYourIdentityOptions(nil, 0, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetSelectYourIdentityOptionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{IdentityOption: identity.Passport}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &selectYourIdentityOptionsData{
			App:  testAppData,
			Form: &selectYourIdentityOptionsForm{Selected: identity.Passport},
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Execute, 0, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSelectYourIdentityOptionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := SelectYourIdentityOptions(template.Execute, 0, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSelectYourIdentityOptions(t *testing.T) {
	form := url.Values{
		"option": {"passport"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{
			LpaID:          "lpa-id",
			IdentityOption: identity.Passport,
		}).
		Return(nil)

	err := SelectYourIdentityOptions(nil, 0, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.YourChosenIdentityOptions.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostSelectYourIdentityOptionsNone(t *testing.T) {
	for pageIndex, nextPath := range map[int]string{
		0: page.Paths.CertificateProvider.SelectYourIdentityOptions1.Format("lpa-id"),
		1: page.Paths.CertificateProvider.SelectYourIdentityOptions2.Format("lpa-id"),
		2: page.Paths.CertificateProviderStart.Format(),
	} {
		t.Run(fmt.Sprintf("Page%d", pageIndex), func(t *testing.T) {
			form := url.Values{
				"option": {"none"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("Get", r.Context()).
				Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)

			err := SelectYourIdentityOptions(nil, pageIndex, certificateProviderStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, nextPath, resp.Header.Get("Location"))
		})
	}
}

func TestPostSelectYourIdentityOptionsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"option": {"passport"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := SelectYourIdentityOptions(nil, 0, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostSelectYourIdentityOptionsWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &selectYourIdentityOptionsData{
			App:    testAppData,
			Form:   &selectYourIdentityOptionsForm{},
			Errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Execute, 0, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadSelectYourIdentityOptionsForm(t *testing.T) {
	form := url.Values{
		"option": {"passport"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readSelectYourIdentityOptionsForm(r)

	assert.Equal(t, identity.Passport, result.Selected)
	assert.False(t, result.None)
}

func TestSelectYourIdentityOptionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form      *selectYourIdentityOptionsForm
		errors    validation.List
		pageIndex int
	}{
		"valid": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.EasyID,
			},
		},
		"none": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.UnknownOption,
				None:     true,
			},
		},
		"missing": {
			form:   &selectYourIdentityOptionsForm{},
			errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		},
		"invalid": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.UnknownOption,
			},
			errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		},
		"missing after first page": {
			form:      &selectYourIdentityOptionsForm{},
			errors:    validation.With("option", validation.SelectError{Label: "whichDocumentYouWillUse"}),
			pageIndex: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(tc.pageIndex))
		})
	}
}
