package lpadata

//go:generate go tool enumerator -type CanBeUsedWhen -linecomment -trimprefix -empty
type CanBeUsedWhen uint8

const (
	CanBeUsedWhenCapacityLost CanBeUsedWhen = iota + 1 // when-capacity-lost
	CanBeUsedWhenHasCapacity                           // when-has-capacity
)
