package actor

type ShareCodeData struct {
	PK, SK                string
	SessionID             string
	LpaID                 string
	AttorneyUID           UID
	IsReplacementAttorney bool
	IsTrustCorporation    bool
	DonorFullname         string
	DonorFirstNames       string
}
