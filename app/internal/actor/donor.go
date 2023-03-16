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

func (d Donor) PossessiveFullName() string {
	if d.LastName == "" {
		format := "%s’s"

		if d.FirstNames[len(d.FirstNames)-1:] == "s" {
			format = "%s’"
		}

		return fmt.Sprintf(format, d.FirstNames)
	} else {
		format := "%s %s’s"

		if d.LastName[len(d.LastName)-1:] == "s" {
			format = "%s %s’"
		}

		return fmt.Sprintf(format, d.FirstNames, d.LastName)
	}
}
