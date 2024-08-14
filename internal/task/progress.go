package task

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

type Localizer interface {
	Concat(list []string, joiner string) string
	Count(messageID string, count int) string
	Format(messageID string, data map[string]interface{}) string
	FormatCount(messageID string, count int, data map[string]any) string
	FormatDate(t date.TimeOrDate) string
	FormatTime(t time.Time) string
	FormatDateTime(t time.Time) string
	Possessive(s string) string
	SetShowTranslationKeys(s bool)
	ShowTranslationKeys() bool
	T(messageID string) string
}

type ProgressTracker struct {
	Localizer Localizer
}

type ProgressTask struct {
	State State
	Label string
}

type Progress struct {
	isOrganisation            bool
	Paid                      ProgressTask
	ConfirmedID               ProgressTask
	DonorSigned               ProgressTask
	CertificateProviderSigned ProgressTask
	AttorneysSigned           ProgressTask
	LpaSubmitted              ProgressTask
	NoticesOfIntentSent       ProgressTask
	StatutoryWaitingPeriod    ProgressTask
	LpaRegistered             ProgressTask
}

func (p Progress) ToSlice() []ProgressTask {
	var list []ProgressTask
	if p.isOrganisation {
		list = append(list, p.Paid, p.ConfirmedID)
	}

	list = append(list, p.DonorSigned, p.CertificateProviderSigned, p.AttorneysSigned, p.LpaSubmitted)

	if p.NoticesOfIntentSent.State.IsCompleted() {
		list = append(list, p.NoticesOfIntentSent)
	}

	list = append(list, p.StatutoryWaitingPeriod, p.LpaRegistered)

	return list
}

func (pt ProgressTracker) Progress(lpa *lpadata.Lpa) Progress {
	var labels map[string]string

	if lpa.IsOrganisationDonor {
		labels = map[string]string{
			"paid": pt.Localizer.Format(
				"donorFullNameHasPaid",
				map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
			),
			"confirmedID": pt.Localizer.Format(
				"donorFullNameHasConfirmedTheirIdentity",
				map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
			),
			"donorSigned": pt.Localizer.Format(
				"donorFullNameHasSignedTheLPA",
				map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
			),
			"certificateProviderSigned": pt.Localizer.T("theCertificateProviderHasDeclared"),
			"attorneysSigned":           pt.Localizer.T("allAttorneysHaveSignedTheLpa"),
			"lpaSubmitted":              pt.Localizer.T("opgHasReceivedTheLPA"),
			"noticesOfIntentSent":       "weSentAnEmailTheLpaIsReadyToRegister",
			"statutoryWaitingPeriod":    pt.Localizer.T("theWaitingPeriodHasStarted"),
			"lpaRegistered":             pt.Localizer.T("theLpaHasBeenRegistered"),
		}
	} else {
		labels = map[string]string{
			"donorSigned":            pt.Localizer.T("youveSignedYourLpa"),
			"attorneysSigned":        pt.Localizer.Count("attorneysHaveDeclared", len(lpa.Attorneys.Attorneys)),
			"lpaSubmitted":           pt.Localizer.T("weHaveReceivedYourLpa"),
			"noticesOfIntentSent":    "weSentAnEmailYourLpaIsReadyToRegister",
			"statutoryWaitingPeriod": pt.Localizer.T("yourWaitingPeriodHasStarted"),
			"lpaRegistered":          pt.Localizer.T("yourLpaHasBeenRegistered"),
		}

		if lpa.CertificateProvider.FirstNames != "" {
			labels["certificateProviderSigned"] = pt.Localizer.Format(
				"certificateProviderHasDeclared",
				map[string]interface{}{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
			)
		} else {
			labels["certificateProviderSigned"] = pt.Localizer.T("yourCertificateProviderHasDeclared")
		}
	}

	progress := Progress{
		isOrganisation: lpa.IsOrganisationDonor,
		Paid: ProgressTask{
			State: StateNotStarted,
			Label: labels["paid"],
		},
		ConfirmedID: ProgressTask{
			State: StateNotStarted,
			Label: labels["confirmedID"],
		},
		DonorSigned: ProgressTask{
			State: StateNotStarted,
			Label: labels["donorSigned"],
		},
		CertificateProviderSigned: ProgressTask{
			State: StateNotStarted,
			Label: labels["certificateProviderSigned"],
		},
		AttorneysSigned: ProgressTask{
			State: StateNotStarted,
			Label: labels["attorneysSigned"],
		},
		LpaSubmitted: ProgressTask{
			State: StateNotStarted,
			Label: labels["lpaSubmitted"],
		},
		NoticesOfIntentSent: ProgressTask{
			State: StateNotStarted,
		},
		StatutoryWaitingPeriod: ProgressTask{
			State: StateNotStarted,
			Label: labels["statutoryWaitingPeriod"],
		},
		LpaRegistered: ProgressTask{
			State: StateNotStarted,
			Label: labels["lpaRegistered"],
		},
	}

	if lpa.IsOrganisationDonor {
		progress.Paid.State = StateInProgress
		if !lpa.Paid {
			return progress
		}

		progress.Paid.State = StateCompleted
		progress.ConfirmedID.State = StateInProgress

		if lpa.Donor.IdentityCheck.CheckedAt.IsZero() {
			return progress
		}

		progress.ConfirmedID.State = StateCompleted
		progress.DonorSigned.State = StateInProgress

		if lpa.SignedAt.IsZero() {
			return progress
		}
	} else {
		progress.DonorSigned.State = StateInProgress
		if lpa.SignedAt.IsZero() {
			return progress
		}
	}

	progress.DonorSigned.State = StateCompleted
	progress.CertificateProviderSigned.State = StateInProgress

	if lpa.CertificateProvider.SignedAt.IsZero() {
		return progress
	}

	progress.CertificateProviderSigned.State = StateCompleted
	progress.AttorneysSigned.State = StateInProgress

	if !lpa.AllAttorneysSigned() {
		return progress
	}

	progress.AttorneysSigned.State = StateCompleted
	progress.LpaSubmitted.State = StateInProgress

	if !lpa.Submitted {
		return progress
	}

	progress.LpaSubmitted.State = StateCompleted

	if lpa.PerfectAt.IsZero() {
		return progress
	}

	progress.NoticesOfIntentSent.Label = pt.Localizer.Format(labels["noticesOfIntentSent"], map[string]any{
		"SentOn": pt.Localizer.FormatDate(lpa.PerfectAt),
	})
	progress.NoticesOfIntentSent.State = StateCompleted
	progress.StatutoryWaitingPeriod.State = StateInProgress

	if lpa.RegisteredAt.IsZero() {
		return progress
	}

	progress.StatutoryWaitingPeriod.State = StateCompleted
	progress.LpaRegistered.State = StateCompleted

	return progress
}
