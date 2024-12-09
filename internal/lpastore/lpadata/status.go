package lpadata

//go:generate enumerator -type Status -trimprefix -linecomment
type Status uint8

const (
	StatusInProgress             Status = iota // in-progress
	StatusStatutoryWaitingPeriod               // statutory-waiting-period
	StatusRegistered                           // registered
	StatusCannotRegister                       // cannot-register
	StatusWithdrawn                            // withdrawn
	StatusCancelled                            // cancelled
	StatusDoNotRegister                        // do-not-register
	StatusExpired                              // expired
)
