package donordata

//go:generate enumerator -type LifeSustainingTreatment -linecomment -trimprefix -empty
type LifeSustainingTreatment uint8

const (
	LifeSustainingTreatmentOptionA LifeSustainingTreatment = iota + 1 // option-a
	LifeSustainingTreatmentOptionB                                    // option-b
)
