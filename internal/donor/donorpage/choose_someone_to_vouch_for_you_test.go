package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseSomeoneToVouchForYou(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseSomeoneToVouchForYouData{
			App: testAppData,
			Form: &form.YesNoForm{
				YesNo:     form.Yes,
				FieldName: form.FieldNames.YesNo,
				Options:   form.YesNoValues,
			},
		}).
		Return(nil)

	err := ChooseSomeoneToVouchForYou(template.Execute, nil)(testAppData, w, r, &donordata.Provided{WantVoucher: form.Yes})

	assert.Nil(t, err)
}

func TestGetChooseSomeoneToVouchForYouWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseSomeoneToVouchForYou(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})

	assert.Error(t, err)
}

func TestPostChooseSomeoneToVouchForYou(t *testing.T) {
	testcases := map[form.YesNo]string{
		form.Yes: donor.PathEnterVoucher.Format("lpa-id"),
		form.No:  donor.PathWhatYouCanDoNow.Format("lpa-id"),
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
				Put(r.Context(), &donordata.Provided{
					LpaID:       "lpa-id",
					WantVoucher: yesNo,
				}).
				Return(nil)

			err := ChooseSomeoneToVouchForYou(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, path, resp.Header.Get("Location"))
		})
	}
}

func TestPostChooseSomeoneToVouchForYouWhenDonorStoreError(t *testing.T) {
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

	err := ChooseSomeoneToVouchForYou(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
