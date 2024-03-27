package donor

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

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRole(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:     testAppData,
			Form:    &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
			Options: actor.ChannelValues,
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
			CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Paper},
			Form:                &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{CarryOutBy: actor.Paper},
			Options:             actor.ChannelValues,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Paper},
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
			Options: actor.ChannelValues,
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
			carryOutBy: actor.Paper,
		},
		{
			carryOutBy: actor.Online,
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

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"carry-out-by": {actor.Paper.String()},
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
	form := url.Values{
		"carry-out-by": {actor.Online.String()},
		"email":        {"someone@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)

	assert.Equal(t, actor.Online, result.CarryOutBy)
	assert.Equal(t, "someone@example.com", result.Email)
}

func TestHowWouldCertificateProviderPreferToCarryOutTheirRoleFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
		errors validation.List
	}{
		"paper": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.Paper,
			},
		},
		"online": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.Online,
				Email:      "someone@example.com",
			},
		},
		"online email invalid": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.Online,
				Email:      "what",
			},
			errors: validation.With("email", validation.EmailError{Label: "certificateProvidersEmail"}),
		},
		"online email missing": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: actor.Online,
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
