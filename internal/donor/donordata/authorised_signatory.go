package donordata

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
)

// AuthorisedSignatory contains details of the person who will sign the LPA on the donor's behalf
type AuthorisedSignatory struct {
	UID        actoruid.UID
	FirstNames string
	LastName   string
}

func (s AuthorisedSignatory) FullName() string {
	return fmt.Sprintf("%s %s", s.FirstNames, s.LastName)
}
