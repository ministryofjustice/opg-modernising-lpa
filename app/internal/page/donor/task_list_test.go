package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		lpa      *page.Lpa
		expected func([]taskListSection) []taskListSection
	}{
		"empty": {
			lpa: &page.Lpa{ID: "lpa-id"},
			expected: func(sections []taskListSection) []taskListSection {
				return sections
			},
		},
		"hw": {
			lpa: &page.Lpa{ID: "lpa-id", Type: page.LpaTypeHealthWelfare},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items[3] = taskListItem{
					Name: "lifeSustainingTreatment",
					Path: page.Paths.LifeSustainingTreatment.Format("lpa-id"),
				}

				return sections
			},
		},
		"confirmed identity": {
			lpa: &page.Lpa{
				ID:                    "lpa-id",
				DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.ReadYourLpa.Format("lpa-id")},
				}

				return sections
			},
		},
		"mixed": {
			lpa: &page.Lpa{
				ID: "lpa-id",
				Donor: actor.Donor{
					FirstNames: "this",
				},
				Attorneys:            actor.Attorneys{Attorneys: []actor.Attorney{{}, {}}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
				Tasks: page.Tasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskInProgress,
					WhenCanTheLpaBeUsed:        actor.TaskInProgress,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskInProgress,
					CheckYourLpa:               actor.TaskCompleted,
					PayForLpa:                  actor.PaymentTaskInProgress,
				},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items = []taskListItem{
					{Name: "provideYourDetails", Path: page.Paths.YourDetails.Format("lpa-id"), State: actor.TaskCompleted},
					{Name: "chooseYourAttorneys", Path: page.Paths.ChooseAttorneysGuidance.Format("lpa-id"), State: actor.TaskCompleted, Count: 2},
					{Name: "chooseYourReplacementAttorneys", Path: page.Paths.DoYouWantReplacementAttorneys.Format("lpa-id"), State: actor.TaskInProgress, Count: 1},
					{Name: "chooseWhenTheLpaCanBeUsed", Path: page.Paths.WhenCanTheLpaBeUsed.Format("lpa-id"), State: actor.TaskInProgress},
					{Name: "addRestrictionsToTheLpa", Path: page.Paths.Restrictions.Format("lpa-id"), State: actor.TaskCompleted},
					{Name: "chooseYourCertificateProvider", Path: page.Paths.WhatACertificateProviderDoes.Format("lpa-id"), State: actor.TaskInProgress},
					{Name: "peopleToNotifyAboutYourLpa", Path: page.Paths.DoYouWantToNotifyPeople.Format("lpa-id")},
					{Name: "checkAndSendToYourCertificateProvider", Path: page.Paths.CheckYourLpa.Format("lpa-id"), State: actor.TaskCompleted},
				}

				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: page.Paths.AboutPayment.Format("lpa-id"), PaymentState: actor.PaymentTaskInProgress},
				}

				return sections
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &taskListData{
					App: testAppData,
					Lpa: tc.lpa,
					Sections: tc.expected([]taskListSection{
						{
							Heading: "fillInTheLpa",
							Items: []taskListItem{
								{Name: "provideYourDetails", Path: page.Paths.YourDetails.Format("lpa-id")},
								{Name: "chooseYourAttorneys", Path: page.Paths.ChooseAttorneysGuidance.Format("lpa-id")},
								{Name: "chooseYourReplacementAttorneys", Path: page.Paths.DoYouWantReplacementAttorneys.Format("lpa-id")},
								{Name: "chooseWhenTheLpaCanBeUsed", Path: page.Paths.WhenCanTheLpaBeUsed.Format("lpa-id")},
								{Name: "addRestrictionsToTheLpa", Path: page.Paths.Restrictions.Format("lpa-id")},
								{Name: "chooseYourCertificateProvider", Path: page.Paths.WhatACertificateProviderDoes.Format("lpa-id")},
								{Name: "peopleToNotifyAboutYourLpa", Path: page.Paths.DoYouWantToNotifyPeople.Format("lpa-id")},
								{Name: "checkAndSendToYourCertificateProvider", Path: page.Paths.CheckYourLpa.Format("lpa-id")},
							},
						},
						{
							Heading: "payForTheLpa",
							Items: []taskListItem{
								{Name: "payForTheLpa", Path: page.Paths.AboutPayment.Format("lpa-id")},
							},
						},
						{
							Heading: "confirmYourIdentityAndSign",
							Items: []taskListItem{
								{Name: "confirmYourIdentityAndSign", Path: page.Paths.HowToConfirmYourIdentityAndSign.Format("lpa-id")},
							},
						},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute)(testAppData, w, r, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
