package donordata

//go:generate go tool enumerator -type NoVoucherDecision -linecomment -empty
type NoVoucherDecision uint8

const (
	ProveOwnIdentity NoVoucherDecision = iota + 1 // prove-own-identity
	SelectNewVoucher                              // select-new-voucher
	WithdrawLPA                                   // withdraw-lpa
	ApplyToCOP                                    // apply-to-cop
)
