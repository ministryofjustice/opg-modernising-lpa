package lpadata

//go:generate go tool enumerator -type LpaType -linecomment -trimprefix -empty
type LpaType uint8

const (
	LpaTypePersonalWelfare    LpaType = iota + 1 // personal-welfare
	LpaTypePropertyAndAffairs                    // property-and-affairs
)
