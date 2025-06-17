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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseAttorneysSummary(t *testing.T) {
	testcases := map[string]*donordata.Provided{
		"attorney": {
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
		},
		"trust corporation": {
			Attorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{Name: "a"}},
		},
	}

	for name, donor := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			service := newMockAttorneyService(t)
			service.EXPECT().
				Reusable(r.Context(), donor).
				Return([]donordata.Attorney{}, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &chooseAttorneysSummaryData{
					App:     testAppData,
					Donor:   donor,
					Options: donordata.YesNoMaybeValues,
				}).
				Return(nil)

			err := ChooseAttorneysSummary(template.Execute, service, nil)(testAppData, w, r, donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChooseAttorneysSummaryWhenReuseStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockAttorneyService(t)
	service.EXPECT().
		Reusable(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := ChooseAttorneysSummary(nil, service, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
	})
	assert.Equal(t, expectedError, err)
}

func TestGetChooseAttorneysSummaryWhenNoAttorneysOrTrustCorporation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysSummary(nil, nil, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneys.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseAttorneysSummaryAddAttorney(t *testing.T) {
	testcases := map[string]struct {
		addMoreFormValue donordata.YesNoMaybe
		expectedUrl      string
		Attorneys        donordata.Attorneys
	}{
		"add attorney - with attorney": {
			addMoreFormValue: donordata.Yes,
			expectedUrl:      donor.PathEnterAttorney.Format("lpa-id") + "?addAnother=1&id=" + testUID.String(),
			Attorneys:        donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: actoruid.New()}}},
		},
		"choose attorney": {
			addMoreFormValue: donordata.Maybe,
			expectedUrl:      donor.PathChooseAttorneys.Format("lpa-id"),
			Attorneys:        donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: actoruid.New()}}},
		},
		"do not add attorney - with single attorney": {
			addMoreFormValue: donordata.No,
			expectedUrl:      donor.PathTaskList.Format("lpa-id"),
			Attorneys:        donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: actoruid.New()}}},
		},
		"do not add attorney - with multiple attorneys": {
			addMoreFormValue: donordata.No,
			expectedUrl:      donor.PathHowShouldAttorneysMakeDecisions.Format("lpa-id"),
			Attorneys:        donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: actoruid.New()}, {UID: actoruid.New()}}},
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			f := url.Values{
				"option": {tc.addMoreFormValue.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			service := newMockAttorneyService(t)
			service.EXPECT().
				Reusable(mock.Anything, mock.Anything).
				Return([]donordata.Attorney{}, nil)

			err := ChooseAttorneysSummary(nil, service, testUIDFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", Attorneys: tc.Attorneys})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedUrl, resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseAttorneysSummaryFormValidation(t *testing.T) {
	f := url.Values{
		"option": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("option", validation.SelectError{Label: "yesToAddAnotherAttorney"})

	service := newMockAttorneyService(t)
	service.EXPECT().
		Reusable(mock.Anything, mock.Anything).
		Return([]donordata.Attorney{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *chooseAttorneysSummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChooseAttorneysSummary(template.Execute, service, nil)(testAppData, w, r, &donordata.Provided{Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
