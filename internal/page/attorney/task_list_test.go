package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	certificateProviderAgreed := func(t *testing.T, r *http.Request) *mockCertificateProviderStore {
		certificateProviderStore := newMockCertificateProviderStore(t)
		certificateProviderStore.
			On("GetAny", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
			Return(&actor.CertificateProviderProvidedDetails{
				Certificate: actor.Certificate{Agreed: time.Now()},
			}, nil)

		return certificateProviderStore
	}

	testCases := map[string]struct {
		lpa                      *page.Lpa
		attorney                 *actor.AttorneyProvidedDetails
		certificateProviderStore func(t *testing.T, r *http.Request) *mockCertificateProviderStore
		appData                  page.AppData
		expected                 func([]taskListItem) []taskListItem
	}{
		"empty": {
			lpa:      &page.Lpa{ID: "lpa-id"},
			attorney: &actor.AttorneyProvidedDetails{},
			certificateProviderStore: func(t *testing.T, r *http.Request) *mockCertificateProviderStore {
				return nil
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				return items
			},
		},
		"trust corporation": {
			lpa:      &page.Lpa{ID: "lpa-id"},
			attorney: &actor.AttorneyProvidedDetails{},
			certificateProviderStore: func(t *testing.T, r *http.Request) *mockCertificateProviderStore {
				return nil
			},
			appData: testTrustCorporationAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].Path = page.Paths.Attorney.ConfirmYourDetails.Format("lpa-id")

				return items
			},
		},
		"trust corporation with two signatories": {
			lpa: &page.Lpa{
				ID:        "lpa-id",
				Submitted: time.Now(),
			},
			attorney: &actor.AttorneyProvidedDetails{
				WouldLikeSecondSignatory: form.Yes,
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
			},
			certificateProviderStore: certificateProviderAgreed,
			appData:                  testTrustCorporationAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[0].Path = page.Paths.Attorney.ConfirmYourDetails.Format("lpa-id")
				items[1].State = actor.TaskCompleted
				items[2].Name = "signTheLpaSignatory1"
				items[2].Path = page.Paths.Attorney.Sign.Format("lpa-id")

				return append(items, taskListItem{
					Name: "signTheLpaSignatory2",
					Path: page.Paths.Attorney.Sign.Format("lpa-id") + "?second",
				})
			},
		},
		"tasks completed not signed": {
			lpa: &page.Lpa{
				ID:        "lpa-id",
				Submitted: time.Now(),
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
			},
			certificateProviderStore: func(t *testing.T, r *http.Request) *mockCertificateProviderStore {
				certificateProviderStore := newMockCertificateProviderStore(t)
				certificateProviderStore.
					On("GetAny", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
					Return(&actor.CertificateProviderProvidedDetails{}, nil)

				return certificateProviderStore
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted

				return items
			},
		},
		"tasks completed and signed": {
			lpa: &page.Lpa{
				ID:        "lpa-id",
				Submitted: time.Now(),
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
			},
			certificateProviderStore: certificateProviderAgreed,
			appData:                  testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted
				items[2].Path = page.Paths.Attorney.Sign.Format("lpa-id")

				return items
			},
		},
		"completed": {
			lpa: &page.Lpa{
				ID:        "lpa-id",
				Submitted: time.Now(),
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
					SignTheLpa:         actor.TaskCompleted,
				},
			},
			certificateProviderStore: certificateProviderAgreed,
			appData:                  testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted
				items[2].State = actor.TaskCompleted
				items[2].Path = page.Paths.Attorney.Sign.Format("lpa-id")

				return items
			},
		},
		"completed replacement": {
			lpa: &page.Lpa{
				ID:        "lpa-id",
				Submitted: time.Now(),
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
					SignTheLpa:         actor.TaskCompleted,
				},
			},
			certificateProviderStore: certificateProviderAgreed,
			appData:                  testReplacementAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted
				items[2].State = actor.TaskCompleted
				items[2].Path = page.Paths.Attorney.Sign.Format("lpa-id")

				return items
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &taskListData{
					App: tc.appData,
					Lpa: tc.lpa,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourDetails", Path: page.Paths.Attorney.MobileNumber.Format("lpa-id")},
						{Name: "readTheLpa", Path: page.Paths.Attorney.ReadTheLpa.Format("lpa-id")},
						{Name: "signTheLpa"},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute, donorStore, tc.certificateProviderStore(t, r))(tc.appData, w, r, tc.attorney)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := TaskList(nil, donorStore, nil)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(nil, expectedError)

	err := TaskList(nil, donorStore, certificateProviderStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskCompleted, ReadTheLpa: actor.TaskCompleted}})

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenCertificateProviderNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(nil, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(nil)

	err := TaskList(template.Execute, donorStore, certificateProviderStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskCompleted, ReadTheLpa: actor.TaskCompleted}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute, donorStore, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
