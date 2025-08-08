package attorneypage

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type companyNumberData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *companyNumberForm
}

type companyNumberForm struct {
	CompanyNumber string
}

func CompanyNumber(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *attorneydata.Provided, lpa *lpadata.Lpa) error {
		if !appData.IsTrustCorporation() {
			return attorney.PathTaskList.Redirect(w, r, appData, provided.LpaID)
		}

		data := &companyNumberData{
			App:  appData,
			Form: &companyNumberForm{CompanyNumber: provided.CompanyNumber},
		}

		if r.Method == http.MethodPost {
			data.Form = readCompanyNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Tasks.ConfirmYourDetails = task.StateInProgress
				provided.CompanyNumber = data.Form.CompanyNumber

				if err := attorneyStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return attorney.PathPhoneNumber.Redirect(w, r, appData, provided.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

func readCompanyNumberForm(r *http.Request) *companyNumberForm {
	return &companyNumberForm{
		CompanyNumber: page.PostFormString(r, "company-number"),
	}
}

func (f *companyNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("company-number", "yourCompanyNumber", strings.TrimSpace(f.CompanyNumber),
		validation.Empty())

	return errors
}
