package page

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

const testRandomString = "123"

var (
	expectedError = errors.New("err")
	testAddress   = place.Address{Line1: "1"}
	TestAppData   = appcontext.Data{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
	}
	testRandomStringFn = func(int) string { return testRandomString }
)
