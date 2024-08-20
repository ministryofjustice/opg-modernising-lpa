package voucherpage

import (
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
)

var (
	expectedError = errors.New("err")
	testAppData   = appcontext.Data{LpaID: "lpa-id"}
	testNow       = time.Now()
	testNowFn     = func() time.Time { return testNow }
)
