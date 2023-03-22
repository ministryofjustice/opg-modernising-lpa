package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type CertificateProvider struct {
	FirstNames              string
	LastName                string
	Email                   string
	Address                 place.Address
	Mobile                  string
	DateOfBirth             date.Date
	CarryOutBy              string
	Relationship            string
	RelationshipDescription string
	RelationshipLength      string
	DeclaredFullName        string
}

func (c CertificateProvider) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}
