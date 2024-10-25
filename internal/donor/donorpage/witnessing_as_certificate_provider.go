package donorpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingAsCertificateProviderData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *witnessingAsCertificateProviderForm
	Donor  *donordata.Provided
}

func WitnessingAsCertificateProvider(
	tmpl template.Template,
	donorStore DonorStore,
	shareCodeSender ShareCodeSender,
	lpaStoreClient LpaStoreClient,
	eventClient EventClient,
	now func() time.Time,
) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &witnessingAsCertificateProviderData{
			App:   appData,
			Donor: provided,
			Form:  &witnessingAsCertificateProviderForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readWitnessingAsCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if provided.WitnessCodeLimiter == nil {
				provided.WitnessCodeLimiter = donordata.NewLimiter(time.Minute, 5, 10)
			}

			if !provided.WitnessCodeLimiter.Allow(now()) {
				data.Errors.Add("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"})
			} else {
				code, found := provided.CertificateProviderCodes.Find(data.Form.Code, now())
				if !found {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"})
				} else if code.HasExpired(now()) {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeExpired"})
				}
			}

			if data.Errors.None() {
				provided.Tasks.SignTheLpa = task.StateCompleted

				provided.WitnessCodeLimiter = nil
				if provided.WitnessedByCertificateProviderAt.IsZero() {
					provided.WitnessedByCertificateProviderAt = now()
				}
			}

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			if data.Errors.None() {
				if provided.Tasks.PayForLpa.IsCompleted() {
					if err := shareCodeSender.SendCertificateProviderPrompt(r.Context(), appData, provided); err != nil {
						return err
					}

					if err := eventClient.SendCertificateProviderStarted(r.Context(), event.CertificateProviderStarted{
						UID: provided.LpaUID,
					}); err != nil {
						return err
					}

					if err := lpaStoreClient.SendLpa(r.Context(), provided); err != nil {
						return err
					}
				}

				return donor.PathYouHaveSubmittedYourLpa.Redirect(w, r, appData, provided)
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
