package page

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
)

type ProgressTracker struct {
	Localizer Localizer
}

type ProgressTask struct {
	State actor.TaskState
	Label string
}

type Progress struct {
	Paid                      ProgressTask
	ConfirmedID               ProgressTask
	DonorSigned               ProgressTask
	CertificateProviderSigned ProgressTask
	AttorneysSigned           ProgressTask
	LpaSubmitted              ProgressTask
	StatutoryWaitingPeriod    ProgressTask
	LpaRegistered             ProgressTask
}

func (pt ProgressTracker) Progress(donor *lpastore.Lpa, certificateProvider *actor.CertificateProviderProvidedDetails, attorneys []*actor.AttorneyProvidedDetails) Progress {
	var labels map[string]string

	if donor.IsOrganisationDonor {
		labels = map[string]string{
			"paid": pt.Localizer.Format(
				"donorFullNameHasPaid",
				map[string]interface{}{"DonorFullName": donor.Donor.FullName()},
			),
			"confirmedID": pt.Localizer.Format(
				"donorFullNameHasConfirmedTheirIdentity",
				map[string]interface{}{"DonorFullName": donor.Donor.FullName()},
			),
			"donorSigned": pt.Localizer.Format(
				"donorFullNameHasSignedTheLPA",
				map[string]interface{}{"DonorFullName": donor.Donor.FullName()},
			),
			"certificateProviderSigned": pt.Localizer.T("theCertificateProviderHasDeclared"),
			"attorneysSigned":           pt.Localizer.T("allAttorneysHaveSignedTheLpa"),
			"lpaSubmitted":              pt.Localizer.T("opgHasReceivedTheLPA"),
			"statutoryWaitingPeriod":    pt.Localizer.T("theWaitingPeriodHasStarted"),
			"lpaRegistered":             pt.Localizer.T("theLpaHasBeenRegistered"),
		}
	} else {
		labels = map[string]string{
			"paid":                   "",
			"confirmedID":            "",
			"donorSigned":            pt.Localizer.T("youveSignedYourLpa"),
			"attorneysSigned":        pt.Localizer.Count("attorneysHaveDeclared", len(donor.Attorneys.Attorneys)),
			"lpaSubmitted":           pt.Localizer.T("weHaveReceivedYourLpa"),
			"statutoryWaitingPeriod": pt.Localizer.T("yourWaitingPeriodHasStarted"),
			"lpaRegistered":          pt.Localizer.T("yourLpaHasBeenRegistered"),
		}

		if donor.CertificateProvider.FirstNames != "" {
			labels["certificateProviderSigned"] = pt.Localizer.Format(
				"certificateProviderHasDeclared",
				map[string]interface{}{"CertificateProviderFullName": donor.CertificateProvider.FullName()},
			)
		} else {
			labels["certificateProviderSigned"] = pt.Localizer.T("yourCertificateProviderHasDeclared")
		}
	}

	progress := Progress{
		Paid: ProgressTask{
			State: actor.TaskNotStarted,
			Label: labels["paid"],
		},
		ConfirmedID: ProgressTask{
			State: actor.TaskNotStarted,
			Label: labels["confirmedID"],
		},
		DonorSigned: ProgressTask{
			State: actor.TaskNotStarted,
			Label: labels["donorSigned"],
		},
		CertificateProviderSigned: ProgressTask{
			State: actor.TaskNotStarted,
			Label: labels["certificateProviderSigned"],
		},
		AttorneysSigned: ProgressTask{
			State: actor.TaskNotStarted,
			Label: labels["attorneysSigned"],
		},
		LpaSubmitted: ProgressTask{
			State: actor.TaskNotStarted,
			Label: labels["lpaSubmitted"],
		},
		StatutoryWaitingPeriod: ProgressTask{
			State: actor.TaskNotStarted,
			Label: labels["statutoryWaitingPeriod"],
		},
		LpaRegistered: ProgressTask{
			State: actor.TaskNotStarted,
			Label: labels["lpaRegistered"],
		},
	}

	if donor.IsOrganisationDonor {
		progress.Paid.State = actor.TaskInProgress
		if !donor.Paid {
			return progress
		}

		progress.Paid.State = actor.TaskCompleted
		progress.ConfirmedID.State = actor.TaskInProgress

		if !donor.DonorIdentityConfirmed {
			return progress
		}

		progress.ConfirmedID.State = actor.TaskCompleted
		progress.DonorSigned.State = actor.TaskInProgress

		if donor.SignedAt.IsZero() {
			return progress
		}
	} else {
		progress.DonorSigned.State = actor.TaskInProgress
		if donor.SignedAt.IsZero() {
			return progress
		}
	}

	progress.DonorSigned.State = actor.TaskCompleted
	progress.CertificateProviderSigned.State = actor.TaskInProgress

	if !certificateProvider.Signed(donor.SignedAt) {
		return progress
	}

	progress.CertificateProviderSigned.State = actor.TaskCompleted
	progress.AttorneysSigned.State = actor.TaskInProgress

	if !donor.AllAttorneysSigned(attorneys) {
		return progress
	}

	progress.AttorneysSigned.State = actor.TaskCompleted
	progress.LpaSubmitted.State = actor.TaskInProgress

	if !donor.Submitted {
		return progress
	}

	progress.LpaSubmitted.State = actor.TaskCompleted
	progress.StatutoryWaitingPeriod.State = actor.TaskInProgress

	if donor.RegisteredAt.IsZero() {
		return progress
	}

	progress.StatutoryWaitingPeriod.State = actor.TaskCompleted
	progress.LpaRegistered.State = actor.TaskCompleted

	return progress
}
