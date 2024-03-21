package supporter

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type donorAccessData struct {
	App       page.AppData
	Errors    validation.List
	Form      *donorAccessForm
	Donor     *actor.DonorProvidedDetails
	ShareCode *actor.ShareCodeData
}

func DonorAccess(tmpl template.Template, donorStore DonorStore, shareCodeStore ShareCodeStore, notifyClient NotifyClient, appPublicURL string, randomString func(int) string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation, member *actor.Member) error {
		donor, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &donorAccessData{
			App:   appData,
			Donor: donor,
			Form:  &donorAccessForm{Email: donor.Donor.Email},
		}

		shareCodeData, err := shareCodeStore.GetDonor(r.Context())
		if err == nil {
			data.ShareCode = &shareCodeData

			switch page.PostFormString(r, "action") {
			case "recall":
				if err := shareCodeStore.Delete(r.Context(), shareCodeData); err != nil {
					return err
				}

				return page.Paths.Supporter.ViewLPA.RedirectQuery(w, r, appData, appData.LpaID, url.Values{
					"inviteRecalledFor": {shareCodeData.InviteSentTo},
				})

			case "remove":
				if donor.Tasks.PayForLpa.IsCompleted() {
					return errors.New("cannot remove LPA access when donor has paid")
				}

				if err := shareCodeStore.Delete(r.Context(), shareCodeData); err != nil {
					return err
				}

				if err := donorStore.DeleteLink(r.Context(), shareCodeData); err != nil {
					return err
				}

				return page.Paths.Supporter.ViewLPA.RedirectQuery(w, r, appData, appData.LpaID, url.Values{
					"accessRemovedFor": {shareCodeData.InviteSentTo},
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

				shareCode := randomString(12)
				shareCodeData := actor.ShareCodeData{
					SessionID:    organisation.ID,
					LpaID:        appData.LpaID,
					ActorUID:     donor.Donor.UID,
					InviteSentTo: data.Form.Email,
				}

				if err := shareCodeStore.PutDonor(r.Context(), shareCode, shareCodeData); err != nil {
					return err
				}

				if err := notifyClient.SendEmail(r.Context(), data.Form.Email, notify.DonorAccessEmail{
					SupporterFullName: member.FullName(),
					OrganisationName:  organisation.Name,
					LpaType:           localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
					DonorName:         donor.Donor.FullName(),
					URL:               appPublicURL + page.Paths.Start.Format(),
					ShareCode:         shareCode,
				}); err != nil {
					return err
				}

				return page.Paths.Supporter.ViewLPA.RedirectQuery(w, r, appData, appData.LpaID, url.Values{
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
