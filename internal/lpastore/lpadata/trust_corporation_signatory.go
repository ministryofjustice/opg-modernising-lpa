package lpadata

import "time"

type TrustCorporationSignatory struct {
	FirstNames        string    `json:"firstNames"`
	LastName          string    `json:"lastName"`
	ProfessionalTitle string    `json:"professionalTitle"`
	SignedAt          time.Time `json:"signedAt"`
}
