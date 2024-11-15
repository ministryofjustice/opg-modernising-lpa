package voucherdata

import "github.com/ministryofjustice/opg-modernising-lpa/internal/task"

type Tasks struct {
	ConfirmYourName     task.State
	VerifyDonorDetails  task.State
	ConfirmYourIdentity task.IdentityState
	SignTheDeclaration  task.State
}
