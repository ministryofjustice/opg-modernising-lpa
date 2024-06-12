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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
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
	appPublicURL             string
}

func (n *checkYourLpaNotifier) Notify(ctx context.Context, appData page.AppData, donor *actor.DonorProvidedDetails, wasCompleted bool) error {
	if donor.CertificateProvider.CarryOutBy.IsPaper() {
		return n.sendPaperNotification(ctx, appData, donor, wasCompleted)
	}

	return n.sendOnlineNotification(ctx, appData, donor, wasCompleted)
}

func (n *checkYourLpaNotifier) sendPaperNotification(ctx context.Context, appData page.AppData, donor *actor.DonorProvidedDetails, wasCompleted bool) error {
	var sms notify.SMS
	if wasCompleted {
		sms = notify.CertificateProviderActingOnPaperDetailsChangedSMS{
			DonorFullName:   donor.Donor.FullName(),
			DonorFirstNames: donor.Donor.FirstNames,
			LpaUID:          donor.LpaUID,
		}
	} else {
		sms = notify.CertificateProviderActingOnPaperMeetingPromptSMS{
			DonorFullName:                   donor.Donor.FullName(),
			DonorFirstNames:                 donor.Donor.FirstNames,
			LpaType:                         localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
			CertificateProviderStartPageURL: n.appPublicURL + appData.Lang.URL(page.Paths.CertificateProviderStart.Format()),
		}
	}

	return n.notifyClient.SendActorSMS(ctx, donor.CertificateProvider.Mobile, donor.LpaUID, sms)
}

func (n *checkYourLpaNotifier) sendOnlineNotification(ctx context.Context, appData page.AppData, donor *actor.DonorProvidedDetails, wasCompleted bool) error {
	if !wasCompleted {
		return n.shareCodeSender.SendCertificateProviderInvite(ctx, appData, page.CertificateProviderInvite{
			LpaKey:                      donor.PK,
			LpaOwnerKey:                 donor.SK,
			LpaUID:                      donor.LpaUID,
			Type:                        donor.Type,
			Donor:                       donor.Donor,
			CertificateProviderUID:      donor.CertificateProvider.UID,
			CertificateProviderFullName: donor.CertificateProvider.FullName(),
			CertificateProviderEmail:    donor.CertificateProvider.Email,
		})
	}

	certificateProvider, err := n.certificateProviderStore.GetAny(ctx)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return err
	}

	var sms notify.SMS

	if certificateProvider.Tasks.ConfirmYourDetails.NotStarted() {
		sms = notify.CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS{
			LpaType:       localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
			DonorFullName: donor.Donor.FullName(),
		}
	} else {
		sms = notify.CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS{
			LpaType:                 localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
			DonorFullNamePossessive: appData.Localizer.Possessive(donor.Donor.FullName()),
			DonorFirstNames:         donor.Donor.FirstNames,
		}
	}

	return n.notifyClient.SendActorSMS(ctx, donor.CertificateProvider.Mobile, donor.LpaUID, sms)
}

func CheckYourLpa(tmpl template.Template, donorStore DonorStore, shareCodeSender ShareCodeSender, notifyClient NotifyClient, certificateProviderStore CertificateProviderStore, now func() time.Time, appPublicURL string) Handler {
	notifier := &checkYourLpaNotifier{
		notifyClient:             notifyClient,
		shareCodeSender:          shareCodeSender,
		certificateProviderStore: certificateProviderStore,
		appPublicURL:             appPublicURL,
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		newHash, err := donor.GenerateCheckedHash()
		if err != nil {
			return err
		}

		data := &checkYourLpaData{
			App:   appData,
			Donor: donor,
			Form: &checkYourLpaForm{
				CheckedAndHappy: !donor.CheckedAt.IsZero(),
			},
			Completed:   donor.Tasks.CheckYourLpa.Completed(),
			CanContinue: donor.CheckedHash != newHash,
		}

		if r.Method == http.MethodPost && data.CanContinue {
			data.Form = readCheckYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Tasks.CheckYourLpa = actor.TaskCompleted
				donor.CheckedAt = now()
				donor.CheckedHash = newHash

				if err := notifier.Notify(r.Context(), appData, donor, data.Completed); err != nil {
					return err
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
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
