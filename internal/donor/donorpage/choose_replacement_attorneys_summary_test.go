package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseReplacementAttorneysSummary(t *testing.T) {
	testcases := map[string]donordata.Attorneys{
		"attorneys":         {Attorneys: []donordata.Attorney{{}}},
		"trust corporation": {TrustCorporation: donordata.TrustCorporation{Name: "a"}},
	}

	for name, attorneys := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			donor := &donordata.Provided{ReplacementAttorneys: attorneys}

			service := newMockAttorneyService(t)
			service.EXPECT().
				Reusable(r.Context(), donor).
				Return([]donordata.Attorney{}, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &chooseReplacementAttorneysSummaryData{
					App:     testAppData,
					Donor:   donor,
					Options: donordata.YesNoMaybeValues,
				}).
				Return(nil)

			err := ChooseReplacementAttorneysSummary(template.Execute, service, nil)(testAppData, w, r, donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChooseReplacementAttorneysSummaryWhenReuseStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockAttorneyService(t)
	service.EXPECT().
		Reusable(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := ChooseReplacementAttorneysSummary(nil, service, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
	})
	assert.Equal(t, expectedError, err)
}

func TestGetChooseReplacementAttorneysSummaryWhenNoReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneysSummary(nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Tasks: donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathDoYouWantReplacementAttorneys.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysSummaryAddAttorney(t *testing.T) {
	testcases := map[donordata.YesNoMaybe]string{
		donordata.Yes:   donor.PathEnterReplacementAttorney.Format("lpa-id") + "?addAnother=1&id=" + testUID.String(),
		donordata.Maybe: donor.PathChooseReplacementAttorneys.Format("lpa-id"),
	}

	for option, redirect := range testcases {
		t.Run(option.String(), func(t *testing.T) {
			f := url.Values{
				"option": {option.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			provided := &donordata.Provided{LpaID: "lpa-id", ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}}}

			service := newMockAttorneyService(t)
			service.EXPECT().
				Reusable(r.Context(), provided).
				Return(nil, nil)

			err := ChooseReplacementAttorneysSummary(nil, service, testUIDFn)(testAppData, w, r, provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseReplacementAttorneysSummaryDoNotAddAttorney(t *testing.T) {
	attorney1 := donordata.Attorney{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}
	attorney2 := donordata.Attorney{FirstNames: "x", LastName: "y", Address: place.Address{Line1: "z"}, DateOfBirth: date.New("2000", "1", "1")}

	testcases := map[string]struct {
		redirectUrl          donor.Path
		attorneys            donordata.Attorneys
		replacementAttorneys donordata.Attorneys
		howAttorneysAct      lpadata.AttorneysAct
		decisionDetails      string
	}{
		"with multiple attorneys acting jointly and severally and single replacement attorney": {
			redirectUrl:          donor.PathHowShouldReplacementAttorneysStepIn,
			attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			replacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1}},
			howAttorneysAct:      lpadata.JointlyAndSeverally,
		},
		"with multiple attorneys acting jointly and severally and multiple replacement attorney": {
			redirectUrl:          donor.PathHowShouldReplacementAttorneysStepIn,
			attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			replacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			howAttorneysAct:      lpadata.JointlyAndSeverally,
		},
		"with multiple attorneys acting jointly and multiple replacement attorneys": {
			redirectUrl:          donor.PathHowShouldReplacementAttorneysMakeDecisions,
			attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			replacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			howAttorneysAct:      lpadata.Jointly,
		},
		"with multiple attorneys acting jointly for some decisions and jointly and severally for other decisions and single replacement attorney": {
			redirectUrl:          donor.PathTaskList,
			attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			replacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1}},
			howAttorneysAct:      lpadata.JointlyForSomeSeverallyForOthers,
			decisionDetails:      "some words",
		},
		"with multiple attorneys acting jointly for some decisions, and jointly and severally for other decisions and multiple replacement attorneys": {
			redirectUrl:          donor.PathTaskList,
			attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			replacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			howAttorneysAct:      lpadata.JointlyForSomeSeverallyForOthers,
			decisionDetails:      "some words",
		},
		"with multiple attorneys acting jointly and single replacement attorneys": {
			redirectUrl:          donor.PathTaskList,
			attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1, attorney2}},
			replacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney1}},
			howAttorneysAct:      lpadata.Jointly,
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			f := url.Values{
				"option": {donordata.No.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			service := newMockAttorneyService(t)
			service.EXPECT().
				Reusable(mock.Anything, mock.Anything).
				Return(nil, nil)

			err := ChooseReplacementAttorneysSummary(nil, service, nil)(testAppData, w, r, &donordata.Provided{
				LpaID:                "lpa-id",
				ReplacementAttorneys: tc.replacementAttorneys,
				AttorneyDecisions: donordata.AttorneyDecisions{
					How:     tc.howAttorneysAct,
					Details: tc.decisionDetails,
				},
				Attorneys: tc.attorneys,
				Tasks: donordata.Tasks{
					YourDetails:     task.StateCompleted,
					ChooseAttorneys: task.StateCompleted,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirectUrl.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseReplacementAttorneySummaryFormValidation(t *testing.T) {
	f := url.Values{
		"option": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("option", validation.SelectError{Label: "yesToAddAnotherReplacementAttorney"})

	service := newMockAttorneyService(t)
	service.EXPECT().
		Reusable(mock.Anything, mock.Anything).
		Return(nil, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *chooseReplacementAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseReplacementAttorneysSummary(template.Execute, service, nil)(testAppData, w, r, &donordata.Provided{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
