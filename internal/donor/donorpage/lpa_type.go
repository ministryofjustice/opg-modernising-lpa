package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lpaTypeData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *lpaTypeForm
	Options     lpadata.LpaTypeOptions
	CanTaskList bool
}

func LpaType(tmpl template.Template, donorStore DonorStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &lpaTypeData{
			App: appData,
			Form: &lpaTypeForm{
				LpaType: provided.Type,
			},
			Options:     lpadata.LpaTypeValues,
			CanTaskList: !provided.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readLpaTypeForm(r)
			data.Errors = data.Form.Validate(provided.Attorneys.TrustCorporation.Name != "" || provided.ReplacementAttorneys.TrustCorporation.Name != "")

			if data.Errors.None() {
				session, err := appcontext.SessionFromContext(r.Context())
				if err != nil {
					return err
				}

				provided.Type = data.Form.LpaType
				if provided.Type.IsPersonalWelfare() {
					provided.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenCapacityLost
				}
				provided.Tasks.YourDetails = task.StateCompleted
				provided.HasSentApplicationUpdatedEvent = false

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if err := eventClient.SendUidRequested(r.Context(), event.UidRequested{
					LpaID:          provided.LpaID,
					DonorSessionID: session.SessionID,
					OrganisationID: session.OrganisationID,
					Type:           provided.Type.String(),
					Donor: uid.DonorDetails{
						Name:     provided.Donor.FullName(),
						Dob:      provided.Donor.DateOfBirth,
						Postcode: provided.Donor.Address.Postcode,
					},
				}); err != nil {
					return err
				}

				return donor.PathTaskList.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type lpaTypeForm struct {
	LpaType lpadata.LpaType
}

func readLpaTypeForm(r *http.Request) *lpaTypeForm {
	lpaType, _ := lpadata.ParseLpaType(page.PostFormString(r, "lpa-type"))

	return &lpaTypeForm{
		LpaType: lpaType,
	}
}

func (f *lpaTypeForm) Validate(hasTrustCorporation bool) validation.List {
	var errors validation.List

	errors.Enum("lpa-type", "theTypeOfLpaToMake", f.LpaType,
		validation.Selected())

	if f.LpaType.IsPersonalWelfare() && hasTrustCorporation {
		errors.Add("lpa-type", validation.CustomError{Label: "youMustDeleteTrustCorporationToChangeLpaType"})
	}

	return errors
}
