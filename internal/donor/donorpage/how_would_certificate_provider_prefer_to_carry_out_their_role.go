package donorpage

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldCertificateProviderPreferToCarryOutTheirRoleData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider donordata.CertificateProvider
	Form                *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
	Options             lpadata.ChannelOptions
}

func HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpl template.Template, donorStore DonorStore, certificateProviderStore CertificateProviderStore, reuseStore ReuseStore, accessCodeStore AccessCodeStore, accessCodeSender AccessCodeSender, now func() time.Time) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if _, err := certificateProviderStore.OneByUID(r.Context(), provided.LpaUID); err == nil {
			return donor.PathCertificateProviderSummary.Redirect(w, r, appData, provided)
		} else if !errors.Is(err, dynamo.NotFoundError{}) {
			return fmt.Errorf("get certificate provider: %w", err)
		}

		data := &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                 appData,
			CertificateProvider: provided.CertificateProvider,
			Form:                newHowWouldCertificateProviderPreferToCarryOutTheirRoleForm(appData.Localizer),
			Options:             lpadata.ChannelValues,
		}

		data.Form.CarryOutBy.SetInput(provided.CertificateProvider.CarryOutBy)
		data.Form.Email.SetInput(provided.CertificateProvider.Email)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			emailChanged := provided.CertificateProvider.Email != data.Form.Email.Value

			provided.CertificateProvider.CarryOutBy = data.Form.CarryOutBy.Value
			provided.CertificateProvider.Email = data.Form.Email.Value

			if emailChanged && !provided.CertificateProviderInvitedAt.IsZero() && !provided.Tasks.SignTheLpa.IsCompleted() {
				if err := accessCodeStore.DeleteByActor(r.Context(), provided.CertificateProvider.UID); err != nil {
					return fmt.Errorf("deleting certificate provider access code: %w", err)
				}

				if err := accessCodeSender.SendCertificateProviderInvite(r.Context(), appData, provided); err != nil {
					return fmt.Errorf("sending certificate provider access code: %w", err)
				}

				provided.CertificateProviderInvitedAt = now()
			}

			if err := reuseStore.PutCertificateProvider(r.Context(), provided.CertificateProvider); err != nil {
				return fmt.Errorf("put certificate provider reuse data: %w", err)
			}

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return donor.PathCertificateProviderAddress.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}

type howWouldCertificateProviderPreferToCarryOutTheirRoleForm struct {
	newforms.Form
	CarryOutBy *newforms.Enum[lpadata.Channel, lpadata.ChannelOptions, *lpadata.Channel]
	Email      *newforms.String
}

func newHowWouldCertificateProviderPreferToCarryOutTheirRoleForm(l Localizer) *howWouldCertificateProviderPreferToCarryOutTheirRoleForm {
	return &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
		CarryOutBy: newforms.NewEnum[lpadata.Channel]("carry-out-by", l.T("howYourCertificateProviderWouldPreferToCarryOutTheirRole"), lpadata.ChannelValues).
			Selected(),
		Email: newforms.NewString("email", l.T("certificateProvidersEmail")).
			NotEmpty().
			Email(),
	}
}

func (f *howWouldCertificateProviderPreferToCarryOutTheirRoleForm) Parse(r *http.Request) bool {
	ok := f.ParsePostForm(r, f.CarryOutBy)

	if f.CarryOutBy.Value.IsOnline() {
		ok = f.ParsePostForm(r, f.Email) && ok
	}

	return ok
}
