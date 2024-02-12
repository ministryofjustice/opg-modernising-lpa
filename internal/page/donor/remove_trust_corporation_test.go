package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveTrustCorporation(t *testing.T) {
	trustCorporation := actor.TrustCorporation{Name: "hey ltd"}

	testcases := map[string]struct {
		isReplacement bool
		titleLabel    string
		donor         *actor.DonorProvidedDetails
	}{
		"attorney": {
			titleLabel: "removeTrustCorporation",
			donor:      &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{TrustCorporation: trustCorporation}},
		},
		"replacement": {
			isReplacement: true,
			titleLabel:    "removeReplacementTrustCorporation",
			donor:         &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{TrustCorporation: trustCorporation}},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &removeAttorneyData{
					App:        testAppData,
					TitleLabel: tc.titleLabel,
					Name:       "hey ltd",
					Form:       form.NewYesNoForm(form.YesNoUnknown),
				}).
				Return(nil)

			err := RemoveTrustCorporation(template.Execute, nil, tc.isReplacement)(testAppData, w, r, tc.donor)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostRemoveTrustCorporation(t *testing.T) {
	attorney := actor.Attorney{UID: actoruid.New(), Email: "a"}
	trustCorporation := actor.TrustCorporation{Name: "a"}

	testcases := map[string]struct {
		isReplacement bool
		donor         *actor.DonorProvidedDetails
		updatedDonor  *actor.DonorProvidedDetails
		redirect      page.LpaPath
	}{
		"many left": {
			donor: &actor.DonorProvidedDetails{
				LpaID:             "lpa-id",
				Attorneys:         actor.Attorneys{TrustCorporation: trustCorporation, Attorneys: []actor.Attorney{attorney, attorney}},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:             "lpa-id",
				Attorneys:         actor.Attorneys{Attorneys: []actor.Attorney{attorney, attorney}},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
				Tasks:             actor.DonorTasks{ChooseAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
		"replacement many left": {
			isReplacement: true,
			donor: &actor.DonorProvidedDetails{
				LpaID:                        "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{TrustCorporation: trustCorporation, Attorneys: []actor.Attorney{attorney, attorney}},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:                        "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{Attorneys: []actor.Attorney{attorney, attorney}},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
				Tasks:                        actor.DonorTasks{ChooseReplacementAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"one left": {
			donor: &actor.DonorProvidedDetails{
				LpaID:             "lpa-id",
				Attorneys:         actor.Attorneys{TrustCorporation: trustCorporation, Attorneys: []actor.Attorney{attorney}},
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:     "lpa-id",
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorney}},
				Tasks:     actor.DonorTasks{ChooseAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
		"replacement one left": {
			isReplacement: true,
			donor: &actor.DonorProvidedDetails{
				LpaID:                        "lpa-id",
				ReplacementAttorneys:         actor.Attorneys{TrustCorporation: trustCorporation, Attorneys: []actor.Attorney{attorney}},
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:                "lpa-id",
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorney}},
				Tasks:                actor.DonorTasks{ChooseReplacementAttorneys: actor.TaskInProgress},
			},
			redirect: page.Paths.ChooseReplacementAttorneysSummary,
		},
		"none left": {
			donor: &actor.DonorProvidedDetails{LpaID: "lpa-id", Attorneys: actor.Attorneys{TrustCorporation: trustCorporation}},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:     "lpa-id",
				Attorneys: actor.Attorneys{},
			},
			redirect: page.Paths.ChooseAttorneysSummary,
		},
		"replacement none left": {
			isReplacement: true,
			donor:         &actor.DonorProvidedDetails{LpaID: "lpa-id", ReplacementAttorneys: actor.Attorneys{TrustCorporation: trustCorporation}},
			updatedDonor: &actor.DonorProvidedDetails{
				LpaID:                "lpa-id",
				ReplacementAttorneys: actor.Attorneys{},
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
			r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.updatedDonor).
				Return(nil)

			err := RemoveTrustCorporation(template.Execute, donorStore, tc.isReplacement)(testAppData, w, r, tc.donor)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostRemoveTrustCorporationWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)

	attorneyWithAddress := actor.Attorney{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := actor.Attorney{
		UID:     uid,
		Address: place.Address{},
	}

	err := RemoveTrustCorporation(template.Execute, nil, false)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveTrustCorporationErrorOnPutStore(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)

	attorneyWithAddress := actor.Attorney{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	attorneyWithoutAddress := actor.Attorney{
		UID:     uid,
		Address: place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := RemoveTrustCorporation(template.Execute, donorStore, false)(testAppData, w, r, &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveTrustCorporationFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyWithoutAddress := actor.Attorney{
		UID:     uid,
		Address: place.Address{},
	}

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToRemoveTrustCorporation"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveTrustCorporation(template.Execute, nil, false)(testAppData, w, r, &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorneyWithoutAddress}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
