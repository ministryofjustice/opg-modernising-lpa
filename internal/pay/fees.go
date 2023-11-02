package pay

//go:generate enumerator -type FeeType
type FeeType uint8

const (
	FullFee FeeType = iota
	HalfFee
	NoFee
	HardshipFee
	RepeatApplicationFee
)

//go:generate enumerator -type PreviousFee -empty
type PreviousFee uint8

const (
	PreviousFeeFull PreviousFee = iota + 1
	PreviousFeeHalf
	PreviousFeeExemption
	PreviousFeeHardship
)

func Cost(feeType FeeType, previousFee PreviousFee) int {
	switch feeType {
	case FullFee:
		return 8200
	case HalfFee:
		return 4100
	case RepeatApplicationFee:
		switch previousFee {
		case PreviousFeeFull:
			return 4100
		case PreviousFeeHalf:
			return 2050
		default:
			return 0
		}
	default:
		return 0
	}
}
