package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		appData          page.AppData
		donor            *actor.DonorProvidedDetails
		evidenceReceived bool
		expected         func([]taskListSection) []taskListSection
	}{
		"empty": {
			appData: testAppData,
			donor:   &actor.DonorProvidedDetails{LpaID: "lpa-id", Donor: actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}}},
			expected: func(sections []taskListSection) []taskListSection {
				return sections
			},
		},
		"cannot sign": {
			appData: testAppData,
			donor:   &actor.DonorProvidedDetails{LpaID: "lpa-id", Donor: actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}, CanSign: form.No}},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items[8].Hidden = false

				return sections
			},
		},
		"evidence received": {
			appData:          testAppData,
			donor:            &actor.DonorProvidedDetails{LpaID: "lpa-id", Donor: actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}}},
			evidenceReceived: true,
			expected: func(sections []taskListSection) []taskListSection {
				return sections
			},
		},
		"more evidence required": {
			appData:          testAppData,
			donor:            &actor.DonorProvidedDetails{LpaID: "lpa-id", Donor: actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}}, Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}},
			evidenceReceived: true,
			expected: func(sections []taskListSection) []taskListSection {
				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: page.Paths.UploadEvidence.Format("lpa-id"), PaymentState: actor.PaymentTaskMoreEvidenceRequired},
				}

				return sections
			},
		},
		"fee denied": {
			appData:          testAppData,
			donor:            &actor.DonorProvidedDetails{LpaID: "lpa-id", Donor: actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}}, Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}},
			evidenceReceived: true,
			expected: func(sections []taskListSection) []taskListSection {
				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: page.Paths.FeeDenied.Format("lpa-id"), PaymentState: actor.PaymentTaskDenied},
				}

				return sections
			},
		},
		"fee approved": {
			appData:          testAppData,
			donor:            &actor.DonorProvidedDetails{LpaID: "lpa-id", Donor: actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}}, Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskApproved}},
			evidenceReceived: true,
			expected: func(sections []taskListSection) []taskListSection {
				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: page.Paths.FeeApproved.Format("lpa-id"), PaymentState: actor.PaymentTaskApproved},
				}

				return sections
			},
		},
		"personal welfare": {
			appData: testAppData,
			donor:   &actor.DonorProvidedDetails{LpaID: "lpa-id", Type: actor.LpaTypePersonalWelfare, Donor: actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}}},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items[3] = taskListItem{
					Name: "lifeSustainingTreatment",
					Path: page.Paths.LifeSustainingTreatment.Format("lpa-id"),
				}

				return sections
			},
		},
		"confirmed identity": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:                 "lpa-id",
				Donor:                 actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}},
				DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed, LastName: "a"},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.ReadYourLpa.Format("lpa-id")},
				}

				return sections
			},
		},
		"confirmed identity does not match LPA": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:                 "lpa-id",
				Donor:                 actor.Donor{LastName: "b", Address: place.Address{Line1: "x"}},
				DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed, LastName: "a"},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.OneLoginIdentityDetails.Format("lpa-id")},
				}

				return sections
			},
		},
		"failed identity": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:                 "lpa-id",
				Donor:                 actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}},
				DonorIdentityUserData: identity.UserData{Status: identity.StatusFailed, LastName: "a"},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.RegisterWithCourtOfProtection.Format("lpa-id")},
				}

				return sections
			},
		},
		"insufficient evidence for identity": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:                 "lpa-id",
				Donor:                 actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}},
				DonorIdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence, LastName: "a"},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.UnableToConfirmIdentity.Format("lpa-id")},
				}

				return sections
			},
		},
		"insufficient evidence for identity with voucher details": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:                 "lpa-id",
				Donor:                 actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}},
				DonorIdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence, LastName: "a"},
				Voucher:               actor.Voucher{FirstNames: "a"},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.ReadYourLpa.Format("lpa-id")},
				}

				return sections
			},
		},
		"does not want a voucher": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:                 "lpa-id",
				Donor:                 actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}},
				DonorIdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence, LastName: "a"},
				WantVoucher:           form.No,
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.WhatYouCanDoNow.Format("lpa-id")},
				}

				return sections
			},
		},
		"is applying to court of protection": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:                            "lpa-id",
				Donor:                            actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}},
				DonorIdentityUserData:            identity.UserData{Status: identity.StatusInsufficientEvidence, LastName: "a"},
				WantVoucher:                      form.No,
				RegisteringWithCourtOfProtection: true,
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.WhatHappensNextRegisteringWithCourtOfProtection.Format("lpa-id")},
				}

				return sections
			},
		},
		"is applying to court of protection and has signed": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:                            "lpa-id",
				Donor:                            actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}},
				DonorIdentityUserData:            identity.UserData{Status: identity.StatusInsufficientEvidence, LastName: "a"},
				WantVoucher:                      form.No,
				RegisteringWithCourtOfProtection: true,
				SignedAt:                         testNow,
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[2].Items = []taskListItem{
					{Name: "confirmYourIdentityAndSign", Path: page.Paths.YouHaveSubmittedYourLpa.Format("lpa-id")},
				}

				return sections
			},
		},
		"attorneys under 18": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{LastName: "a", Address: place.Address{Line1: "xx"}},
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{FirstNames: "aa", LastName: "bb", DateOfBirth: date.Today().AddDate(-17, 0, 0), Address: place.Address{Line1: "zz"}},
				}},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items[1] = taskListItem{
					Name:  "chooseYourAttorneys",
					Path:  page.Paths.ChooseAttorneysGuidance.Format("lpa-id"),
					State: actor.TaskNotStarted,
					Count: 1,
				}

				sections[0].Items[9] = taskListItem{Name: "checkAndSendToYourCertificateProvider", Path: page.Paths.YouCannotSignYourLpaYet.Format("lpa-id")}

				return sections
			},
		},
		"certificate provider has similar name": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:               "lpa-id",
				Donor:               actor.Donor{LastName: "a", Address: place.Address{Line1: "x"}},
				CertificateProvider: actor.CertificateProvider{LastName: "a"},
			},
			expected: func(sections []taskListSection) []taskListSection {
				sections[0].Items[9].Path = page.Paths.ConfirmYourCertificateProviderIsNotRelated.Format("lpa-id")

				return sections
			},
		},
		"mixed": {
			appData: testAppData,
			donor: &actor.DonorProvidedDetails{
				LpaID:               "lpa-id",
				Donor:               actor.Donor{FirstNames: "this"},
				CertificateProvider: actor.CertificateProvider{LastName: "a", Address: place.Address{Line1: "x"}},
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{DateOfBirth: date.Today().AddDate(-20, 0, 0)},
					{DateOfBirth: date.Today().AddDate(-20, 0, 0)},
				}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{DateOfBirth: date.Today().AddDate(-20, 0, 0)},
				}},
				Tasks: actor.DonorTasks{
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
					{Name: "addCorrespondent", Path: page.Paths.AddCorrespondent.Format("lpa-id")},
					{Name: "chooseYourSignatoryAndIndependentWitness", Path: page.Paths.GettingHelpSigning.Format("lpa-id"), Hidden: true},
					{Name: "checkAndSendToYourCertificateProvider", Path: page.Paths.CheckYourLpa.Format("lpa-id"), State: actor.TaskCompleted},
				}

				sections[1].Items = []taskListItem{
					{Name: "payForTheLpa", Path: page.Paths.AboutPayment.Format("lpa-id"), PaymentState: actor.PaymentTaskInProgress},
				}

				return sections
			},
		},
		"supporter": {
			appData: page.AppData{
				SessionID:     "session-id",
				LpaID:         "lpa-id",
				Lang:          localize.En,
				SupporterData: &page.SupporterData{},
			},
			donor: &actor.DonorProvidedDetails{
				LpaID:               "lpa-id",
				Donor:               actor.Donor{FirstNames: "this"},
				CertificateProvider: actor.CertificateProvider{LastName: "a", Address: place.Address{Line1: "x"}},
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{DateOfBirth: date.Today().AddDate(-20, 0, 0)},
					{DateOfBirth: date.Today().AddDate(-20, 0, 0)},
				}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
					{DateOfBirth: date.Today().AddDate(-20, 0, 0)},
				}},
				Tasks: actor.DonorTasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskInProgress,
					WhenCanTheLpaBeUsed:        actor.TaskInProgress,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskInProgress,
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
					{Name: "addCorrespondent", Path: page.Paths.AddCorrespondent.Format("lpa-id"), State: actor.TaskNotStarted},
					{Name: "chooseYourSignatoryAndIndependentWitness", Path: page.Paths.GettingHelpSigning.Format("lpa-id"), Hidden: true},
				}

				return sections[0:1]
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &taskListData{
					App:              tc.appData,
					Donor:            tc.donor,
					EvidenceReceived: tc.evidenceReceived,
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
								{Name: "addCorrespondent", Path: page.Paths.AddCorrespondent.Format("lpa-id")},
								{Name: "chooseYourSignatoryAndIndependentWitness", Path: page.Paths.GettingHelpSigning.Format("lpa-id"), Hidden: true},
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

			evidenceReceivedStore := newMockEvidenceReceivedStore(t)
			evidenceReceivedStore.EXPECT().
				Get(r.Context()).
				Return(tc.evidenceReceived, nil)

			err := TaskList(template.Execute, evidenceReceivedStore)(tc.appData, w, r, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenEvidenceReceivedStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	evidenceReceivedStore := newMockEvidenceReceivedStore(t)
	evidenceReceivedStore.EXPECT().
		Get(r.Context()).
		Return(false, expectedError)

	err := TaskList(nil, evidenceReceivedStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	evidenceReceivedStore := newMockEvidenceReceivedStore(t)
	evidenceReceivedStore.EXPECT().
		Get(r.Context()).
		Return(false, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute, evidenceReceivedStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
