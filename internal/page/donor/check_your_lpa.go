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
	Lpa         *actor.Lpa
	Form        *checkYourLpaForm
	Completed   bool
	CanContinue bool
}

type checkYourLpaNotifier struct {
	notifyClient             NotifyClient
	shareCodeSender          ShareCodeSender
	certificateProviderStore CertificateProviderStore
}

func (n *checkYourLpaNotifier) Notify(ctx context.Context, appData page.AppData, lpa *actor.Lpa, wasCompleted bool) error {
	if lpa.CertificateProvider.CarryOutBy.IsPaper() {
		return n.sendPaperNotification(ctx, appData, lpa, wasCompleted)
	}

	return n.sendOnlineNotification(ctx, appData, lpa, wasCompleted)
}

func (n *checkYourLpaNotifier) sendPaperNotification(ctx context.Context, appData page.AppData, lpa *actor.Lpa, wasCompleted bool) error {
	sms := notify.Sms{
		PhoneNumber: lpa.CertificateProvider.Mobile,
		Personalisation: map[string]string{
			"donorFullName":   lpa.Donor.FullName(),
			"donorFirstNames": lpa.Donor.FirstNames,
		},
	}

	if wasCompleted {
		sms.TemplateID = n.notifyClient.TemplateID(notify.CertificateProviderPaperLpaDetailsChangedSMS)
		sms.Personalisation["lpaId"] = lpa.ID
	} else {
		sms.TemplateID = n.notifyClient.TemplateID(notify.CertificateProviderPaperMeetingPromptSMS)
		sms.Personalisation["lpaType"] = appData.Localizer.T(lpa.Type.LegalTermTransKey())
	}

	_, err := n.notifyClient.Sms(ctx, sms)
	return err
}

func (n *checkYourLpaNotifier) sendOnlineNotification(ctx context.Context, appData page.AppData, lpa *actor.Lpa, wasCompleted bool) error {
	if !wasCompleted {
		return n.shareCodeSender.SendCertificateProvider(ctx, notify.CertificateProviderInviteEmail, appData, true, lpa)
	}

	certificateProvider, err := n.certificateProviderStore.GetAny(ctx)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return err
	}

	sms := notify.Sms{
		PhoneNumber: lpa.CertificateProvider.Mobile,
	}

	if certificateProvider.Tasks.ConfirmYourDetails.NotStarted() {
		sms.TemplateID = n.notifyClient.TemplateID(notify.CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS)
		sms.Personalisation = map[string]string{
			"donorFullName": lpa.Donor.FullName(),
			"lpaType":       appData.Localizer.T(lpa.Type.LegalTermTransKey()),
		}
	} else {
		sms.TemplateID = n.notifyClient.TemplateID(notify.CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS)
		sms.Personalisation = map[string]string{
			"donorFullNamePossessive": appData.Localizer.Possessive(lpa.Donor.FullName()),
			"lpaType":                 appData.Localizer.T(lpa.Type.LegalTermTransKey()),
			"lpaId":                   lpa.ID,
			"donorFirstNames":         lpa.Donor.FirstNames,
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

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.Lpa) error {
		data := &checkYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &checkYourLpaForm{
				CheckedAndHappy: !lpa.CheckedAt.IsZero(),
			},
			Completed:   lpa.Tasks.CheckYourLpa.Completed(),
			CanContinue: lpa.CheckedHash != lpa.Hash,
		}

		if r.Method == http.MethodPost && data.CanContinue {
			data.Form = readCheckYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.Tasks.CheckYourLpa = actor.TaskCompleted
				lpa.CheckedAt = now()

				newHash, err := lpa.GenerateHash()
				if err != nil {
					return err
				}
				lpa.CheckedHash = newHash

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if err := notifier.Notify(r.Context(), appData, lpa, data.Completed); err != nil {
					return err
				}

				if !data.Completed {
					return page.Paths.LpaDetailsSaved.RedirectQuery(w, r, appData, lpa, url.Values{"firstCheck": {"1"}})
				}

				return page.Paths.LpaDetailsSaved.Redirect(w, r, appData, lpa)
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
