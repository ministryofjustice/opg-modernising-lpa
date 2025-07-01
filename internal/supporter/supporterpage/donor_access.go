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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type donorAccessData struct {
	App        appcontext.Data
	Errors     validation.List
	Form       *donorAccessForm
	Donor      *donordata.Provided
	AccessCode *accesscodedata.Link
}

func DonorAccess(logger Logger, tmpl template.Template, donorStore DonorStore, accessCodeStore AccessCodeStore, notifyClient NotifyClient, donorStartURL string, generate func() (accesscodedata.PlainText, accesscodedata.Hashed)) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, member *supporterdata.Member) error {
		donor, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &donorAccessData{
			App:   appData,
			Donor: donor,
			Form:  &donorAccessForm{Email: donor.Donor.Email},
		}

		accessCodeData, err := accessCodeStore.GetDonor(r.Context())
		if err == nil {
			data.AccessCode = &accessCodeData

			switch page.PostFormString(r, "action") {
			case "recall":
				if err := accessCodeStore.Delete(r.Context(), accessCodeData); err != nil {
					return err
				}

				return supporter.PathViewLPA.RedirectQuery(w, r, appData, appData.LpaID, url.Values{
					"inviteRecalledFor": {accessCodeData.InviteSentTo},
				})

			case "remove":
				if donor.Tasks.PayForLpa.IsCompleted() {
					return errors.New("cannot remove LPA access when donor has paid")
				}

				if err := donorStore.DeleteDonorAccess(r.Context(), accessCodeData); err != nil {
					return err
				}
				logger.InfoContext(r.Context(), "donor access removed", slog.String("lpa_id", appData.LpaID))

				return supporter.PathViewLPA.RedirectQuery(w, r, appData, appData.LpaID, url.Values{
					"accessRemovedFor": {accessCodeData.InviteSentTo},
				})

			default:
				return tmpl(w, data)
			}
		}

		if !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}

		if r.Method == http.MethodPost {
			data.Form = readDonorAccessForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.Email != donor.Donor.Email {
					donor.Donor.Email = data.Form.Email

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}
				}

				plainCode, hashedCode := generate()
				accessCodeData := accesscodedata.Link{
					LpaOwnerKey:  dynamo.LpaOwnerKey(organisation.PK),
					LpaKey:       dynamo.LpaKey(appData.LpaID),
					LpaUID:       donor.LpaUID,
					ActorUID:     donor.Donor.UID,
					InviteSentTo: data.Form.Email,
				}

				if err := accessCodeStore.PutDonor(r.Context(), hashedCode, accessCodeData); err != nil {
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
					"inviteSentTo": {data.Form.Email},
				})
			}
		}

		return tmpl(w, data)
	}
}

type donorAccessForm struct {
	Email string
}

func readDonorAccessForm(r *http.Request) *donorAccessForm {
	return &donorAccessForm{
		Email: page.PostFormString(r, "email"),
	}
}

func (f *donorAccessForm) Validate() validation.List {
	var errors validation.List

	errors.String("email", "email", f.Email,
		validation.Empty(),
		validation.Email())

	return errors
}
