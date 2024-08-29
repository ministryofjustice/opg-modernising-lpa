package progress

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

//go:generate enumerator -type StepName -empty
type StepName uint8

const (
	FeeEvidenceSubmitted StepName = iota + 1
	FeeEvidenceNotification
	FeeEvidenceApproved
	DonorPaid
	DonorProvedID
	DonorSignedLPA
	CertificateProvided
	AllAttorneysSignedLPA
	NoticesOfIntentSent
	LpaSubmitted
	StatutoryWaitingPeriodFinished
	LpaRegistered
)

type Step struct {
	Name StepName
	// When the notification was completed
	Completed time.Time
	// Notification indicates the step relies on receiving a notification to be shown
	Notification bool
}

func (s Step) DonorLabel(l Localizer, lpa *lpadata.Lpa) string {
	switch s.Name {
	case FeeEvidenceSubmitted:
		return l.T("yourLPAFeeEvidenceHasBeenSubmitted")
	case FeeEvidenceNotification:
		return l.Format(
			"weEmailedYouOnAbout",
			map[string]interface{}{
				"On":    l.FormatDate(s.Completed),
				"About": l.T("yourFee"),
			},
		)
	case FeeEvidenceApproved:
		return l.T("yourLPAFeeEvidenceHasBeenApproved")
	case DonorSignedLPA:
		return l.T("youveSignedYourLpa")
	case CertificateProvided:
		if lpa.CertificateProvider.FirstNames != "" {
			return l.Format(
				"certificateProviderHasDeclared",
				map[string]interface{}{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
			)
		} else {
			return l.T("yourCertificateProviderHasDeclared")
		}
	case AllAttorneysSignedLPA:
		return l.Count("attorneysHaveDeclared", len(lpa.Attorneys.Attorneys))
	case LpaSubmitted:
		return l.T("weHaveReceivedYourLpa")
	case NoticesOfIntentSent:
		return l.Format("weSentAnEmailYourLpaIsReadyToRegister", map[string]any{
			"SentOn": l.FormatDate(lpa.PerfectAt),
		})
	case StatutoryWaitingPeriodFinished:
		return l.T("yourWaitingPeriodHasStarted")
	case LpaRegistered:
		return l.T("yourLpaHasBeenRegistered")
	default:
		return ""
	}
}

func (s Step) SupporterLabel(l Localizer, lpa *lpadata.Lpa) string {
	switch s.Name {
	case FeeEvidenceSubmitted:
		donorFullNamePossessive := l.Possessive(lpa.Donor.FullName())
		return l.Format(
			"donorNamesLPAFeeEvidenceHasBeenSubmitted",
			map[string]interface{}{"DonorFullNamePossessive": donorFullNamePossessive},
		)
	case FeeEvidenceNotification:
		return l.Format(
			"weEmailedDonorNameOnAbout",
			map[string]interface{}{
				"On":            l.FormatDate(s.Completed),
				"About":         l.T("theFee"),
				"DonorFullName": lpa.Donor.FullName(),
			},
		)
	case FeeEvidenceApproved:
		donorFullNamePossessive := l.Possessive(lpa.Donor.FullName())
		return l.Format(
			"donorNamesLPAFeeEvidenceHasBeenApproved",
			map[string]interface{}{"DonorFullNamePossessive": donorFullNamePossessive},
		)
	case DonorPaid:
		return l.Format(
			"donorFullNameHasPaid",
			map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
		)
	case DonorProvedID:
		return l.Format(
			"donorFullNameHasConfirmedTheirIdentity",
			map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
		)
	case DonorSignedLPA:
		return l.Format(
			"donorFullNameHasSignedTheLPA",
			map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
		)
	case CertificateProvided:
		return l.T("theCertificateProviderHasDeclared")
	case AllAttorneysSignedLPA:
		return l.T("allAttorneysHaveSignedTheLpa")
	case LpaSubmitted:
		return l.T("opgHasReceivedTheLPA")
	case NoticesOfIntentSent:
		return l.Format(
			"weSentAnEmailTheLpaIsReadyToRegister",
			map[string]any{"SentOn": l.FormatDate(lpa.PerfectAt)})
	case StatutoryWaitingPeriodFinished:
		return l.T("theWaitingPeriodHasStarted")
	case LpaRegistered:
		return l.T("theLpaHasBeenRegistered")
	default:
		return ""
	}
}

func (s Step) Show() bool {
	return (s.Notification && !s.Completed.IsZero()) || !s.Notification
}
