package pay

const (
	FeeFull    = 9200
	FeeHalf    = 4600
	FeeQuarter = 2300
)

//go:generate go tool enumerator -type FeeType
type FeeType uint8

const (
	FullFee FeeType = iota
	HalfFee
	QuarterFee
	NoFee
	HardshipFee
	RepeatApplicationFee
)

//go:generate go tool enumerator -type PreviousFee -empty -trimprefix
type PreviousFee uint8

const (
	PreviousFeeFull PreviousFee = iota + 1
	PreviousFeeHalf
	PreviousFeeExemption
	PreviousFeeHardship
)

//go:generate go tool enumerator -type CostOfRepeatApplication -empty -trimprefix
type CostOfRepeatApplication uint8

const (
	CostOfRepeatApplicationNoFee CostOfRepeatApplication = iota + 1
	CostOfRepeatApplicationHalfFee
)

func Cost(feeType FeeType, previousFee PreviousFee, costOfRepeatApplication CostOfRepeatApplication) int {
	switch feeType {
	case FullFee:
		return FeeFull
	case HalfFee:
		return FeeHalf
	case QuarterFee:
		return FeeQuarter
	case RepeatApplicationFee:
		if costOfRepeatApplication.IsNoFee() {
			return 0
		}

		switch previousFee {
		case PreviousFeeFull:
			return FeeHalf
		case PreviousFeeHalf:
			return FeeQuarter
		default:
			return 0
		}
	default:
		return 0
	}
}
