package attorneypage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type signData struct {
	App                         appcontext.Data
	Errors                      validation.List
	Lpa                         *lpadata.Lpa
	Attorney                    lpadata.Attorney
	TrustCorporation            lpadata.TrustCorporation
	IsReplacement               bool
	IsSecondSignatory           bool
	LpaCanBeUsedWhenHasCapacity bool
	Form                        *signForm
}

func Sign(tmpl template.Template, attorneyStore AttorneyStore, lpaStoreClient LpaStoreClient, now func() time.Time) Handler {
	signAttorney := func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *attorneydata.Provided, lpa *lpadata.Lpa) error {
		data := &signData{
			App:                         appData,
			Lpa:                         lpa,
			IsReplacement:               appData.IsReplacementAttorney(),
			LpaCanBeUsedWhenHasCapacity: lpa.WhenCanTheLpaBeUsed.IsHasCapacity(),
			Form:                        &signForm{},
		}

		attorneys := lpa.Attorneys
		if appData.IsReplacementAttorney() {
			attorneys = lpa.ReplacementAttorneys
		}

		{
			attorney, ok := attorneys.Get(appData.AttorneyUID)
			if !ok {
				return page.PathDashboard.Redirect(w, r, appData)
			}

			data.Attorney = attorney
		}

		if r.Method == http.MethodPost {
			data.Form = readSignForm(r, lpa.Language)
			data.Errors = data.Form.Validate(appData.IsTrustCorporation(), appData.IsReplacementAttorney())

			if data.Errors.None() {
				provided.Tasks.SignTheLpa = task.StateCompleted
				provided.SignedAt = now()

				if !provided.PhoneSet {
					if _, mobile, _ := lpa.Attorney(provided.UID); mobile != "" {
						provided.Phone = mobile
					}
				}

				if data.Attorney.SignedAt == nil || data.Attorney.SignedAt.IsZero() {
					if err := lpaStoreClient.SendAttorney(r.Context(), lpa, provided); err != nil {
						return err
					}
				} else {
					provided.SignedAt = *data.Attorney.SignedAt
				}

				if err := attorneyStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return attorney.PathWhatHappensNext.Redirect(w, r, appData, provided.LpaID)
			}
		}

		return tmpl(w, data)
	}

	signTrustCorporation := func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided, lpa *lpadata.Lpa) error {
		signatoryIndex := 0
		if r.URL.Query().Has("second") {
			signatoryIndex = 1
		}

		signatory := attorneyProvidedDetails.AuthorisedSignatories[signatoryIndex]

		data := &signData{
			App:                         appData,
			Lpa:                         lpa,
			IsReplacement:               appData.IsReplacementAttorney(),
			IsSecondSignatory:           signatoryIndex == 1,
			LpaCanBeUsedWhenHasCapacity: lpa.WhenCanTheLpaBeUsed.IsHasCapacity(),
			Form: &signForm{
				FirstNames:        signatory.FirstNames,
				LastName:          signatory.LastName,
				ProfessionalTitle: signatory.ProfessionalTitle,
			},
		}

		if appData.IsReplacementAttorney() {
			data.TrustCorporation = lpa.ReplacementAttorneys.TrustCorporation
		} else {
			data.TrustCorporation = lpa.Attorneys.TrustCorporation
		}

		if r.Method == http.MethodPost {
			data.Form = readSignForm(r, lpa.Language)
			data.Errors = data.Form.Validate(appData.IsTrustCorporation(), appData.IsReplacementAttorney())

			if data.Errors.None() {
				if signatoryIndex == 1 {
					attorneyProvidedDetails.Tasks.SignTheLpaSecond = task.StateCompleted
				} else {
					attorneyProvidedDetails.Tasks.SignTheLpa = task.StateCompleted
				}

				attorneyProvidedDetails.AuthorisedSignatories[signatoryIndex] = attorneydata.TrustCorporationSignatory{
					FirstNames:        data.Form.FirstNames,
					LastName:          data.Form.LastName,
					ProfessionalTitle: data.Form.ProfessionalTitle,
					SignedAt:          now(),
				}

				if len(data.TrustCorporation.Signatories) == 0 {
					if signatoryIndex == 1 {
						if err := lpaStoreClient.SendAttorney(r.Context(), lpa, attorneyProvidedDetails); err != nil {
							return err
						}
					}
				} else {
					attorneyProvidedDetails.AuthorisedSignatories[signatoryIndex].SignedAt = data.TrustCorporation.Signatories[signatoryIndex].SignedAt
				}

				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				if signatoryIndex == 0 {
					return attorney.PathWouldLikeSecondSignatory.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
				} else {
					return attorney.PathWhatHappensNext.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
				}
			}
		}

		return tmpl(w, data)
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided, lpa *lpadata.Lpa) error {
		if attorneyProvidedDetails.Signed() {
			return attorney.PathWhatHappensNext.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
		}

		if !lpa.SignedForDonor() || lpa.CertificateProvider.SignedAt.IsZero() {
			return attorney.PathTaskList.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
		}

		if appData.IsTrustCorporation() {
			return signTrustCorporation(appData, w, r, attorneyProvidedDetails, lpa)
		} else {
			return signAttorney(appData, w, r, attorneyProvidedDetails, lpa)
		}
	}
}

type signForm struct {
	FirstNames        string
	LastName          string
	ProfessionalTitle string
	Confirm           bool
	WrongLanguage     bool

	lpaLanguage localize.Lang
}

func readSignForm(r *http.Request, lang localize.Lang) *signForm {
	return &signForm{
		FirstNames:        page.PostFormString(r, "first-names"),
		LastName:          page.PostFormString(r, "last-name"),
		ProfessionalTitle: page.PostFormString(r, "professional-title"),
		Confirm:           page.PostFormString(r, "confirm") == "1",
		WrongLanguage:     page.PostFormString(r, "wrong-language") == "1",
		lpaLanguage:       lang,
	}
}

func (f *signForm) Validate(isTrustCorporation, isReplacement bool) validation.List {
	var errors validation.List

	if isTrustCorporation {
		errors.String("first-names", "firstNames", f.FirstNames,
			validation.Empty())
		errors.String("last-name", "lastName", f.LastName,
			validation.Empty())
		errors.String("professional-title", "professionalTitle", f.ProfessionalTitle,
			validation.Empty())
		errors.Bool("confirm", "youMustSelectTheBoxToSignAttorney", f.Confirm,
			validation.Selected().CustomError())
	} else if isReplacement {
		errors.Bool("confirm", "youMustSelectTheBoxToSignReplacementAttorney", f.Confirm,
			validation.Selected().CustomError())
	} else {
		errors.Bool("confirm", "youMustSelectTheBoxToSignAttorney", f.Confirm,
			validation.Selected().CustomError())
	}

	if f.Confirm && f.WrongLanguage {
		errors.Add("confirm", toSignLpaYouMustViewInLanguageError{LpaLanguage: f.lpaLanguage})
	}

	return errors
}

type toSignLpaYouMustViewInLanguageError struct {
	LpaLanguage localize.Lang
}

func (e toSignLpaYouMustViewInLanguageError) Format(l validation.Localizer) string {
	return l.Format("toSignLpaYouMustViewInLanguage", map[string]any{
		"InLang": l.T("in:" + e.LpaLanguage.String()),
	})
}
