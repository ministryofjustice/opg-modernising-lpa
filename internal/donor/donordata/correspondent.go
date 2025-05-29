package donordata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Correspondent struct {
	UID          actoruid.UID
	FirstNames   string
	LastName     string
	Email        string
	Organisation string
	Phone        string
	WantAddress  form.YesNo
	Address      place.Address
}

func (c Correspondent) FullName() string {
	return c.FirstNames + " " + c.LastName
}

func (c Correspondent) NameHasChanged(firstNames, lastName string) bool {
	return c.FirstNames != firstNames || c.LastName != lastName
}
