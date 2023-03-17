package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Donor struct {
	FirstNames  string
	LastName    string
	Email       string
	OtherNames  string
	DateOfBirth date.Date
	Address     place.Address
}

func (d Donor) FullName() string {
	return fmt.Sprintf("%s %s", d.FirstNames, d.LastName)
}
