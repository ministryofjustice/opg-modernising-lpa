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
	Path          string
	State         task.State
	PaymentState  task.PaymentState
	IdentityState task.IdentityState
	Count         int
	Hidden        bool
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

		chooseAttorneysLink := donor.PathChooseAttorneysGuidance.Format(provided.LpaID)
		if provided.Attorneys.Len() > 0 {
			chooseAttorneysLink = donor.PathChooseAttorneysSummary.Format(provided.LpaID)
		}

		chooseReplacementAttorneysLink := donor.PathDoYouWantReplacementAttorneys.Format(provided.LpaID)
		if provided.ReplacementAttorneys.Len() > 0 {
			chooseReplacementAttorneysLink = donor.PathChooseReplacementAttorneysSummary.Format(provided.LpaID)
		}

		section1 := taskListSection{
			Heading: "fillInTheLpa",
			Items: []taskListItem{
				{
					Name:  "provideYourDetails",
					Path:  donor.PathYourDetails.Format(provided.LpaID),
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
					Path:  donor.PathRestrictions.Format(provided.LpaID),
					State: provided.Tasks.Restrictions,
				},
				{
					Name:  "chooseYourCertificateProvider",
					Path:  donor.PathWhatACertificateProviderDoes.Format(provided.LpaID),
					State: provided.Tasks.CertificateProvider,
				},
				{
					Name:  "peopleToNotifyAboutYourLpa",
					Path:  donor.PathDoYouWantToNotifyPeople.Format(provided.LpaID),
					State: provided.Tasks.PeopleToNotify,
					Count: len(provided.PeopleToNotify),
				},
				{
					Name:  "addCorrespondent",
					Path:  donor.PathAddCorrespondent.Format(provided.LpaID),
					State: provided.Tasks.AddCorrespondent,
				},
				{
					Name:   "chooseYourSignatoryAndIndependentWitness",
					Path:   donor.PathGettingHelpSigning.Format(provided.LpaID),
					State:  provided.Tasks.ChooseYourSignatory,
					Hidden: !provided.Donor.CanSign.IsNo(),
				},
			},
		}

		var sections []taskListSection
		if appData.SupporterData != nil {
			sections = []taskListSection{section1}
		} else {
			section1.Items = append(section1.Items, taskListItem{
				Name:  "checkAndSendToYourCertificateProvider",
				Path:  taskListCheckLpaPath(provided).Format(provided.LpaID),
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
			Path:  donor.PathLifeSustainingTreatment.Format(provided.LpaID),
			State: provided.Tasks.LifeSustainingTreatment,
		}
	}

	return taskListItem{
		Name:  "chooseWhenTheLpaCanBeUsed",
		Path:  donor.PathWhenCanTheLpaBeUsed.Format(provided.LpaID),
		State: provided.Tasks.WhenCanTheLpaBeUsed,
	}
}

func taskListCheckLpaPath(provided *donordata.Provided) donor.Path {
	if len(provided.Under18ActorDetails()) > 0 {
		return donor.PathYouCannotSignYourLpaYet
	} else if provided.CertificateProviderSharesDetails() {
		return donor.PathConfirmYourCertificateProviderIsNotRelated
	} else {
		return donor.PathCheckYourLpa
	}
}

func taskListPaymentSection(provided *donordata.Provided) taskListSection {
	var paymentPath string
	switch provided.Tasks.PayForLpa {
	case task.PaymentStateApproved:
		paymentPath = donor.PathFeeApproved.Format(provided.LpaID)
	case task.PaymentStateDenied:
		paymentPath = donor.PathFeeDenied.Format(provided.LpaID)
	case task.PaymentStateMoreEvidenceRequired:
		paymentPath = donor.PathUploadEvidence.Format(provided.LpaID)
	default:
		paymentPath = donor.PathAboutPayment.Format(provided.LpaID)
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

		if !provided.WitnessedByCertificateProviderAt.IsZero() {
			signTheLpaPath = donor.PathYouHaveSubmittedYourLpa
		} else if !provided.SignedAt.IsZero() {
			signTheLpaPath = donor.PathWitnessingYourSignature
		}

	case identity.StatusFailed:
		confirmYourIdentityPath = donor.PathRegisterWithCourtOfProtection

		if provided.RegisteringWithCourtOfProtection {
			confirmYourIdentityPath = donor.PathWhatHappensNextRegisteringWithCourtOfProtection

			if !provided.WitnessedByCertificateProviderAt.IsZero() {
				signTheLpaPath = donor.PathYouHaveSubmittedYourLpa
			} else if !provided.SignedAt.IsZero() {
				signTheLpaPath = donor.PathWitnessingYourSignature
			} else {
				signTheLpaPath = donor.PathHowToSignYourLpa
			}
		}

	case identity.StatusExpired:
		confirmYourIdentityPath = donor.PathWhatYouCanDoNowExpired

	case identity.StatusInsufficientEvidence:
		if !provided.WitnessedByCertificateProviderAt.IsZero() {
			signTheLpaPath = donor.PathYouHaveSubmittedYourLpa
		} else if !provided.SignedAt.IsZero() {
			signTheLpaPath = donor.PathWitnessingYourSignature
		}

		if provided.RegisteringWithCourtOfProtection {
			confirmYourIdentityPath = donor.PathWhatHappensNextRegisteringWithCourtOfProtection
		} else if provided.Voucher.Allowed {
			confirmYourIdentityPath = donor.PathWeHaveContactedVoucher
		} else if provided.WantVoucher.IsYes() {
			confirmYourIdentityPath = donor.PathEnterVoucher
		} else if provided.WantVoucher.IsNo() {
			confirmYourIdentityPath = donor.PathWhatYouCanDoNow
		} else {
			confirmYourIdentityPath = donor.PathUnableToConfirmIdentity
		}
	}

	return taskListSection{
		Heading: "confirmYourIdentityAndSign",
		Items: []taskListItem{
			{
				Name:          "confirmYourIdentity",
				Path:          confirmYourIdentityPath.Format(provided.LpaID),
				IdentityState: provided.Tasks.ConfirmYourIdentity,
			},
			{
				Name:  "signTheLpa",
				Path:  signTheLpaPath.Format(provided.LpaID),
				State: provided.Tasks.SignTheLpa,
			},
		},
	}
}
