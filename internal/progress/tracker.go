package progress

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

type ProgressTracker struct {
	Localizer      Localizer
	PaidFullFee    bool
	Supporter      bool
	CompletedSteps []Step
}

func NewProgressTracker(localizer Localizer) *ProgressTracker {
	return &ProgressTracker{Localizer: localizer}
}

func (pt *ProgressTracker) Init(paidFullFee, isSupporter bool, completedSteps []Step) {
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

	if len(pt.CompletedSteps) == 0 {
		return allSteps[0], allSteps[1:]
	}

	removeMap := make(map[StepName]bool)

	for _, toRemove := range pt.CompletedSteps {
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

	for _, step := range pt.CompletedSteps {
		if seen[step.Name] && step.Show() {
			filteredCompleted = append(filteredCompleted, step)
		}
	}

	sort.Slice(filteredCompleted, func(a, b int) bool {
		return pt.CompletedSteps[a].Completed.Before(pt.CompletedSteps[b].Completed)
	})

	return filteredCompleted
}

func (pt *ProgressTracker) IsSupporter() bool {
	return pt.Supporter
}
