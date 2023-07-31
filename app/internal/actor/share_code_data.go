package actor

type ShareCodeData struct {
	PK, SK                string
	SessionID             string
	LpaID                 string
	Identity              bool
	AttorneyID            string
	IsReplacementAttorney bool
	IsTrustCorporation    bool
	DonorFullname         string
	DonorFirstNames       string
}
