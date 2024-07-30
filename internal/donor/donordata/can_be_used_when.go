package donordata

//go:generate enumerator -type CanBeUsedWhen -linecomment -trimprefix
type CanBeUsedWhen uint8

const (
	CanBeUsedWhenUnknown      CanBeUsedWhen = iota
	CanBeUsedWhenCapacityLost               // when-capacity-lost
	CanBeUsedWhenHasCapacity                // when-has-capacity
)
