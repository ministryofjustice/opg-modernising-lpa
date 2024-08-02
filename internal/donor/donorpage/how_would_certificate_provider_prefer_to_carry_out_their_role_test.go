package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRole(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:     testAppData,
			Form:    &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
			Options: lpadata.ChannelValues,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                 testAppData,
			CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper},
			Form:                &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{CarryOutBy: lpadata.ChannelPaper},
			Options:             lpadata.ChannelValues,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:     testAppData,
			Form:    &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
			Options: lpadata.ChannelValues,
		}).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRole(t *testing.T) {
	testCases := []struct {
		carryOutBy lpadata.Channel
		email      string
	}{
		{
			carryOutBy: lpadata.ChannelPaper,
		},
		{
			carryOutBy: lpadata.ChannelOnline,
			email:      "someone@example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.carryOutBy.String(), func(t *testing.T) {
			form := url.Values{
				"carry-out-by": {tc.carryOutBy.String()},
				"email":        {tc.email},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:               "lpa-id",
					CertificateProvider: donordata.CertificateProvider{CarryOutBy: tc.carryOutBy, Email: tc.email},
				}).
				Return(nil)

			err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProviderAddress.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleChangingFromOnlineToPaper(t *testing.T) {
	form := url.Values{
		"carry-out-by": {lpadata.ChannelPaper.String()},
		"email":        {"a@b.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:               "lpa-id",
			CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper, Email: ""},
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID:               "lpa-id",
		CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelOnline, Email: "a@b.com"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"carry-out-by": {lpadata.ChannelPaper.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *howWouldCertificateProviderPreferToCarryOutTheirRoleData) bool {
			return assert.Equal(t, validation.With("carry-out-by", validation.SelectError{Label: "howYourCertificateProviderWouldPreferToCarryOutTheirRole"}), data.Errors)
		})).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowWouldCertificateProviderPreferToCarryOutTheirRoleForm(t *testing.T) {
	testcases := map[string]struct {
		carryOutBy   lpadata.Channel
		email        string
		formValues   url.Values
		expectedForm *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
	}{
		"online with email": {
			formValues: url.Values{
				"carry-out-by": {lpadata.ChannelOnline.String()},
				"email":        {"a@b.com"},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelOnline,
				Email:      "a@b.com",
			},
		},
		"paper": {
			formValues: url.Values{
				"carry-out-by": {lpadata.ChannelPaper.String()},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelPaper,
			},
		},
		"paper with email": {
			formValues: url.Values{
				"carry-out-by": {lpadata.ChannelPaper.String()},
				"email":        {"a@b.com"},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelPaper,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.formValues.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			result := readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)

			assert.Equal(t, tc.expectedForm, result)
		})
	}
}

func TestHowWouldCertificateProviderPreferToCarryOutTheirRoleFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
		errors validation.List
	}{
		"paper": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelPaper,
			},
		},
		"online": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelOnline,
				Email:      "someone@example.com",
			},
		},
		"online email invalid": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelOnline,
				Email:      "what",
			},
			errors: validation.With("email", validation.EmailError{Label: "certificateProvidersEmail"}),
		},
		"online email missing": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelOnline,
			},
			errors: validation.With("email", validation.EnterError{Label: "certificateProvidersEmail"}),
		},
		"missing": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.Channel(0),
				Error:      expectedError,
			},
			errors: validation.With("carry-out-by", validation.SelectError{Label: "howYourCertificateProviderWouldPreferToCarryOutTheirRole"}),
		},
		"invalid": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.Channel(99),
				Error:      expectedError,
			},
			errors: validation.With("carry-out-by", validation.SelectError{Label: "howYourCertificateProviderWouldPreferToCarryOutTheirRole"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
