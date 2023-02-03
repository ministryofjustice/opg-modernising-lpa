package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowDoYouKnowYourCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howDoYouKnowYourCertificateProviderData{
			App:  appData,
			Form: &howDoYouKnowYourCertificateProviderForm{},
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowDoYouKnowYourCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := HowDoYouKnowYourCertificateProvider(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowDoYouKnowYourCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := CertificateProvider{
		Relationship: "friend",
	}
	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			CertificateProvider: certificateProvider,
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			Form:                &howDoYouKnowYourCertificateProviderForm{How: "friend"},
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetHowDoYouKnowYourCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	err := HowDoYouKnowYourCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostHowDoYouKnowYourCertificateProvider(t *testing.T) {
	testCases := map[string]struct {
		form                url.Values
		certificateProvider CertificateProvider
		taskState           TaskState
		redirect            string
	}{
		"legal-professional": {
			form: url.Values{"how": {"legal-professional"}},
			certificateProvider: CertificateProvider{
				FirstNames:   "John",
				Relationship: "legal-professional",
			},
			taskState: TaskCompleted,
			redirect:  "/lpa/lpa-id" + Paths.DoYouWantToNotifyPeople,
		},
		"health-professional": {
			form: url.Values{"how": {"health-professional"}},
			certificateProvider: CertificateProvider{
				FirstNames:   "John",
				Relationship: "health-professional",
			},
			taskState: TaskCompleted,
			redirect:  "/lpa/lpa-id" + Paths.DoYouWantToNotifyPeople,
		},
		"other": {
			form: url.Values{"how": {"other"}, "description": {"This"}},
			certificateProvider: CertificateProvider{
				FirstNames:              "John",
				Relationship:            "other",
				RelationshipDescription: "This",
				RelationshipLength:      "gte-2-years",
			},
			taskState: TaskInProgress,
			redirect:  "/lpa/lpa-id" + Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - friend": {
			form: url.Values{"how": {"friend"}},
			certificateProvider: CertificateProvider{
				FirstNames:         "John",
				Relationship:       "friend",
				RelationshipLength: "gte-2-years",
			},
			taskState: TaskInProgress,
			redirect:  "/lpa/lpa-id" + Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - neighbour": {
			form: url.Values{"how": {"neighbour"}},
			certificateProvider: CertificateProvider{
				FirstNames:         "John",
				Relationship:       "neighbour",
				RelationshipLength: "gte-2-years",
			},
			taskState: TaskInProgress,
			redirect:  "/lpa/lpa-id" + Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - colleague": {
			form: url.Values{"how": {"colleague"}},
			certificateProvider: CertificateProvider{
				FirstNames:         "John",
				Relationship:       "colleague",
				RelationshipLength: "gte-2-years",
			},
			taskState: TaskInProgress,
			redirect:  "/lpa/lpa-id" + Paths.HowLongHaveYouKnownCertificateProvider,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
					CertificateProvider: CertificateProvider{FirstNames: "John", Relationship: "what", RelationshipLength: "gte-2-years"},
					Tasks: Tasks{
						YourDetails:     TaskCompleted,
						ChooseAttorneys: TaskCompleted,
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &Lpa{
					CertificateProvider: tc.certificateProvider,
					Tasks: Tasks{
						YourDetails:         TaskCompleted,
						ChooseAttorneys:     TaskCompleted,
						CertificateProvider: tc.taskState,
					},
				}).
				Return(nil)

			err := HowDoYouKnowYourCertificateProvider(nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostHowDoYouKnowYourCertificateProviderWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how": {"friend"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := HowDoYouKnowYourCertificateProvider(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowDoYouKnowYourCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howDoYouKnowYourCertificateProviderData{
			App:    appData,
			Form:   &howDoYouKnowYourCertificateProviderForm{},
			Errors: validation.With("how", validation.SelectError{Label: "howYouKnowCertificateProvider"}),
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadHowDoYouKnowYourCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"how":         {"friend"},
		"description": {"What"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

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
