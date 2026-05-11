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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
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
		attorneys := lpa.Attorneys
		if appData.IsReplacementAttorney() {
			attorneys = lpa.ReplacementAttorneys
		}

		thisAttorney, ok := attorneys.Get(appData.AttorneyUID)
		if !ok {
			return page.PathDashboard.Redirect(w, r, appData)
		}

		data := &signData{
			App:                         appData,
			Lpa:                         lpa,
			IsReplacement:               appData.IsReplacementAttorney(),
			LpaCanBeUsedWhenHasCapacity: lpa.WhenCanTheLpaBeUsed.IsHasCapacity(),
			Attorney:                    thisAttorney,
			Form: newSignForm(
				appData.Localizer,
				appData.IsTrustCorporation(),
				appData.IsReplacementAttorney(),
				lpa.Language,
				thisAttorney.FullName(),
				"",
			),
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
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

		trustCorporation := lpa.Attorneys.TrustCorporation
		if appData.IsReplacementAttorney() {
			trustCorporation = lpa.ReplacementAttorneys.TrustCorporation
		}

		data := &signData{
			App:                         appData,
			Lpa:                         lpa,
			IsReplacement:               appData.IsReplacementAttorney(),
			IsSecondSignatory:           signatoryIndex == 1,
			LpaCanBeUsedWhenHasCapacity: lpa.WhenCanTheLpaBeUsed.IsHasCapacity(),
			TrustCorporation:            trustCorporation,
			Form: newSignForm(
				appData.Localizer,
				appData.IsTrustCorporation(),
				appData.IsReplacementAttorney(),
				lpa.Language,
				"",
				trustCorporation.Name,
			),
		}

		data.Form.FirstNames.Input = signatory.FirstNames
		data.Form.LastName.Input = signatory.LastName
		data.Form.ProfessionalTitle.Input = signatory.ProfessionalTitle

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				if signatoryIndex == 1 {
					attorneyProvidedDetails.Tasks.SignTheLpaSecond = task.StateCompleted
				} else {
					attorneyProvidedDetails.Tasks.SignTheLpa = task.StateCompleted
				}

				attorneyProvidedDetails.AuthorisedSignatories[signatoryIndex] = attorneydata.TrustCorporationSignatory{
					FirstNames:        data.Form.FirstNames.Value,
					LastName:          data.Form.LastName.Value,
					ProfessionalTitle: data.Form.ProfessionalTitle.Value,
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
	FirstNames        *newforms.String
	LastName          *newforms.String
	ProfessionalTitle *newforms.String
	Confirm           *newforms.Bool
	WrongLanguage     *newforms.Bool
	Errors            []newforms.Field

	isTrustCorporation bool
	lpaLanguage        localize.Lang
}

func newSignForm(l Localizer, isTrustCorporation, isReplacement bool, lang localize.Lang, attorneyFullName, trustCorporationName string) *signForm {
	confirmError := l.T("youMustSelectTheBoxToSignAttorney")

	// TODO: this copies the previous logic, but do we not want to show
	// replacement content to a trust corporation??
	if isReplacement && !isTrustCorporation {
		confirmError = l.T("youMustSelectTheBoxToSignReplacementAttorney")
	}

	confirmLabel := l.Format("iAttorneyConfirmTheseStatements", map[string]any{
		"AttorneyFullName": attorneyFullName,
	})
	if isTrustCorporation {
		confirmLabel = l.Format("iTrustCorporationConfirmTheseStatements", map[string]any{
			"TrustCorporationName": trustCorporationName,
		})
	}

	if isTrustCorporation {
		return &signForm{
			FirstNames: newforms.NewString("first-names", l.T("firstNames")).
				NotEmpty(),
			LastName: newforms.NewString("last-name", l.T("lastName")).
				NotEmpty(),
			ProfessionalTitle: newforms.NewString("professional-title", l.T("professionalTitle")).
				NotEmpty(),
			Confirm: newforms.NewBool("confirm", confirmLabel).
				True(confirmError),
			WrongLanguage:      newforms.NewBool("wrong-language", ""),
			isTrustCorporation: true,
			lpaLanguage:        lang,
		}
	} else {
		return &signForm{
			Confirm: newforms.NewBool("confirm", confirmLabel).
				True(confirmError),
			WrongLanguage:      newforms.NewBool("wrong-language", ""),
			isTrustCorporation: false,
			lpaLanguage:        lang,
		}
	}
}

func (f *signForm) Parse(r *http.Request) bool {
	if f.isTrustCorporation {
		f.Errors = newforms.ParsePostForm(r,
			f.FirstNames,
			f.LastName,
			f.ProfessionalTitle,
			f.Confirm,
			f.WrongLanguage,
		)
	} else {
		f.Errors = newforms.ParsePostForm(r,
			f.Confirm,
			f.WrongLanguage,
		)
	}

	if f.Confirm.Value && f.WrongLanguage.Value {
		f.Confirm.Error = toSignLpaYouMustViewInLanguageError{LpaLanguage: f.lpaLanguage}
		f.Errors = append(f.Errors, f.Confirm.Field)
	}

	return len(f.Errors) == 0
}

type toSignLpaYouMustViewInLanguageError struct {
	LpaLanguage localize.Lang
}

func (e toSignLpaYouMustViewInLanguageError) Format(l newforms.Localizer) string {
	return l.Format("toSignLpaYouMustViewInLanguage", map[string]any{
		"InLang": l.T("in:" + e.LpaLanguage.String()),
	})
}
