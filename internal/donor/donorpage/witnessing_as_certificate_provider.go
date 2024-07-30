package donorpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingAsCertificateProviderData struct {
	App    page.AppData
	Errors validation.List
	Form   *witnessingAsCertificateProviderForm
	Donor  *actor.DonorProvidedDetails
}

func WitnessingAsCertificateProvider(
	tmpl template.Template,
	donorStore DonorStore,
	shareCodeSender ShareCodeSender,
	lpaStoreClient LpaStoreClient,
	now func() time.Time,
) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &witnessingAsCertificateProviderData{
			App:   appData,
			Donor: donor,
			Form:  &witnessingAsCertificateProviderForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readWitnessingAsCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if donor.WitnessCodeLimiter == nil {
				donor.WitnessCodeLimiter = actor.NewLimiter(time.Minute, 5, 10)
			}

			if !donor.WitnessCodeLimiter.Allow(now()) {
				data.Errors.Add("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"})
			} else {
				code, found := donor.CertificateProviderCodes.Find(data.Form.Code)
				if !found {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"})
				} else if code.HasExpired() {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeExpired"})
				}
			}

			if data.Errors.None() {
				donor.Tasks.ConfirmYourIdentityAndSign = actor.IdentityTaskCompleted
				if donor.RegisteringWithCourtOfProtection {
					donor.Tasks.ConfirmYourIdentityAndSign = actor.IdentityTaskPending
				}

				donor.WitnessCodeLimiter = nil
				if donor.WitnessedByCertificateProviderAt.IsZero() {
					donor.WitnessedByCertificateProviderAt = now()
				}
			}

			if err := donorStore.Put(r.Context(), donor); err != nil {
				return err
			}

			if data.Errors.None() {
				if donor.Tasks.PayForLpa.IsCompleted() {
					if err := shareCodeSender.SendCertificateProviderPrompt(r.Context(), appData, donor); err != nil {
						return err
					}

					if err := lpaStoreClient.SendLpa(r.Context(), donor); err != nil {
						return err
					}
				}

				return page.Paths.YouHaveSubmittedYourLpa.Redirect(w, r, appData, donor)
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
