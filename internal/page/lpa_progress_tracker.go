package page

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type ProgressTracker struct {
	Localizer Localizer
}

type ProgressTask struct {
	State task.State
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

	if p.NoticesOfIntentSent.State.Completed() {
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
			State: task.StateNotStarted,
			Label: labels["paid"],
		},
		ConfirmedID: ProgressTask{
			State: task.StateNotStarted,
			Label: labels["confirmedID"],
		},
		DonorSigned: ProgressTask{
			State: task.StateNotStarted,
			Label: labels["donorSigned"],
		},
		CertificateProviderSigned: ProgressTask{
			State: task.StateNotStarted,
			Label: labels["certificateProviderSigned"],
		},
		AttorneysSigned: ProgressTask{
			State: task.StateNotStarted,
			Label: labels["attorneysSigned"],
		},
		LpaSubmitted: ProgressTask{
			State: task.StateNotStarted,
			Label: labels["lpaSubmitted"],
		},
		NoticesOfIntentSent: ProgressTask{
			State: task.StateNotStarted,
		},
		StatutoryWaitingPeriod: ProgressTask{
			State: task.StateNotStarted,
			Label: labels["statutoryWaitingPeriod"],
		},
		LpaRegistered: ProgressTask{
			State: task.StateNotStarted,
			Label: labels["lpaRegistered"],
		},
	}

	if lpa.IsOrganisationDonor {
		progress.Paid.State = task.StateInProgress
		if !lpa.Paid {
			return progress
		}

		progress.Paid.State = task.StateCompleted
		progress.ConfirmedID.State = task.StateInProgress

		if lpa.Donor.IdentityCheck.CheckedAt.IsZero() {
			return progress
		}

		progress.ConfirmedID.State = task.StateCompleted
		progress.DonorSigned.State = task.StateInProgress

		if lpa.SignedAt.IsZero() {
			return progress
		}
	} else {
		progress.DonorSigned.State = task.StateInProgress
		if lpa.SignedAt.IsZero() {
			return progress
		}
	}

	progress.DonorSigned.State = task.StateCompleted
	progress.CertificateProviderSigned.State = task.StateInProgress

	if lpa.CertificateProvider.SignedAt.IsZero() {
		return progress
	}

	progress.CertificateProviderSigned.State = task.StateCompleted
	progress.AttorneysSigned.State = task.StateInProgress

	if !lpa.AllAttorneysSigned() {
		return progress
	}

	progress.AttorneysSigned.State = task.StateCompleted
	progress.LpaSubmitted.State = task.StateInProgress

	if !lpa.Submitted {
		return progress
	}

	progress.LpaSubmitted.State = task.StateCompleted

	if lpa.PerfectAt.IsZero() {
		return progress
	}

	progress.NoticesOfIntentSent.Label = pt.Localizer.Format(labels["noticesOfIntentSent"], map[string]any{
		"SentOn": pt.Localizer.FormatDate(lpa.PerfectAt),
	})
	progress.NoticesOfIntentSent.State = task.StateCompleted
	progress.StatutoryWaitingPeriod.State = task.StateInProgress

	if lpa.RegisteredAt.IsZero() {
		return progress
	}

	progress.StatutoryWaitingPeriod.State = task.StateCompleted
	progress.LpaRegistered.State = task.StateCompleted

	return progress
}
