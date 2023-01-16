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
		lpa      *Lpa
		expected func([]taskListSection) []taskListSection
	}{
		"empty": {
			lpa: &Lpa{},
			expected: func(sections []taskListSection) []taskListSection {
				return sections
			},
		},
		"mixed": {
			lpa: &Lpa{
				You: Person{
					FirstNames: "this",
				},
				Attorneys:            []Attorney{{}, {}},
				ReplacementAttorneys: []Attorney{{}},
				Tasks: Tasks{
					YourDetails:                TaskCompleted,
					ChooseAttorneys:            TaskCompleted,
					ChooseReplacementAttorneys: TaskInProgress,
					WhenCanTheLpaBeUsed:        TaskInProgress,
					Restrictions:               TaskCompleted,
					CertificateProvider:        TaskInProgress,
					CheckYourLpa:               TaskCompleted,
					PayForLpa:                  TaskInProgress,
				},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items = []taskListItem{
					{Name: "provideYourDetails", Path: Paths.YourDetails, State: TaskCompleted},
					{Name: "chooseYourAttorneys", Path: Paths.ChooseAttorneys, State: TaskCompleted, Count: 2},
					{Name: "chooseYourReplacementAttorneys", Path: Paths.DoYouWantReplacementAttorneys, State: TaskInProgress, Count: 1},
					{Name: "chooseWhenTheLpaCanBeUsed", Path: Paths.WhenCanTheLpaBeUsed, State: TaskInProgress},
					{Name: "addRestrictionsToTheLpa", Path: Paths.Restrictions, State: TaskCompleted},
					{Name: "chooseYourCertificateProvider", Path: Paths.WhoDoYouWantToBeCertificateProviderGuidance, State: TaskInProgress},
					{Name: "peopleToNotify", Path: Paths.DoYouWantToNotifyPeople},
					{Name: "checkAndSendToYourCertificateProvider", Path: Paths.CheckYourLpa, State: TaskCompleted},
				}

				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: Paths.AboutPayment, State: TaskInProgress},
				}

				return sections
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(tc.lpa, nil)

			template := &mockTemplate{}
			template.
				On("Func", w, &taskListData{
					App: appData,
					Lpa: tc.lpa,
					Sections: tc.expected([]taskListSection{
						{
							Heading: "fillInTheLpa",
							Items: []taskListItem{
								{Name: "provideYourDetails", Path: Paths.YourDetails},
								{Name: "chooseYourAttorneys", Path: Paths.ChooseAttorneys},
								{Name: "chooseYourReplacementAttorneys", Path: Paths.DoYouWantReplacementAttorneys},
								{Name: "chooseWhenTheLpaCanBeUsed", Path: Paths.WhenCanTheLpaBeUsed},
								{Name: "addRestrictionsToTheLpa", Path: Paths.Restrictions},
								{Name: "chooseYourCertificateProvider", Path: Paths.WhoDoYouWantToBeCertificateProviderGuidance},
								{Name: "peopleToNotify", Path: Paths.DoYouWantToNotifyPeople},
								{Name: "checkAndSendToYourCertificateProvider", Path: Paths.CheckYourLpa},
							},
						},
						{
							Heading: "payForTheLpa",
							Items: []taskListItem{
								{Name: "payForTheLpa", Path: Paths.AboutPayment},
							},
						},
						{
							Heading: "confirmYourIdentityAndSign",
							Items: []taskListItem{
								{Name: "confirmYourIdentityAndSign", Path: Paths.HowToConfirmYourIdentityAndSign},
							},
						},
						{
							Heading: "registerTheLpa",
							Items: []taskListItem{
								{Name: "registerTheLpa"},
							},
						},
					}),
				}).
				Return(nil)

			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			err := TaskList(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, template, lpaStore)
		})
	}
}

func TestGetTaskListWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := TaskList(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := TaskList(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}
