package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveAttorney(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	attorney := donordata.Attorney{
		UID:        uid,
		FirstNames: "John",
		LastName:   "Smith",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &removeAttorneyData{
			App:        testAppData,
			TitleLabel: "removeAnAttorney",
			Name:       "John Smith",
			Form:       form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := RemoveAttorney(template.Execute, testAttorneyService(t))(testAppData, w, r, &donordata.Provided{Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemoveAttorneyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	template := newMockTemplate(t)

	attorney := donordata.Attorney{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	err := RemoveAttorney(template.Execute, testAttorneyService(t))(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveAttorney(t *testing.T) {
	attorneyWithEmail := donordata.Attorney{UID: actoruid.New(), Email: "a"}
	attorneyWithAddress := donordata.Attorney{UID: actoruid.New(), Address: place.Address{Line1: "1 Road way"}}
	attorneyWithoutAddress := donordata.Attorney{UID: actoruid.New()}

	testcases := map[string]struct {
		donor    *donordata.Provided
		redirect donor.Path
	}{
		"many left": {
			donor: &donordata.Provided{
				LpaID:             "lpa-id",
				Attorneys:         donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithEmail, attorneyWithAddress, attorneyWithoutAddress}},
				AttorneyDecisions: donordata.AttorneyDecisions{How: lpadata.Jointly},
			},
			redirect: donor.PathChooseAttorneysSummary,
		},
		"one left": {
			donor: &donordata.Provided{
				LpaID:             "lpa-id",
				Attorneys:         donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithAddress, attorneyWithoutAddress}},
				AttorneyDecisions: donordata.AttorneyDecisions{How: lpadata.Jointly},
			},
			redirect: donor.PathChooseAttorneysSummary,
		},
		"none left": {
			donor:    &donordata.Provided{LpaID: "lpa-id", Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress}}},
			redirect: donor.PathChooseAttorneysSummary,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {form.Yes.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id="+attorneyWithoutAddress.UID.String(), strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			service := testAttorneyService(t)
			service.EXPECT().
				Delete(r.Context(), tc.donor, attorneyWithoutAddress).
				Return(nil)

			err := RemoveAttorney(nil, service)(testAppData, w, r, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostRemoveAttorneyWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	attorneyWithAddress := donordata.Attorney{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := donordata.Attorney{
		UID:     actoruid.New(),
		Address: place.Address{},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+attorneyWithoutAddress.UID.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := RemoveAttorney(nil, testAttorneyService(t))(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveAttorneyWhenServiceErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	attorneyWithAddress := donordata.Attorney{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := donordata.Attorney{
		UID:     actoruid.New(),
		Address: place.Address{},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+attorneyWithoutAddress.UID.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := testAttorneyService(t)
	service.EXPECT().
		Delete(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveAttorney(nil, service)(testAppData, w, r, &donordata.Provided{Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})
	assert.ErrorIs(t, err, expectedError)
}

func TestRemoveAttorneyFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToRemoveAttorney"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveAttorney(template.Execute, testAttorneyService(t))(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{
			Attorneys: []donordata.Attorney{{
				UID:     uid,
				Address: place.Address{},
			}},
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
