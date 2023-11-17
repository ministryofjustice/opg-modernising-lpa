package actor

//go:generate enumerator -type LpaType -linecomment -trimprefix -empty
type LpaType uint8

const (
	LpaTypeHealthWelfare   LpaType = iota + 1 // hw
	LpaTypePropertyFinance                    // pfa
)

func (e LpaType) LegalTermTransKey() string {
	switch e {
	case LpaTypePropertyFinance:
		return "pfaLegalTerm"
	case LpaTypeHealthWelfare:
		return "hwLegalTerm"
	}
	return ""
}

//go:generate enumerator -type CanBeUsedWhen -linecomment -trimprefix -empty
type CanBeUsedWhen uint8

const (
	CanBeUsedWhenCapacityLost CanBeUsedWhen = iota + 1 // when-capacity-lost
	CanBeUsedWhenHasCapacity                           // when-has-capacity
)

//go:generate enumerator -type LifeSustainingTreatment -linecomment -trimprefix -empty
type LifeSustainingTreatment uint8

const (
	LifeSustainingTreatmentOptionA LifeSustainingTreatment = iota + 1 // option-a
	LifeSustainingTreatmentOptionB                                    // option-b
)

//go:generate enumerator -type ReplacementAttorneysStepIn -linecomment -trimprefix -empty
type ReplacementAttorneysStepIn uint8

const (
	ReplacementAttorneysStepInWhenAllCanNoLongerAct ReplacementAttorneysStepIn = iota + 1 // all
	ReplacementAttorneysStepInWhenOneCanNoLongerAct                                       // one
	ReplacementAttorneysStepInAnotherWay                                                  // other
)
