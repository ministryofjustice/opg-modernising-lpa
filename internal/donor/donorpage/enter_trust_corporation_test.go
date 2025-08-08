package donorpage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterTrustCorporation(t *testing.T) {
	testcases := map[bool]struct {
		provided  *donordata.Provided
		enterPath donor.Path
	}{
		false: {
			provided: &donordata.Provided{
				LpaID: "lpa-id",
				Attorneys: donordata.Attorneys{
					TrustCorporation: donordata.TrustCorporation{Name: "X"},
				},
			},
			enterPath: donor.PathEnterAttorney,
		},
		true: {
			provided: &donordata.Provided{
				LpaID: "lpa-id",
				ReplacementAttorneys: donordata.Attorneys{
					TrustCorporation: donordata.TrustCorporation{Name: "X"},
				},
			},
			enterPath: donor.PathEnterReplacementAttorney,
		},
	}

	for isReplacement, tc := range testcases {
		t.Run(fmt.Sprint(isReplacement), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			service := newMockAttorneyService(t)
			service.EXPECT().
				IsReplacement().
				Return(isReplacement)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &enterTrustCorporationData{
					App: testAppData,
					Form: &enterTrustCorporationForm{
						Name: "X",
					},
					LpaID:               "lpa-id",
					ChooseAttorneysPath: tc.enterPath.FormatQuery("lpa-id", url.Values{"id": {testUID.String()}}),
				}).
				Return(nil)

			err := EnterTrustCorporation(template.Execute, service, testUIDFn)(testAppData, w, r, tc.provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetEnterTrustCorporationWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterTrustCorporation(template.Execute, testAttorneyService(t), testUIDFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporation(t *testing.T) {
	form := url.Values{
		"name":  {"Co co."},
		"email": {"name@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	trustCorporation := donordata.TrustCorporation{
		Name:  "Co co.",
		Email: "name@example.com",
	}

	provided := &donordata.Provided{
		LpaID: "lpa-id",
	}

	service := testAttorneyService(t)
	service.EXPECT().
		PutTrustCorporation(r.Context(), provided, trustCorporation).
		Return(nil)

	err := EnterTrustCorporation(nil, service, testUIDFn)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterTrustCorporationAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterTrustCorporationWhenAddressSet(t *testing.T) {
	form := url.Values{
		"name":  {"Co co."},
		"email": {"name@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	trustCorporation := donordata.TrustCorporation{
		Name:    "Co co.",
		Email:   "name@example.com",
		Address: place.Address{Line1: "123"},
	}

	provided := &donordata.Provided{
		LpaID: "lpa-id",
		Attorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
			Name:    "Other co.",
			Email:   "name@example.com",
			Address: place.Address{Line1: "123"},
		}},
	}

	service := testAttorneyService(t)
	service.EXPECT().
		PutTrustCorporation(r.Context(), provided, trustCorporation).
		Return(nil)

	err := EnterTrustCorporation(nil, service, testUIDFn)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterTrustCorporationAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterTrustCorporationWhenValidationError(t *testing.T) {
	form := url.Values{
		"email": {"name@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterTrustCorporationData) bool {
			return assert.Equal(t, validation.With("name", validation.EnterError{Label: "trustCorporationName"}), data.Errors)
		})).
		Return(nil)

	err := EnterTrustCorporation(template.Execute, testAttorneyService(t), testUIDFn)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationWhenReuseStoreErrors(t *testing.T) {
	form := url.Values{
		"name":  {"Co co."},
		"email": {"name@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := testAttorneyService(t)
	service.EXPECT().
		PutTrustCorporation(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterTrustCorporation(nil, service, testUIDFn)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Attorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
			Name:    "Other co.",
			Email:   "name@example.com",
			Address: place.Address{Line1: "123"},
		}},
	})
	assert.Equal(t, expectedError, err)
}

func TestReadEnterTrustCorporationForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"name":  {"  Yoyodyne "},
		"email": {"contact@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readEnterTrustCorporationForm(r)

	assert.Equal("Yoyodyne", result.Name)
	assert.Equal("contact@example.com", result.Email)
}

func TestEnterTrustCorporationFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form                  *enterTrustCorporationForm
		isReplacement         bool
		otherTrustCorporation donordata.TrustCorporation
		errors                validation.List
	}{
		"valid": {
			form: &enterTrustCorporationForm{
				Name:  "A",
				Email: "a@b.c",
			},
		},
		"missing all": {
			form: &enterTrustCorporationForm{},
			errors: validation.
				With("name", validation.EnterError{Label: "trustCorporationName"}),
		},
		"invalid email": {
			form: &enterTrustCorporationForm{
				Name:  "A",
				Email: "person@",
			},
			errors: validation.With("email", validation.EmailError{Label: "trustCorporationEmailAddress"}),
		},
		"name matches original": {
			form: &enterTrustCorporationForm{
				Name:  "A",
				Email: "person@whatever.com",
			},
			isReplacement:         true,
			otherTrustCorporation: donordata.TrustCorporation{Name: "A"},
			errors:                validation.With("name", trustCorporationCannotAlsoBeError{Name: "A", Replacement: true}),
		},
		"name matches replacement": {
			form: &enterTrustCorporationForm{
				Name:  "A",
				Email: "person@whatever.com",
			},
			otherTrustCorporation: donordata.TrustCorporation{Name: "A"},
			errors:                validation.With("name", trustCorporationCannotAlsoBeError{Name: "A"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(tc.isReplacement, tc.otherTrustCorporation))
		})
	}
}

func TestTrustCorporationCannotAlsoBeError(t *testing.T) {
	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("aReplacementAttorney").
		Return("A")
	localizer.EXPECT().
		T("anOriginalAttorney").
		Return("B")
	localizer.EXPECT().
		Format("errorTrustCorporationCannotAlsoBe", map[string]any{
			"Name":        "X",
			"IsAppointed": "A",
			"CannotBe":    "B",
		}).
		Return("result")

	error := trustCorporationCannotAlsoBeError{Name: "X"}
	assert.Equal(t, "result", error.Format(localizer))
}

func TestTrustCorporationCannotAlsoBeErrorWhenReplacementService(t *testing.T) {
	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("aReplacementAttorney").
		Return("A")
	localizer.EXPECT().
		T("anOriginalAttorney").
		Return("B")
	localizer.EXPECT().
		Format("errorTrustCorporationCannotAlsoBe", map[string]any{
			"Name":        "X",
			"IsAppointed": "B",
			"CannotBe":    "A",
		}).
		Return("result")

	error := trustCorporationCannotAlsoBeError{Name: "X", Replacement: true}
	assert.Equal(t, "result", error.Format(localizer))
}
