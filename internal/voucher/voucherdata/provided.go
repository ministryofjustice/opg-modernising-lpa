package voucherdata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

// Provided contains the information a voucher has given
type Provided struct {
	PK dynamo.LpaKeyType
	SK dynamo.VoucherKeyType
	// LpaID is for the LPA the voucher is provided a vouch for
	LpaID string
	// UpdatedAt is the time that this data was last updated
	UpdatedAt time.Time
	// Tasks shows the state of the actions the voucher will do
	Tasks Tasks
	// Email is the email address of the voucher
	Email string
	// FirstNames is the first names provided by the voucher. If set it overrides
	// that provided by the donor.
	FirstNames string
	// LastName is a last name provided by the voucher. If set it overrides that
	// provided by the donor.
	LastName string
	// DonorDetailsMatch records whether the voucher confirms that the details
	// presented to them match the donor they expected to vouch for.
	DonorDetailsMatch form.YesNo
}

type Tasks struct {
	ConfirmYourName     task.State
	VerifyDonorDetails  task.State
	ConfirmYourIdentity task.State
	SignTheDeclaration  task.State
}
