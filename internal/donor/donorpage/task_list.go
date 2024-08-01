package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App              page.AppData
	Errors           validation.List
	Donor            *donordata.DonorProvidedDetails
	Sections         []taskListSection
	EvidenceReceived bool
}

type taskListItem struct {
	Name          string
	Path          string
	State         actor.TaskState
	PaymentState  task.PaymentState
	IdentityState actor.IdentityTask
	Count         int
	Hidden        bool
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template, evidenceReceivedStore EvidenceReceivedStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		evidenceReceived, err := evidenceReceivedStore.Get(r.Context())
		if err != nil {
			return err
		}

		chooseAttorneysLink := page.Paths.ChooseAttorneysGuidance.Format(donor.LpaID)
		if donor.Attorneys.Len() > 0 {
			chooseAttorneysLink = page.Paths.ChooseAttorneysSummary.Format(donor.LpaID)
		}

		chooseReplacementAttorneysLink := page.Paths.DoYouWantReplacementAttorneys.Format(donor.LpaID)
		if donor.ReplacementAttorneys.Len() > 0 {
			chooseReplacementAttorneysLink = page.Paths.ChooseReplacementAttorneysSummary.Format(donor.LpaID)
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
					Path:  chooseAttorneysLink,
					State: donor.Tasks.ChooseAttorneys,
					Count: donor.Attorneys.Len(),
				},
				{
					Name:  "chooseYourReplacementAttorneys",
					Path:  chooseReplacementAttorneysLink,
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

func taskListTypeSpecificStep(donor *donordata.DonorProvidedDetails) taskListItem {
	if donor.Type == donordata.LpaTypePersonalWelfare {
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

func taskListCheckLpaPath(donor *donordata.DonorProvidedDetails) page.LpaPath {
	if len(donor.Under18ActorDetails()) > 0 {
		return page.Paths.YouCannotSignYourLpaYet
	} else if donor.CertificateProviderSharesDetails() {
		return page.Paths.ConfirmYourCertificateProviderIsNotRelated
	} else {
		return page.Paths.CheckYourLpa
	}
}

func taskListPaymentSection(donor *donordata.DonorProvidedDetails) taskListSection {
	var paymentPath string
	switch donor.Tasks.PayForLpa {
	case task.PaymentStateApproved:
		paymentPath = page.Paths.FeeApproved.Format(donor.LpaID)
	case task.PaymentStateDenied:
		paymentPath = page.Paths.FeeDenied.Format(donor.LpaID)
	case task.PaymentStateMoreEvidenceRequired:
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

func taskListSignSection(donor *donordata.DonorProvidedDetails) taskListSection {
	var signPath page.LpaPath

	switch donor.DonorIdentityUserData.Status {
	case identity.StatusConfirmed:
		if !donor.SignedAt.IsZero() {
			signPath = page.Paths.YouHaveSubmittedYourLpa
		} else if donor.DonorIdentityConfirmed() {
			signPath = page.Paths.ReadYourLpa
		} else {
			signPath = page.Paths.OneLoginIdentityDetails
		}

	case identity.StatusFailed:
		signPath = page.Paths.RegisterWithCourtOfProtection

	case identity.StatusInsufficientEvidence:
		if !donor.SignedAt.IsZero() {
			signPath = page.Paths.YouHaveSubmittedYourLpa
		} else if donor.RegisteringWithCourtOfProtection {
			signPath = page.Paths.WhatHappensNextRegisteringWithCourtOfProtection
		} else if donor.Voucher.FirstNames != "" {
			signPath = page.Paths.ReadYourLpa
		} else if donor.WantVoucher.IsYes() {
			signPath = page.Paths.EnterVoucher
		} else if donor.WantVoucher.IsNo() {
			signPath = page.Paths.WhatYouCanDoNow
		} else {
			signPath = page.Paths.UnableToConfirmIdentity
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
