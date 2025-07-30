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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: provided.CertificateProvider.CarryOutBy,
				Email:      provided.CertificateProvider.Email,
			},
			Options: lpadata.ChannelValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				emailChanged := provided.CertificateProvider.Email != data.Form.Email

				provided.CertificateProvider.CarryOutBy = data.Form.CarryOutBy
				provided.CertificateProvider.Email = data.Form.Email

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
		}

		return tmpl(w, data)
	}
}

type howWouldCertificateProviderPreferToCarryOutTheirRoleForm struct {
	CarryOutBy lpadata.Channel
	Email      string
}

func readHowWouldCertificateProviderPreferToCarryOutTheirRole(r *http.Request) *howWouldCertificateProviderPreferToCarryOutTheirRoleForm {
	channel, _ := lpadata.ParseChannel(page.PostFormString(r, "carry-out-by"))

	email := page.PostFormString(r, "email")
	if channel.IsPaper() {
		email = ""
	}

	return &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
		CarryOutBy: channel,
		Email:      email,
	}
}

func (f *howWouldCertificateProviderPreferToCarryOutTheirRoleForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("carry-out-by", "howYourCertificateProviderWouldPreferToCarryOutTheirRole", f.CarryOutBy,
		validation.Selected())

	if f.CarryOutBy.IsOnline() {
		errors.String("email", "certificateProvidersEmail", f.Email,
			validation.Empty(),
			validation.Email())
	}

	return errors
}
