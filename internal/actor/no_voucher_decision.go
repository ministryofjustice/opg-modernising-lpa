package actor

//go:generate enumerator -type NoVoucherDecision -linecomment -empty
type NoVoucherDecision uint8

const (
	ProveOwnID       NoVoucherDecision = iota + 1 // prove-own-id
	SelectNewVoucher                              // select-new-voucher
	WithdrawLPA                                   // withdraw-lpa
	ApplyToCOP                                    // apply-to-cop
)
