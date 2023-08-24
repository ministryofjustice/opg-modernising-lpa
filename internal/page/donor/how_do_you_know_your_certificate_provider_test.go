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

func TestGetHowDoYouKnowYourCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howDoYouKnowYourCertificateProviderData{
			App:     testAppData,
			Form:    &howDoYouKnowYourCertificateProviderForm{},
			Options: actor.CertificateProviderRelationshipValues,
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowYourCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{
		Relationship: actor.Personally,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howDoYouKnowYourCertificateProviderData{
			App:                 testAppData,
			CertificateProvider: certificateProvider,
			Form:                &howDoYouKnowYourCertificateProviderForm{How: actor.Personally},
			Options:             actor.CertificateProviderRelationshipValues,
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &page.Lpa{
		CertificateProvider: certificateProvider,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowYourCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowDoYouKnowYourCertificateProvider(t *testing.T) {
	testCases := map[string]struct {
		form                       url.Values
		certificateProviderDetails actor.CertificateProvider
		taskState                  actor.TaskState
		redirect                   page.LpaPath
	}{
		"professionally": {
			form: url.Values{"how": {actor.Professionally.String()}},
			certificateProviderDetails: actor.CertificateProvider{
				FirstNames:   "John",
				Relationship: actor.Professionally,
			},
			taskState: actor.TaskCompleted,
			redirect:  page.Paths.DoYouWantToNotifyPeople,
		},
		"personally": {
			form: url.Values{"how": {actor.Personally.String()}},
			certificateProviderDetails: actor.CertificateProvider{
				FirstNames:   "John",
				Relationship: actor.Personally,
			},
			taskState: actor.TaskInProgress,
			redirect:  page.Paths.HowLongHaveYouKnownCertificateProvider,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                  "lpa-id",
					CertificateProvider: tc.certificateProviderDetails,
					Tasks: page.Tasks{
						YourDetails:         actor.TaskCompleted,
						ChooseAttorneys:     actor.TaskCompleted,
						CertificateProvider: tc.taskState,
					},
				}).
				Return(nil)

			err := HowDoYouKnowYourCertificateProvider(nil, donorStore)(testAppData, w, r, &page.Lpa{
				ID:                  "lpa-id",
				CertificateProvider: actor.CertificateProvider{FirstNames: "John"},
				Tasks: page.Tasks{
					YourDetails:     actor.TaskCompleted,
					ChooseAttorneys: actor.TaskCompleted,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowDoYouKnowYourCertificateProviderWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how": {actor.Personally.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := HowDoYouKnowYourCertificateProvider(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostHowDoYouKnowYourCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *howDoYouKnowYourCertificateProviderData) bool {
			return assert.Equal(t, validation.With("how", validation.SelectError{Label: "howYouKnowCertificateProvider"}), data.Errors)
		})).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowDoYouKnowYourCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"how": {actor.Personally.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowDoYouKnowYourCertificateProviderForm(r)

	assert.Equal(t, actor.Personally, result.How)
}

func TestHowDoYouKnowYourCertificateProviderFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howDoYouKnowYourCertificateProviderForm
		errors validation.List
	}{
		"valid": {
			form: &howDoYouKnowYourCertificateProviderForm{},
		},
		"invalid": {
			form: &howDoYouKnowYourCertificateProviderForm{
				Error: expectedError,
			},
			errors: validation.With("how", validation.SelectError{Label: "howYouKnowCertificateProvider"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
