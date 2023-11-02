package donor

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type paymentConfirmationData struct {
	App              page.AppData
	Errors           validation.List
	PaymentReference string
	FeeType          pay.FeeType
	PreviousFee      pay.PreviousFee
	EvidenceDelivery pay.EvidenceDelivery
}

func PaymentConfirmation(logger Logger, tmpl template.Template, payClient PayClient, donorStore DonorStore, sessionStore sessions.Store, evidenceS3Client S3Client, now func() time.Time, documentStore DocumentStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		paymentSession, err := sesh.Payment(sessionStore, r)
		if err != nil {
			return err
		}

		paymentId := paymentSession.PaymentID

		payment, err := payClient.GetPayment(paymentId)
		if err != nil {
			logger.Print(fmt.Sprintf("unable to retrieve payment info: %s", err.Error()))
			return err
		}

		lpa.PaymentDetails = append(lpa.PaymentDetails, page.Payment{
			PaymentReference: payment.Reference,
			PaymentId:        payment.PaymentId,
			Amount:           payment.Amount,
		})

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: payment.Reference,
			FeeType:          lpa.FeeType,
			PreviousFee:      lpa.PreviousFee,
			EvidenceDelivery: lpa.EvidenceDelivery,
		}

		if err := sesh.ClearPayment(sessionStore, r, w); err != nil {
			logger.Print(fmt.Sprintf("unable to expire cookie in session: %s", err.Error()))
		}

		if lpa.FeeType.IsFullFee() {
			lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
		} else {
			lpa.Tasks.PayForLpa = actor.PaymentTaskPending

			documents, err := documentStore.GetAll(r.Context())
			if err != nil {
				return err
			}

			for _, document := range documents {
				if document.Sent.IsZero() {
					err := evidenceS3Client.PutObjectTagging(r.Context(), document.Key, []types.Tag{
						{Key: aws.String("replicate"), Value: aws.String("true")},
					})

					if err != nil {
						logger.Print(fmt.Sprintf("error tagging evidence: %s", err.Error()))
						return err
					}

					document.Sent = now()
					if err := documentStore.Put(r.Context(), document); err != nil {
						return err
					}
				}
			}
		}

		if err := donorStore.Put(r.Context(), lpa); err != nil {
			logger.Print(fmt.Sprintf("unable to update lpa in donorStore: %s", err.Error()))
			return err
		}

		return tmpl(w, data)
	}
}
