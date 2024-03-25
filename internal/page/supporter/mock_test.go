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
	testLpaAppData       = page.AppData{LpaID: "lpa-id"}
	testOrgMemberAppData = page.AppData{
		SessionID:         "session-id",
		Lang:              localize.En,
		LoginSessionEmail: "supporter@example.com",
		SupporterData: &page.SupporterData{
			OrganisationName:    "My organisation",
			Permission:          actor.PermissionNone,
			LoggedInSupporterID: "supporter-id",
		},
	}
	testRandomString   = "12345"
	testRandomStringFn = func(int) string { return testRandomString }
)
