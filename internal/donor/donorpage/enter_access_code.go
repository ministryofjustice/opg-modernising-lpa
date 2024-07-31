package donorpage

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterAccessCodeData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterAccessCodeForm
}

func EnterAccessCode(logger Logger, tmpl template.Template, shareCodeStore ShareCodeStore, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := enterAccessCodeData{
			App:  appData,
			Form: &enterAccessCodeForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterAccessCodeForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeDonor, data.Form.AccessCode)
				if err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectAccessCode"})
						return tmpl(w, data)
					} else {
						return err
					}
				}

				if err := donorStore.Link(r.Context(), shareCode, appData.LoginSessionEmail); err != nil {
					return err
				}
				logger.InfoContext(r.Context(), "donor access added", slog.String("lpa_id", shareCode.LpaKey.ID()))

				return page.Paths.Dashboard.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}

type enterAccessCodeForm struct {
	AccessCode    string
	AccessCodeRaw string
}

func readEnterAccessCodeForm(r *http.Request) *enterAccessCodeForm {
	return &enterAccessCodeForm{
		AccessCode:    page.PostFormReferenceNumber(r, "reference-number"),
		AccessCodeRaw: page.PostFormString(r, "reference-number"),
	}
}

func (f *enterAccessCodeForm) Validate() validation.List {
	var errors validation.List

	errors.String("reference-number", "accessCode", f.AccessCode,
		validation.Empty(),
		validation.StringLength(12))

	return errors
}
