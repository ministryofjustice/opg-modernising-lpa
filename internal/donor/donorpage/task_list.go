package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App              appcontext.Data
	Errors           validation.List
	Donor            *donordata.Provided
	Sections         []taskListSection
	EvidenceReceived bool
}

type taskListItem struct {
	Name          string
	Path          donor.Path
	State         task.State
	PaymentState  task.PaymentState
	IdentityState task.IdentityState
	Count         int
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template, evidenceReceivedStore EvidenceReceivedStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		evidenceReceived, err := evidenceReceivedStore.Get(r.Context())
		if err != nil {
			return err
		}

		chooseAttorneysLink := donor.PathChooseAttorneysGuidance
		if provided.Attorneys.Len() > 0 {
			chooseAttorneysLink = donor.PathChooseAttorneysSummary
		}

		chooseReplacementAttorneysLink := donor.PathDoYouWantReplacementAttorneys
		if provided.ReplacementAttorneys.Len() > 0 {
			chooseReplacementAttorneysLink = donor.PathChooseReplacementAttorneysSummary
		}

		section1 := taskListSection{
			Heading: "fillInTheLpa",
			Items: []taskListItem{
				{
					Name:  "provideYourDetails",
					Path:  donor.PathYourDetails,
					State: provided.Tasks.YourDetails,
				},
				{
					Name:  "chooseYourAttorneys",
					Path:  chooseAttorneysLink,
					State: provided.Tasks.ChooseAttorneys,
					Count: provided.Attorneys.Len(),
				},
				{
					Name:  "chooseYourReplacementAttorneys",
					Path:  chooseReplacementAttorneysLink,
					State: provided.Tasks.ChooseReplacementAttorneys,
					Count: provided.ReplacementAttorneys.Len(),
				},
				taskListTypeSpecificStep(provided),
				{
					Name:  "addRestrictionsToTheLpa",
					Path:  donor.PathRestrictions,
					State: provided.Tasks.Restrictions,
				},
				{
					Name:  "chooseYourCertificateProvider",
					Path:  donor.PathWhatACertificateProviderDoes,
					State: provided.Tasks.CertificateProvider,
				},
				{
					Name:  "peopleToNotifyAboutYourLpa",
					Path:  donor.PathDoYouWantToNotifyPeople,
					State: provided.Tasks.PeopleToNotify,
					Count: len(provided.PeopleToNotify),
				},
				{
					Name:  "addCorrespondent",
					Path:  donor.PathAddCorrespondent,
					State: provided.Tasks.AddCorrespondent,
				},
			},
		}

		if provided.Donor.CanSign.IsNo() {
			section1.Items = append(section1.Items, taskListItem{
				Name:  "chooseYourSignatoryAndIndependentWitness",
				Path:  donor.PathGettingHelpSigning,
				State: provided.Tasks.ChooseYourSignatory,
			})
		}

		var sections []taskListSection
		if appData.SupporterData != nil {
			sections = []taskListSection{section1}
		} else {
			section1.Items = append(section1.Items, taskListItem{
				Name:  "checkAndSendToYourCertificateProvider",
				Path:  donor.PathCheckYourLpa,
				State: provided.Tasks.CheckYourLpa,
			})
			sections = []taskListSection{section1, taskListPaymentSection(provided), taskListSignSection(provided)}
		}

		return tmpl(w, &taskListData{
			App:              appData,
			Donor:            provided,
			EvidenceReceived: evidenceReceived,
			Sections:         sections,
		})
	}
}

func taskListTypeSpecificStep(provided *donordata.Provided) taskListItem {
	if provided.Type == lpadata.LpaTypePersonalWelfare {
		return taskListItem{
			Name:  "lifeSustainingTreatment",
			Path:  donor.PathLifeSustainingTreatment,
			State: provided.Tasks.LifeSustainingTreatment,
		}
	}

	return taskListItem{
		Name:  "chooseWhenTheLpaCanBeUsed",
		Path:  donor.PathWhenCanTheLpaBeUsed,
		State: provided.Tasks.WhenCanTheLpaBeUsed,
	}
}

func taskListPaymentSection(provided *donordata.Provided) taskListSection {
	var paymentPath donor.Path
	switch provided.Tasks.PayForLpa {
	case task.PaymentStateApproved, task.PaymentStateDenied:
		paymentPath = donor.PathPayFee
	case task.PaymentStateMoreEvidenceRequired, task.PaymentStatePending:
		paymentPath = donor.PathPendingPayment
	case task.PaymentStateCompleted:
		paymentPath = ""
	default:
		paymentPath = donor.PathAboutPayment
	}

	return taskListSection{
		Heading: "payForTheLpa",
		Items: []taskListItem{
			{
				Name:         "payForTheLpa",
				Path:         paymentPath,
				PaymentState: provided.Tasks.PayForLpa,
			},
		},
	}
}

func taskListSignSection(provided *donordata.Provided) taskListSection {
	confirmYourIdentityPath := donor.PathConfirmYourIdentity
	signTheLpaPath := donor.PathHowToSignYourLpa

	switch provided.IdentityUserData.Status {
	case identity.StatusConfirmed:
		confirmYourIdentityPath = donor.PathIdentityDetails

		if !provided.SignedAt.IsZero() {
			signTheLpaPath = donor.PathWitnessingYourSignature
		}

	case identity.StatusFailed:
		confirmYourIdentityPath = donor.PathRegisterWithCourtOfProtection

		if provided.RegisteringWithCourtOfProtection {
			confirmYourIdentityPath = donor.PathWhatHappensNextRegisteringWithCourtOfProtection

			if !provided.SignedAt.IsZero() {
				signTheLpaPath = donor.PathWitnessingYourSignature
			} else {
				signTheLpaPath = donor.PathHowToSignYourLpa
			}
		}

	case identity.StatusExpired:
		confirmYourIdentityPath = donor.PathWhatYouCanDoNowExpired

	case identity.StatusInsufficientEvidence:
		if !provided.SignedAt.IsZero() {
			signTheLpaPath = donor.PathWitnessingYourSignature
		}

		if provided.RegisteringWithCourtOfProtection {
			confirmYourIdentityPath = donor.PathWhatHappensNextRegisteringWithCourtOfProtection
		} else if provided.Voucher.Allowed {
			confirmYourIdentityPath = donor.PathWeHaveContactedVoucher
		} else if provided.WantVoucher.IsYes() {
			confirmYourIdentityPath = donor.PathEnterVoucher
		} else if provided.WantVoucher.IsNo() || provided.WantVoucher.IsUnknown() && provided.FailedVouchAttempts > 0 {
			confirmYourIdentityPath = donor.PathWhatYouCanDoNow
		} else {
			confirmYourIdentityPath = donor.PathUnableToConfirmIdentity
		}

	case identity.StatusUnknown:
		if provided.Tasks.ConfirmYourIdentity.IsInProgress() {
			confirmYourIdentityPath = donor.PathHowWillYouConfirmYourIdentity
		} else if provided.Tasks.ConfirmYourIdentity.IsPending() {
			confirmYourIdentityPath = donor.PathCompletingYourIdentityConfirmation
		}
	}

	if !provided.WitnessedByCertificateProviderAt.IsZero() {
		signTheLpaPath = ""
	}

	return taskListSection{
		Heading: "confirmYourIdentityAndSign",
		Items: []taskListItem{
			{
				Name:          "confirmYourIdentity",
				Path:          confirmYourIdentityPath,
				IdentityState: provided.Tasks.ConfirmYourIdentity,
			},
			{
				Name:  "signTheLpa",
				Path:  signTheLpaPath,
				State: provided.Tasks.SignTheLpa,
			},
		},
	}
}
