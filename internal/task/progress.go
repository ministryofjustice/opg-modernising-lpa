package task

import (
	"slices"
	"sort"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
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

type Progress struct {
	CompletedSteps []Step
}

func (p *Progress) Complete(name StepName, completed time.Time) {
	p.CompletedSteps = append(p.CompletedSteps, Step{Name: name, Completed: completed})
}

type ProgressTracker struct {
	Localizer      Localizer
	PaidFullFee    bool
	Supporter      bool
	CompletedSteps *Progress
}

func NewProgressTracker(localizer Localizer) *ProgressTracker {
	return &ProgressTracker{Localizer: localizer}
}

func (pt *ProgressTracker) Init(paidFullFee, isSupporter bool, completedSteps *Progress) {
	pt.PaidFullFee = paidFullFee
	pt.Supporter = isSupporter
	pt.CompletedSteps = completedSteps
}

func (pt *ProgressTracker) DonorSteps() []Step {
	steps := []Step{
		{Name: DonorSignedLPA},
		{Name: CertificateProvided},
		{Name: AllAttorneysSignedLPA},
		{Name: LpaSubmitted},
		{Name: NoticesOfIntentSent, Notification: true},
		{Name: StatutoryWaitingPeriodFinished},
		{Name: LpaRegistered},
	}

	if !pt.PaidFullFee {
		steps = slices.Insert(steps, 0,
			Step{Name: FeeEvidenceSubmitted},
			Step{Name: FeeEvidenceNotification, Notification: true},
			Step{Name: FeeEvidenceApproved},
		)
	}

	return steps
}

func (pt *ProgressTracker) SupporterSteps() []Step {
	steps := []Step{
		{Name: DonorPaid},
		{Name: DonorProvedID},
		{Name: DonorSignedLPA},
		{Name: CertificateProvided},
		{Name: AllAttorneysSignedLPA},
		{Name: LpaSubmitted},
		{Name: NoticesOfIntentSent, Notification: true},
		{Name: StatutoryWaitingPeriodFinished},
		{Name: LpaRegistered},
	}

	if !pt.PaidFullFee {
		steps = slices.Insert(steps, 0,
			Step{Name: FeeEvidenceSubmitted},
			Step{Name: FeeEvidenceNotification, Notification: true},
			Step{Name: FeeEvidenceApproved},
		)
	}

	return steps
}

func (pt *ProgressTracker) Remaining() (inProgress Step, remaining []Step) {
	allSteps := pt.DonorSteps()
	if pt.Supporter {
		allSteps = pt.SupporterSteps()
	}

	if len(pt.CompletedSteps.CompletedSteps) == 0 {
		return allSteps[0], allSteps[1:]
	}

	removeMap := make(map[StepName]bool)

	for _, toRemove := range pt.CompletedSteps.CompletedSteps {
		removeMap[toRemove.Name] = true
	}

	for _, step := range allSteps {
		if _, ok := removeMap[step.Name]; !ok && step.Show() {
			remaining = append(remaining, step)
		}
	}

	return remaining[0], remaining[1:]
}

func (pt *ProgressTracker) Completed() []Step {
	var filteredCompleted []Step

	steps := pt.DonorSteps()
	if pt.Supporter {
		steps = pt.SupporterSteps()
	}

	//Filter out supporter steps appearing in donor steps
	seen := make(map[StepName]bool)
	for _, step := range steps {
		seen[step.Name] = true
	}

	for _, step := range pt.CompletedSteps.CompletedSteps {
		if seen[step.Name] && step.Show() {
			filteredCompleted = append(filteredCompleted, step)
		}
	}

	sort.Slice(filteredCompleted, func(a, b int) bool {
		return pt.CompletedSteps.CompletedSteps[a].Completed.Before(pt.CompletedSteps.CompletedSteps[b].Completed)
	})

	return filteredCompleted
}

func (pt *ProgressTracker) IsSupporter() bool {
	return pt.Supporter
}

type ProgressTask struct {
	State     State
	Label     string
	Completed time.Time
}

//type Progress struct {
//	isOrganisation            bool
//	FeeEvidenceSubmitted      ProgressTask
//	FeeEvidenceNotification   ProgressTask
//	FeeEvidenceApproved       ProgressTask
//	Paid                      ProgressTask
//	ConfirmedID               ProgressTask
//	DonorSigned               ProgressTask
//	CertificateProviderSigned ProgressTask
//	AttorneysSigned           ProgressTask
//	LpaSubmitted              ProgressTask
//	NoticesOfIntentSent       ProgressTask
//	StatutoryWaitingPeriod    ProgressTask
//	LpaRegistered             ProgressTask
//}

//func (p Progress) ToSlice() []ProgressTask {
//	var list []ProgressTask
//
//	if !p.FeeEvidenceSubmitted.State.IsNotStarted() {
//		list = append(list, p.FeeEvidenceSubmitted)
//
//		if !p.isOrganisation {
//			if p.FeeEvidenceNotification.State.IsNotStarted() {
//				list = append(list, p.FeeEvidenceApproved, p.DonorSigned)
//			} else {
//				if !p.DonorSigned.State.IsCompleted() {
//					list = append(list, p.FeeEvidenceNotification, p.FeeEvidenceApproved, p.DonorSigned)
//				} else if p.FeeEvidenceNotification.Completed.Before(p.DonorSigned.Completed) {
//					list = append(list, p.FeeEvidenceNotification, p.DonorSigned, p.FeeEvidenceApproved)
//				} else {
//					list = append(list, p.DonorSigned, p.FeeEvidenceNotification, p.FeeEvidenceApproved)
//				}
//			}
//		} else {
//			if p.FeeEvidenceNotification.State.IsNotStarted() {
//				list = append(list, p.FeeEvidenceApproved, p.Paid, p.ConfirmedID, p.DonorSigned)
//			} else {
//				if !p.DonorSigned.State.IsCompleted() {
//					list = append(list, p.FeeEvidenceNotification, p.FeeEvidenceApproved, p.Paid, p.ConfirmedID, p.DonorSigned)
//				} else if p.FeeEvidenceNotification.Completed.Before(p.DonorSigned.Completed) {
//					list = append(list, p.FeeEvidenceNotification, p.DonorSigned, p.FeeEvidenceApproved, p.Paid, p.ConfirmedID)
//				} else {
//					list = append(list, p.DonorSigned, p.FeeEvidenceNotification, p.FeeEvidenceApproved, p.Paid, p.ConfirmedID)
//				}
//			}
//		}
//	} else {
//		if p.isOrganisation {
//			list = append(list, p.Paid, p.ConfirmedID, p.DonorSigned)
//		} else {
//			list = append(list, p.DonorSigned)
//		}
//	}
//
//	list = append(list, p.CertificateProviderSigned, p.AttorneysSigned, p.LpaSubmitted)
//
//	if p.NoticesOfIntentSent.State.IsCompleted() {
//		list = append(list, p.NoticesOfIntentSent)
//	}
//
//	list = append(list, p.StatutoryWaitingPeriod, p.LpaRegistered)
//
//	return list
//}

//func (pt *ProgressTracker) Progress(lpa *lpadata.Lpa, donorTasks DonorTasks, notifications notification.Notifications, feeType pay.FeeType) Progress {
//	var labels map[string]string
//
//	if lpa.IsOrganisationDonor {
//		labels = map[string]string{
//			"paid": pt.Localizer.Format(
//				"donorFullNameHasPaid",
//				map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
//			),
//			"confirmedID": pt.Localizer.Format(
//				"donorFullNameHasConfirmedTheirIdentity",
//				map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
//			),
//			"donorSigned": pt.Localizer.Format(
//				"donorFullNameHasSignedTheLPA",
//				map[string]interface{}{"DonorFullName": lpa.Donor.FullName()},
//			),
//			"certificateProviderSigned": pt.Localizer.T("theCertificateProviderHasDeclared"),
//			"attorneysSigned":           pt.Localizer.T("allAttorneysHaveSignedTheLpa"),
//			"lpaSubmitted":              pt.Localizer.T("opgHasReceivedTheLPA"),
//			"noticesOfIntentSent":       "weSentAnEmailTheLpaIsReadyToRegister",
//			"statutoryWaitingPeriod":    pt.Localizer.T("theWaitingPeriodHasStarted"),
//			"lpaRegistered":             pt.Localizer.T("theLpaHasBeenRegistered"),
//		}
//
//		if !feeType.IsFullFee() {
//			donorFullNamePossessive := pt.Localizer.Possessive(lpa.Donor.FullName())
//
//			labels["feeEvidenceSubmitted"] = pt.Localizer.Format(
//				"donorNamesLPAFeeEvidenceHasBeenSubmitted",
//				map[string]interface{}{"DonorFullNamePossessive": donorFullNamePossessive},
//			)
//
//			if !notifications.FeeEvidence.Received.IsZero() {
//				labels["feeEvidenceNotification"] = pt.Localizer.Format(
//					"weEmailedDonorNameOnAbout",
//					map[string]interface{}{
//						"On":            pt.Localizer.FormatDate(notifications.FeeEvidence.Received),
//						"About":         pt.Localizer.T("theFee"),
//						"DonorFullName": lpa.Donor.FullName(),
//					},
//				)
//			}
//
//			labels["feeEvidenceApproved"] = pt.Localizer.Format(
//				"donorNamesLPAFeeEvidenceHasBeenApproved",
//				map[string]interface{}{"DonorFullNamePossessive": donorFullNamePossessive},
//			)
//		}
//	} else {
//		labels = map[string]string{
//			"donorSigned":            pt.Localizer.T("youveSignedYourLpa"),
//			"attorneysSigned":        pt.Localizer.Count("attorneysHaveDeclared", len(lpa.Attorneys.Attorneys)),
//			"lpaSubmitted":           pt.Localizer.T("weHaveReceivedYourLpa"),
//			"noticesOfIntentSent":    "weSentAnEmailYourLpaIsReadyToRegister",
//			"statutoryWaitingPeriod": pt.Localizer.T("yourWaitingPeriodHasStarted"),
//			"lpaRegistered":          pt.Localizer.T("yourLpaHasBeenRegistered"),
//		}
//
//		if !feeType.IsFullFee() {
//			labels["feeEvidenceSubmitted"] = pt.Localizer.T("yourLPAFeeEvidenceHasBeenSubmitted")
//
//			if !notifications.FeeEvidence.Received.IsZero() {
//				labels["feeEvidenceNotification"] = pt.Localizer.Format(
//					"weEmailedYouOnAbout",
//					map[string]interface{}{
//						"On":    pt.Localizer.FormatDate(notifications.FeeEvidence.Received),
//						"About": pt.Localizer.T("yourFee"),
//					},
//				)
//			}
//
//			labels["feeEvidenceApproved"] = pt.Localizer.T("yourLPAFeeEvidenceHasBeenApproved")
//		}
//
//		if lpa.CertificateProvider.FirstNames != "" {
//			labels["certificateProviderSigned"] = pt.Localizer.Format(
//				"certificateProviderHasDeclared",
//				map[string]interface{}{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
//			)
//		} else {
//			labels["certificateProviderSigned"] = pt.Localizer.T("yourCertificateProviderHasDeclared")
//		}
//	}
//
//	progress := Progress{
//		isOrganisation: lpa.IsOrganisationDonor,
//		FeeEvidenceSubmitted: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["feeEvidenceSubmitted"],
//		},
//		FeeEvidenceNotification: ProgressTask{
//			State:     StateNotStarted,
//			Label:     labels["feeEvidenceNotification"],
//			Completed: notifications.FeeEvidence.Received,
//		},
//		FeeEvidenceApproved: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["feeEvidenceApproved"],
//		},
//		Paid: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["paid"],
//		},
//		ConfirmedID: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["confirmedID"],
//		},
//		DonorSigned: ProgressTask{
//			State:     StateNotStarted,
//			Label:     labels["donorSigned"],
//			Completed: lpa.SignedAt,
//		},
//		CertificateProviderSigned: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["certificateProviderSigned"],
//		},
//		AttorneysSigned: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["attorneysSigned"],
//		},
//		LpaSubmitted: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["lpaSubmitted"],
//		},
//		NoticesOfIntentSent: ProgressTask{
//			State: StateNotStarted,
//		},
//		StatutoryWaitingPeriod: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["statutoryWaitingPeriod"],
//		},
//		LpaRegistered: ProgressTask{
//			State: StateNotStarted,
//			Label: labels["lpaRegistered"],
//		},
//	}
//
//	if !feeType.IsFullFee() {
//		progress.FeeEvidenceSubmitted.State = StateCompleted
//
//		if !notifications.FeeEvidence.Received.IsZero() {
//			progress.FeeEvidenceSubmitted.State = StateCompleted
//			progress.FeeEvidenceNotification.State = StateCompleted
//		}
//
//		// donor signed task can move positions when fee evidence tasks exist
//		if !lpa.SignedAt.IsZero() {
//			progress.DonorSigned.State = StateCompleted
//		}
//
//		progress.FeeEvidenceApproved.State = StateInProgress
//
//		if !donorTasks.PayForLpa.IsApproved() && !donorTasks.PayForLpa.IsCompleted() {
//			return progress
//		}
//
//		progress.FeeEvidenceApproved.State = StateCompleted
//	}
//
//	if lpa.IsOrganisationDonor {
//		progress.Paid.State = StateInProgress
//		if !lpa.Paid {
//			return progress
//		}
//
//		progress.Paid.State = StateCompleted
//		progress.ConfirmedID.State = StateInProgress
//
//		if lpa.Donor.IdentityCheck.CheckedAt.IsZero() {
//			return progress
//		}
//
//		progress.ConfirmedID.State = StateCompleted
//		progress.DonorSigned.State = StateInProgress
//
//		if lpa.SignedAt.IsZero() {
//			return progress
//		}
//	} else {
//		progress.DonorSigned.State = StateInProgress
//		if lpa.SignedAt.IsZero() {
//			return progress
//		}
//	}
//
//	progress.DonorSigned.State = StateCompleted
//	progress.CertificateProviderSigned.State = StateInProgress
//
//	if lpa.CertificateProvider.SignedAt.IsZero() {
//		return progress
//	}
//
//	progress.CertificateProviderSigned.State = StateCompleted
//	progress.AttorneysSigned.State = StateInProgress
//
//	if !lpa.AllAttorneysSigned() {
//		return progress
//	}
//
//	progress.AttorneysSigned.State = StateCompleted
//	progress.LpaSubmitted.State = StateInProgress
//
//	if !lpa.Submitted {
//		return progress
//	}
//
//	progress.LpaSubmitted.State = StateCompleted
//
//	if lpa.PerfectAt.IsZero() {
//		return progress
//	}
//
//	progress.NoticesOfIntentSent.Label = pt.Localizer.Format(labels["noticesOfIntentSent"], map[string]any{
//		"SentOn": pt.Localizer.FormatDate(lpa.PerfectAt),
//	})
//	progress.NoticesOfIntentSent.State = StateCompleted
//	progress.StatutoryWaitingPeriod.State = StateInProgress
//
//	if lpa.RegisteredAt.IsZero() {
//		return progress
//	}
//
//	progress.StatutoryWaitingPeriod.State = StateCompleted
//	progress.LpaRegistered.State = StateCompleted
//
//	return progress
//}
