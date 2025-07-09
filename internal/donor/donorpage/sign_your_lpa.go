package donorpage

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type signYourLpaData struct {
	App                  appcontext.Data
	Errors               validation.List
	Donor                *donordata.Provided
	Form                 *signYourLpaForm
	WantToSignFormValue  string
	WantToApplyFormValue string
}

const (
	WantToSignLpa     = "want-to-sign"
	WantToApplyForLpa = "want-to-apply"
)

func SignYourLpa(tmpl template.Template, donorStore DonorStore, scheduledStore ScheduledStore, now func() time.Time) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if !provided.SignedAt.IsZero() {
			return donor.PathWitnessingYourSignature.Redirect(w, r, appData, provided)
		}

		data := &signYourLpaData{
			App:   appData,
			Donor: provided,
			Form: &signYourLpaForm{
				WantToApply: provided.WantToApplyForLpa,
				WantToSign:  provided.WantToSignLpa,
			},
			WantToSignFormValue:  WantToSignLpa,
			WantToApplyFormValue: WantToApplyForLpa,
		}

		if r.Method == http.MethodPost {
			if appData.Page == donor.PathSignTheLpaOnBehalf.Format(appData.LpaID) {
				data.Form = readSignYourLpaForm(r, provided.Donor.LpaLanguagePreference, provided.Donor.FullName())
			} else {
				data.Form = readSignYourLpaForm(r, provided.Donor.LpaLanguagePreference, "")
			}
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.WantToApplyForLpa = data.Form.WantToApply
				provided.WantToSignLpa = data.Form.WantToSign
				provided.SignedAt = now()

				if err := scheduledStore.Create(r.Context(), scheduled.Event{
					At:                provided.SignedAt.AddDate(0, 3, 1),
					Action:            scheduleddata.ActionRemindCertificateProviderToComplete,
					TargetLpaKey:      provided.PK,
					TargetLpaOwnerKey: provided.SK,
					LpaUID:            provided.LpaUID,
				}, scheduled.Event{
					At:                provided.SignedAt.AddDate(0, 21, 1),
					Action:            scheduleddata.ActionRemindCertificateProviderToComplete,
					TargetLpaKey:      provided.PK,
					TargetLpaOwnerKey: provided.SK,
					LpaUID:            provided.LpaUID,
				}); err != nil {
					return fmt.Errorf("could not schedule certificate provider prompt: %w", err)
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathWitnessingYourSignature.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type signYourLpaForm struct {
	WantToApply   bool
	WantToSign    bool
	WrongLanguage bool

	lpaLanguage   localize.Lang
	donorFullName string
}

func readSignYourLpaForm(r *http.Request, lang localize.Lang, donorFullName string) *signYourLpaForm {
	r.ParseForm()

	form := &signYourLpaForm{
		WrongLanguage: page.PostFormString(r, "wrong-language") == "1",
		lpaLanguage:   lang,
		donorFullName: donorFullName,
	}

	for _, checkBox := range r.PostForm["sign-lpa"] {
		if checkBox == WantToSignLpa {
			form.WantToSign = true
		}

		if checkBox == WantToApplyForLpa {
			form.WantToApply = true
		}
	}

	return form
}

func (f *signYourLpaForm) Validate() validation.List {
	var errors validation.List

	if !f.WantToApply || !f.WantToSign {
		errors.Add("sign-lpa", validation.SelectError{Label: "bothBoxesToSignAndApply"})
	} else if f.WrongLanguage {
		errors.Add("sign-lpa", youMustViewAndSignInLanguageError{LpaLanguage: f.lpaLanguage, DonorFullName: f.donorFullName})
	}

	return errors
}

type youMustViewAndSignInLanguageError struct {
	LpaLanguage   localize.Lang
	DonorFullName string
}

func (e youMustViewAndSignInLanguageError) Format(l validation.Localizer) string {
	if e.DonorFullName != "" {
		return l.Format("errorYouMustViewAndSignDonorsLpaInLanguage", map[string]any{
			"InLang":        l.T("in:" + e.LpaLanguage.String()),
			"DonorFullName": e.DonorFullName,
		})
	}

	return l.Format("youMustViewAndSignInLanguage", map[string]any{
		"InLang": l.T("in:" + e.LpaLanguage.String()),
	})
}
