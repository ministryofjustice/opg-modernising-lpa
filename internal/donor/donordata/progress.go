package donordata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/progress"
)

type Progress struct {
	CompletedSteps []progress.Step
}

func (p *Progress) Complete(name progress.StepName, completed time.Time) {
	p.CompletedSteps = append(p.CompletedSteps, progress.Step{Name: name, Completed: completed})
}
