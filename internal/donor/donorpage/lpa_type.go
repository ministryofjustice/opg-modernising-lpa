package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
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
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &lpaTypeData{
			App: appData,
			Form: &lpaTypeForm{
				LpaType: donor.Type,
			},
			Options:     lpadata.LpaTypeValues,
			CanTaskList: !donor.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readLpaTypeForm(r)
			data.Errors = data.Form.Validate(donor.Attorneys.TrustCorporation.Name != "" || donor.ReplacementAttorneys.TrustCorporation.Name != "")

			if data.Errors.None() {
				session, err := appcontext.SessionFromContext(r.Context())
				if err != nil {
					return err
				}

				donor.Type = data.Form.LpaType
				if donor.Type.IsPersonalWelfare() {
					donor.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenCapacityLost
				}
				donor.Tasks.YourDetails = task.StateCompleted
				donor.HasSentApplicationUpdatedEvent = false

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if err := eventClient.SendUidRequested(r.Context(), event.UidRequested{
					LpaID:          donor.LpaID,
					DonorSessionID: session.SessionID,
					OrganisationID: session.OrganisationID,
					Type:           donor.Type.String(),
					Donor: uid.DonorDetails{
						Name:     donor.Donor.FullName(),
						Dob:      donor.Donor.DateOfBirth,
						Postcode: donor.Donor.Address.Postcode,
					},
				}); err != nil {
					return err
				}

				return page.Paths.TaskList.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type lpaTypeForm struct {
	LpaType lpadata.LpaType
	Error   error
}

func readLpaTypeForm(r *http.Request) *lpaTypeForm {
	lpaType, err := lpadata.ParseLpaType(page.PostFormString(r, "lpa-type"))

	return &lpaTypeForm{
		LpaType: lpaType,
		Error:   err,
	}
}

func (f *lpaTypeForm) Validate(hasTrustCorporation bool) validation.List {
	var errors validation.List

	errors.Error("lpa-type", "theTypeOfLpaToMake", f.Error,
		validation.Selected())

	if f.LpaType.IsPersonalWelfare() && hasTrustCorporation {
		errors.Add("lpa-type", validation.CustomError{Label: "youMustDeleteTrustCorporationToChangeLpaType"})
	}

	return errors
}