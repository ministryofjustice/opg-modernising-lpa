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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterCorrespondentDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterCorrespondentDetailsData{
			App:  testAppData,
			Form: &enterCorrespondentDetailsForm{WantAddress: form.NewYesNoForm(form.YesNoUnknown)},
		}).
		Return(nil)

	err := EnterCorrespondentDetails(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterCorrespondentDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterCorrespondentDetailsData{
			App: testAppData,
			Form: &enterCorrespondentDetailsForm{
				FirstNames:  "John",
				WantAddress: form.NewYesNoForm(form.YesNoUnknown),
			},
		}).
		Return(nil)

	err := EnterCorrespondentDetails(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Correspondent: donordata.Correspondent{
			FirstNames: "John",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterCorrespondentDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterCorrespondentDetails(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentDetails(t *testing.T) {
	f := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"email":               {"email@example.com"},
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	correspondent := donordata.Correspondent{
		FirstNames:  "John",
		LastName:    "Doe",
		Email:       "email@example.com",
		WantAddress: form.No,
	}

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:         "lpa-id",
			LpaUID:        "lpa-uid",
			Donor:         donordata.Donor{FirstNames: "John", LastName: "Smith"},
			Correspondent: correspondent,
		}).
		Return(nil)

	err := EnterCorrespondentDetails(nil, service)(testAppData, w, r, &donordata.Provided{
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
		Donor:  donordata.Donor{FirstNames: "John", LastName: "Smith"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCorrespondentSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterCorrespondentDetailsWhenWantsAddress(t *testing.T) {
	f := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"email":               {"email@example.com"},
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?from=/what", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	correspondent := donordata.Correspondent{
		FirstNames:  "John",
		LastName:    "Doe",
		Email:       "email@example.com",
		WantAddress: form.Yes,
	}

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:         "lpa-id",
			Donor:         donordata.Donor{FirstNames: "John", LastName: "Smith"},
			Correspondent: correspondent,
		}).
		Return(nil)

	err := EnterCorrespondentDetails(nil, service)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "John", LastName: "Smith"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterCorrespondentAddress.FormatQuery("lpa-id", url.Values{"from": {"/what"}}), resp.Header.Get("Location"))
}

func TestPostEnterCorrespondentDetailsWhenNameMatchesDonor(t *testing.T) {
	testcases := map[form.YesNo]donor.Path{
		form.Yes: donor.PathEnterCorrespondentAddress,
		form.No:  donor.PathCorrespondentSummary,
	}

	for wantAddress, redirect := range testcases {
		t.Run(wantAddress.String(), func(t *testing.T) {
			f := url.Values{
				"first-names":         {"John"},
				"last-name":           {"Smith"},
				"email":               {"email@example.com"},
				form.FieldNames.YesNo: {wantAddress.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?from=/what", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			correspondent := donordata.Correspondent{
				FirstNames:  "John",
				LastName:    "Smith",
				Email:       "email@example.com",
				WantAddress: wantAddress,
			}

			service := newMockCorrespondentService(t)
			service.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:         "lpa-id",
					Donor:         donordata.Donor{FirstNames: "John", LastName: "Smith"},
					Correspondent: correspondent,
				}).
				Return(nil)

			appData := appcontext.Data{Page: "/abc"}

			err := EnterCorrespondentDetails(nil, service)(appData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "John", LastName: "Smith"},
			})
			resp := w.Result()

			expectedRedirect := donor.PathWarningInterruption.FormatQuery("lpa-id", url.Values{
				"next":        {redirect.Format("lpa-id")},
				"warningFrom": {"/abc"},
				"actor":       {actor.TypeCorrespondent.String()},
			})

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterCorrespondentDetailsWhenValidationError(t *testing.T) {
	form := url.Values{
		"last-name":           {"Doe"},
		"email":               {"email@example.com"},
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterCorrespondentDetailsData) bool {
			return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
		})).
		Return(nil)

	err := EnterCorrespondentDetails(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{
			FirstNames: "John",
			LastName:   "Doe",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentDetailsWhenServiceErrors(t *testing.T) {
	form := url.Values{
		"first-names":         {"John"},
		"last-name":           {"Doe"},
		"email":               {"email@example.com"},
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := EnterCorrespondentDetails(nil, service)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
	})

	assert.Equal(t, expectedError, err)
}

func TestReadEnterCorrespondentDetailsForm(t *testing.T) {
	assert := assert.New(t)

	f := url.Values{
		"first-names": {"  John "},
		"last-name":   {"Doe"},
		"email":       {"email@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readEnterCorrespondentDetailsForm(r, donordata.Donor{FirstNames: "Dave", LastName: "Smith", Email: "email@example.com"})

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
	assert.Equal("email@example.com", result.Email)
	assert.True(result.DonorEmailMatch)
	assert.Equal("Dave Smith", result.DonorFullName)
}

func TestEnterCorrespondentDetailsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *enterCorrespondentDetailsForm
		errors validation.List
	}{
		"valid": {
			form: &enterCorrespondentDetailsForm{
				FirstNames:  "A",
				LastName:    "B",
				Email:       "email@example.com",
				WantAddress: form.NewYesNoForm(form.Yes),
			},
		},
		"max length": {
			form: &enterCorrespondentDetailsForm{
				FirstNames:  strings.Repeat("x", 53),
				LastName:    strings.Repeat("x", 61),
				Email:       "email@example.com",
				WantAddress: form.NewYesNoForm(form.Yes),
			},
		},
		"missing all": {
			form: &enterCorrespondentDetailsForm{
				WantAddress: form.NewYesNoForm(form.YesNoUnknown),
			},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("email", validation.EnterError{Label: "email"}),
		},
		"too long": {
			form: &enterCorrespondentDetailsForm{
				FirstNames:  strings.Repeat("x", 54),
				LastName:    strings.Repeat("x", 62),
				Email:       "email@example.com",
				WantAddress: form.NewYesNoForm(form.Yes),
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
		"invalid contact": {
			form: &enterCorrespondentDetailsForm{
				FirstNames:  "A",
				LastName:    "B",
				Email:       "email",
				Phone:       "phone",
				WantAddress: form.NewYesNoForm(form.Yes),
			},
			errors: validation.
				With("email", validation.EmailError{Label: "email"}).
				With("phone", validation.PhoneError{Tmpl: "errorPhone", Label: "phoneNumber"}),
		},
		"matching donor email": {
			form: &enterCorrespondentDetailsForm{
				FirstNames:      "A",
				LastName:        "B",
				Email:           "email@example.com",
				WantAddress:     form.NewYesNoForm(form.No),
				DonorEmailMatch: true,
				DonorFullName:   "Other Person",
			},
			errors: validation.
				With("email", validation.CustomFormattedError{Label: "youProvidedThisEmailForDonorError", Data: map[string]any{"DonorFullName": "Other Person"}}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
