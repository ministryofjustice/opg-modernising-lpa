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

type Tracker struct {
	Localizer      Localizer
	PaidFullFee    bool
	Supporter      bool
	CompletedSteps []Step
}

func NewTracker(localizer Localizer) *Tracker {
	return &Tracker{Localizer: localizer}
}

func (t *Tracker) Init(paidFullFee, isSupporter bool, completedSteps []Step) {
	t.PaidFullFee = paidFullFee
	t.Supporter = isSupporter
	t.CompletedSteps = completedSteps
}

func (t *Tracker) DonorSteps() []Step {
	steps := []Step{
		{Name: DonorSignedLPA},
		{Name: CertificateProvided},
		{Name: AllAttorneysSignedLPA},
		{Name: LpaSubmitted},
		{Name: NoticesOfIntentSent, Notification: true},
		{Name: StatutoryWaitingPeriodFinished},
		{Name: LpaRegistered},
	}

	if !t.PaidFullFee {
		steps = slices.Insert(steps, 0,
			Step{Name: FeeEvidenceSubmitted},
			Step{Name: FeeEvidenceNotification, Notification: true},
			Step{Name: FeeEvidenceApproved},
		)
	}

	return steps
}

func (t *Tracker) SupporterSteps() []Step {
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

	if !t.PaidFullFee {
		steps = slices.Insert(steps, 0,
			Step{Name: FeeEvidenceSubmitted},
			Step{Name: FeeEvidenceNotification, Notification: true},
			Step{Name: FeeEvidenceApproved},
		)
	}

	return steps
}

func (t *Tracker) Remaining() (inProgress Step, remaining []Step) {
	allSteps := t.DonorSteps()
	if t.Supporter {
		allSteps = t.SupporterSteps()
	}

	if len(t.CompletedSteps) == 0 {
		return allSteps[0], allSteps[1:]
	}

	removeMap := make(map[StepName]bool)

	for _, toRemove := range t.CompletedSteps {
		removeMap[toRemove.Name] = true
	}

	for _, step := range allSteps {
		if _, ok := removeMap[step.Name]; !ok && step.Show() {
			remaining = append(remaining, step)
		}
	}

	return remaining[0], remaining[1:]
}

func (t *Tracker) Completed() []Step {
	var filteredCompleted []Step

	steps := t.DonorSteps()
	if t.Supporter {
		steps = t.SupporterSteps()
	}

	//Filter out supporter steps appearing in donor steps
	seen := make(map[StepName]bool)
	for _, step := range steps {
		seen[step.Name] = true
	}

	for _, step := range t.CompletedSteps {
		if seen[step.Name] && step.Show() {
			filteredCompleted = append(filteredCompleted, step)
		}
	}

	sort.Slice(filteredCompleted, func(a, b int) bool {
		return t.CompletedSteps[a].Completed.Before(t.CompletedSteps[b].Completed)
	})

	return filteredCompleted
}

func (t *Tracker) IsSupporter() bool {
	return t.Supporter
}
