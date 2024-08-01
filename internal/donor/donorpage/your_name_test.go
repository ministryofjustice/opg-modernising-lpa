package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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

	err := YourName(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
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

	err := YourName(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{
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

	err := YourName(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Donor: actor.Donor{FirstNames: "John"}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourName(t *testing.T) {
	testCases := map[string]struct {
		url      string
		form     url.Values
		person   actor.Donor
		redirect string
	}{
		"valid": {
			url: "/",
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"other-names": {"Fawn"},
			},
			person: actor.Donor{
				FirstNames: "John",
				LastName:   "Doe",
				OtherNames: "Fawn",
				Email:      "what",
			},
			redirect: page.Paths.YourDateOfBirth.Format("lpa-id"),
		},
		"warning ignored": {
			url: "/",
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Bloggs"},
				"ignore-name-warning": {"1|4|Jane|Bloggs"},
			},
			person: actor.Donor{
				FirstNames: "Jane",
				LastName:   "Bloggs",
				Email:      "what",
			},
			redirect: page.Paths.YourDateOfBirth.Format("lpa-id"),
		},
		"making another lpa": {
			url: "/?makingAnotherLPA=1",
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"other-names": {"Fawn"},
			},
			person: actor.Donor{
				FirstNames: "John",
				LastName:   "Doe",
				OtherNames: "Fawn",
				Email:      "what",
			},
			redirect: page.Paths.WeHaveUpdatedYourDetails.Format("lpa-id") + "?detail=name",
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
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:               "lpa-id",
					Donor:               tc.person,
					CertificateProvider: actor.CertificateProvider{FirstNames: "Jane", LastName: "Bloggs"},
				}).
				Return(nil)

			err := YourName(nil, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID:                          "lpa-id",
				Donor:                          actor.Donor{FirstNames: "John"},
				CertificateProvider:            actor.CertificateProvider{FirstNames: "Jane", LastName: "Bloggs"},
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
		redirect page.LpaPath
	}{
		"making first": {
			url:      "/",
			redirect: page.Paths.YourDateOfBirth,
		},
		"making another": {
			url:      "/?makingAnotherLPA=1",
			redirect: page.Paths.MakeANewLPA,
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

			err := YourName(nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{
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

			err := YourName(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
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

	err := YourName(nil, nil, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
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

	err := YourName(nil, nil, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
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

	err := YourName(nil, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestDonorMatches(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: actor.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  actor.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, donorMatches(donor, "x", "y"))
	assert.Equal(t, actor.TypeNone, donorMatches(donor, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, donorMatches(donor, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, donorMatches(donor, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, donorMatches(donor, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, donorMatches(donor, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, donorMatches(donor, "k", "l"))
	assert.Equal(t, actor.TypePersonToNotify, donorMatches(donor, "m", "n"))
	assert.Equal(t, actor.TypePersonToNotify, donorMatches(donor, "O", "P"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, donorMatches(donor, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, donorMatches(donor, "i", "w"))
}

func TestDonorMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, donorMatches(donor, "", ""))
}
