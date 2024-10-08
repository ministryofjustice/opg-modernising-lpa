package supporterpage

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
)

var (
	expectedError        = errors.New("err")
	testAppData          = appcontext.Data{}
	testLpaAppData       = appcontext.Data{LpaID: "lpa-id"}
	testOrgMemberAppData = appcontext.Data{
		SessionID:         "session-id",
		Lang:              localize.En,
		LoginSessionEmail: "supporter@example.com",
		SupporterData: &appcontext.SupporterData{
			OrganisationName:    "My organisation",
			Permission:          supporterdata.PermissionNone,
			LoggedInSupporterID: "supporter-id",
		},
	}
	testRandomString   = "12345"
	testRandomStringFn = func(int) string { return testRandomString }
)
