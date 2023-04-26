package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		lpa      *page.Lpa
		appData  page.AppData
		expected func([]taskListItem) []taskListItem
	}{
		"empty": {
			lpa:     &page.Lpa{},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				return items
			},
		},
		"donor and certificate provider signed": {
			lpa: &page.Lpa{
				Submitted:   time.Now(),
				Certificate: page.Certificate{Agreed: time.Now()},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[2].Path = page.Paths.Attorney.Sign

				return items
			},
		},
		"completed": {
			lpa: &page.Lpa{
				Submitted:   time.Now(),
				Certificate: page.Certificate{Agreed: time.Now()},
				AttorneyTasks: map[string]page.AttorneyTasks{
					"attorney-id": {
						ConfirmYourDetails: page.TaskCompleted,
						ReadTheLpa:         page.TaskCompleted,
						SignTheLpa:         page.TaskCompleted,
					},
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = page.TaskCompleted
				items[1].State = page.TaskCompleted
				items[2].State = page.TaskCompleted
				items[2].Path = page.Paths.Attorney.Sign

				return items
			},
		},
		"completed replacement": {
			lpa: &page.Lpa{
				Submitted:   time.Now(),
				Certificate: page.Certificate{Agreed: time.Now()},
				ReplacementAttorneyTasks: map[string]page.AttorneyTasks{
					"attorney-id": {
						ConfirmYourDetails: page.TaskCompleted,
						ReadTheLpa:         page.TaskCompleted,
						SignTheLpa:         page.TaskCompleted,
					},
				},
			},
			appData: testReplacementAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = page.TaskCompleted
				items[1].State = page.TaskCompleted
				items[2].State = page.TaskCompleted
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

			template := newMockTemplate(t)
			template.
				On("Execute", w, &taskListData{
					App: tc.appData,
					Lpa: tc.lpa,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourDetails", Path: page.Paths.Attorney.CheckYourName},
						{Name: "readTheLpa", Path: page.Paths.Attorney.NextPage},
						{Name: "signTheLpa"},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := TaskList(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
