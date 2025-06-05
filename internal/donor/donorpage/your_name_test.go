package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourNameData{
			App:  testAppData,
			Form: &yourNameForm{},
		}).
		Return(nil)

	err := YourName(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourNameFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourNameData{
			App: testAppData,
			Form: &yourNameForm{
				FirstNames: "John",
				LastName:   "Doe",
				OtherNames: "Fawn",
			},
		}).
		Return(nil)

	err := YourName(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{
			FirstNames: "John",
			LastName:   "Doe",
			OtherNames: "Fawn",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := YourName(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{Donor: donordata.Donor{FirstNames: "John"}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourName(t *testing.T) {
	testCases := map[string]struct {
		url      string
		form     url.Values
		person   donordata.Donor
		redirect string
	}{
		"valid": {
			url: "/",
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"other-names": {"Fawn"},
			},
			person: donordata.Donor{
				FirstNames: "John",
				LastName:   "Doe",
				OtherNames: "Fawn",
				Email:      "what",
			},
			redirect: donor.PathYourDateOfBirth.Format("lpa-id"),
		},
		"shares name": {
			url: "/",
			form: url.Values{
				"first-names": {"Jane"},
				"last-name":   {"Bloggs"},
			},
			person: donordata.Donor{
				FirstNames: "Jane",
				LastName:   "Bloggs",
				Email:      "what",
			},
			redirect: donor.PathWarningInterruption.FormatQuery(
				"lpa-id",
				url.Values{
					"warningFrom": {"/abc"},
					"next":        {donor.PathYourDateOfBirth.Format("lpa-id")},
					"actor":       {actor.TypeDonor.String()},
				},
			),
		},
		"making another lpa": {
			url: "/?makingAnotherLPA=1",
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"other-names": {"Fawn"},
			},
			person: donordata.Donor{
				FirstNames: "John",
				LastName:   "Doe",
				OtherNames: "Fawn",
				Email:      "what",
			},
			redirect: donor.PathWeHaveUpdatedYourDetails.Format("lpa-id") + "?detail=name",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Login(r).
				Return(&sesh.LoginSession{Email: "what"}, nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:               "lpa-id",
					Donor:               tc.person,
					CertificateProvider: donordata.CertificateProvider{FirstNames: "Jane", LastName: "Bloggs"},
				}).
				Return(nil)

			appData := appcontext.Data{Page: "/abc"}
			err := YourName(nil, donorStore, sessionStore)(appData, w, r, &donordata.Provided{
				LpaID:                          "lpa-id",
				Donor:                          donordata.Donor{FirstNames: "John"},
				CertificateProvider:            donordata.CertificateProvider{FirstNames: "Jane", LastName: "Bloggs"},
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourNameWhenDetailsNotChanged(t *testing.T) {
	testcases := map[string]struct {
		url      string
		redirect donor.Path
	}{
		"making first": {
			url:      "/",
			redirect: donor.PathYourDateOfBirth,
		},
		"making another": {
			url:      "/?makingAnotherLPA=1",
			redirect: donor.PathMakeANewLPA,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"other-names": {"Fawn"},
			}

			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := YourName(nil, nil, nil)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					FirstNames: "John",
					LastName:   "Doe",
					OtherNames: "Fawn",
				},
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourNameWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourNameData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
			},
			dataMatcher: func(t *testing.T, data *yourNameData) bool {
				return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.
				On("Execute", w, mock.MatchedBy(func(data *yourNameData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := YourName(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostYourNameWhenSessionStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	err := YourName(nil, nil, sessionStore)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostYourNameWhenSessionMissingEmail(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{}, nil)

	err := YourName(nil, nil, sessionStore)(testAppData, w, r, &donordata.Provided{})
	assert.EqualError(t, err, "no email in login session")
}

func TestPostYourNameWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Email: "what"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourName(nil, donorStore, sessionStore)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}
