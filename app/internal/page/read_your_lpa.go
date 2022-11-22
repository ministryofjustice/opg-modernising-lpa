package page

import (
	"crypto/subtle"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type readYourLpaData struct {
	App                            AppData
	Errors                         map[string]string
	Lpa                            *Lpa
	EnteredSignature               bool
	Form                           *readYourLpaForm
	HowAttorneysMakeDecisionsPath  string
	ChooseAttorneysPath            string
	WhenCanLpaBeUsedPath           string
	RestrictionsPath               string
	CertificatesProviderPath       string
	AttorneyDetailsPath            string
	AttorneyAddressPath            string
	RemoveAttorneyPath             string
	ReplacementAttorneyDetailsPath string
	ReplacementAttorneyAddressPath string
	RemoveReplacementAttorneyPath  string
}

func ReadYourLpa(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &readYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &readYourLpaForm{
				Checked:   lpa.CheckedAgain,
				Confirm:   lpa.ConfirmFreeWill,
				Signature: lpa.EnteredSignatureCode,
			},
			HowAttorneysMakeDecisionsPath:  howShouldAttorneysMakeDecisionsPath,
			ChooseAttorneysPath:            chooseAttorneysPath,
			WhenCanLpaBeUsedPath:           whenCanTheLpaBeUsedPath,
			RestrictionsPath:               restrictionsPath,
			CertificatesProviderPath:       certificateProviderDetailsPath,
			AttorneyDetailsPath:            chooseAttorneysPath,
			AttorneyAddressPath:            chooseAttorneysAddressPath,
			RemoveAttorneyPath:             removeAttorneyPath,
			ReplacementAttorneyDetailsPath: chooseReplacementAttorneysPath,
			ReplacementAttorneyAddressPath: chooseReplacementAttorneysAddressPath,
			RemoveReplacementAttorneyPath:  removeReplacementAttorneyPath,
		}

		if r.Method == http.MethodPost {
			data.Form = readReadYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				if cmp := subtle.ConstantTimeCompare([]byte(lpa.SignatureCode), []byte(data.Form.Signature)); cmp != 1 {
					data.Errors["signature"] = "enterCorrectSignatureCode"
					return tmpl(w, data)
				}

				lpa.CheckedAgain = data.Form.Checked
				lpa.ConfirmFreeWill = data.Form.Confirm
				lpa.EnteredSignatureCode = data.Form.Signature
				lpa.Tasks.ConfirmYourIdentityAndSign = TaskCompleted

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				appData.Lang.Redirect(w, r, signingConfirmationPath, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type readYourLpaForm struct {
	Checked   bool
	Confirm   bool
	Signature string
}

func readReadYourLpaForm(r *http.Request) *readYourLpaForm {
	r.ParseForm()

	return &readYourLpaForm{
		Checked:   postFormString(r, "checked") == "1",
		Confirm:   postFormString(r, "confirm") == "1",
		Signature: postFormString(r, "signature"),
	}
}

func (f *readYourLpaForm) Validate() map[string]string {
	errors := map[string]string{}

	if !f.Checked {
		errors["checked"] = "selectReadAndCheckedLpa"
	}

	if !f.Confirm {
		errors["confirm"] = "selectConfirmMadeThisLpaOfOwnFreeWill"
	}

	if f.Signature == "" {
		errors["signature"] = "enterSignatureCode"
	}

	return errors
}
