package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/forms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type companyNumberData struct {
	App  appcontext.Data
	Form *companyNumberForm
}

func CompanyNumber(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *attorneydata.Provided, lpa *lpadata.Lpa) error {
		if !appData.IsTrustCorporation() {
			return attorney.PathTaskList.Redirect(w, r, appData, provided.LpaID)
		}

		data := &companyNumberData{
			App:  appData,
			Form: newCompanyNumberForm(appData.Localizer),
		}

		data.Form.CompanyNumber.Set(provided.CompanyNumber)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			provided.Tasks.ConfirmYourDetails = task.StateInProgress
			provided.CompanyNumber = data.Form.CompanyNumber.Value

			if err := attorneyStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return attorney.PathPhoneNumber.Redirect(w, r, appData, provided.LpaID)
		}

		return tmpl(w, data)
	}
}

type companyNumberForm struct {
	forms.Form
	CompanyNumber *forms.String
}

func newCompanyNumberForm(l Localizer) *companyNumberForm {
	return &companyNumberForm{
		CompanyNumber: forms.NewString("company-number", l.T("enterYourCompanyNumber")).
			NotEmpty().WithErrorLabel(l.T("yourCompanyNumber")),
	}
}

func (f *companyNumberForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.CompanyNumber)
}
