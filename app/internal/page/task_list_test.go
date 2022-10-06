package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		lpa      Lpa
		expected func([]taskListSection) []taskListSection
	}{
		"start": {
			expected: func(sections []taskListSection) []taskListSection {
				return sections
			},
		},
		"in-progress": {
			lpa: Lpa{
				You: Person{
					FirstNames: "this",
				},
				Attorney: Attorney{
					FirstNames: "this",
				},
				Tasks: Tasks{
					WhenCanTheLpaBeUsed:        TaskInProgress,
					Restrictions:               TaskInProgress,
					CertificateProvider:        TaskInProgress,
					CheckYourLpa:               TaskInProgress,
					PayForLpa:                  TaskInProgress,
					ConfirmYourIdentityAndSign: TaskInProgress,
				},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items = []taskListItem{
					{Name: "provideDonorDetails", Path: yourDetailsPath, InProgress: true},
					{Name: "chooseYourAttorneys", Path: chooseAttorneysPath, InProgress: true},
					{Name: "chooseYourReplacementAttorneys", Path: wantReplacementAttorneysPath},
					{Name: "chooseWhenTheLpaCanBeUsed", Path: whenCanTheLpaBeUsedPath, InProgress: true},
					{Name: "addRestrictionsToTheLpa", Path: restrictionsPath, InProgress: true},
					{Name: "chooseYourCertificateProvider", Path: whoDoYouWantToBeCertificateProviderGuidancePath, InProgress: true},
					{Name: "checkAndSendToYourCertificateProvider", Path: checkYourLpaPath, InProgress: true},
				}

				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: aboutPaymentPath, InProgress: true},
				}

				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: selectYourIdentityOptionsPath, InProgress: true},
				}

				return sections
			},
		},
		"complete": {
			lpa: Lpa{
				You: Person{
					Address: Address{
						Line1: "this",
					},
				},
				Attorney: Attorney{
					Address: Address{
						Line1: "this",
					},
				},
				Contact:                  []string{"this"},
				WantReplacementAttorneys: "this",
				Checked:                  true,
				HappyToShare:             true,
				Tasks: Tasks{
					WhenCanTheLpaBeUsed:        TaskCompleted,
					Restrictions:               TaskCompleted,
					CertificateProvider:        TaskCompleted,
					CheckYourLpa:               TaskCompleted,
					PayForLpa:                  TaskCompleted,
					ConfirmYourIdentityAndSign: TaskCompleted,
				},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items = []taskListItem{
					{Name: "provideDonorDetails", Path: yourDetailsPath, Completed: true},
					{Name: "chooseYourAttorneys", Path: chooseAttorneysPath, Completed: true},
					{Name: "chooseYourReplacementAttorneys", Path: wantReplacementAttorneysPath, Completed: true},
					{Name: "chooseWhenTheLpaCanBeUsed", Path: whenCanTheLpaBeUsedPath, Completed: true},
					{Name: "addRestrictionsToTheLpa", Path: restrictionsPath, Completed: true},
					{Name: "chooseYourCertificateProvider", Path: whoDoYouWantToBeCertificateProviderGuidancePath, Completed: true},
					{Name: "checkAndSendToYourCertificateProvider", Path: checkYourLpaPath, Completed: true},
				}

				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: aboutPaymentPath, Completed: true},
				}

				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: selectYourIdentityOptionsPath, Completed: true},
				}

				return sections
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			dataStore := &mockDataStore{data: tc.lpa}
			dataStore.
				On("Get", mock.Anything, "session-id").
				Return(nil)

			template := &mockTemplate{}
			template.
				On("Func", w, &taskListData{
					App: appData,
					Sections: tc.expected([]taskListSection{
						{
							Heading: "fillInTheLpa",
							Items: []taskListItem{
								{Name: "provideDonorDetails", Path: yourDetailsPath},
								{Name: "chooseYourAttorneys", Path: chooseAttorneysPath},
								{Name: "chooseYourReplacementAttorneys", Path: wantReplacementAttorneysPath},
								{Name: "chooseWhenTheLpaCanBeUsed", Path: whenCanTheLpaBeUsedPath},
								{Name: "addRestrictionsToTheLpa", Path: restrictionsPath},
								{Name: "chooseYourCertificateProvider", Path: whoDoYouWantToBeCertificateProviderGuidancePath},
								{Name: "checkAndSendToYourCertificateProvider", Path: checkYourLpaPath},
							},
						},
						{
							Heading: "payForTheLpa",
							Items: []taskListItem{
								{
									Name: "payForTheLpa",
									Path: aboutPaymentPath,
								},
							},
						},
						{
							Heading: "confirmYourIdentityAndSign",
							Items: []taskListItem{
								{
									Name: "confirmYourIdentityAndSign",
									Path: selectYourIdentityOptionsPath,
								},
							},
						},
						{
							Heading: "registerTheLpa",
							Items: []taskListItem{
								{Name: "registerTheLpa", Disabled: true},
							},
						},
					}),
				}).
				Return(nil)

			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			err := TaskList(template.Func, dataStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, template, dataStore)
		})
	}
}

func TestGetTaskListWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := TaskList(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := TaskList(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}
