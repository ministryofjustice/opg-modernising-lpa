package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	certificateProviderAgreed := func(t *testing.T, r *http.Request) *mockCertificateProviderStore {
		certificateProviderStore := newMockCertificateProviderStore(t)
		certificateProviderStore.
			On("Get", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
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
		"tasks completed not signed": {
			lpa: &page.Lpa{
				ID:        "lpa-id",
				Submitted: time.Now(),
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ReadTheLpa: actor.TaskCompleted,
				},
			},
			certificateProviderStore: func(t *testing.T, r *http.Request) *mockCertificateProviderStore {
				certificateProviderStore := newMockCertificateProviderStore(t)
				certificateProviderStore.
					On("Get", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
					Return(&actor.CertificateProviderProvidedDetails{}, nil)

				return certificateProviderStore
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
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
					ReadTheLpa: actor.TaskCompleted,
				},
			},
			certificateProviderStore: certificateProviderAgreed,
			appData:                  testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[1].State = actor.TaskCompleted
				items[2].Path = page.Paths.Attorney.Sign

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
				items[2].Path = page.Paths.Attorney.Sign

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
				items[2].Path = page.Paths.Attorney.Sign

				return items
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(tc.attorney, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &taskListData{
					App: tc.appData,
					Lpa: tc.lpa,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourDetails", Path: page.Paths.Attorney.CheckYourName},
						{Name: "readTheLpa", Path: page.Paths.Attorney.ReadTheLpa},
						{Name: "signTheLpa"},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute, lpaStore, tc.certificateProviderStore(t, r), attorneyStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := TaskList(nil, lpaStore, nil, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{
			Tasks: actor.AttorneyTasks{
				ReadTheLpa: actor.TaskCompleted,
			},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(nil, expectedError)

	err := TaskList(nil, lpaStore, certificateProviderStore, attorneyStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenCertificateProviderNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{
			Tasks: actor.AttorneyTasks{
				ReadTheLpa: actor.TaskCompleted,
			},
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(nil, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(nil)

	err := TaskList(template.Execute, lpaStore, certificateProviderStore, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute, lpaStore, nil, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
