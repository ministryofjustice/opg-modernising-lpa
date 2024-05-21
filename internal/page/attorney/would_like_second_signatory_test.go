package attorney

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
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

	err := WouldLikeSecondSignatory(template.Execute, nil, nil, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWouldLikeSecondSignatoryWhenAlreadySigned(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WouldLikeSecondSignatory(nil, nil, nil, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{
		LpaID:     "lpa-id",
		Confirmed: time.Now(),
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.WhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
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

	err := WouldLikeSecondSignatory(template.Execute, nil, nil, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
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
		Put(r.Context(), &actor.AttorneyProvidedDetails{
			LpaID:                    "lpa-id",
			WouldLikeSecondSignatory: form.Yes,
		}).
		Return(nil)

	err := WouldLikeSecondSignatory(nil, attorneyStore, nil, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Sign.Format("lpa-id")+"?second=", resp.Header.Get("Location"))
}

func TestPostWouldLikeSecondSignatoryWhenNo(t *testing.T) {
	donor := &lpastore.Lpa{SignedAt: time.Now()}
	updatedAttorney := &actor.AttorneyProvidedDetails{
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

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorney(r.Context(), donor, updatedAttorney).
		Return(nil)

	err := WouldLikeSecondSignatory(nil, attorneyStore, lpaStoreResolvingService, lpaStoreClient)(testAppData, w, r, &actor.AttorneyProvidedDetails{
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.WhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWouldLikeSecondSignatoryWhenNoAndSignedInLpaStore(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *lpastore.Lpa
	}{
		"trust corporation": {
			appData: testTrustCorporationAppData,
			lpa: &lpastore.Lpa{
				Attorneys: lpastore.Attorneys{TrustCorporation: lpastore.TrustCorporation{
					Signatories: []lpastore.TrustCorporationSignatory{{SignedAt: time.Now()}},
				}},
			},
		},
		"replacement trust corporation": {
			appData: testReplacementTrustCorporationAppData,
			lpa: &lpastore.Lpa{
				ReplacementAttorneys: lpastore.Attorneys{TrustCorporation: lpastore.TrustCorporation{
					Signatories: []lpastore.TrustCorporationSignatory{{SignedAt: time.Now()}},
				}},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			updatedAttorney := &actor.AttorneyProvidedDetails{
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

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.lpa, nil)

			err := WouldLikeSecondSignatory(nil, attorneyStore, lpaStoreResolvingService, nil)(tc.appData, w, r, &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.WhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostWouldLikeSecondSignatoryWhenLpaStoreClientErrors(t *testing.T) {
	donor := &lpastore.Lpa{SignedAt: time.Now()}

	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorney(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := WouldLikeSecondSignatory(nil, nil, lpaStoreResolvingService, lpaStoreClient)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestPostWouldLikeSecondSignatoryWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := WouldLikeSecondSignatory(nil, nil, lpaStoreResolvingService, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestPostWouldLikeSecondSignatoryWhenAttorneyStoreErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorney(r.Context(), mock.Anything, mock.Anything).
		Return(nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := WouldLikeSecondSignatory(nil, attorneyStore, lpaStoreResolvingService, lpaStoreClient)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
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

	err := WouldLikeSecondSignatory(template.Execute, nil, nil, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
