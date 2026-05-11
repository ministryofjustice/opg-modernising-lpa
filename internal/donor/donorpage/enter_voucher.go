package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/names"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
)

type enterVoucherData struct {
	App  appcontext.Data
	Form *enterVoucherForm
}

func EnterVoucher(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &enterVoucherData{
			App:  appData,
			Form: newEnterVoucherForm(appData.Localizer),
		}

		data.Form.FirstNames.SetInput(provided.Voucher.FirstNames)
		data.Form.LastName.SetInput(provided.Voucher.LastName)
		data.Form.Email.SetInput(provided.Voucher.Email)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			if provided.Voucher.UID.IsZero() {
				provided.Voucher.UID = newUID()
			}

			if provided.Voucher.FirstNames != data.Form.FirstNames.Value || provided.Voucher.LastName != data.Form.LastName.Value {
				provided.Voucher.FirstNames = data.Form.FirstNames.Value
				provided.Voucher.LastName = data.Form.LastName.Value
				provided.Voucher.Allowed = len(provided.Voucher.Matches(provided)) == 0 && !names.Equal(provided.Voucher.LastName, provided.Donor.LastName)
			}

			provided.Voucher.Email = data.Form.Email.Value

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			if !provided.Voucher.Allowed {
				return donor.PathConfirmPersonAllowedToVouch.Redirect(w, r, appData, provided)
			}

			return donor.PathCheckYourDetails.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}

type enterVoucherForm struct {
	newforms.Form
	FirstNames *newforms.String
	LastName   *newforms.String
	Email      *newforms.String
}

func newEnterVoucherForm(l Localizer) *enterVoucherForm {
	return &enterVoucherForm{
		FirstNames: newforms.NewString("first-names", l.T("firstNames")).
			NotEmpty().
			MaxLength(53),
		LastName: newforms.NewString("last-name", l.T("lastName")).
			NotEmpty().
			MaxLength(61),
		Email: newforms.NewString("email", l.T("email")).
			NotEmpty().
			Email(),
	}
}

func (f *enterVoucherForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r,
		f.FirstNames,
		f.LastName,
		f.Email,
	)
}
