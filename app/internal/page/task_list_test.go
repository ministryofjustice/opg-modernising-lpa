package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		expected func([]taskListSection) []taskListSection
	}{
		"start": {
			lpa: &Lpa{},
			expected: func(sections []taskListSection) []taskListSection {
				return sections
			},
		},
		"in-progress": {
			lpa: &Lpa{
				You: Person{
					FirstNames: "this",
				},
				Attorneys: []Attorney{
					{FirstNames: "this"},
				},
				ReplacementAttorneys: []Attorney{
					{FirstNames: "this"},
				},
				Tasks: Tasks{
					WhenCanTheLpaBeUsed:        TaskInProgress,
					Restrictions:               TaskInProgress,
					CertificateProvider:        TaskInProgress,
					CheckYourLpa:               TaskInProgress,
					PayForLpa:                  TaskInProgress,
					ConfirmYourIdentityAndSign: TaskInProgress,
					PeopleToNotify:             TaskInProgress,
				},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items = []taskListItem{
					{Name: ProvideYourDetailsTask, Path: appData.Paths.YourDetails, InProgress: true},
					{Name: ChooseYourAttorneysTask, Path: appData.Paths.ChooseAttorneys, InProgress: true, Count: 1},
					{Name: ChooseYourReplacementAttorneysTask, Path: appData.Paths.DoYouWantReplacementAttorneys, InProgress: true, Count: 1},
					{Name: ChooseWhenTheLpaCanBeUsedTask, Path: appData.Paths.WhenCanTheLpaBeUsed, InProgress: true},
					{Name: AddRestrictionsToLpaTask, Path: appData.Paths.Restrictions, InProgress: true},
					{Name: ChooseCertificateProviderTask, Path: appData.Paths.WhoDoYouWantToBeCertificateProviderGuidance, InProgress: true},
					{Name: PeopleToNotifyTask, Path: appData.Paths.DoYouWantToNotifyPeople, InProgress: true, Count: 0},
					{Name: CheckAndSendToCertificateProviderTask, Path: appData.Paths.CheckYourLpa, InProgress: true},
				}

				sections[1].Items = []taskListItem{
					{Name: PayForTheLpaTask, Path: appData.Paths.AboutPayment, InProgress: true},
				}

				sections[2].Items = []taskListItem{
					{Name: ConfirmYourIdentityAndSignTask, Path: appData.Paths.SelectYourIdentityOptions, InProgress: true},
				}

				return sections
			},
		},
		"complete": {
			lpa: &Lpa{
				You: Person{
					Address: place.Address{
						Line1: "this",
					},
				},
				Attorneys: []Attorney{
					validAttorney,
					validAttorney,
				},
				ReplacementAttorneys: []Attorney{
					validAttorney,
				},
				PeopleToNotify: []PersonToNotify{
					validPersonToNotify,
					validPersonToNotify,
					validPersonToNotify,
				},
				Contact:                                     []string{"this"},
				HowAttorneysMakeDecisions:                   "jointly",
				WantReplacementAttorneys:                    "yes",
				HowReplacementAttorneysMakeDecisions:        "mixed",
				HowReplacementAttorneysMakeDecisionsDetails: "some details",
				Checked:      true,
				HappyToShare: true,
				Tasks: Tasks{
					WhenCanTheLpaBeUsed:        TaskCompleted,
					Restrictions:               TaskCompleted,
					CertificateProvider:        TaskCompleted,
					CheckYourLpa:               TaskCompleted,
					PayForLpa:                  TaskCompleted,
					ConfirmYourIdentityAndSign: TaskCompleted,
					PeopleToNotify:             TaskCompleted,
				},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items = []taskListItem{
					{Name: ProvideYourDetailsTask, Path: appData.Paths.YourDetails, Completed: true},
					{Name: ChooseYourAttorneysTask, Path: appData.Paths.ChooseAttorneys, Completed: true, Count: 2},
					{Name: ChooseYourReplacementAttorneysTask, Path: appData.Paths.DoYouWantReplacementAttorneys, Completed: true, Count: 1},
					{Name: ChooseWhenTheLpaCanBeUsedTask, Path: appData.Paths.WhenCanTheLpaBeUsed, Completed: true},
					{Name: AddRestrictionsToLpaTask, Path: appData.Paths.Restrictions, Completed: true},
					{Name: ChooseCertificateProviderTask, Path: appData.Paths.WhoDoYouWantToBeCertificateProviderGuidance, Completed: true},
					{Name: PeopleToNotifyTask, Path: appData.Paths.DoYouWantToNotifyPeople, Completed: true, Count: 3},
					{Name: CheckAndSendToCertificateProviderTask, Path: appData.Paths.CheckYourLpa, Completed: true},
				}

				sections[1].Items = []taskListItem{
					{Name: PayForTheLpaTask, Path: appData.Paths.AboutPayment, Completed: true},
				}

				sections[2].Items = []taskListItem{
					{Name: ConfirmYourIdentityAndSignTask, Path: appData.Paths.SelectYourIdentityOptions, Completed: true},
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
							Heading: FillInLpaSection,
							Items: []taskListItem{
								{Name: ProvideYourDetailsTask, Path: appData.Paths.YourDetails},
								{Name: ChooseYourAttorneysTask, Path: appData.Paths.ChooseAttorneys},
								{Name: ChooseYourReplacementAttorneysTask, Path: appData.Paths.DoYouWantReplacementAttorneys},
								{Name: ChooseWhenTheLpaCanBeUsedTask, Path: appData.Paths.WhenCanTheLpaBeUsed},
								{Name: AddRestrictionsToLpaTask, Path: appData.Paths.Restrictions},
								{Name: ChooseCertificateProviderTask, Path: appData.Paths.WhoDoYouWantToBeCertificateProviderGuidance},
								{Name: PeopleToNotifyTask, Path: appData.Paths.DoYouWantToNotifyPeople},
								{Name: CheckAndSendToCertificateProviderTask, Path: appData.Paths.CheckYourLpa},
							},
						},
						{
							Heading: PayForLpaSection,
							Items: []taskListItem{
								{Name: PayForTheLpaTask, Path: appData.Paths.AboutPayment},
							},
						},
						{
							Heading: ConfirmYourIdentityAndSignSection,
							Items: []taskListItem{
								{
									Name: ConfirmYourIdentityAndSignTask, Path: appData.Paths.SelectYourIdentityOptions,
								},
							},
						},
						{
							Heading: RegisterTheLpaSection,
							Items: []taskListItem{
								{Name: RegisterTheLpaTask},
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
