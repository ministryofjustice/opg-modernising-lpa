package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterPersonToNotify(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterPersonToNotifyData{
			App:  testAppData,
			Form: &enterPersonToNotifyForm{},
		}).
		Return(nil)

	err := EnterPersonToNotify(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterPersonToNotifyFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)

	err := EnterPersonToNotify(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.Provided{
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
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetEnterPersonToNotifyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterPersonToNotifyData{
			App:  testAppData,
			Form: &enterPersonToNotifyForm{},
		}).
		Return(expectedError)

	err := EnterPersonToNotify(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterPersonToNotifyPeopleLimitReached(t *testing.T) {
	personToNotify := donordata.PersonToNotify{
		FirstNames: "John",
		LastName:   "Doe",
		UID:        actoruid.New(),
	}

	testcases := map[string]struct {
		addedPeople donordata.PeopleToNotify
		expectedUrl donor.Path
	}{
		"5 people": {
			addedPeople: donordata.PeopleToNotify{
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
				personToNotify,
			},
			expectedUrl: donor.PathChoosePeopleToNotifySummary,
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
			expectedUrl: donor.PathChoosePeopleToNotifySummary,
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			err := EnterPersonToNotify(nil, nil, testUIDFn)(testAppData, w, r, &donordata.Provided{
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

func TestPostEnterPersonToNotifyPersonDoesNotExists(t *testing.T) {
	testCases := map[string]struct {
		form             url.Values
		personToNotify   donordata.PersonToNotify
		expectedRedirect string
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
			expectedRedirect: donor.PathEnterPersonToNotifyAddress.FormatQuery("lpa-id", url.Values{
				"id": {testUID.String()},
			}),
		},
		"with name warning": {
			form: url.Values{
				"first-names": {"Jane"},
				"last-name":   {"Doe"},
			},
			personToNotify: donordata.PersonToNotify{
				FirstNames: "Jane",
				LastName:   "Doe",
				UID:        testUID,
			},
			expectedRedirect: donor.PathWarningInterruption.FormatQuery("lpa-id", url.Values{
				"id":          {testUID.String()},
				"warningFrom": {"/abc"},
				"next": {donor.PathEnterPersonToNotifyAddress.FormatQuery(
					"lpa-id",
					url.Values{"id": {testUID.String()}}),
				},
				"actor": {actor.TypePersonToNotify.String()},
			}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:          "lpa-id",
					Donor:          donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
					PeopleToNotify: donordata.PeopleToNotify{tc.personToNotify},
					Tasks:          donordata.Tasks{PeopleToNotify: task.StateInProgress},
				}).
				Return(nil)

			appData := appcontext.Data{Page: "/abc"}
			err := EnterPersonToNotify(nil, donorStore, testUIDFn)(appData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterPersonToNotifyPersonExists(t *testing.T) {
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
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			PeopleToNotify: donordata.PeopleToNotify{{
				FirstNames: "Johnny",
				LastName:   "Dear",
				UID:        uid,
			}},
			Tasks: donordata.Tasks{PeopleToNotify: task.StateInProgress},
		}).
		Return(nil)

	err := EnterPersonToNotify(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.Provided{
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
	assert.Equal(t, donor.PathEnterPersonToNotifyAddress.Format("lpa-id")+"?id="+uid.String(), resp.Header.Get("Location"))
}

func TestPostEnterPersonToNotifyWhenInputRequired(t *testing.T) {
	form := url.Values{
		"last-name": {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterPersonToNotifyData) bool {
			return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
		})).
		Return(nil)

	err := EnterPersonToNotify(template.Execute, nil, testUIDFn)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyWhenStoreErrors(t *testing.T) {
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

	err := EnterPersonToNotify(nil, donorStore, testUIDFn)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestReadEnterPersonToNotifyForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names": {"  John "},
		"last-name":   {"Doe"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readEnterPersonToNotifyForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
}

func TestEnterPersonToNotifyFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *enterPersonToNotifyForm
		errors validation.List
	}{
		"valid": {
			form: &enterPersonToNotifyForm{
				FirstNames: "A",
				LastName:   "B",
			},
		},
		"max length": {
			form: &enterPersonToNotifyForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
			},
		},
		"missing all": {
			form: &enterPersonToNotifyForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}),
		},
		"too long": {
			form: &enterPersonToNotifyForm{
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
