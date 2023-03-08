package certificateprovider

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReferenceNumberData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterReferenceNumberForm
	Lpa    *page.Lpa
}

func EnterReferenceNumber(tmpl template.Template, lpaStore LpaStore, dataStore page.DataStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := enterReferenceNumberData{
			App:  appData,
			Form: &enterReferenceNumberForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReferenceNumberForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				referenceNumber := data.Form.ReferenceNumber

				var v page.ShareCodeData
				if err := dataStore.Get(r.Context(), "SHARECODE#"+referenceNumber, "#METADATA#"+referenceNumber, &v); err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"})
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

type enterReferenceNumberForm struct {
	ReferenceNumber string
}

func (f *enterReferenceNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("reference-number", "twelveCharactersReferenceNumber", strings.ReplaceAll(f.ReferenceNumber, " ", ""),
		validation.Empty(),
	)

	errors.String("reference-number", "referenceNumberMustBeTwelveCharacters", strings.ReplaceAll(f.ReferenceNumber, " ", ""),
		validation.StringLength(12),
	)

	return errors
}

func readEnterReferenceNumberForm(r *http.Request) *enterReferenceNumberForm {
	return &enterReferenceNumberForm{
		ReferenceNumber: page.PostFormString(r, "reference-number"),
	}
}
