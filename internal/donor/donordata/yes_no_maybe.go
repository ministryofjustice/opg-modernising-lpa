package donordata

//go:generate go tool enumerator -type YesNoMaybe -linecomment -empty
type YesNoMaybe uint8

const (
	Yes YesNoMaybe = iota + 1
	No
	Maybe
)
