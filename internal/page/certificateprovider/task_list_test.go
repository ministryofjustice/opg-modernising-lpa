package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		donor               *lpastore.Lpa
		certificateProvider *actor.CertificateProviderProvidedDetails
		appData             page.AppData
		expected            func([]taskListItem) []taskListItem
	}{
		"empty": {
			donor:               &lpastore.Lpa{LpaID: "lpa-id"},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			appData:             testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[1].Disabled = true
				items[2].Disabled = true

				return items
			},
		},
		"paid": {
			donor: &lpastore.Lpa{
				LpaID: "lpa-id",
				Paid:  true,
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].Disabled = true
				items[2].Disabled = true

				return items
			},
		},
		"submitted": {
			donor: &lpastore.Lpa{
				LpaID:    "lpa-id",
				SignedAt: time.Now(),
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].Disabled = true
				items[2].Disabled = true

				return items
			},
		},
		"identity confirmed": {
			donor: &lpastore.Lpa{
				LpaID:    "lpa-id",
				SignedAt: time.Now(),
				Paid:     true,
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails:    actor.TaskCompleted,
					ConfirmYourIdentity:   actor.TaskCompleted,
					ProvideTheCertificate: actor.TaskCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted
				items[1].Path = page.Paths.CertificateProvider.ReadTheLpa.Format("lpa-id")
				items[2].State = actor.TaskCompleted

				return items
			},
		},
		"all": {
			donor: &lpastore.Lpa{
				LpaID:    "lpa-id",
				SignedAt: time.Now(),
				Paid:     true,
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails:    actor.TaskCompleted,
					ConfirmYourIdentity:   actor.TaskCompleted,
					ProvideTheCertificate: actor.TaskCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted
				items[1].Path = page.Paths.CertificateProvider.ReadTheLpa.Format("lpa-id")
				items[2].State = actor.TaskCompleted

				return items
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.donor, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				Get(r.Context()).
				Return(tc.certificateProvider, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &taskListData{
					App: tc.appData,
					Lpa: tc.donor,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourDetails", Path: page.Paths.CertificateProvider.EnterDateOfBirth.Format("lpa-id")},
						{Name: "confirmYourIdentity", Path: page.Paths.CertificateProvider.ProveYourIdentity.Format("lpa-id")},
						{Name: "provideYourCertificate", Path: page.Paths.CertificateProvider.ReadTheLpa.Format("lpa-id")},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute, lpaStoreResolvingService, certificateProviderStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, expectedError)

	err := TaskList(nil, lpaStoreResolvingService, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{LpaID: "lpa-id"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	err := TaskList(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{LpaID: "lpa-id"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
