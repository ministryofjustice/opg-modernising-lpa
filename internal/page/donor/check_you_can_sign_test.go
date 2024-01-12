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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCheckYouCanSign(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYouCanSignData{
			App:     testAppData,
			Errors:  nil,
			Form:    &form.YesNoForm{YesNo: form.No},
			Options: form.YesNoValues,
		}).
		Return(nil)

	err := CheckYouCanSign(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{CanSign: form.No},
	})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYouCanSign(t *testing.T) {
	testcases := map[form.YesNo]page.LpaPath{
		form.Yes: page.Paths.YourAddress,
		form.No:  page.Paths.NeedHelpSigningConfirmation,
	}

	for yesNo, redirect := range testcases {
		t.Run(yesNo.String(), func(t *testing.T) {

			form := url.Values{
				"yes-no": {yesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{LpaID: "lpa-id", Donor: actor.Donor{CanSign: yesNo}}).
				Return(nil)

			err := CheckYouCanSign(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYouCanSignErrorOnPutStore(t *testing.T) {
	form := url.Values{
		"yes-no": {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := CheckYouCanSign(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCheckYouCanSignFormValidation(t *testing.T) {
	form := url.Values{
		"yes-no": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("yes-no", validation.SelectError{Label: "yesIfYouWillBeAbleToSignYourself"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *checkYouCanSignData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := CheckYouCanSign(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
