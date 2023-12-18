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
	Donor            *actor.DonorProvidedDetails
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
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		signTaskPage := page.Paths.HowToConfirmYourIdentityAndSign

		if donor.DonorIdentityConfirmed() {
			signTaskPage = page.Paths.ReadYourLpa
		}

		typeSpecificStep := taskListItem{
			Name:  "chooseWhenTheLpaCanBeUsed",
			Path:  page.Paths.WhenCanTheLpaBeUsed.Format(donor.LpaID),
			State: donor.Tasks.WhenCanTheLpaBeUsed,
		}
		if donor.Type == actor.LpaTypePersonalWelfare {
			typeSpecificStep = taskListItem{
				Name:  "lifeSustainingTreatment",
				Path:  page.Paths.LifeSustainingTreatment.Format(donor.LpaID),
				State: donor.Tasks.LifeSustainingTreatment,
			}
		}

		evidenceReceived, err := evidenceReceivedStore.Get(r.Context())
		if err != nil {
			return err
		}

		var paymentPath string
		switch donor.Tasks.PayForLpa {
		case actor.PaymentTaskDenied:
			paymentPath = page.Paths.FeeDenied.Format(donor.LpaID)
		case actor.PaymentTaskMoreEvidenceRequired:
			paymentPath = page.Paths.UploadEvidence.Format(donor.LpaID)
		default:
			paymentPath = page.Paths.AboutPayment.Format(donor.LpaID)
		}

		checkPath := page.Paths.CheckYourLpa

		if len(donor.Under18ActorDetails()) > 0 {
			checkPath = page.Paths.YouCannotSignYourLpaYet
		} else if donor.CertificateProviderSharesDetails() {
			checkPath = page.Paths.ConfirmYourCertificateProviderIsNotRelated
		}

		data := &taskListData{
			App:              appData,
			Donor:            donor,
			EvidenceReceived: evidenceReceived,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{
							Name:  "provideYourDetails",
							Path:  page.Paths.YourDetails.Format(donor.LpaID),
							State: donor.Tasks.YourDetails,
						},
						{
							Name:  "chooseYourAttorneys",
							Path:  page.Paths.ChooseAttorneysGuidance.Format(donor.LpaID),
							State: donor.Tasks.ChooseAttorneys,
							Count: donor.Attorneys.Len(),
						},
						{
							Name:  "chooseYourReplacementAttorneys",
							Path:  page.Paths.DoYouWantReplacementAttorneys.Format(donor.LpaID),
							State: donor.Tasks.ChooseReplacementAttorneys,
							Count: donor.ReplacementAttorneys.Len(),
						},
						typeSpecificStep,
						{
							Name:  "addRestrictionsToTheLpa",
							Path:  page.Paths.Restrictions.Format(donor.LpaID),
							State: donor.Tasks.Restrictions,
						},
						{
							Name:  "chooseYourCertificateProvider",
							Path:  page.Paths.WhatACertificateProviderDoes.Format(donor.LpaID),
							State: donor.Tasks.CertificateProvider,
						},
						{
							Name:  "peopleToNotifyAboutYourLpa",
							Path:  page.Paths.DoYouWantToNotifyPeople.Format(donor.LpaID),
							State: donor.Tasks.PeopleToNotify,
							Count: len(donor.PeopleToNotify),
						},
						{
							Name:   "chooseYourSignatoryAndIndependentWitness",
							Path:   page.Paths.GettingHelpSigning.Format(donor.LpaID),
							State:  donor.Tasks.ChooseYourSignatory,
							Hidden: !donor.Donor.CanSign.IsNo(),
						},
						{
							Name:  "checkAndSendToYourCertificateProvider",
							Path:  checkPath.Format(donor.LpaID),
							State: donor.Tasks.CheckYourLpa,
						},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{
							Name:         "payForTheLpa",
							Path:         paymentPath,
							PaymentState: donor.Tasks.PayForLpa,
						},
					},
				},
				{
					Heading: "confirmYourIdentityAndSign",
					Items: []taskListItem{
						{
							Name:  "confirmYourIdentityAndSign",
							Path:  signTaskPage.Format(donor.LpaID),
							State: donor.Tasks.ConfirmYourIdentityAndSign,
						},
					},
				},
			},
		}

		return tmpl(w, data)
	}
}
