package donor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetWhatHappensAfterNoFee(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, whatHappensAfterNoFeeData{App: testAppData}).
		Return(nil)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending},
			Evidence: []page.Evidence{
				{Key: "evidence-key", Sent: now},
				{Key: "another-evidence-key", Sent: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
		}).
		Return(nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObjectTagging", r.Context(), "evidence-key", []types.Tag{
			{Key: aws.String("replicate"), Value: aws.String("true")},
		}).
		Return(nil)

	err := WhatHappensAfterNoFee(template.Execute, donorStore, s3Client, nil, func() time.Time { return now })(testAppData, w, r, &page.Lpa{
		Evidence: []page.Evidence{
			{Key: "evidence-key"},
			{Key: "another-evidence-key", Sent: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
		},
	})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhatHappensAfterNoFeeWhenS3ClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObjectTagging", r.Context(), "evidence-key", []types.Tag{
			{Key: aws.String("replicate"), Value: aws.String("true")},
		}).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", fmt.Sprintf("error tagging evidence: %s", expectedError.Error())).
		Return(nil)

	err := WhatHappensAfterNoFee(nil, nil, s3Client, logger, nil)(testAppData, w, r, &page.Lpa{
		Evidence: []page.Evidence{
			{Key: "evidence-key"},
		},
	})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhatHappensAfterNoFeeWhenDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending},
			Evidence: []page.Evidence{
				{Key: "evidence-key", Sent: now},
				{Key: "another-evidence-key", Sent: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
		}).
		Return(expectedError)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObjectTagging", r.Context(), "evidence-key", []types.Tag{
			{Key: aws.String("replicate"), Value: aws.String("true")},
		}).
		Return(nil)

	err := WhatHappensAfterNoFee(nil, donorStore, s3Client, nil, func() time.Time { return now })(testAppData, w, r, &page.Lpa{
		Evidence: []page.Evidence{
			{Key: "evidence-key"},
			{Key: "another-evidence-key", Sent: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
		},
	})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhatHappensAfterNoFeeWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, whatHappensAfterNoFeeData{App: testAppData}).
		Return(expectedError)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending},
			Evidence: []page.Evidence{
				{Key: "evidence-key", Sent: now},
				{Key: "another-evidence-key", Sent: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
		}).
		Return(nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObjectTagging", r.Context(), "evidence-key", []types.Tag{
			{Key: aws.String("replicate"), Value: aws.String("true")},
		}).
		Return(nil)

	err := WhatHappensAfterNoFee(template.Execute, donorStore, s3Client, nil, func() time.Time { return now })(testAppData, w, r, &page.Lpa{
		Evidence: []page.Evidence{
			{Key: "evidence-key"},
			{Key: "another-evidence-key", Sent: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
		},
	})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
