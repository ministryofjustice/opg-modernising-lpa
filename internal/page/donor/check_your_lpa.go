package donor

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYourLpaData struct {
	App         page.AppData
	Errors      validation.List
	Donor       *actor.DonorProvidedDetails
	Form        *checkYourLpaForm
	Completed   bool
	CanContinue bool
}

type checkYourLpaNotifier struct {
	notifyClient             NotifyClient
	shareCodeSender          ShareCodeSender
	certificateProviderStore CertificateProviderStore
}

func (n *checkYourLpaNotifier) Notify(ctx context.Context, appData page.AppData, donor *actor.DonorProvidedDetails, wasCompleted bool) error {
	if donor.CertificateProvider.CarryOutBy.IsPaper() {
		return n.sendPaperNotification(ctx, appData, donor, wasCompleted)
	}

	return n.sendOnlineNotification(ctx, appData, donor, wasCompleted)
}

func (n *checkYourLpaNotifier) sendPaperNotification(ctx context.Context, appData page.AppData, donor *actor.DonorProvidedDetails, wasCompleted bool) error {
	sms := notify.Sms{
		PhoneNumber: donor.CertificateProvider.Mobile,
		Personalisation: map[string]string{
			"donorFullName":   donor.Donor.FullName(),
			"donorFirstNames": donor.Donor.FirstNames,
		},
	}

	if wasCompleted {
		sms.TemplateID = n.notifyClient.TemplateID(notify.CertificateProviderPaperLpaDetailsChangedSMS)
		sms.Personalisation["lpaId"] = donor.LpaID
	} else {
		sms.TemplateID = n.notifyClient.TemplateID(notify.CertificateProviderPaperMeetingPromptSMS)
		sms.Personalisation["lpaType"] = appData.Localizer.T(donor.Type.LegalTermTransKey())
		sms.Personalisation["CPLandingPageLink"] = "www.gov.uk/opg/certificate-provider"
	}

	_, err := n.notifyClient.Sms(ctx, sms)
	return err
}

func (n *checkYourLpaNotifier) sendOnlineNotification(ctx context.Context, appData page.AppData, donor *actor.DonorProvidedDetails, wasCompleted bool) error {
	if !wasCompleted {
		return n.shareCodeSender.SendCertificateProvider(ctx, notify.CertificateProviderInviteEmail, appData, true, donor)
	}

	certificateProvider, err := n.certificateProviderStore.GetAny(ctx)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return err
	}

	sms := notify.Sms{
		PhoneNumber: donor.CertificateProvider.Mobile,
	}

	if certificateProvider.Tasks.ConfirmYourDetails.NotStarted() {
		sms.TemplateID = n.notifyClient.TemplateID(notify.CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS)
		sms.Personalisation = map[string]string{
			"donorFullName": donor.Donor.FullName(),
			"lpaType":       appData.Localizer.T(donor.Type.LegalTermTransKey()),
		}
	} else {
		sms.TemplateID = n.notifyClient.TemplateID(notify.CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS)
		sms.Personalisation = map[string]string{
			"donorFullNamePossessive": appData.Localizer.Possessive(donor.Donor.FullName()),
			"lpaType":                 appData.Localizer.T(donor.Type.LegalTermTransKey()),
			"lpaId":                   donor.LpaID,
			"donorFirstNames":         donor.Donor.FirstNames,
		}
	}

	_, err = n.notifyClient.Sms(ctx, sms)
	return err
}

func CheckYourLpa(tmpl template.Template, donorStore DonorStore, shareCodeSender ShareCodeSender, notifyClient NotifyClient, certificateProviderStore CertificateProviderStore, now func() time.Time) Handler {
	notifier := &checkYourLpaNotifier{
		notifyClient:             notifyClient,
		shareCodeSender:          shareCodeSender,
		certificateProviderStore: certificateProviderStore,
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &checkYourLpaData{
			App:   appData,
			Donor: donor,
			Form: &checkYourLpaForm{
				CheckedAndHappy: !donor.CheckedAt.IsZero(),
			},
			Completed:   donor.Tasks.CheckYourLpa.Completed(),
			CanContinue: donor.CheckedHash != donor.Hash,
		}

		if r.Method == http.MethodPost && data.CanContinue {
			data.Form = readCheckYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Tasks.CheckYourLpa = actor.TaskCompleted
				donor.CheckedAt = now()

				newHash, err := donor.GenerateHash()
				if err != nil {
					return err
				}
				donor.CheckedHash = newHash

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if err := notifier.Notify(r.Context(), appData, donor, data.Completed); err != nil {
					return err
				}

				if !data.Completed {
					return page.Paths.LpaDetailsSaved.RedirectQuery(w, r, appData, donor, url.Values{"firstCheck": {"1"}})
				}

				return page.Paths.LpaDetailsSaved.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type checkYourLpaForm struct {
	CheckedAndHappy bool
}

func readCheckYourLpaForm(r *http.Request) *checkYourLpaForm {
	return &checkYourLpaForm{
		CheckedAndHappy: page.PostFormString(r, "checked-and-happy") == "1",
	}
}

func (f *checkYourLpaForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("checked-and-happy", "theBoxIfYouHaveCheckedAndHappyToShareLpa", f.CheckedAndHappy,
		validation.Selected())

	return errors
}
