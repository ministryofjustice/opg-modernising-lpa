package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		lpa      *lpastore.Lpa
		attorney *actor.AttorneyProvidedDetails
		appData  page.AppData
		expected func([]taskListItem) []taskListItem
	}{
		"empty": {
			lpa:      &lpastore.Lpa{LpaID: "lpa-id"},
			attorney: &actor.AttorneyProvidedDetails{},
			appData:  testAppData,
			expected: func(items []taskListItem) []taskListItem {
				return items
			},
		},
		"trust corporation": {
			lpa:      &lpastore.Lpa{LpaID: "lpa-id"},
			attorney: &actor.AttorneyProvidedDetails{},
			appData:  testTrustCorporationAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].Path = page.Paths.Attorney.MobileNumber.Format("lpa-id")

				return items
			},
		},
		"trust corporation with two signatories": {
			lpa: &lpastore.Lpa{
				LpaID:               "lpa-id",
				SignedAt:            time.Now(),
				CertificateProvider: lpastore.CertificateProvider{SignedAt: time.Now()},
			},
			attorney: &actor.AttorneyProvidedDetails{
				WouldLikeSecondSignatory: form.Yes,
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
			},
			appData: testTrustCorporationAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[0].Path = page.Paths.Attorney.MobileNumber.Format("lpa-id")
				items[1].State = actor.TaskCompleted
				items[2].Name = "signTheLpaSignatory1"
				items[2].Path = page.Paths.Attorney.RightsAndResponsibilities.Format("lpa-id")

				return append(items, taskListItem{
					Name: "signTheLpaSignatory2",
					Path: page.Paths.Attorney.Sign.Format("lpa-id") + "?second",
				})
			},
		},
		"tasks completed not signed": {
			lpa: &lpastore.Lpa{
				LpaID:    "lpa-id",
				SignedAt: time.Now(),
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted

				return items
			},
		},
		"tasks completed and signed": {
			lpa: &lpastore.Lpa{
				LpaID:               "lpa-id",
				SignedAt:            time.Now(),
				CertificateProvider: lpastore.CertificateProvider{SignedAt: time.Now()},
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted
				items[2].Path = page.Paths.Attorney.RightsAndResponsibilities.Format("lpa-id")

				return items
			},
		},
		"completed": {
			lpa: &lpastore.Lpa{
				LpaID:               "lpa-id",
				SignedAt:            time.Now(),
				CertificateProvider: lpastore.CertificateProvider{SignedAt: time.Now()},
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
					SignTheLpa:         actor.TaskCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted
				items[2].State = actor.TaskCompleted
				items[2].Path = page.Paths.Attorney.RightsAndResponsibilities.Format("lpa-id")

				return items
			},
		},
		"completed replacement": {
			lpa: &lpastore.Lpa{
				LpaID:               "lpa-id",
				SignedAt:            time.Now(),
				CertificateProvider: lpastore.CertificateProvider{SignedAt: time.Now()},
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
					SignTheLpa:         actor.TaskCompleted,
				},
			},
			appData: testReplacementAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = actor.TaskCompleted
				items[1].State = actor.TaskCompleted
				items[2].State = actor.TaskCompleted
				items[2].Path = page.Paths.Attorney.RightsAndResponsibilities.Format("lpa-id")

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
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &taskListData{
					App: tc.appData,
					Lpa: tc.lpa,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourDetails", Path: page.Paths.Attorney.MobileNumber.Format("lpa-id")},
						{Name: "readTheLpa", Path: page.Paths.Attorney.ReadTheLpa.Format("lpa-id")},
						{Name: "signTheLpa"},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute, lpaStoreResolvingService)(tc.appData, w, r, tc.attorney)
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

	err := TaskList(nil, lpaStoreResolvingService)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{LpaID: "lpa-id"}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute, lpaStoreResolvingService)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
