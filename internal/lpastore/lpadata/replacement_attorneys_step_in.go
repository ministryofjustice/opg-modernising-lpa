package lpadata

//go:generate enumerator -type ReplacementAttorneysStepIn -linecomment -trimprefix -empty
type ReplacementAttorneysStepIn uint8

const (
	ReplacementAttorneysStepInWhenAllCanNoLongerAct ReplacementAttorneysStepIn = iota + 1 // all-can-no-longer-act
	ReplacementAttorneysStepInWhenOneCanNoLongerAct                                       // one-can-no-longer-act
	ReplacementAttorneysStepInAnotherWay                                                  // another-way
)
