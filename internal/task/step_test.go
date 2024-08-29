package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	yesterday = time.Date(2023, time.July, 1, 4, 5, 6, 1, time.UTC)
	today     = time.Date(2023, time.July, 2, 4, 5, 6, 1, time.UTC)
	tomorrow  = time.Date(2023, time.July, 3, 4, 5, 6, 1, time.UTC)
)

func TestProgressCompleted(t *testing.T) {
	progress := Progress{CompletedSteps: []Step{
		{Name: DonorPaid},
		{Name: FeeEvidenceNotification, Completed: today},
		{Name: FeeEvidenceSubmitted, Completed: yesterday},
		{Name: FeeEvidenceApproved, Completed: tomorrow},
	}}

	assert.Equal(t, []Step{
		{Name: FeeEvidenceSubmitted, Completed: yesterday},
		{Name: FeeEvidenceNotification, Completed: today},
		{Name: FeeEvidenceApproved, Completed: tomorrow},
	}, progress.Completed(true, false))
}

func TestInProgress(t *testing.T) {
	progress := Progress{CompletedSteps: []Step{
		{Name: FeeEvidenceNotification, Completed: today},
		{Name: DonorPaid},
	}}

	assert.Equal(t, Step{Name: DonorPaid}, progress.InProgress(true))
}

func TestProgressRemainingDonorSteps(t *testing.T) {
	progress := Progress{CompletedSteps: []Step{
		{Name: FeeEvidenceNotification, Completed: today},
		{Name: FeeEvidenceSubmitted, Completed: yesterday},
		{Name: FeeEvidenceApproved},
	}}

	assert.Equal(t, []Step{
		{Name: DonorSignedLPA},
		{Name: CertificateProvided},
		{Name: AllAttorneysSignedLPA},
		{Name: LpaSubmitted},
		{Name: NoticesOfIntentSent},
		{Name: StatutoryWaitingPeriodFinished},
		{Name: LpaRegistered},
	}, progress.RemainingDonorSteps(true))
}

func TestProgressRemainingSupporterSteps(t *testing.T) {
	progress := Progress{CompletedSteps: []Step{
		{Name: FeeEvidenceNotification, Completed: today},
		{Name: FeeEvidenceSubmitted, Completed: yesterday},
		{Name: FeeEvidenceApproved},
	}}

	assert.Equal(t, []Step{
		{Name: DonorPaid},
		{Name: DonorProvedID},
		{Name: DonorSignedLPA},
		{Name: CertificateProvided},
		{Name: AllAttorneysSignedLPA},
		{Name: LpaSubmitted},
		{Name: NoticesOfIntentSent},
		{Name: StatutoryWaitingPeriodFinished},
		{Name: LpaRegistered},
	}, progress.RemainingSupporterSteps())
}
