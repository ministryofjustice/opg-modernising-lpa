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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveReplacementAttorney(t *testing.T) {
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
			TitleLabel: "doYouWantToRemoveReplacementAttorney",
			Name:       "John Smith",
			Form:       form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := RemoveReplacementAttorney(template.Execute, nil)(testAppData, w, r, &donordata.Provided{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemoveReplacementAttorneyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	attorney := donordata.Attorney{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	err := RemoveReplacementAttorney(nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveReplacementAttorney(t *testing.T) {
	attorneyWithEmail := donordata.Attorney{UID: actoruid.New(), Email: "a"}
	attorneyWithAddress := donordata.Attorney{UID: actoruid.New(), Address: place.Address{Line1: "1 Road way"}}
	attorneyWithoutAddress := donordata.Attorney{UID: actoruid.New()}

	testcases := map[string]struct {
		donor        *donordata.Provided
		updatedDonor *donordata.Provided
		redirect     page.LpaPath
	}{
		"many left": {
			donor: &donordata.Provided{
				LpaID:                        "lpa-id",
				ReplacementAttorneys:         donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithEmail, attorneyWithAddress, attorneyWithoutAddress}},
				ReplacementAttorneyDecisions: donordata.AttorneyDecisions{How: donordata.Jointly},
			},
			updatedDonor: &donordata.Provided{
				LpaID:                        "lpa-id",
				ReplacementAttorneys:         donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithEmail, attorneyWithAddress}},
				ReplacementAttorneyDecisions: donordata.AttorneyDecisions{How: donordata.Jointly},
				Tasks:                        donordata.Tasks{ChooseReplacementAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"one left": {
			donor: &donordata.Provided{
				LpaID:                        "lpa-id",
				ReplacementAttorneys:         donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithAddress, attorneyWithoutAddress}},
				ReplacementAttorneyDecisions: donordata.AttorneyDecisions{How: donordata.Jointly},
			},
			updatedDonor: &donordata.Provided{
				LpaID:                "lpa-id",
				ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithAddress}},
				Tasks:                donordata.Tasks{ChooseReplacementAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"none left": {
			donor: &donordata.Provided{LpaID: "lpa-id", ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress}}},
			updatedDonor: &donordata.Provided{
				LpaID:                "lpa-id",
				ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{}},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
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

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.updatedDonor).
				Return(nil)

			err := RemoveReplacementAttorney(nil, donorStore)(testAppData, w, r, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostRemoveReplacementAttorneyWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyWithAddress := donordata.Attorney{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := donordata.Attorney{
		UID:     uid,
		Address: place.Address{},
	}

	err := RemoveReplacementAttorney(nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveReplacementAttorneyErrorOnPutStore(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyWithAddress := donordata.Attorney{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := donordata.Attorney{
		UID:     uid,
		Address: place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := RemoveReplacementAttorney(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		WantReplacementAttorneys: form.Yes,
		ReplacementAttorneys:     donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress, attorneyWithAddress}},
	})

	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveReplacementAttorneyFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyWithoutAddress := donordata.Attorney{
		UID:     uid,
		Address: place.Address{},
	}

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToRemoveReplacementAttorney"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveReplacementAttorney(template.Execute, nil)(testAppData, w, r, &donordata.Provided{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
