package lpadata

//go:generate go tool enumerator -type AttorneyStatus -trimprefix -linecomment
type AttorneyStatus uint8

const (
	AttorneyStatusActive   AttorneyStatus = iota // active
	AttorneyStatusInactive                       // inactive
	AttorneyStatusRemoved                        // removed
)
