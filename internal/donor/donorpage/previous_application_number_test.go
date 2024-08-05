package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPreviousApplicationNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &previousApplicationNumberData{
			App:  testAppData,
			Form: &previousApplicationNumberForm{},
		}).
		Return(nil)

	err := PreviousApplicationNumber(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPreviousApplicationNumberFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &previousApplicationNumberData{
			App: testAppData,
			Form: &previousApplicationNumberForm{
				PreviousApplicationNumber: "ABC",
			},
		}).
		Return(nil)

	err := PreviousApplicationNumber(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{PreviousApplicationNumber: "ABC"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPreviousApplicationNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := PreviousApplicationNumber(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostPreviousApplicationNumber(t *testing.T) {
	testcases := map[string]page.LpaPath{
		"7": page.Paths.PreviousFee,
		"M": page.Paths.EvidenceSuccessfullyUploaded,
	}

	for start, redirect := range testcases {
		t.Run(start, func(t *testing.T) {
			form := url.Values{
				"previous-application-number": {start},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:                     "lpa-id",
					LpaUID:                    "lpa-uid",
					PreviousApplicationNumber: start,
				}).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendPreviousApplicationLinked(r.Context(), event.PreviousApplicationLinked{
					UID:                       "lpa-uid",
					PreviousApplicationNumber: start,
				}).
				Return(nil)

			err := PreviousApplicationNumber(nil, donorStore, eventClient)(testAppData, w, r, &donordata.Provided{
				LpaID:  "lpa-id",
				LpaUID: "lpa-uid",
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostPreviousApplicationNumberWhenEventErrors(t *testing.T) {
	form := url.Values{
		"previous-application-number": {"7"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPreviousApplicationLinked(r.Context(), mock.Anything).
		Return(expectedError)

	err := PreviousApplicationNumber(nil, donorStore, eventClient)(testAppData, w, r, &donordata.Provided{
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
	})

	assert.Equal(t, expectedError, err)
}

func TestPostPreviousApplicationNumberWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"previous-application-number": {"MABC"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := PreviousApplicationNumber(nil, donorStore, nil)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostPreviousApplicationNumberWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *previousApplicationNumberData) bool {
			return assert.Equal(t, validation.With("previous-application-number", validation.EnterError{Label: "previousApplicationNumber"}), data.Errors)
		})).
		Return(nil)

	err := PreviousApplicationNumber(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadPreviousApplicationNumberForm(t *testing.T) {
	form := url.Values{
		"previous-application-number": {"ABC"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readPreviousApplicationNumberForm(r)

	assert.Equal(t, "ABC", result.PreviousApplicationNumber)
}

func TestPreviousApplicationNumberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *previousApplicationNumberForm
		errors validation.List
	}{
		"valid modernised": {
			form: &previousApplicationNumberForm{
				PreviousApplicationNumber: "M",
			},
		},
		"valid old": {
			form: &previousApplicationNumberForm{
				PreviousApplicationNumber: "7",
			},
		},
		"invalid": {
			form: &previousApplicationNumberForm{
				PreviousApplicationNumber: "x",
			},
			errors: validation.With("previous-application-number", validation.ReferenceNumberError{Label: "previousApplicationNumber"}),
		},
		"empty": {
			form:   &previousApplicationNumberForm{},
			errors: validation.With("previous-application-number", validation.EnterError{Label: "previousApplicationNumber"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
