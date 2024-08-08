package donorpage

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYourLpaData struct {
	App                          appcontext.Data
	Errors                       validation.List
	Donor                        *donordata.Provided
	Form                         *checkYourLpaForm
	CertificateProviderContacted bool
	CanContinue                  bool
}

type checkYourLpaNotifier struct {
	notifyClient             NotifyClient
	shareCodeSender          ShareCodeSender
	certificateProviderStore CertificateProviderStore
	appPublicURL             string
}

func (n *checkYourLpaNotifier) Notify(ctx context.Context, appData appcontext.Data, donor *donordata.Provided, wasCompleted bool) error {
	if donor.CertificateProvider.CarryOutBy.IsPaper() {
		return n.sendPaperNotification(ctx, appData, donor, wasCompleted)
	}

	return n.sendOnlineNotification(ctx, appData, donor, wasCompleted)
}

func (n *checkYourLpaNotifier) sendPaperNotification(ctx context.Context, appData appcontext.Data, provided *donordata.Provided, wasCompleted bool) error {
	var sms notify.SMS
	if wasCompleted {
		sms = notify.CertificateProviderActingOnPaperDetailsChangedSMS{
			DonorFullName:   provided.Donor.FullName(),
			DonorFirstNames: provided.Donor.FirstNames,
			LpaUID:          provided.LpaUID,
		}
	} else {
		sms = notify.CertificateProviderActingOnPaperMeetingPromptSMS{
			DonorFullName:                   provided.Donor.FullName(),
			DonorFirstNames:                 provided.Donor.FirstNames,
			LpaType:                         localize.LowerFirst(appData.Localizer.T(provided.Type.String())),
			CertificateProviderStartPageURL: n.appPublicURL + appData.Lang.URL(page.PathCertificateProviderStart.Format()),
		}
	}

	return n.notifyClient.SendActorSMS(ctx, provided.CertificateProvider.Mobile, provided.LpaUID, sms)
}

func (n *checkYourLpaNotifier) sendOnlineNotification(ctx context.Context, appData appcontext.Data, donor *donordata.Provided, wasCompleted bool) error {
	if !wasCompleted {
		return n.shareCodeSender.SendCertificateProviderInvite(ctx, appData, sharecode.CertificateProviderInvite{
			LpaKey:                      donor.PK,
			LpaOwnerKey:                 donor.SK,
			LpaUID:                      donor.LpaUID,
			Type:                        donor.Type,
			DonorFirstNames:             donor.Donor.FirstNames,
			DonorFullName:               donor.Donor.FullName(),
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

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &checkYourLpaData{
			App:   appData,
			Donor: provided,
			Form: &checkYourLpaForm{
				CheckedAndHappy: !provided.CheckedAt.IsZero(),
			},
			CertificateProviderContacted: !provided.CheckedAt.IsZero(),
			CanContinue:                  provided.CheckedHashChanged(),
		}

		if r.Method == http.MethodPost && data.CanContinue {
			data.Form = readCheckYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Tasks.CheckYourLpa = task.StateCompleted
				provided.CheckedAt = now()
				if err := provided.UpdateCheckedHash(); err != nil {
					return err
				}

				if err := notifier.Notify(r.Context(), appData, provided, data.CertificateProviderContacted); err != nil {
					return err
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if !data.CertificateProviderContacted {
					return donor.PathLpaDetailsSaved.RedirectQuery(w, r, appData, provided, url.Values{"firstCheck": {"1"}})
				}

				return donor.PathLpaDetailsSaved.Redirect(w, r, appData, provided)
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
