package actor

import (
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// TODO Once updated all cp handlers, look at re-working cp.CertificateProviderDetails to be a details type rather than actual CP
type CertificateProvider struct {
	ID                      string
	LpaID                   string
	UpdatedAt               time.Time
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
	IdentityOption          identity.Option
	IdentityUserData        identity.UserData
}

func (c CertificateProvider) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}

func (c *CertificateProvider) CertificateProviderIdentityConfirmed() bool {
	return c.IdentityUserData.OK && c.IdentityUserData.Provider != identity.UnknownOption &&
		c.IdentityUserData.MatchName(c.FirstNames, c.LastName) &&
		c.IdentityUserData.DateOfBirth.Equals(c.DateOfBirth)
}
