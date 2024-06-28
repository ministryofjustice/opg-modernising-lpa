package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhatIsVouching(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whatIsVouchingData{
			App: testAppData,
			Form: &form.YesNoForm{
				YesNo:     form.Yes,
				FieldName: form.FieldNames.YesNo,
				Options:   form.YesNoValues,
			},
		}).
		Return(nil)

	err := WhatIsVouching(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{WantVoucher: form.Yes})

	assert.Nil(t, err)
}

func TestGetWhatIsVouchingWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatIsVouching(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Error(t, err)
}

func TestPostWhatIsVouching(t *testing.T) {
	testcases := map[form.YesNo]string{
		form.Yes: page.Paths.EnterVoucher.Format("lpa-id"),
		form.No:  page.Paths.WhatYouCanDoNow.Format("lpa-id"),
	}

	for yesNo, path := range testcases {
		t.Run(yesNo.String(), func(t *testing.T) {
			f := url.Values{
				"yes-no": {yesNo.String()},
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{LpaID: "lpa-id", WantVoucher: yesNo}).
				Return(nil)

			err := WhatIsVouching(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, path, resp.Header.Get("Location"))
		})
	}
}

func TestPostWhatIsVouchingWhenDonorStoreError(t *testing.T) {
	f := url.Values{
		"yes-no": {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatIsVouching(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
