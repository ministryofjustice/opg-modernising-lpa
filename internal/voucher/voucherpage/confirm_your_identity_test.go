package voucherpage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmYourIdentity(t *testing.T) {
	testcases := []bool{
		true, false,
	}

	for _, enabled := range testcases {
		t.Run(fmt.Sprintf("enabled=%t", enabled), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpa := &lpadata.Lpa{LpaID: "lpa-id"}

			resolvingService := newMockLpaStoreResolvingService(t)
			resolvingService.EXPECT().
				Get(r.Context()).
				Return(lpa, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &confirmYourIdentityData{
					App:                  testAppData,
					Errors:               nil,
					LowConfidenceEnabled: enabled,
					Lpa:                  lpa,
				}).
				Return(nil)

			err := ConfirmYourIdentity(template.Execute, enabled, resolvingService)(testAppData, w, r, nil)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestConfirmYourIdentityResolvingServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	resolvingService := newMockLpaStoreResolvingService(t)
	resolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := ConfirmYourIdentity(nil, true, resolvingService)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestConfirmYourIdentityTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	resolvingService := newMockLpaStoreResolvingService(t)
	resolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ConfirmYourIdentity(template.Execute, true, resolvingService)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
