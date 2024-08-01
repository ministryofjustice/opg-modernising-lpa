package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotify(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &choosePeopleToNotifyData{
			App:  testAppData,
			Form: &choosePeopleToNotifyForm{},
		}).
		Return(nil)

	err := ChoosePeopleToNotify(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)

	err := ChoosePeopleToNotify(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID: "lpa-id",
		PeopleToNotify: donordata.PeopleToNotify{
			{
				UID:        actoruid.New(),
				Address:    testAddress,
				FirstNames: "Johnny",
				LastName:   "Jones",
			},
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetChoosePeopleToNotifyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &choosePeopleToNotifyData{
			App:  testAppData,
			Form: &choosePeopleToNotifyForm{},
		}).
		Return(expectedError)

	err := ChoosePeopleToNotify(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyPeopleLimitReached(t *testing.T) {
	personToNotify := donordata.PersonToNotify{
		FirstNames: "John",
		LastName:   "Doe",
		UID:        actoruid.New(),
	}

	testcases := map[string]struct {
		addedPeople donordata.PeopleToNotify
		expectedUrl page.LpaPath
	}{
		"5 people": {
			addedPeople: donordata.PeopleToNotify{
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
			},
			expectedUrl: page.Paths.ChoosePeopleToNotifySummary,
		},
		"6 people": {
			addedPeople: donordata.PeopleToNotify{
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
			},
			expectedUrl: page.Paths.ChoosePeopleToNotifySummary,
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			err := ChoosePeopleToNotify(nil, nil, testUIDFn)(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaID:          "lpa-id",
				PeopleToNotify: tc.addedPeople,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostChoosePeopleToNotifyPersonDoesNotExists(t *testing.T) {
	testCases := map[string]struct {
		form           url.Values
		personToNotify donordata.PersonToNotify
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
			},
			personToNotify: donordata.PersonToNotify{
				FirstNames: "John",
				LastName:   "Doe",
				UID:        testUID,
			},
		},
		"name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypePersonToNotify, actor.TypeDonor, "Jane", "Doe").String()},
			},
			personToNotify: donordata.PersonToNotify{
				FirstNames: "Jane",
				LastName:   "Doe",
				UID:        testUID,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.DonorProvidedDetails{
					LpaID:          "lpa-id",
					Donor:          donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
					PeopleToNotify: donordata.PeopleToNotify{tc.personToNotify},
					Tasks:          donordata.DonorTasks{PeopleToNotify: actor.TaskInProgress},
				}).
				Return(nil)

			err := ChoosePeopleToNotify(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.ChoosePeopleToNotifyAddress.Format("lpa-id")+"?id="+testUID.String(), resp.Header.Get("Location"))
		})
	}
}

func TestPostChoosePeopleToNotifyPersonExists(t *testing.T) {
	form := url.Values{
		"first-names": {"Johnny"},
		"last-name":   {"Dear"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID: "lpa-id",
			PeopleToNotify: donordata.PeopleToNotify{{
				FirstNames: "Johnny",
				LastName:   "Dear",
				UID:        uid,
			}},
			Tasks: donordata.DonorTasks{PeopleToNotify: actor.TaskInProgress},
		}).
		Return(nil)

	err := ChoosePeopleToNotify(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID: "lpa-id",
		PeopleToNotify: donordata.PeopleToNotify{{
			FirstNames: "John",
			LastName:   "Doe",
			UID:        uid,
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChoosePeopleToNotifyAddress.Format("lpa-id")+"?id="+uid.String(), resp.Header.Get("Location"))
}

func TestPostChoosePeopleToNotifyWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *choosePeopleToNotifyData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Nil(t, data.NameWarning) &&
					assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
		"name warning": {
			form: url.Values{
				"first-names": {"Jane"},
				"last-name":   {"Doe"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypePersonToNotify, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"ignore-name-warning": {"errorDonorMatchesActor|aPersonToNotify|John|Doe"},
			},
			dataMatcher: func(t *testing.T, data *choosePeopleToNotifyData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypePersonToNotify, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, mock.MatchedBy(func(data *choosePeopleToNotifyData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := ChoosePeopleToNotify(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.DonorProvidedDetails{
				Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostChoosePeopleToNotifyWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ChoosePeopleToNotify(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestReadChoosePeopleToNotifyForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names": {"  John "},
		"last-name":   {"Doe"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChoosePeopleToNotifyForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
}

func TestChoosePeopleToNotifyFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *choosePeopleToNotifyForm
		errors validation.List
	}{
		"valid": {
			form: &choosePeopleToNotifyForm{
				FirstNames: "A",
				LastName:   "B",
			},
		},
		"max length": {
			form: &choosePeopleToNotifyForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
			},
		},
		"missing all": {
			form: &choosePeopleToNotifyForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}),
		},
		"too long": {
			form: &choosePeopleToNotifyForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestPersonToNotifyMatches(t *testing.T) {
	uid := actoruid.New()
	donor := &donordata.DonorProvidedDetails{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{UID: uid, FirstNames: "o", LastName: "p"},
		},
	}

	assert.Equal(t, actor.TypeNone, personToNotifyMatches(donor, uid, "x", "y"))
	assert.Equal(t, actor.TypeDonor, personToNotifyMatches(donor, uid, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, personToNotifyMatches(donor, uid, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, personToNotifyMatches(donor, uid, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, personToNotifyMatches(donor, uid, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, personToNotifyMatches(donor, uid, "i", "j"))
	assert.Equal(t, actor.TypeNone, personToNotifyMatches(donor, uid, "k", "L"))
	assert.Equal(t, actor.TypePersonToNotify, personToNotifyMatches(donor, uid, "m", "n"))
	assert.Equal(t, actor.TypeNone, personToNotifyMatches(donor, uid, "o", "p"))
}

func TestPersonToNotifyMatchesEmptyNamesIgnored(t *testing.T) {
	uid := actoruid.New()
	donor := &donordata.DonorProvidedDetails{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "", LastName: ""},
			{UID: uid, FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, personToNotifyMatches(donor, uid, "", ""))
}
