package voucherpage

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
)

var (
	expectedError = errors.New("err")
	testAppData   = appcontext.Data{LpaID: "lpa-id"}
)
