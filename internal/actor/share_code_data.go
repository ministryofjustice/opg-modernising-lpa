package actor

type ShareCodeData struct {
	PK, SK                string
	SessionID             string
	LpaID                 string
	AttorneyID            string
	IsReplacementAttorney bool
	IsTrustCorporation    bool
	DonorFullname         string
	DonorFirstNames       string
}
