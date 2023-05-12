package actor

import (
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type CertificateProviderProvidedDetails struct {
	LpaID            string
	UpdatedAt        time.Time
	FirstNames       string
	LastName         string
	Email            string
	Address          place.Address
	Mobile           string
	DateOfBirth      date.Date
	DeclaredFullName string
	IdentityOption   identity.Option
	IdentityUserData identity.UserData
	Certificate      Certificate
}

type Certificate struct {
	AgreeToStatement bool
	Agreed           time.Time
}

func (c CertificateProviderProvidedDetails) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}

func (c *CertificateProviderProvidedDetails) CertificateProviderIdentityConfirmed() bool {
	return c.IdentityUserData.OK && c.IdentityUserData.Provider != identity.UnknownOption &&
		c.IdentityUserData.MatchName(c.FirstNames, c.LastName) &&
		c.IdentityUserData.DateOfBirth.Equals(c.DateOfBirth)
}
