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

func TestGetRemoveTrustCorporation(t *testing.T) {
	trustCorporation := donordata.TrustCorporation{Name: "hey ltd"}

	testcases := map[string]struct {
		isReplacement bool
		titleLabel    string
		donor         *donordata.Provided
	}{
		"attorney": {
			titleLabel: "removeTrustCorporation",
			donor:      &donordata.Provided{Attorneys: donordata.Attorneys{TrustCorporation: trustCorporation}},
		},
		"replacement": {
			isReplacement: true,
			titleLabel:    "removeReplacementTrustCorporation",
			donor:         &donordata.Provided{ReplacementAttorneys: donordata.Attorneys{TrustCorporation: trustCorporation}},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

			service := newMockAttorneyService(t)
			service.EXPECT().
				IsReplacement().
				Return(tc.isReplacement)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &removeAttorneyData{
					App:        testAppData,
					TitleLabel: tc.titleLabel,
					Name:       "hey ltd",
					Form:       form.NewYesNoForm(form.YesNoUnknown),
				}).
				Return(nil)

			err := RemoveTrustCorporation(template.Execute, service)(testAppData, w, r, tc.donor)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostRemoveTrustCorporation(t *testing.T) {
	attorney := donordata.Attorney{UID: actoruid.New(), Email: "a"}
	trustCorporation := donordata.TrustCorporation{Name: "a"}

	testcases := map[string]struct {
		isReplacement bool
		donor         *donordata.Provided
		redirect      donor.Path
	}{
		"many left": {
			donor: &donordata.Provided{
				LpaID:             "lpa-id",
				Attorneys:         donordata.Attorneys{TrustCorporation: trustCorporation, Attorneys: []donordata.Attorney{attorney, attorney}},
				AttorneyDecisions: donordata.AttorneyDecisions{How: lpadata.Jointly},
			},
			redirect: donor.PathChooseAttorneysSummary,
		},
		"replacement many left": {
			isReplacement: true,
			donor: &donordata.Provided{
				LpaID:                        "lpa-id",
				Attorneys:                    donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}},
				ReplacementAttorneys:         donordata.Attorneys{TrustCorporation: trustCorporation, Attorneys: []donordata.Attorney{attorney, attorney}},
				ReplacementAttorneyDecisions: donordata.AttorneyDecisions{How: lpadata.Jointly},
			},
			redirect: donor.PathChooseReplacementAttorneysSummary,
		},
		"one left": {
			donor: &donordata.Provided{
				LpaID:             "lpa-id",
				Attorneys:         donordata.Attorneys{TrustCorporation: trustCorporation, Attorneys: []donordata.Attorney{attorney}},
				AttorneyDecisions: donordata.AttorneyDecisions{How: lpadata.Jointly},
			},
			redirect: donor.PathChooseAttorneysSummary,
		},
		"replacement one left": {
			isReplacement: true,
			donor: &donordata.Provided{
				LpaID:                        "lpa-id",
				ReplacementAttorneys:         donordata.Attorneys{TrustCorporation: trustCorporation, Attorneys: []donordata.Attorney{attorney}},
				ReplacementAttorneyDecisions: donordata.AttorneyDecisions{How: lpadata.Jointly},
			},
			redirect: donor.PathChooseReplacementAttorneysSummary,
		},
		"none left": {
			donor:    &donordata.Provided{LpaID: "lpa-id", Attorneys: donordata.Attorneys{TrustCorporation: trustCorporation}},
			redirect: donor.PathChooseAttorneysSummary,
		},
		"replacement none left": {
			isReplacement: true,
			donor:         &donordata.Provided{LpaID: "lpa-id", ReplacementAttorneys: donordata.Attorneys{TrustCorporation: trustCorporation}},
			redirect:      donor.PathChooseReplacementAttorneysSummary,
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

			service := newMockAttorneyService(t)
			service.EXPECT().
				IsReplacement().
				Return(tc.isReplacement)
			service.EXPECT().
				DeleteTrustCorporation(r.Context(), tc.donor).
				Return(nil)

			err := RemoveTrustCorporation(nil, service)(testAppData, w, r, tc.donor)

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

	err := RemoveTrustCorporation(template.Execute, testAttorneyService(t))(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveTrustCorporationWhenServiceErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)

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

	service := testAttorneyService(t)
	service.EXPECT().
		DeleteTrustCorporation(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveTrustCorporation(template.Execute, service)(testAppData, w, r, &donordata.Provided{Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress, attorneyWithAddress}}})
	assert.Equal(t, expectedError, err)
}

func TestRemoveTrustCorporationFormValidation(t *testing.T) {
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

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToRemoveTrustCorporation"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removeAttorneyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveTrustCorporation(template.Execute, testAttorneyService(t))(testAppData, w, r, &donordata.Provided{Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorneyWithoutAddress}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
