package lpadata

//go:generate enumerator -type AttorneyStatus -trimprefix -linecomment
type AttorneyStatus uint8

const (
	AttorneyStatusActive   AttorneyStatus = iota // active
	AttorneyStatusInactive                       // inactive
	AttorneyStatusRemoved                        // removed
)
