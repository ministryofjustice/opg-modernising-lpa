package attorney

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

var (
	expectedError = errors.New("err")
	testAppData   = page.AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     page.Paths,
	}
)
