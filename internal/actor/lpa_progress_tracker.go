package actor

type ProgressTracker struct {
	Localizer Localizer
}

type ProgressTask struct {
	State TaskState
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

func (pt ProgressTracker) Progress(donor *DonorProvidedDetails, certificateProvider *CertificateProviderProvidedDetails, attorneys []*AttorneyProvidedDetails) Progress {
	var labels map[string]string

	if donor.IsOrganisationDonor() {
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
			State: TaskNotStarted,
			Label: labels["paid"],
		},
		ConfirmedID: ProgressTask{
			State: TaskNotStarted,
			Label: labels["confirmedID"],
		},
		DonorSigned: ProgressTask{
			State: TaskNotStarted,
			Label: labels["donorSigned"],
		},
		CertificateProviderSigned: ProgressTask{
			State: TaskNotStarted,
			Label: labels["certificateProviderSigned"],
		},
		AttorneysSigned: ProgressTask{
			State: TaskNotStarted,
			Label: labels["attorneysSigned"],
		},
		LpaSubmitted: ProgressTask{
			State: TaskNotStarted,
			Label: labels["lpaSubmitted"],
		},
		StatutoryWaitingPeriod: ProgressTask{
			State: TaskNotStarted,
			Label: labels["statutoryWaitingPeriod"],
		},
		LpaRegistered: ProgressTask{
			State: TaskNotStarted,
			Label: labels["lpaRegistered"],
		},
	}

	if donor.IsOrganisationDonor() {
		progress.Paid.State = TaskInProgress
		if !donor.Tasks.PayForLpa.IsCompleted() {
			return progress
		}

		progress.Paid.State = TaskCompleted
		progress.ConfirmedID.State = TaskInProgress

		if !donor.DonorIdentityConfirmed() {
			return progress
		}

		progress.ConfirmedID.State = TaskCompleted
		progress.DonorSigned.State = TaskInProgress

		if donor.SignedAt.IsZero() {
			return progress
		}
	} else {
		progress.DonorSigned.State = TaskInProgress
		if donor.SignedAt.IsZero() {
			return progress
		}
	}

	progress.DonorSigned.State = TaskCompleted
	progress.CertificateProviderSigned.State = TaskInProgress

	if !certificateProvider.Signed(donor.SignedAt) {
		return progress
	}

	progress.CertificateProviderSigned.State = TaskCompleted
	progress.AttorneysSigned.State = TaskInProgress

	if !donor.AllAttorneysSigned(attorneys) {
		return progress
	}

	progress.AttorneysSigned.State = TaskCompleted
	progress.LpaSubmitted.State = TaskInProgress

	if donor.SubmittedAt.IsZero() {
		return progress
	}

	progress.LpaSubmitted.State = TaskCompleted
	progress.StatutoryWaitingPeriod.State = TaskInProgress

	if donor.RegisteredAt.IsZero() {
		return progress
	}

	progress.StatutoryWaitingPeriod.State = TaskCompleted
	progress.LpaRegistered.State = TaskCompleted

	return progress
}
