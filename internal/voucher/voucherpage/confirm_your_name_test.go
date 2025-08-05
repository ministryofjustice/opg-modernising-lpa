package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmYourName(t *testing.T) {
	testcases := map[string]struct {
		providedFirstNames, providedLastName string
		lpaFirstNames, lpaLastName           string
		changed                              bool
	}{
		"initial": {
			lpaFirstNames: "V",
			lpaLastName:   "W",
		},
		"set to initial": {
			providedFirstNames: "V",
			providedLastName:   "W",
			lpaFirstNames:      "V",
			lpaLastName:        "W",
		},
		"set to different": {
			providedFirstNames: "V",
			providedLastName:   "W",
			lpaFirstNames:      "A",
			lpaLastName:        "B",
			changed:            true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpadata.Lpa{
					Voucher: lpadata.Voucher{FirstNames: tc.lpaFirstNames, LastName: tc.lpaLastName},
				}, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &confirmYourNameData{
					App: testAppData,
					Lpa: &lpadata.Lpa{
						Voucher: lpadata.Voucher{FirstNames: tc.lpaFirstNames, LastName: tc.lpaLastName},
					},
					FirstNames: "V",
					LastName:   "W",
					Changed:    tc.changed,
				}).
				Return(nil)

			err := ConfirmYourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{
				LpaID:      "lpa-id",
				FirstNames: tc.providedFirstNames,
				LastName:   tc.providedLastName,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Empty(t, resp.Cookies())
		})
	}
}

func TestGetConfirmYourNameWhenChanged(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "banner", Value: "1", MaxAge: 60})

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
		}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmYourNameData{
			App: testAppData,
			Lpa: &lpadata.Lpa{
				Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
			},
			FirstNames: "A",
			LastName:   "B",
			Changed:    true,
			ShowBanner: true,
		}).
		Return(nil)

	err := ConfirmYourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{
		LpaID:      "lpa-id",
		FirstNames: "A",
		LastName:   "B",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if assert.Len(t, resp.Cookies(), 1) {
		cookie := resp.Cookies()[0]

		assert.Equal(t, "banner", cookie.Name)
		assert.Equal(t, "1", cookie.Value)
		assert.Equal(t, -1, cookie.MaxAge)
	}
}

func TestGetConfirmYourNameWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, expectedError)

	err := ConfirmYourName(nil, lpaStoreResolvingService, nil)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmYourNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), &voucherdata.Provided{
			LpaID: "lpa-id",
			Tasks: voucherdata.Tasks{ConfirmYourName: task.StateCompleted},
		}).
		Return(nil)

	err := ConfirmYourName(nil, lpaStoreResolvingService, voucherStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourNameWhenDonorLastNameMatch(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), &voucherdata.Provided{
			LpaID:    "lpa-id",
			LastName: "Smith",
			Tasks:    voucherdata.Tasks{ConfirmYourName: task.StateInProgress},
		}).
		Return(nil)

	err := ConfirmYourName(nil, lpaStoreResolvingService, voucherStore)(testAppData, w, r, &voucherdata.Provided{
		LpaID:    "lpa-id",
		LastName: "Smith",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathConfirmAllowedToVouch.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourNameWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourName(nil, lpaStoreResolvingService, voucherStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
