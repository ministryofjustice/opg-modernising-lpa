package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		lpa      *page.Lpa
		expected func([]taskListSection) []taskListSection
	}{
		"empty": {
			lpa: &page.Lpa{},
			expected: func(sections []taskListSection) []taskListSection {
				return sections
			},
		},
		"mixed": {
			lpa: &page.Lpa{
				Donor: actor.Person{
					FirstNames: "this",
				},
				Attorneys:            actor.Attorneys{{}, {}},
				ReplacementAttorneys: actor.Attorneys{{}},
				Tasks: page.Tasks{
					YourDetails:                page.TaskCompleted,
					ChooseAttorneys:            page.TaskCompleted,
					ChooseReplacementAttorneys: page.TaskInProgress,
					WhenCanTheLpaBeUsed:        page.TaskInProgress,
					Restrictions:               page.TaskCompleted,
					CertificateProvider:        page.TaskInProgress,
					CheckYourLpa:               page.TaskCompleted,
					PayForLpa:                  page.TaskInProgress,
				},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items = []taskListItem{
					{Name: "provideYourDetails", Path: page.Paths.YourDetails, State: page.TaskCompleted},
					{Name: "chooseYourAttorneys", Path: page.Paths.ChooseAttorneys, State: page.TaskCompleted, Count: 2},
					{Name: "chooseYourReplacementAttorneys", Path: page.Paths.DoYouWantReplacementAttorneys, State: page.TaskInProgress, Count: 1},
					{Name: "chooseWhenTheLpaCanBeUsed", Path: page.Paths.WhenCanTheLpaBeUsed, State: page.TaskInProgress},
					{Name: "addRestrictionsToTheLpa", Path: page.Paths.Restrictions, State: page.TaskCompleted},
					{Name: "chooseYourCertificateProvider", Path: page.Paths.WhoDoYouWantToBeCertificateProviderGuidance, State: page.TaskInProgress},
					{Name: "peopleToNotify", Path: page.Paths.DoYouWantToNotifyPeople},
					{Name: "checkAndSendToYourCertificateProvider", Path: page.Paths.CheckYourLpa, State: page.TaskCompleted},
				}

				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: page.Paths.AboutPayment, State: page.TaskInProgress},
				}

				return sections
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
					App: testAppData,
					Lpa: tc.lpa,
					Sections: tc.expected([]taskListSection{
						{
							Heading: "fillInTheLpa",
							Items: []taskListItem{
								{Name: "provideYourDetails", Path: page.Paths.YourDetails},
								{Name: "chooseYourAttorneys", Path: page.Paths.ChooseAttorneys},
								{Name: "chooseYourReplacementAttorneys", Path: page.Paths.DoYouWantReplacementAttorneys},
								{Name: "chooseWhenTheLpaCanBeUsed", Path: page.Paths.WhenCanTheLpaBeUsed},
								{Name: "addRestrictionsToTheLpa", Path: page.Paths.Restrictions},
								{Name: "chooseYourCertificateProvider", Path: page.Paths.WhoDoYouWantToBeCertificateProviderGuidance},
								{Name: "peopleToNotify", Path: page.Paths.DoYouWantToNotifyPeople},
								{Name: "checkAndSendToYourCertificateProvider", Path: page.Paths.CheckYourLpa},
							},
						},
						{
							Heading: "payForTheLpa",
							Items: []taskListItem{
								{Name: "payForTheLpa", Path: page.Paths.AboutPayment},
							},
						},
						{
							Heading: "confirmYourIdentityAndSign",
							Items: []taskListItem{
								{Name: "confirmYourIdentityAndSign", Path: page.Paths.HowToConfirmYourIdentityAndSign},
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

			err := TaskList(template.Execute, lpaStore)(testAppData, w, r)
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
