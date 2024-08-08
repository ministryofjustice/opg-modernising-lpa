package lpadata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type TrustCorporation struct {
	UID                       actoruid.UID
	Name                      string
	CompanyNumber             string
	Email                     string
	Address                   place.Address
	Mobile                    string
	Signatories               []TrustCorporationSignatory
	ContactLanguagePreference localize.Lang
	Channel                   Channel
}
