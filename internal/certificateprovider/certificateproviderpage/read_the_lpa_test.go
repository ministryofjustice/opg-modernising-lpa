package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetReadTheLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &readTheLpaData{App: testAppData, Lpa: donor}).
		Return(nil)

	err := ReadTheLpa(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, expectedError)

	err := ReadTheLpa(nil, lpaStoreResolvingService, nil)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetReadTheLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ReadTheLpa(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestPostReadTheLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{
			LpaID:    "lpa-id",
			SignedAt: time.Now(),
			Paid:     true,
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), &actor.CertificateProviderProvidedDetails{
			Tasks: actor.CertificateProviderTasks{
				ReadTheLpa: actor.TaskCompleted,
			},
		}).
		Return(nil)

	err := ReadTheLpa(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.WhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostReadTheLpaWhenNotReady(t *testing.T) {
	testcases := map[string]*lpastore.Lpa{
		"not submitted": {
			LpaID: "lpa-id",
			Paid:  true,
		},
		"not paid": {
			LpaID:    "lpa-id",
			SignedAt: time.Now(),
		},
	}

	for name, donor := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(donor, nil)

			err := ReadTheLpa(nil, lpaStoreResolvingService, nil)(testAppData, w, r, nil)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProvider.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostReadTheLpaWithAttorneyWhenCertificateStorePutErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{
			LpaID:    "lpa-id",
			SignedAt: time.Now(),
			Paid:     true,
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ReadTheLpa(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
