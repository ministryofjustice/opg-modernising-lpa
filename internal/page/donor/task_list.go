package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App              page.AppData
	Errors           validation.List
	Lpa              *page.Lpa
	Sections         []taskListSection
	EvidenceReceived bool
}

type taskListItem struct {
	Name         string
	Path         string
	State        actor.TaskState
	PaymentState actor.PaymentTask
	Count        int
	Hidden       bool
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template, evidenceReceivedStore EvidenceReceivedStore) Handler {
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

		evidenceReceived, err := evidenceReceivedStore.Get(r.Context())
		if err != nil {
			return err
		}

		var paymentPath string
		switch lpa.Tasks.PayForLpa {
		case actor.PaymentTaskDenied:
			paymentPath = page.Paths.FeeDenied.Format(lpa.ID)
		case actor.PaymentTaskMoreEvidenceRequired:
			paymentPath = page.Paths.UploadEvidence.Format(lpa.ID)
		default:
			paymentPath = page.Paths.AboutPayment.Format(lpa.ID)
		}

		checkPath := page.Paths.CheckYourLpa
		if lpa.CertificateProviderSharesDetails() {
			checkPath = page.Paths.ConfirmYourCertificateProviderIsNotRelated
		}

		data := &taskListData{
			App:              appData,
			Lpa:              lpa,
			EvidenceReceived: evidenceReceived,
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
							Count: lpa.Attorneys.Len(),
						},
						{
							Name:  "chooseYourReplacementAttorneys",
							Path:  page.Paths.DoYouWantReplacementAttorneys.Format(lpa.ID),
							State: lpa.Tasks.ChooseReplacementAttorneys,
							Count: lpa.ReplacementAttorneys.Len(),
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
							Name:   "chooseYourSignatoryAndIndependentWitness",
							Path:   page.Paths.GettingHelpSigning.Format(lpa.ID),
							State:  lpa.Tasks.ChooseYourSignatory,
							Hidden: !lpa.Donor.CanSign.IsNo(),
						},
						{
							Name:  "checkAndSendToYourCertificateProvider",
							Path:  checkPath.Format(lpa.ID),
							State: lpa.Tasks.CheckYourLpa,
						},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{
							Name:         "payForTheLpa",
							Path:         paymentPath,
							PaymentState: lpa.Tasks.PayForLpa,
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
