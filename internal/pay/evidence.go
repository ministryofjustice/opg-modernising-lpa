package pay

//go:generate go tool enumerator -type EvidenceDelivery -linecomment -empty
type EvidenceDelivery uint8

const (
	Upload EvidenceDelivery = iota + 1 // upload
	Post                               // post
)
