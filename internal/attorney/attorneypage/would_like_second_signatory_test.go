package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWouldLikeSecondSignatory(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &wouldLikeSecondSignatoryData{
			App:  testAppData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := WouldLikeSecondSignatory(template.Execute, nil, nil)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWouldLikeSecondSignatoryWhenAlreadySigned(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WouldLikeSecondSignatory(nil, nil, nil)(testAppData, w, r, &attorneydata.Provided{
		LpaID:    "lpa-id",
		SignedAt: time.Now(),
	}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, attorney.PathWhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetWouldLikeSecondSignatoryWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &wouldLikeSecondSignatoryData{
			App:  testAppData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(expectedError)

	err := WouldLikeSecondSignatory(template.Execute, nil, nil)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWouldLikeSecondSignatoryWhenYes(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), &attorneydata.Provided{
			LpaID:                    "lpa-id",
			WouldLikeSecondSignatory: form.Yes,
		}).
		Return(nil)

	err := WouldLikeSecondSignatory(nil, attorneyStore, nil)(testAppData, w, r, &attorneydata.Provided{
		LpaID: "lpa-id",
	}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, attorney.PathSign.Format("lpa-id")+"?second=", resp.Header.Get("Location"))
}

func TestPostWouldLikeSecondSignatoryWhenNo(t *testing.T) {
	lpa := &lpadata.Lpa{SignedAt: time.Now()}
	updatedAttorney := &attorneydata.Provided{
		LpaID:                    "lpa-id",
		WouldLikeSecondSignatory: form.No,
	}

	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), updatedAttorney).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorney(r.Context(), lpa, updatedAttorney).
		Return(nil)

	err := WouldLikeSecondSignatory(nil, attorneyStore, lpaStoreClient)(testAppData, w, r, &attorneydata.Provided{
		LpaID: "lpa-id",
	}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, attorney.PathWhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWouldLikeSecondSignatoryWhenNoAndSignedInLpaStore(t *testing.T) {
	testcases := map[string]struct {
		appData appcontext.Data
		lpa     *lpadata.Lpa
	}{
		"trust corporation": {
			appData: testTrustCorporationAppData,
			lpa: &lpadata.Lpa{
				Attorneys: lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{
					Signatories: []lpadata.TrustCorporationSignatory{{SignedAt: time.Now()}},
				}},
			},
		},
		"replacement trust corporation": {
			appData: testReplacementTrustCorporationAppData,
			lpa: &lpadata.Lpa{
				ReplacementAttorneys: lpadata.Attorneys{TrustCorporation: lpadata.TrustCorporation{
					Signatories: []lpadata.TrustCorporationSignatory{{SignedAt: time.Now()}},
				}},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			updatedAttorney := &attorneydata.Provided{
				LpaID:                    "lpa-id",
				WouldLikeSecondSignatory: form.No,
			}

			f := url.Values{
				form.FieldNames.YesNo: {form.No.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Put(r.Context(), updatedAttorney).
				Return(nil)

			err := WouldLikeSecondSignatory(nil, attorneyStore, nil)(tc.appData, w, r, &attorneydata.Provided{
				LpaID: "lpa-id",
			}, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathWhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostWouldLikeSecondSignatoryWhenLpaStoreClientErrors(t *testing.T) {
	lpa := &lpadata.Lpa{SignedAt: time.Now()}

	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorney(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := WouldLikeSecondSignatory(nil, nil, lpaStoreClient)(testAppData, w, r, &attorneydata.Provided{}, lpa)
	assert.Equal(t, expectedError, err)
}

func TestPostWouldLikeSecondSignatoryWhenAttorneyStoreErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorney(r.Context(), mock.Anything, mock.Anything).
		Return(nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := WouldLikeSecondSignatory(nil, attorneyStore, lpaStoreClient)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestPostWouldLikeSecondSignatoryWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesIfWouldLikeSecondSignatory"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *wouldLikeSecondSignatoryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := WouldLikeSecondSignatory(template.Execute, nil, nil)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
