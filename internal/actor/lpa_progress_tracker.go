package actor

type ProgressTracker struct {
	Localizer Localizer
}

func (pt ProgressTracker) Progress(donor *DonorProvidedDetails, certificateProvider *CertificateProviderProvidedDetails, attorneys []*AttorneyProvidedDetails) Progress {
	var (
		paidLabel,
		confirmedIDLabel,
		donorSignedLabel,
		certificateProviderSignedLabel,
		attorneysSignedLabel,
		lpaSubmittedLabel,
		statutoryWaitingPeriodLabel,
		lpaRegisteredLabel string
	)

	if donor.CertificateProvider.FirstNames != "" {
		certificateProviderSignedLabel = pt.Localizer.Format(
			"certificateProviderHasDeclared",
			map[string]interface{}{"CertificateProviderFullName": donor.CertificateProvider.FullName()},
		)
	} else {
		if donor.IsOrganisationDonor() {
			certificateProviderSignedLabel = pt.Localizer.T("theCertificateProviderHasDeclared")
		} else {
			certificateProviderSignedLabel = pt.Localizer.T("yourCertificateProviderHasDeclared")
		}
	}

	if donor.IsOrganisationDonor() {
		paidLabel = pt.Localizer.Format(
			"donorFullNameHasPaid",
			map[string]interface{}{"DonorFullName": donor.Donor.FullName()},
		)

		confirmedIDLabel = pt.Localizer.Format(
			"donorFullNameHasConfirmedTheirIdentity",
			map[string]interface{}{"DonorFullName": donor.Donor.FullName()},
		)

		donorSignedLabel = pt.Localizer.Format(
			"donorFullNameHasSignedTheLPA",
			map[string]interface{}{"DonorFullName": donor.Donor.FullName()},
		)
		attorneysSignedLabel = pt.Localizer.T("allAttorneysHaveSignedTheLpa")
		lpaSubmittedLabel = pt.Localizer.T("opgHasReceivedTheLPA")
		statutoryWaitingPeriodLabel = pt.Localizer.T("theWaitingPeriodHasStarted")
		lpaRegisteredLabel = pt.Localizer.T("theLpaHasBeenRegistered")
	} else {
		donorSignedLabel = pt.Localizer.T("youveSignedYourLpa")
		attorneysSignedLabel = pt.Localizer.Count("attorneysHaveDeclared", len(donor.Attorneys.Attorneys))
		lpaSubmittedLabel = pt.Localizer.T("weHaveReceivedYourLpa")
		statutoryWaitingPeriodLabel = pt.Localizer.T("yourWaitingPeriodHasStarted")
		lpaRegisteredLabel = pt.Localizer.T("yourLpaHasBeenRegistered")
	}

	progress := Progress{
		Paid: ProgressTask{
			State: TaskNotStarted,
			Label: paidLabel,
		},
		ConfirmedID: ProgressTask{
			State: TaskNotStarted,
			Label: confirmedIDLabel,
		},
		DonorSigned: ProgressTask{
			State: TaskNotStarted,
			Label: donorSignedLabel,
		},
		CertificateProviderSigned: ProgressTask{
			State: TaskNotStarted,
			Label: certificateProviderSignedLabel,
		},
		AttorneysSigned: ProgressTask{
			State: TaskNotStarted,
			Label: attorneysSignedLabel,
		},
		LpaSubmitted: ProgressTask{
			State: TaskNotStarted,
			Label: lpaSubmittedLabel,
		},
		StatutoryWaitingPeriod: ProgressTask{
			State: TaskNotStarted,
			Label: statutoryWaitingPeriodLabel,
		},
		LpaRegistered: ProgressTask{
			State: TaskNotStarted,
			Label: lpaRegisteredLabel,
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

//localizer := newMockLocalizer(t)
//localizer.EXPECT().
//T(mock.Anything).
//Return("translated")
//localizer.EXPECT().
//Format(mock.Anything, mock.Anything).
//Return("translated")
