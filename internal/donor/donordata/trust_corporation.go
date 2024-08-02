package donordata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// TrustCorporation contains details about a trust corporation, provided by the applicant
type TrustCorporation struct {
	// UID for the actor
	UID actoruid.UID
	// Name of the company
	Name string
	// CompanyNumber as registered by Companies House
	CompanyNumber string
	// Email to contact the company
	Email string
	// Address of the company
	Address place.Address
}

func (tc TrustCorporation) Channel() lpadata.Channel {
	if tc.Email != "" {
		return lpadata.ChannelOnline
	}

	return lpadata.ChannelPaper
}
