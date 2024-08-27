package task

import (
	"slices"
	"sort"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

//go:generate enumerator -type StepName -empty
type StepName uint8

const (
	FeeEvidenceSubmitted StepName = iota + 1
	FeeEvidenceNotification
	FeeEvidenceApproved
	DonorPaid
	DonorProvedID
	DonorSignedLPA
	CertificateProvided
	AllAttorneysSignedLPA
	NoticesOfIntentSent
	LpaSubmitted
	StatutoryWaitingPeriodFinished
	LpaRegistered
)

type Progress2 struct {
	Steps []Step
}

type Step struct {
	Name      StepName
	Completed time.Time
	State     State
}

func (p *Progress2) RemainingDonorSteps(fullFee bool) []Step {
	steps := []Step{
		{Name: DonorSignedLPA},
		{Name: CertificateProvided},
		{Name: AllAttorneysSignedLPA},
		{Name: LpaSubmitted},
		{Name: NoticesOfIntentSent},
		{Name: StatutoryWaitingPeriodFinished},
		{Name: LpaRegistered},
	}

	if !fullFee {
		steps = slices.Insert(steps, 0,
			Step{Name: FeeEvidenceSubmitted},
			Step{Name: FeeEvidenceApproved},
		)
	}

	filteredSteps := make([]Step, 0)
	removeMap := make(map[StepName]bool)

	for _, toRemove := range p.Steps {
		removeMap[toRemove.Name] = true
	}
	for _, step := range steps {
		if _, ok := removeMap[step.Name]; !ok {
			filteredSteps = append(filteredSteps, step)
		}
	}

	return filteredSteps
}

func (p *Progress2) RemainingSupporterSteps() []Step {
	steps := []Step{
		{Name: DonorPaid},
		{Name: DonorProvedID},
		{Name: DonorSignedLPA},
		{Name: CertificateProvided},
		{Name: AllAttorneysSignedLPA},
		{Name: LpaSubmitted},
		{Name: NoticesOfIntentSent},
		{Name: StatutoryWaitingPeriodFinished},
		{Name: LpaRegistered},
	}

	filteredSteps := make([]Step, 0)
	removeMap := make(map[StepName]bool)

	for _, toRemove := range p.Steps {
		removeMap[toRemove.Name] = true
	}
	for _, step := range steps {
		if _, ok := removeMap[step.Name]; !ok {
			filteredSteps = append(filteredSteps, step)
		}
	}

	return filteredSteps
}

func (p *Progress2) Complete(name StepName, now time.Time) {
	p.Steps = append(p.Steps, Step{Name: name, Completed: now, State: StateCompleted})
}

func (p *Progress2) Completed() []Step {
	var completed []Step
	for _, step := range p.Steps {
		if !step.Completed.IsZero() {
			completed = append(completed, step)
		}
	}

	sort.Slice(completed, func(a, b int) bool {
		return p.Steps[a].Completed.Before(p.Steps[b].Completed)
	})

	return completed
}

func (p *Progress2) InProgress(fullFee bool) Step {
	return p.RemainingDonorSteps(fullFee)[0]
}

func (s Step) DonorLabel(l Localizer, lpa *lpadata.Lpa) string {
	switch s.Name {
	case FeeEvidenceSubmitted:
		return l.T("yourLPAFeeEvidenceHasBeenSubmitted")
	case FeeEvidenceNotification:
		return l.Format(
			"weEmailedYouOnAbout",
			map[string]interface{}{
				"On":    l.FormatDate(s.Completed),
				"About": l.T("yourFee"),
			},
		)
	case FeeEvidenceApproved:
		return l.T("yourLPAFeeEvidenceHasBeenApproved")
	case DonorSignedLPA:
		return l.T("youveSignedYourLpa")
	case CertificateProvided:
		if lpa.CertificateProvider.FirstNames != "" {
			return l.Format(
				"certificateProviderHasDeclared",
				map[string]interface{}{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
			)
		} else {
			return l.T("yourCertificateProviderHasDeclared")
		}
	case AllAttorneysSignedLPA:
		return l.Count("attorneysHaveDeclared", len(lpa.Attorneys.Attorneys))
	case LpaSubmitted:
		return l.T("weHaveReceivedYourLpa")
	case NoticesOfIntentSent:
		return l.Format("weSentAnEmailYourLpaIsReadyToRegister", map[string]any{
			"SentOn": l.FormatDate(lpa.PerfectAt),
		})
	case StatutoryWaitingPeriodFinished:
		return l.T("yourWaitingPeriodHasStarted")
	case LpaRegistered:
		return l.T("yourLpaHasBeenRegistered")
	default:
		return ""
	}
}

func (s Step) SupporterLabel() string {
	switch s.Name {
	case FeeEvidenceSubmitted:
		return ""
	case FeeEvidenceApproved:
		return ""
	case DonorPaid:
		return ""
	case DonorProvedID:
		return ""
	case DonorSignedLPA:
		return ""
	case CertificateProvided:
		return ""
	case AllAttorneysSignedLPA:
		return ""
	case LpaSubmitted:
		return ""
	case NoticesOfIntentSent:
		return ""
	case StatutoryWaitingPeriodFinished:
		return ""
	case LpaRegistered:
		return ""
	default:
		return ""
	}
}
