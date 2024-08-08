package page

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

const testRandomString = "123"

var (
	expectedError = errors.New("err")
	TestAppData   = appcontext.Data{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
	}
	testRandomStringFn = func(int) string { return testRandomString }
)
