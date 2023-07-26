package actor

import "github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"

type TrustCorporation struct {
	Name          string
	CompanyNumber string
	Email         string
	Address       place.Address
}
