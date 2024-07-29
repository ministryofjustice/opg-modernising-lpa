package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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
			Options: donordata.ChannelValues,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
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
			CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.ChannelPaper},
			Form:                &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{CarryOutBy: actor.ChannelPaper},
			Options:             donordata.ChannelValues,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.ChannelPaper},
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
			Options: donordata.ChannelValues,
		}).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRole(t *testing.T) {
	testCases := []struct {
		carryOutBy actor.Channel
		email      string
	}{
		{
			carryOutBy: actor.ChannelPaper,
		},
		{
			carryOutBy: actor.ChannelOnline,
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
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:               "lpa-id",
					CertificateProvider: actor.CertificateProvider{CarryOutBy: tc.carryOutBy, Email: tc.email},
				}).
				Return(nil)

			err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProviderAddress.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleChangingFromOnlineToPaper(t *testing.T) {
	form := url.Values{
		"carry-out-by": {actor.ChannelPaper.String()},
		"email":        {"a@b.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:               "lpa-id",
			CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.ChannelPaper, Email: ""},
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:               "lpa-id",
		CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.ChannelOnline, Email: "a@b.com"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"carry-out-by": {actor.ChannelPaper.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

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

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowWouldCertificateProviderPreferToCarryOutTheirRoleForm(t *testing.T) {
	testcases := map[string]struct {
		carryOutBy   actor.Channel
		email        string
		formValues   url.Values
		expectedForm *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
	}{
		"online with email": {
			formValues: url.Values{
				"carry-out-by": {actor.ChannelOnline.String()},
				"email":        {"a@b.com"},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.ChannelOnline,
				Email:      "a@b.com",
			},
		},
		"paper": {
			formValues: url.Values{
				"carry-out-by": {actor.ChannelPaper.String()},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.ChannelPaper,
			},
		},
		"paper with email": {
			formValues: url.Values{
				"carry-out-by": {actor.ChannelPaper.String()},
				"email":        {"a@b.com"},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.ChannelPaper,
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
				CarryOutBy: actor.ChannelPaper,
			},
		},
		"online": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.ChannelOnline,
				Email:      "someone@example.com",
			},
		},
		"online email invalid": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.ChannelOnline,
				Email:      "what",
			},
			errors: validation.With("email", validation.EmailError{Label: "certificateProvidersEmail"}),
		},
		"online email missing": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.ChannelOnline,
			},
			errors: validation.With("email", validation.EnterError{Label: "certificateProvidersEmail"}),
		},
		"missing": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.Channel(0),
				Error:      expectedError,
			},
			errors: validation.With("carry-out-by", validation.SelectError{Label: "howYourCertificateProviderWouldPreferToCarryOutTheirRole"}),
		},
		"invalid": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.Channel(99),
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
