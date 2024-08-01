package donordata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Correspondent struct {
	FirstNames   string
	LastName     string
	Email        string
	Organisation string
	Telephone    string
	WantAddress  form.YesNo
	Address      place.Address
}

func (c Correspondent) FullName() string {
	return c.FirstNames + " " + c.LastName
}
