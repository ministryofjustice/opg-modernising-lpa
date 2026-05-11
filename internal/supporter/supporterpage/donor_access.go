package supporterpage

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type donorAccessData struct {
	App           appcontext.Data
	Errors        validation.List
	Form          *donorAccessForm
	Donor         *donordata.Provided
	SupporterLink *supporterdata.LpaLink
}

func DonorAccess(logger Logger, tmpl template.Template, donorStore DonorStore, accessCodeStore AccessCodeStore, notifyClient NotifyClient, donorStartURL string, generate accesscodedata.Generator) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, member *supporterdata.Member) error {
		donor, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &donorAccessData{
			App:   appData,
			Donor: donor,
			Form:  newDonorAccessForm(appData.Localizer),
		}

		data.Form.Email.Input = donor.Donor.Email

		supporterLink, err := accessCodeStore.GetDonorAccess(r.Context())
		if err == nil {
			data.SupporterLink = &supporterLink

			switch page.PostFormString(r, "action") {
			case "recall":
				if err := accessCodeStore.DeleteDonorAccess(r.Context(), supporterLink); err != nil {
					return err
				}

				return supporter.PathViewLPA.RedirectQuery(w, r, appData, appData.LpaID, url.Values{
					"inviteRecalledFor": {supporterLink.InviteSentTo},
				})

			case "remove":
				if donor.Tasks.PayForLpa.IsCompleted() {
					return errors.New("cannot remove LPA access when donor has paid")
				}

				if err := donorStore.DeleteDonorAccess(r.Context(), supporterLink); err != nil {
					return err
				}
				logger.InfoContext(r.Context(), "donor access removed", slog.String("lpa_id", appData.LpaID))

				return supporter.PathViewLPA.RedirectQuery(w, r, appData, appData.LpaID, url.Values{
					"accessRemovedFor": {supporterLink.InviteSentTo},
				})

			default:
				return tmpl(w, data)
			}
		}

		if !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				if data.Form.Email.Value != donor.Donor.Email {
					donor.Donor.Email = data.Form.Email.Value

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}
				}

				plainCode, hashedCode := generate(donor.Donor.LastName)
				accessCodeData := accesscodedata.Link{
					LpaOwnerKey: dynamo.LpaOwnerKey(organisation.PK),
					LpaKey:      dynamo.LpaKey(appData.LpaID),
					LpaUID:      donor.LpaUID,
					ActorUID:    donor.Donor.UID,
				}

				if err := accessCodeStore.PutDonorAccess(r.Context(), hashedCode, accessCodeData, data.Form.Email.Value); err != nil {
					return err
				}

				if err := notifyClient.SendEmail(r.Context(), notify.ToDonorOnly(donor), notify.DonorAccessEmail{
					SupporterFullName:  member.FullName(),
					OrganisationName:   organisation.Name,
					LpaType:            localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
					LpaReferenceNumber: donor.LpaUID,
					DonorName:          donor.Donor.FullName(),
					URL:                donorStartURL,
					AccessCode:         plainCode.Plain(),
				}); err != nil {
					return err
				}

				return supporter.PathViewLPA.RedirectQuery(w, r, appData, appData.LpaID, url.Values{
					"inviteSentTo": {data.Form.Email.Value},
				})
			}
		}

		return tmpl(w, data)
	}
}

type donorAccessForm struct {
	Email  *newforms.String
	Errors []newforms.Field
}

func newDonorAccessForm(l Localizer) *donorAccessForm {
	return &donorAccessForm{
		Email: newforms.NewString("email", l.T("email")).
			NotEmpty().
			Email(),
	}
}

func (f *donorAccessForm) Parse(r *http.Request) bool {
	f.Errors = newforms.ParsePostForm(r,
		f.Email,
	)

	return len(f.Errors) == 0
}
