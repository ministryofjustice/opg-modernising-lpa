package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRole(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:  appData,
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			CertificateProvider: CertificateProvider{CarryOutBy: "paper"},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                 appData,
			CertificateProvider: CertificateProvider{CarryOutBy: "paper"},
			Form:                &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{CarryOutBy: "paper"},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:  appData,
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRole(t *testing.T) {
	testCases := []struct {
		carryOutBy       string
		email            string
		expectedRedirect string
	}{
		{
			carryOutBy:       "paper",
			expectedRedirect: Paths.CertificateProviderAddress,
		},
		{
			carryOutBy:       "email",
			email:            "someone@example.com",
			expectedRedirect: Paths.HowDoYouKnowYourCertificateProvider,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.carryOutBy, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{
					CertificateProvider: CertificateProvider{CarryOutBy: tc.carryOutBy, Email: tc.email},
				}).
				Return(nil)

			form := url.Values{
				"carry-out-by": {tc.carryOutBy},
				"email":        {tc.email},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"carry-out-by": {"paper"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App: appData,
			Errors: map[string]string{
				"carry-out-by": "selectHowWouldCertificateProviderPreferToCarryOutTheirRole",
			},
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadHowWouldCertificateProviderPreferToCarryOutTheirRoleForm(t *testing.T) {
	form := url.Values{
		"carry-out-by": {"email"},
		"email":        {"someone@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)

	assert.Equal(t, "email", result.CarryOutBy)
	assert.Equal(t, "someone@example.com", result.Email)
}

func TestHowWouldCertificateProviderPreferToCarryOutTheirRoleFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
		errors map[string]string
	}{
		"paper": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: "paper",
			},
			errors: map[string]string{},
		},
		"email": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: "email",
				Email:      "someone@example.com",
			},
			errors: map[string]string{},
		},
		"email invalid": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: "email",
				Email:      "what",
			},
			errors: map[string]string{"email": "emailIncorrectFormat"},
		},
		"email missing": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: "email",
			},
			errors: map[string]string{"email": "enterEmail"},
		},
		"missing": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
			errors: map[string]string{
				"carry-out-by": "selectHowWouldCertificateProviderPreferToCarryOutTheirRole",
			},
		},
		"invalid": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: "what",
			},
			errors: map[string]string{
				"carry-out-by": "selectHowWouldCertificateProviderPreferToCarryOutTheirRole",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
