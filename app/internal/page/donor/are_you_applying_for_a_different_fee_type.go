package donor

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type areYouApplyingForADifferentFeeTypeData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	Options             actor.YesNoOptions
	Form                *areYouApplyingForADifferentFeeTypeForm
}

func AreYouApplyingForADifferentFeeType(logger Logger, tmpl template.Template, sessionStore sessions.Store, payClient PayClient, appPublicUrl string, randomString func(int) string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &areYouApplyingForADifferentFeeTypeData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			Options:             actor.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readAreYouApplyingForADifferentFeeTypeForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.DifferentFee.IsNo() {
					createPaymentBody := pay.CreatePaymentBody{
						Amount:      page.CostOfLpaPence,
						Reference:   randomString(12),
						Description: "Property and Finance LPA",
						ReturnUrl:   appPublicUrl + appData.BuildUrl(page.Paths.PaymentConfirmation.Format(lpa.ID)),
						Email:       lpa.Donor.Email,
						Language:    appData.Lang.String(),
					}

					resp, err := payClient.CreatePayment(createPaymentBody)
					if err != nil {
						logger.Print(fmt.Sprintf("Error creating payment: %s", err.Error()))
						return err
					}

					if err = sesh.SetPayment(sessionStore, r, w, &sesh.PaymentSession{PaymentID: resp.PaymentId}); err != nil {
						return err
					}

					nextUrl := resp.Links["next_url"].Href
					// If URL matches expected domain for GOV UK PAY redirect there. If not,
					// redirect to the confirmation code and carry on with flow.
					if strings.HasPrefix(nextUrl, pay.PaymentPublicServiceUrl) {
						http.Redirect(w, r, nextUrl, http.StatusFound)
					} else {
						appData.Redirect(w, r, lpa, page.Paths.PaymentConfirmation.Format(lpa.ID))
					}
				} else {
					appData.Redirect(w, r, lpa, page.Paths.WhichFeeTypeAreYouApplyingFor.Format(lpa.ID))
				}
			}
		}

		return tmpl(w, data)
	}
}

type areYouApplyingForADifferentFeeTypeForm struct {
	DifferentFee actor.YesNo
	Error        error
}

func readAreYouApplyingForADifferentFeeTypeForm(r *http.Request) *areYouApplyingForADifferentFeeTypeForm {
	differentFee, err := actor.ParseYesNo(page.PostFormString(r, "different-fee"))

	return &areYouApplyingForADifferentFeeTypeForm{
		DifferentFee: differentFee,
		Error:        err,
	}
}

func (f *areYouApplyingForADifferentFeeTypeForm) Validate() validation.List {
	var errors validation.List

	errors.Error("different-type", "whetherApplyingForDifferentFeeType", f.Error,
		validation.Selected())

	return errors
}
