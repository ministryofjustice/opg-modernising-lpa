package voucherpage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
)

type enterReferenceNumberData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *enterReferenceNumberForm
}

func EnterReferenceNumber(tmpl template.Template, shareCodeStore ShareCodeStore, sessionStore SessionStore, voucherStore VoucherStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := enterReferenceNumberData{
			App:  appData,
			Form: &enterReferenceNumberForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReferenceNumberForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				referenceNumber := sharecodedata.HashedFromString(data.Form.ReferenceNumber)

				shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeVoucher, referenceNumber)
				if err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"})
						return tmpl(w, data)
					} else {
						return fmt.Errorf("error getting sharecode: %w", err)
					}
				}

				session, err := sessionStore.Login(r)
				if err != nil {
					return fmt.Errorf("error getting login session: %w", err)
				}

				session.HasLPAs = true
				if err := sessionStore.SetLogin(r, w, session); err != nil {
					return fmt.Errorf("error saving login session: %w", err)
				}

				ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{
					SessionID: session.SessionID(),
					LpaID:     shareCode.LpaKey.ID(),
				})

				if _, err := voucherStore.Create(ctx, shareCode, session.Email); err != nil {
					return fmt.Errorf("error creating voucher: %w", err)
				}

				appData.LpaID = shareCode.LpaKey.ID()
				return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type enterReferenceNumberForm struct {
	ReferenceNumber    string
	ReferenceNumberRaw string
}

func readEnterReferenceNumberForm(r *http.Request) *enterReferenceNumberForm {
	return &enterReferenceNumberForm{
		ReferenceNumber:    page.PostFormReferenceNumber(r, "reference-number"),
		ReferenceNumberRaw: page.PostFormString(r, "reference-number"),
	}
}

func (f *enterReferenceNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("reference-number", "twelveCharactersReferenceNumber", f.ReferenceNumber,
		validation.Empty())

	errors.String("reference-number", "theReferenceNumberYouEnter", f.ReferenceNumber,
		validation.StringLength(12))

	return errors
}
