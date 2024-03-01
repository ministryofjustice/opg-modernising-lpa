package page

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

const RandomString = "123"

var (
	expectedError = errors.New("err")
	TestAppData   = AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
	}
	MockRandomString = func(int) string { return RandomString }
)
