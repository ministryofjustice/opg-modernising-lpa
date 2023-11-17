package donor

import (
	"errors"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingAsCertificateProviderData struct {
	App    page.AppData
	Errors validation.List
	Form   *witnessingAsCertificateProviderForm
	Lpa    *page.Lpa
}

func WitnessingAsCertificateProvider(tmpl template.Template, donorStore DonorStore, shareCodeSender ShareCodeSender, now func() time.Time, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &witnessingAsCertificateProviderData{
			App:  appData,
			Lpa:  lpa,
			Form: &witnessingAsCertificateProviderForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readWitnessingAsCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if lpa.WitnessCodeLimiter == nil {
				lpa.WitnessCodeLimiter = actor.NewLimiter(time.Minute, 5, 10)
			}

			if !lpa.WitnessCodeLimiter.Allow(now()) {
				data.Errors.Add("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"})
			} else {
				code, found := lpa.CertificateProviderCodes.Find(data.Form.Code)
				if !found {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"})
				} else if code.HasExpired() {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeExpired"})
				}
			}

			if data.Errors.None() {
				lpa.WitnessCodeLimiter = nil
				lpa.WitnessedByCertificateProviderAt = now()
				lpa.SignedAt = now()
			}

			if err := donorStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			if data.Errors.None() {
				ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
					SessionID: appData.SessionID,
					LpaID:     appData.LpaID,
				})

				certificateProvider, err := certificateProviderStore.GetAny(ctx)
				if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
					return err
				}

				if err == nil && certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
					if err := shareCodeSender.SendCertificateProvider(r.Context(), notify.CertificateProviderReturnEmail, appData, false, lpa); err != nil {
						return err
					}
				}

				if err := shareCodeSender.SendAttorneys(r.Context(), appData, lpa); err != nil {
					return err
				}

				return page.Paths.YouHaveSubmittedYourLpa.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}

type witnessingAsCertificateProviderForm struct {
	Code string
}

func readWitnessingAsCertificateProviderForm(r *http.Request) *witnessingAsCertificateProviderForm {
	return &witnessingAsCertificateProviderForm{
		Code: page.PostFormString(r, "witness-code"),
	}
}

func (w *witnessingAsCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	errors.String("witness-code", "theCodeWeSentCertificateProvider", w.Code,
		validation.Empty(),
		validation.StringLength(4))

	return errors
}
