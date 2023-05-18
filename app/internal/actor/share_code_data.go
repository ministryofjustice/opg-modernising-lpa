package actor

type ShareCodeData struct {
	SessionID             string
	LpaID                 string
	Identity              bool
	AttorneyID            string
	IsReplacementAttorney bool
	DonorFullname         string
	DonorFirstNames       string
}
