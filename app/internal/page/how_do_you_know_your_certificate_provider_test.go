package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowDoYouKnowYourCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howDoYouKnowYourCertificateProviderData{
			App:  appData,
			Form: &howDoYouKnowYourCertificateProviderForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowDoYouKnowYourCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowDoYouKnowYourCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowDoYouKnowYourCertificateProvider(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowDoYouKnowYourCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	certificateProvider := CertificateProvider{
		Relationship: "friend",
	}
	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
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

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowDoYouKnowYourCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetHowDoYouKnowYourCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

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
			redirect:  appData.Paths.CheckYourLpa,
		},
		"health-professional": {
			form: url.Values{"how": {"health-professional"}},
			certificateProvider: CertificateProvider{
				FirstNames:   "John",
				Relationship: "health-professional",
			},
			taskState: TaskCompleted,
			redirect:  appData.Paths.CheckYourLpa,
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
			redirect:  appData.Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - friend": {
			form: url.Values{"how": {"friend"}},
			certificateProvider: CertificateProvider{
				FirstNames:         "John",
				Relationship:       "friend",
				RelationshipLength: "gte-2-years",
			},
			taskState: TaskInProgress,
			redirect:  appData.Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - neighbour": {
			form: url.Values{"how": {"neighbour"}},
			certificateProvider: CertificateProvider{
				FirstNames:         "John",
				Relationship:       "neighbour",
				RelationshipLength: "gte-2-years",
			},
			taskState: TaskInProgress,
			redirect:  appData.Paths.HowLongHaveYouKnownCertificateProvider,
		},
		"lay - colleague": {
			form: url.Values{"how": {"colleague"}},
			certificateProvider: CertificateProvider{
				FirstNames:         "John",
				Relationship:       "colleague",
				RelationshipLength: "gte-2-years",
			},
			taskState: TaskInProgress,
			redirect:  appData.Paths.HowLongHaveYouKnownCertificateProvider,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{
					Attorneys:                 []Attorney{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)}},
					HowAttorneysMakeDecisions: Jointly,
					CertificateProvider:       CertificateProvider{FirstNames: "John", Relationship: "what", RelationshipLength: "gte-2-years"},
				}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{
					Attorneys:                 []Attorney{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)}},
					HowAttorneysMakeDecisions: Jointly,
					CertificateProvider:       tc.certificateProvider,
					Tasks: Tasks{
						CertificateProvider: tc.taskState,
					},
				}).
				Return(nil)

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

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
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"how": {"friend"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowDoYouKnowYourCertificateProvider(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowDoYouKnowYourCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howDoYouKnowYourCertificateProviderData{
			App:  appData,
			Form: &howDoYouKnowYourCertificateProviderForm{},
			Errors: map[string]string{
				"how": "selectHowYouKnowCertificateProvider",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

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
		errors map[string]string
	}{
		"valid": {
			form: &howDoYouKnowYourCertificateProviderForm{
				How:         "friend",
				Description: "This",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &howDoYouKnowYourCertificateProviderForm{},
			errors: map[string]string{
				"how": "selectHowYouKnowCertificateProvider",
			},
		},
		"invalid-option": {
			form: &howDoYouKnowYourCertificateProviderForm{
				How: "what",
			},
			errors: map[string]string{
				"how": "selectHowYouKnowCertificateProvider",
			},
		},
		"other-missing-description": {
			form: &howDoYouKnowYourCertificateProviderForm{
				How: "other",
			},
			errors: map[string]string{
				"description": "enterDescription",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
