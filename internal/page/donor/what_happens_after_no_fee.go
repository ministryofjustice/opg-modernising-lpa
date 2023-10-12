package donor

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whatHappensAfterNoFeeData struct {
	App    page.AppData
	Errors validation.List
}

func WhatHappensAfterNoFee(tmpl template.Template, donorStore DonorStore, evidenceS3Client S3Client, logger Logger, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		for i, evidence := range lpa.Evidence {
			if evidence.Sent.IsZero() {
				err := evidenceS3Client.PutObjectTagging(r.Context(), evidence.Key, []types.Tag{
					{Key: aws.String("replicate"), Value: aws.String("true")},
				})

				if err != nil {
					logger.Print(fmt.Sprintf("error tagging evidence: %s", err.Error()))
					return err
				}

				lpa.Evidence[i].Sent = now()
			}
		}

		lpa.Tasks.PayForLpa = actor.PaymentTaskPending

		if err := donorStore.Put(r.Context(), lpa); err != nil {
			return err
		}

		return tmpl(w, whatHappensAfterNoFeeData{App: appData})
	}
}
