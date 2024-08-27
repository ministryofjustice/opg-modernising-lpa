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
	progress := Progress2{Steps: []Step{
		{Name: DonorPaid, State: StateInProgress},
		{Name: FeeEvidenceNotification, State: StateCompleted, Completed: today},
		{Name: FeeEvidenceSubmitted, State: StateCompleted, Completed: yesterday},
		{Name: FeeEvidenceApproved, State: StateCompleted, Completed: tomorrow},
	}}

	assert.Equal(t, []Step{
		{Name: FeeEvidenceSubmitted, State: StateCompleted, Completed: yesterday},
		{Name: FeeEvidenceNotification, State: StateCompleted, Completed: today},
		{Name: FeeEvidenceApproved, State: StateCompleted, Completed: tomorrow},
	}, progress.Completed())
}

func TestInProgress(t *testing.T) {
	progress := Progress2{Steps: []Step{
		{Name: FeeEvidenceNotification, State: StateCompleted, Completed: today},
		{Name: DonorPaid, State: StateInProgress},
	}}

	assert.Equal(t, Step{Name: DonorPaid, State: StateInProgress}, progress.InProgress(true))
}

func TestProgressRemainingDonorSteps(t *testing.T) {
	progress := Progress2{Steps: []Step{
		{Name: FeeEvidenceNotification, State: StateCompleted, Completed: today},
		{Name: FeeEvidenceSubmitted, State: StateCompleted, Completed: yesterday},
		{Name: FeeEvidenceApproved, State: StateInProgress},
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
	progress := Progress2{Steps: []Step{
		{Name: FeeEvidenceNotification, State: StateCompleted, Completed: today},
		{Name: FeeEvidenceSubmitted, State: StateCompleted, Completed: yesterday},
		{Name: FeeEvidenceApproved, State: StateInProgress},
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
