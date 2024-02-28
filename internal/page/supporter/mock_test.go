package supporter

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

var (
	expectedError        = errors.New("err")
	testAppData          = page.AppData{}
	testOrgMemberAppData = page.AppData{
		SessionID:           "session-id",
		Lang:                localize.En,
		LoginSessionEmail:   "supporter@example.com",
		IsSupporter:         true,
		OrganisationName:    "My organisation",
		Permission:          actor.None,
		LoggedInSupporterID: "supporter-id",
	}
)
