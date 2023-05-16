package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type CertificateProvider struct {
	FirstNames              string
	LastName                string
	Address                 place.Address
	Mobile                  string
	Email                   string
	CarryOutBy              string
	DateOfBirth             date.Date
	Relationship            string
	RelationshipDescription string
	RelationshipLength      string
}

func (c CertificateProvider) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}
