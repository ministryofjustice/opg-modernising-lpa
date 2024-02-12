package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourDetailsData{
			App:               testAppData,
			Form:              &yourDetailsForm{},
			YesNoMaybeOptions: actor.YesNoMaybeValues,
		}).
		Return(nil)

	err := YourDetails(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourDetailsData{
			App: testAppData,
			Form: &yourDetailsForm{
				FirstNames: "John",
			},
			YesNoMaybeOptions: actor.YesNoMaybeValues,
		}).
		Return(nil)

	err := YourDetails(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{
			FirstNames: "John",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourDetailsDobWarningIsAlwaysShown(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourDetailsData{
			App: testAppData,
			Form: &yourDetailsForm{
				FirstNames: "John",
				LastName:   "Doe",
				OtherNames: "J",
				Dob:        date.New("1900", "01", "02"),
				CanSign:    actor.Yes,
			},
			DobWarning:        "dateOfBirthIsOver100",
			YesNoMaybeOptions: actor.YesNoMaybeValues,
		}).
		Return(nil)

	err := YourDetails(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{
			FirstNames:    "John",
			LastName:      "Doe",
			OtherNames:    "J",
			DateOfBirth:   date.New("1900", "01", "02"),
			ThinksCanSign: actor.Yes,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := YourDetails(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Donor: actor.Donor{FirstNames: "John"}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourDetails(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form     url.Values
		person   actor.Donor
		redirect page.LpaPath
	}{
		"valid": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
				"can-sign":            {actor.Yes.String()},
			},
			person: actor.Donor{
				FirstNames:    "John",
				LastName:      "Doe",
				DateOfBirth:   date.New(validBirthYear, "1", "2"),
				Address:       place.Address{Line1: "abc"},
				Email:         "name@example.com",
				ThinksCanSign: actor.Yes,
				CanSign:       form.Yes,
			},
			redirect: page.Paths.YourAddress,
		},
		"warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
				"can-sign":            {actor.Yes.String()},
			},
			person: actor.Donor{
				FirstNames:    "John",
				LastName:      "Doe",
				DateOfBirth:   date.New("1900", "1", "2"),
				Address:       place.Address{Line1: "abc"},
				Email:         "name@example.com",
				ThinksCanSign: actor.Yes,
				CanSign:       form.Yes,
			},
			redirect: page.Paths.YourAddress,
		},
		"cannot sign": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
				"can-sign":            {actor.No.String()},
			},
			person: actor.Donor{
				FirstNames:    "John",
				LastName:      "Doe",
				DateOfBirth:   date.New(validBirthYear, "1", "2"),
				Address:       place.Address{Line1: "abc"},
				Email:         "name@example.com",
				ThinksCanSign: actor.No,
			},
			redirect: page.Paths.CheckYouCanSign,
		},
		"maybe can sign": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
				"can-sign":            {actor.Maybe.String()},
			},
			person: actor.Donor{
				FirstNames:    "John",
				LastName:      "Doe",
				DateOfBirth:   date.New(validBirthYear, "1", "2"),
				Address:       place.Address{Line1: "abc"},
				Email:         "name@example.com",
				ThinksCanSign: actor.Maybe,
			},
			redirect: page.Paths.CheckYouCanSign,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID: "lpa-id",
					Donor: tc.person,
					Tasks: actor.DonorTasks{YourDetails: actor.TaskInProgress},
				}).
				Return(nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Get(r, "session").
				Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "xyz", Email: "name@example.com"}}}, nil)

			err := YourDetails(nil, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{
					FirstNames: "John",
					Address:    place.Address{Line1: "abc"},
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

func TestPostYourDetailsWhenDetailsNotChanged(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)
	f := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {validBirthYear},
		"can-sign":            {actor.Yes.String()},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			Donor: actor.Donor{
				FirstNames:    "John",
				LastName:      "Doe",
				DateOfBirth:   date.New(validBirthYear, "1", "2"),
				Email:         "name@example.com",
				ThinksCanSign: actor.Yes,
				CanSign:       form.Yes,
			},
			Tasks:                          actor.DonorTasks{YourDetails: actor.TaskInProgress},
			HasSentApplicationUpdatedEvent: true,
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "xyz", Email: "name@example.com"}}}, nil)

	err := YourDetails(nil, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Doe",
			DateOfBirth: date.New(validBirthYear, "1", "2"),
		},
		HasSentApplicationUpdatedEvent: true,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YourAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourDetailsWhenTaskCompleted(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	f := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {validBirthYear},
		"can-sign":            {actor.Yes.String()},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			Donor: actor.Donor{
				FirstNames:    "John",
				LastName:      "Doe",
				DateOfBirth:   date.New(validBirthYear, "1", "2"),
				Address:       place.Address{Line1: "abc"},
				Email:         "name@example.com",
				ThinksCanSign: actor.Yes,
				CanSign:       form.Yes,
			},
			Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted},
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "xyz", Email: "name@example.com"}}}, nil)

	err := YourDetails(nil, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Donor: actor.Donor{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
		Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YourAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourDetailsWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourDetailsData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
				"can-sign":            {actor.Yes.String()},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
		"dob warning": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"can-sign":            {actor.Yes.String()},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"dob warning ignored but other errors": {
			form: url.Values{
				"first-names":         {"John"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
				"can-sign":            {actor.Yes.String()},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"other dob warning ignored": {
			form: url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsUnder18"},
				"can-sign":            {actor.Yes.String()},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
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
				Execute(w, mock.MatchedBy(func(data *yourDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Get(mock.Anything, "session").
				Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "xyz", Email: "name@example.com"}}}, nil)

			err := YourDetails(template.Execute, nil, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostYourDetailsNameWarningOnlyShownWhenDonorAndFormNamesAreDifferent(t *testing.T) {
	f := url.Values{
		"first-names":         {"Jane"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1999"},
		"can-sign":            {actor.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			Donor: actor.Donor{
				FirstNames:    "Jane",
				LastName:      "Doe",
				DateOfBirth:   date.New("1999", "1", "2"),
				Email:         "name@example.com",
				ThinksCanSign: actor.Yes,
				CanSign:       form.Yes,
			},
			Tasks: actor.DonorTasks{YourDetails: actor.TaskInProgress},
			ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
				{FirstNames: "Jane", LastName: "Doe", UID: uid, Address: place.Address{Line1: "abc"}},
			}},
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(mock.Anything, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "xyz", Email: "name@example.com"}}}, nil)

	err := YourDetails(nil, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "Jane", LastName: "Doe", UID: uid, Address: place.Address{Line1: "abc"}},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YourAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourDetailsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
		"can-sign":            {actor.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(mock.Anything, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "xyz", Email: "name@example.com"}}}, nil)

	err := YourDetails(nil, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
	})

	assert.Equal(t, expectedError, err)
}

func TestPostYourDetailsWhenSessionProblem(t *testing.T) {
	testCases := map[string]struct {
		session *sessions.Session
		error   error
	}{
		"store error": {
			session: &sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "xyz", Email: "name@example.com"}}},
			error:   expectedError,
		},
		"missing donor session": {
			session: &sessions.Session{},
			error:   nil,
		},
		"missing email": {
			session: &sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "xyz"}}},
			error:   nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"first-names":         {"John"},
				"last-name":           {"Doe"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1990"},
				"can-sign":            {actor.Yes.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Get(mock.Anything, "session").
				Return(tc.session, tc.error)

			err := YourDetails(nil, nil, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

			assert.NotNil(t, err)
		})
	}
}

func TestReadYourDetailsForm(t *testing.T) {
	assert := assert.New(t)

	f := url.Values{
		"first-names":         {"  John "},
		"last-name":           {"Doe"},
		"other-names":         {"Somebody"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
		"can-sign":            {actor.Yes.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourDetailsForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("Somebody", result.OtherNames)
	assert.Equal(date.New("1990", "1", "2"), result.Dob)
	assert.Equal(actor.Yes, result.CanSign)
	assert.Nil(result.CanSignError)
}

func TestYourDetailsFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *yourDetailsForm
		errors validation.List
	}{
		"valid": {
			form: &yourDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        validDob,
			},
		},
		"max length": {
			form: &yourDetailsForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
				OtherNames: strings.Repeat("x", 50),
				Dob:        validDob,
			},
		},
		"missing all": {
			form: &yourDetailsForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}),
		},
		"too long": {
			form: &yourDetailsForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				OtherNames: strings.Repeat("x", 51),
				Dob:        validDob,
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}).
				With("other-names", validation.StringTooLongError{Label: "otherNamesYouAreKnownBy", Length: 50}),
		},
		"future dob": {
			form: &yourDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}),
		},
		"invalid dob": {
			form: &yourDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid missing dob": {
			form: &yourDetailsForm{
				FirstNames: "A",
				LastName:   "B",
				Dob:        date.New("1", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingMonth: true}),
		},
		"invalid can sign": {
			form: &yourDetailsForm{
				FirstNames:   "A",
				LastName:     "B",
				Dob:          validDob,
				CanSignError: expectedError,
			},
			errors: validation.With("can-sign", validation.SelectError{Label: "yesIfCanSign"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestYourDetailsFormDobWarning(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form    *yourDetailsForm
		warning string
	}{
		"valid": {
			form: &yourDetailsForm{
				Dob: validDob,
			},
		},
		"future-dob": {
			form: &yourDetailsForm{
				Dob: now.AddDate(0, 0, 1),
			},
		},
		"dob-is-18": {
			form: &yourDetailsForm{
				Dob: now.AddDate(-18, 0, 0),
			},
		},
		"dob-under-18": {
			form: &yourDetailsForm{
				Dob: now.AddDate(-18, 0, 1),
			},
			warning: "dateOfBirthIsUnder18",
		},
		"dob-is-100": {
			form: &yourDetailsForm{
				Dob: now.AddDate(-100, 0, 0),
			},
		},
		"dob-over-100": {
			form: &yourDetailsForm{
				Dob: now.AddDate(-100, 0, -1),
			},
			warning: "dateOfBirthIsOver100",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.warning, tc.form.DobWarning())
		})
	}
}

func TestDonorMatches(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
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
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, donorMatches(donor, "", ""))
}
