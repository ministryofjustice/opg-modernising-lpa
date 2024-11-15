package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowWillYouConfirmYourIdentity(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWillYouConfirmYourIdentityData{
			App:     testAppData,
			Form:    &howWillYouConfirmYourIdentityForm{},
			Options: howYouWillConfirmYourIdentityValues,
		}).
		Return(nil)

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowWillYouConfirmYourIdentityWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &voucherdata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostHowWillYouConfirmYourIdentity(t *testing.T) {
	testCases := map[string]struct {
		how      howYouWillConfirmYourIdentity
		provided *voucherdata.Provided
		redirect voucher.Path
	}{
		"post office successful": {
			how:      howYouWillConfirmYourIdentityPostOfficeSuccessfully,
			provided: &voucherdata.Provided{LpaID: "lpa-id"},
			redirect: voucher.PathIdentityWithOneLogin,
		},
		"one login": {
			how:      howYouWillConfirmYourIdentityOneLogin,
			provided: &voucherdata.Provided{LpaID: "lpa-id"},
			redirect: voucher.PathIdentityWithOneLogin,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"how": {tc.how.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := HowWillYouConfirmYourIdentity(nil, nil)(testAppData, w, r, tc.provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowWillYouConfirmYourIdentityWhenAtPostOfficeSelected(t *testing.T) {
	form := url.Values{
		"how": {howYouWillConfirmYourIdentityAtPostOffice.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), &voucherdata.Provided{
			LpaID: "lpa-id",
			Tasks: voucherdata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
		}).
		Return(nil)

	err := HowWillYouConfirmYourIdentity(nil, voucherStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowWillYouConfirmYourIdentityWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how": {howYouWillConfirmYourIdentityAtPostOffice.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowWillYouConfirmYourIdentity(nil, voucherStore)(testAppData, w, r, &voucherdata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostHowWillYouConfirmYourIdentityWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *howWillYouConfirmYourIdentityData) bool {
			return assert.Equal(t, validation.With("how", validation.SelectError{Label: "howYouWillConfirmYourIdentity"}), data.Errors)
		})).
		Return(nil)

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowWillYouConfirmYourIdentityForm(t *testing.T) {
	form := url.Values{
		"how": {howYouWillConfirmYourIdentityAtPostOffice.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowWillYouConfirmYourIdentityForm(r)
	assert.Equal(t, howYouWillConfirmYourIdentityAtPostOffice, result.How)
}

func TestHowWillYouConfirmYourIdentityFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howWillYouConfirmYourIdentityForm
		errors validation.List
	}{
		"valid": {
			form: &howWillYouConfirmYourIdentityForm{
				How: howYouWillConfirmYourIdentityAtPostOffice,
			},
		},
		"invalid": {
			form:   &howWillYouConfirmYourIdentityForm{},
			errors: validation.With("how", validation.SelectError{Label: "howYouWillConfirmYourIdentity"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
