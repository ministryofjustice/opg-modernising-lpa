package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowDoYouKnowYourCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howDoYouKnowYourCertificateProviderData{
			App:  testAppData,
			Form: &howDoYouKnowYourCertificateProviderForm{},
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowYourCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := HowDoYouKnowYourCertificateProvider(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetHowDoYouKnowYourCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := page.CertificateProviderDetails{
		Relationship: "friend",
	}
	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProviderDetails: certificateProvider,
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howDoYouKnowYourCertificateProviderData{
			App:                        testAppData,
			CertificateProviderDetails: certificateProvider,
			Form:                       &howDoYouKnowYourCertificateProviderForm{How: "friend"},
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowYourCertificateProviderWhenTemplateErrors(t *testing.T) {
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

	err := HowDoYouKnowYourCertificateProvider(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowDoYouKnowYourCertificateProvider(t *testing.T) {
	testCases := map[string]struct {
		form                       url.Values
		certificateProviderDetails page.CertificateProviderDetails
		taskState                  page.TaskState
		redirect                   string
	}{
		"legal-professional": {
			form: url.Values{"how": {"legal-professional"}},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:   "John",
				Relationship: "legal-professional",
			},
			taskState: page.TaskCompleted,
			redirect:  "/lpa/lpa-id" + page.Paths.DoYouWantToNotifyPeople,
		},
		"health-professional": {
			form: url.Values{"how": {"health-professional"}},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:   "John",
				Relationship: "health-professional",
			},
			taskState: page.TaskCompleted,
			redirect:  "/lpa/lpa-id" + page.Paths.DoYouWantToNotifyPeople,
		},
		"other": {
			form: url.Values{"how": {"other"}, "description": {"This"}},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:              "John",
				Relationship:            "other",
				RelationshipDescription: "This",
				RelationshipLength:      "gte-2-years",
			},
			taskState: page.TaskInProgress,
			redirect:  "/lpa/lpa-id" + page.Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - friend": {
			form: url.Values{"how": {"friend"}},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:         "John",
				Relationship:       "friend",
				RelationshipLength: "gte-2-years",
			},
			taskState: page.TaskInProgress,
			redirect:  "/lpa/lpa-id" + page.Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - neighbour": {
			form: url.Values{"how": {"neighbour"}},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:         "John",
				Relationship:       "neighbour",
				RelationshipLength: "gte-2-years",
			},
			taskState: page.TaskInProgress,
			redirect:  "/lpa/lpa-id" + page.Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - colleague": {
			form: url.Values{"how": {"colleague"}},
			certificateProviderDetails: page.CertificateProviderDetails{
				FirstNames:         "John",
				Relationship:       "colleague",
				RelationshipLength: "gte-2-years",
			},
			taskState: page.TaskInProgress,
			redirect:  "/lpa/lpa-id" + page.Paths.HowLongHaveYouKnownCertificateProvider,
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
					CertificateProviderDetails: page.CertificateProviderDetails{FirstNames: "John", Relationship: "what", RelationshipLength: "gte-2-years"},
					Tasks: page.Tasks{
						YourDetails:     page.TaskCompleted,
						ChooseAttorneys: page.TaskCompleted,
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					CertificateProviderDetails: tc.certificateProviderDetails,
					Tasks: page.Tasks{
						YourDetails:         page.TaskCompleted,
						ChooseAttorneys:     page.TaskCompleted,
						CertificateProvider: tc.taskState,
					},
				}).
				Return(nil)

			err := HowDoYouKnowYourCertificateProvider(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostHowDoYouKnowYourCertificateProviderWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how": {"friend"},
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

	err := HowDoYouKnowYourCertificateProvider(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostHowDoYouKnowYourCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howDoYouKnowYourCertificateProviderData{
			App:    testAppData,
			Form:   &howDoYouKnowYourCertificateProviderForm{},
			Errors: validation.With("how", validation.SelectError{Label: "howYouKnowCertificateProvider"}),
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowDoYouKnowYourCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"how":         {"friend"},
		"description": {"What"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowDoYouKnowYourCertificateProviderForm(r)

	assert.Equal(t, "friend", result.How)
	assert.Equal(t, "What", result.Description)
}

func TestHowDoYouKnowYourCertificateProviderFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howDoYouKnowYourCertificateProviderForm
		errors validation.List
	}{
		"valid": {
			form: &howDoYouKnowYourCertificateProviderForm{
				How:         "friend",
				Description: "This",
			},
		},
		"missing": {
			form:   &howDoYouKnowYourCertificateProviderForm{},
			errors: validation.With("how", validation.SelectError{Label: "howYouKnowCertificateProvider"}),
		},
		"invalid-option": {
			form: &howDoYouKnowYourCertificateProviderForm{
				How: "what",
			},
			errors: validation.With("how", validation.SelectError{Label: "howYouKnowCertificateProvider"}),
		},
		"other-missing-description": {
			form: &howDoYouKnowYourCertificateProviderForm{
				How: "other",
			},
			errors: validation.With("description", validation.EnterError{Label: "description"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
