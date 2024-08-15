package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourNameData{
			App: testAppData,
			Form: &yourNameForm{
				FirstNames: "V",
				LastName:   "W",
			},
		}).
		Return(nil)

	err := YourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourNameWhenChanged(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourNameData{
			App: testAppData,
			Form: &yourNameForm{
				FirstNames: "A",
				LastName:   "B",
			},
		}).
		Return(nil)

	err := YourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{
		FirstNames: "A",
		LastName:   "B",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := YourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourName(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
		}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), &voucherdata.Provided{
			LpaID:      "lpa-id",
			FirstNames: "John",
			LastName:   "Doe",
		}).
		Return(nil)

	err := YourName(nil, lpaStoreResolvingService, voucherStore)(testAppData, w, r, &voucherdata.Provided{
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathConfirmYourName.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourNameWhenInputRequired(t *testing.T) {
	form := url.Values{
		"last-name": {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *yourNameData) bool {
			return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
		})).
		Return(nil)

	err := YourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourNameWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourName(nil, lpaStoreResolvingService, voucherStore)(testAppData, w, r, &voucherdata.Provided{})
	assert.Equal(t, expectedError, err)
}
