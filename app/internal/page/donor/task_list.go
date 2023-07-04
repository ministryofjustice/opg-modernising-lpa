package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type taskListData struct {
	App      page.AppData
	Errors   validation.List
	Lpa      *page.Lpa
	Sections []taskListSection
}

type taskListItem struct {
	Name  string
	Path  string
	State actor.TaskState
	Count int
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		signTaskPage := page.Paths.HowToConfirmYourIdentityAndSign
		if lpa.DonorIdentityConfirmed() {
			signTaskPage = page.Paths.ReadYourLpa
		}

		typeSpecificStep := taskListItem{
			Name:  "chooseWhenTheLpaCanBeUsed",
			Path:  page.Paths.WhenCanTheLpaBeUsed.Format(lpa.ID),
			State: lpa.Tasks.WhenCanTheLpaBeUsed,
		}
		if lpa.Type == page.LpaTypeHealthWelfare {
			typeSpecificStep = taskListItem{
				Name:  "lifeSustainingTreatment",
				Path:  page.Paths.LifeSustainingTreatment.Format(lpa.ID),
				State: lpa.Tasks.LifeSustainingTreatment,
			}
		}

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{
							Name:  "provideYourDetails",
							Path:  page.Paths.YourDetails.Format(lpa.ID),
							State: lpa.Tasks.YourDetails,
						},
						{
							Name:  "chooseYourAttorneys",
							Path:  page.Paths.ChooseAttorneysGuidance.Format(lpa.ID),
							State: lpa.Tasks.ChooseAttorneys,
							Count: len(lpa.Attorneys),
						},
						{
							Name:  "chooseYourReplacementAttorneys",
							Path:  page.Paths.DoYouWantReplacementAttorneys.Format(lpa.ID),
							State: lpa.Tasks.ChooseReplacementAttorneys,
							Count: len(lpa.ReplacementAttorneys),
						},
						typeSpecificStep,
						{
							Name:  "addRestrictionsToTheLpa",
							Path:  page.Paths.Restrictions.Format(lpa.ID),
							State: lpa.Tasks.Restrictions,
						},
						{
							Name:  "chooseYourCertificateProvider",
							Path:  page.Paths.WhatACertificateProviderDoes.Format(lpa.ID),
							State: lpa.Tasks.CertificateProvider,
						},
						{
							Name:  "peopleToNotifyAboutYourLpa",
							Path:  page.Paths.DoYouWantToNotifyPeople.Format(lpa.ID),
							State: lpa.Tasks.PeopleToNotify,
							Count: len(lpa.PeopleToNotify),
						},
						{
							Name:  "checkAndSendToYourCertificateProvider",
							Path:  page.Paths.CheckYourLpa.Format(lpa.ID),
							State: lpa.Tasks.CheckYourLpa,
						},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{
							Name:  "payForTheLpa",
							Path:  page.Paths.AboutPayment.Format(lpa.ID),
							State: lpa.Tasks.PayForLpa,
						},
					},
				},
				{
					Heading: "confirmYourIdentityAndSign",
					Items: []taskListItem{
						{
							Name:  "confirmYourIdentityAndSign",
							Path:  signTaskPage.Format(lpa.ID),
							State: lpa.Tasks.ConfirmYourIdentityAndSign,
						},
					},
				},
			},
		}

		return tmpl(w, data)
	}
}
