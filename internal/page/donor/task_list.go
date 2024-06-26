package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
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
	Name          string
	Path          string
	State         actor.TaskState
	PaymentState  actor.PaymentTask
	IdentityState actor.IdentityTask
	Count         int
	Hidden        bool
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template, evidenceReceivedStore EvidenceReceivedStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		evidenceReceived, err := evidenceReceivedStore.Get(r.Context())
		if err != nil {
			return err
		}

		section1 := taskListSection{
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
				taskListTypeSpecificStep(donor),
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
					Name:  "addCorrespondent",
					Path:  page.Paths.AddCorrespondent.Format(donor.LpaID),
					State: donor.Tasks.AddCorrespondent,
				},
				{
					Name:   "chooseYourSignatoryAndIndependentWitness",
					Path:   page.Paths.GettingHelpSigning.Format(donor.LpaID),
					State:  donor.Tasks.ChooseYourSignatory,
					Hidden: !donor.Donor.CanSign.IsNo(),
				},
			},
		}

		var sections []taskListSection
		if appData.SupporterData != nil {
			sections = []taskListSection{section1}
		} else {
			section1.Items = append(section1.Items, taskListItem{
				Name:  "checkAndSendToYourCertificateProvider",
				Path:  taskListCheckLpaPath(donor).Format(donor.LpaID),
				State: donor.Tasks.CheckYourLpa,
			})
			sections = []taskListSection{section1, taskListPaymentSection(donor), taskListSignSection(donor)}
		}

		return tmpl(w, &taskListData{
			App:              appData,
			Donor:            donor,
			EvidenceReceived: evidenceReceived,
			Sections:         sections,
		})
	}
}

func taskListTypeSpecificStep(donor *actor.DonorProvidedDetails) taskListItem {
	if donor.Type == actor.LpaTypePersonalWelfare {
		return taskListItem{
			Name:  "lifeSustainingTreatment",
			Path:  page.Paths.LifeSustainingTreatment.Format(donor.LpaID),
			State: donor.Tasks.LifeSustainingTreatment,
		}
	}

	return taskListItem{
		Name:  "chooseWhenTheLpaCanBeUsed",
		Path:  page.Paths.WhenCanTheLpaBeUsed.Format(donor.LpaID),
		State: donor.Tasks.WhenCanTheLpaBeUsed,
	}
}

func taskListCheckLpaPath(donor *actor.DonorProvidedDetails) page.LpaPath {
	if len(donor.Under18ActorDetails()) > 0 {
		return page.Paths.YouCannotSignYourLpaYet
	} else if donor.CertificateProviderSharesDetails() {
		return page.Paths.ConfirmYourCertificateProviderIsNotRelated
	} else {
		return page.Paths.CheckYourLpa
	}
}

func taskListPaymentSection(donor *actor.DonorProvidedDetails) taskListSection {
	var paymentPath string
	switch donor.Tasks.PayForLpa {
	case actor.PaymentTaskApproved:
		paymentPath = page.Paths.FeeApproved.Format(donor.LpaID)
	case actor.PaymentTaskDenied:
		paymentPath = page.Paths.FeeDenied.Format(donor.LpaID)
	case actor.PaymentTaskMoreEvidenceRequired:
		paymentPath = page.Paths.UploadEvidence.Format(donor.LpaID)
	default:
		paymentPath = page.Paths.AboutPayment.Format(donor.LpaID)
	}

	return taskListSection{
		Heading: "payForTheLpa",
		Items: []taskListItem{
			{
				Name:         "payForTheLpa",
				Path:         paymentPath,
				PaymentState: donor.Tasks.PayForLpa,
			},
		},
	}
}

func taskListSignSection(donor *actor.DonorProvidedDetails) taskListSection {
	var signPath page.LpaPath
	switch donor.DonorIdentityUserData.Status {
	case identity.StatusConfirmed:
		signPath = page.Paths.ReadYourLpa
	case identity.StatusFailed:
		signPath = page.Paths.RegisterWithCourtOfProtection
	case identity.StatusInsufficientEvidence:
		signPath = page.Paths.UnableToConfirmIdentity
		if donor.WantVoucher.IsNo() {
			signPath = page.Paths.WhatYouCanDoNow
		}
	default:
		signPath = page.Paths.HowToConfirmYourIdentityAndSign
	}

	return taskListSection{
		Heading: "confirmYourIdentityAndSign",
		Items: []taskListItem{
			{
				Name:          "confirmYourIdentityAndSign",
				Path:          signPath.Format(donor.LpaID),
				IdentityState: donor.Tasks.ConfirmYourIdentityAndSign,
			},
		},
	}
}
