package supporterdata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

// An Organisation contains users associated with a set of permissions that work on the
// same set of LPAs.
type Organisation struct {
	PK, SK dynamo.OrganisationKeyType
	// CreatedAt is when the Organisation was created
	CreatedAt time.Time
	// UpdatedAt is when the Organisation was last updated
	UpdatedAt time.Time
	// DeletedAt is when the Organisation was (soft) deleted
	DeletedAt time.Time
	// ID is a unique identifier for the Organisation
	ID string
	// Name of the Organisation, this is unique across all Organisations
	Name string
}
