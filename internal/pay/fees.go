package pay

const (
	feeFull    = 8200
	feeHalf    = 4100
	feeQuarter = 2050
	feeNoFee   = 0
)

//go:generate enumerator -type FeeType
type FeeType uint8

const (
	FullFee FeeType = iota
	HalfFee
	QuarterFee
	NoFee
	HardshipFee
	RepeatApplicationFee
)

//go:generate enumerator -type PreviousFee -empty -trimprefix
type PreviousFee uint8

const (
	PreviousFeeFull PreviousFee = iota + 1
	PreviousFeeHalf
	PreviousFeeExemption
	PreviousFeeHardship
)

//go:generate enumerator -type CostOfRepeatApplication -empty -trimprefix
type CostOfRepeatApplication uint8

const (
	CostOfRepeatApplicationNoFee CostOfRepeatApplication = iota + 1
	CostOfRepeatApplicationHalfFee
)

func Cost(feeType FeeType, previousFee PreviousFee, costOfRepeatApplication CostOfRepeatApplication) int {
	switch feeType {
	case FullFee:
		return feeFull
	case HalfFee:
		return feeHalf
	case RepeatApplicationFee:
		if costOfRepeatApplication.IsHalfFee() || previousFee.IsFull() {
			return feeHalf
		} else if previousFee.IsHalf() {
			return feeQuarter
		} else {
			return 0
		}
	default:
		return 0
	}
}
