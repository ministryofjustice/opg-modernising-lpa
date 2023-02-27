package certificateprovider

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type enterReferenceCodeData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterReferenceCodeForm
	Lpa    *page.Lpa
}

func EnterReferenceCode(tmpl template.Template, lpaStore LpaStore, dataStore page.DataStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := enterReferenceCodeData{
			App:  appData,
			Form: &enterReferenceCodeForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReferenceCodeForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				shareCode := data.Form.ReferenceCode

				var v page.ShareCodeData
				if err := dataStore.Get(r.Context(), "SHARECODE#"+shareCode, "#METADATA#"+shareCode, &v); err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add("reference-code", validation.CustomError{Label: "incorrectReferenceCode"})
						return tmpl(w, data)
					} else {
						return err
					}
				}

				lpa, err := lpaStore.Get(page.ContextWithSessionData(r.Context(), &page.SessionData{
					SessionID: v.SessionID,
					LpaID:     v.LpaID,
				}))
				if err != nil {
					return err
				}

				query := url.Values{
					"lpaId":     {v.LpaID},
					"sessionId": {v.SessionID},
				}
				if v.Identity {
					query.Add("identity", "1")
				}

				appData.Redirect(w, r, lpa, page.Paths.CertificateProviderLogin+"?"+query.Encode())
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type enterReferenceCodeForm struct {
	ReferenceCode string
}

func (f *enterReferenceCodeForm) Validate() validation.List {
	var errors validation.List

	errors.String("reference-code", "twelveCharactersReferenceCode", strings.ReplaceAll(f.ReferenceCode, " ", ""),
		validation.Empty(),
	)

	errors.String("reference-code", "referenceCodeMustBeTwelveCharacters", strings.ReplaceAll(f.ReferenceCode, " ", ""),
		validation.StringLength(12),
	)

	return errors
}

func readEnterReferenceCodeForm(r *http.Request) *enterReferenceCodeForm {
	return &enterReferenceCodeForm{
		ReferenceCode: page.PostFormString(r, "reference-code"),
	}
}
