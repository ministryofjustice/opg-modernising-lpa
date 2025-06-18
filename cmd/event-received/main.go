// Event received is an AWS Lambda function to handle incoming events.
package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/document"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/s3"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	virusFound = "THREATS_FOUND"
)

var (
	tableName                   = os.Getenv("LPAS_TABLE")
	appPublicURL                = os.Getenv("APP_PUBLIC_URL")
	donorStartURL               = cmp.Or(os.Getenv("DONOR_START_URL"), appPublicURL+page.PathStart.Format())
	certificateProviderStartURL = cmp.Or(os.Getenv("CERTIFICATE_PROVIDER_START_URL"), appPublicURL+page.PathCertificateProviderStart.Format())
	attorneyStartURL            = cmp.Or(os.Getenv("ATTORNEY_START_URL"), appPublicURL+page.PathAttorneyStart.Format())
	awsBaseURL                  = os.Getenv("AWS_BASE_URL")
	notifyBaseURL               = os.Getenv("GOVUK_NOTIFY_BASE_URL")
	evidenceBucketName          = os.Getenv("UPLOADS_S3_BUCKET_NAME")
	uidBaseURL                  = os.Getenv("UID_BASE_URL")
	lpaStoreBaseURL             = os.Getenv("LPA_STORE_BASE_URL")
	lpaStoreSecretARN           = os.Getenv("LPA_STORE_SECRET_ARN")
	eventBusName                = cmp.Or(os.Getenv("EVENT_BUS_NAME"), "default")
	searchEndpoint              = os.Getenv("SEARCH_ENDPOINT")
	searchIndexName             = cmp.Or(os.Getenv("SEARCH_INDEX_NAME"), "lpas")
	searchIndexingEnabled       = os.Getenv("SEARCH_INDEXING_DISABLED") != "1"
	xrayEnabled                 = os.Getenv("XRAY_ENABLED") == "1"
	kmsKeyAlias                 = cmp.Or(os.Getenv("S3_UPLOADS_KMS_KEY_ALIAS"), "alias/custom-key")
	environment                 = os.Getenv("ENVIRONMENT")

	cfg        aws.Config
	httpClient *http.Client
	logger     *slog.Logger
)

type factory interface {
	AppData() (appcontext.Data, error)
	AppPublicURL() string
	DonorStartURL() string
	Bundle() (Bundle, error)
	CertificateProviderStore() CertificateProviderStore
	DynamoClient() dynamodbClient
	EventClient() EventClient
	LpaStoreClient() (LpaStoreClient, error)
	NotifyClient(ctx context.Context) (NotifyClient, error)
	Now() func() time.Time
	ScheduledStore() ScheduledStore
	ShareCodeSender(ctx context.Context) (ShareCodeSender, error)
	UidClient() UidClient
	UidStore() (UidStore, error)
	UuidString() func() string
}

type Handler interface {
	Handle(context.Context, factory, *events.CloudWatchEvent) error
}

type uidEvent struct {
	UID string `json:"uid"`
}

type feeApprovedEvent struct {
	UID          string      `json:"uid"`
	ApprovedType pay.FeeType `json:"approvedType"`
}

type dynamodbClient interface {
	AllByKeys(ctx context.Context, keys []dynamo.Keys) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllByLpaUIDAndPartialSK(ctx context.Context, uid string, partialSK dynamo.SK) ([]dynamo.Keys, error)
	AllByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v any) error
	AllBySK(ctx context.Context, sk dynamo.SK, v any) error
	AllKeysByPK(ctx context.Context, pk dynamo.PK) ([]dynamo.Keys, error)
	AnyByPK(ctx context.Context, pk dynamo.PK, v any) error
	BatchPut(ctx context.Context, items []any) error
	Create(ctx context.Context, v any) error
	CreateOnly(ctx context.Context, v any) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	LatestForActor(ctx context.Context, sk dynamo.SK, v any) error
	Move(ctx context.Context, oldKeys dynamo.Keys, value any) error
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v any) error
	OneByPK(ctx context.Context, pk dynamo.PK, v any) error
	OneByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v any) error
	OneBySK(ctx context.Context, sk dynamo.SK, v any) error
	OneByUID(ctx context.Context, uid string) (dynamo.Keys, error)
	Put(ctx context.Context, v any) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type s3Client interface {
	GetObjectTags(ctx context.Context, key string) ([]types.Tag, error)
}

type DocumentStore interface {
	UpdateScanResults(ctx context.Context, lpaID, objectKey string, virusDetected bool) error
}

type EventClient interface {
	SendApplicationUpdated(ctx context.Context, event event.ApplicationUpdated) error
	SendCertificateProviderStarted(ctx context.Context, event event.CertificateProviderStarted) error
	SendLpaAccessGranted(ctx context.Context, event event.LpaAccessGranted) error
	SendAttorneyStarted(ctx context.Context, event event.AttorneyStarted) error
	SendNotificationSent(ctx context.Context, notificationSentEvent event.NotificationSent) error
	SendPaperFormRequested(ctx context.Context, paperFormRequestedEvent event.PaperFormRequested) error
}

type ScheduledStore interface {
	DeleteAllByUID(ctx context.Context, uid string) error
	DeleteAllActionByUID(ctx context.Context, actions []scheduled.Action, uid string) error
}

type NotifyClient interface {
	EmailGreeting(lpa *lpadata.Lpa) string
	SendActorEmail(context context.Context, to notify.ToEmail, lpaUID string, email notify.Email) error
	SendActorSMS(context context.Context, to notify.ToMobile, lpaUID string, sms notify.SMS) error
}

type Bundle interface {
	For(lang localize.Lang) localize.Localizer
}

type Event struct {
	S3Event  *events.S3Event
	SQSEvent *events.SQSEvent
}

func (e *Event) UnmarshalJSON(data []byte) error {
	var s3 events.S3Event
	if err := json.Unmarshal(data, &s3); err == nil && len(s3.Records) > 0 && s3.Records[0].S3.Bucket.Name != "" {
		e.S3Event = &s3
		return nil
	}

	var sqs events.SQSEvent
	if err := json.Unmarshal(data, &sqs); err == nil && len(sqs.Records) > 0 && sqs.Records[0].MessageId != "" {
		e.SQSEvent = &sqs
		return nil
	}

	return errors.New("unknown event type")
}

func handler(ctx context.Context, event Event) (map[string]any, error) {
	result := map[string]any{}

	dynamoClient, err := dynamo.NewClient(cfg, tableName)
	if err != nil {
		return result, fmt.Errorf("failed to create dynamodb client: %w", err)
	}

	if event.S3Event != nil {
		logger.InfoContext(ctx, "handling s3 event")

		s3Client, err := s3.NewClient(cfg, evidenceBucketName, kmsKeyAlias)
		if err != nil {
			return result, fmt.Errorf("failed to create s3 client: %w", err)
		}

		documentStore := document.NewStore(dynamoClient, nil, nil)

		if err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore); err != nil {
			return result, fmt.Errorf("ObjectTagging:Put: %w", err)
		}

		return result, nil
	}

	factory := &Factory{
		logger:                      logger,
		now:                         time.Now,
		uuidString:                  random.UuidString,
		cfg:                         cfg,
		dynamoClient:                dynamoClient,
		appPublicURL:                appPublicURL,
		donorStartURL:               donorStartURL,
		attorneyStartURL:            attorneyStartURL,
		certificateProviderStartURL: certificateProviderStartURL,
		lpaStoreBaseURL:             lpaStoreBaseURL,
		lpaStoreSecretARN:           lpaStoreSecretARN,
		uidBaseURL:                  uidBaseURL,
		notifyBaseURL:               notifyBaseURL,
		eventBusName:                eventBusName,
		searchEndpoint:              searchEndpoint,
		searchIndexName:             searchIndexName,
		searchIndexingEnabled:       searchIndexingEnabled,
		httpClient:                  httpClient,
		environment:                 environment,
	}

	if event.SQSEvent != nil {
		batchItemFailures := []map[string]any{}
		for _, record := range event.SQSEvent.Records {
			var cloud *events.CloudWatchEvent
			if err := json.Unmarshal([]byte(record.Body), &cloud); err != nil {
				logger.ErrorContext(ctx, "could not unmarshal event", slog.String("messageID", record.MessageId), slog.Any("err", err))
				batchItemFailures = append(batchItemFailures, map[string]any{"itemIdentifier": record.MessageId})
				continue
			}

			if err := handleCloudWatchEvent(ctx, factory, cloud); err != nil {
				logger.ErrorContext(ctx, "error processing event", slog.String("messageID", record.MessageId), slog.Any("err", err))
				batchItemFailures = append(batchItemFailures, map[string]any{"itemIdentifier": record.MessageId})
				continue
			}
		}

		result["batchItemFailures"] = batchItemFailures
		return result, nil
	}

	return result, nil
}

func handleCloudWatchEvent(ctx context.Context, factory *Factory, event *events.CloudWatchEvent) error {
	var handler Handler
	switch event.Source {
	case "opg.poas.sirius":
		handler = &siriusEventHandler{}
	case "opg.poas.makeregister":
		handler = &makeregisterEventHandler{}
	case "opg.poas.lpastore":
		handler = &lpastoreEventHandler{}
	}

	if handler == nil {
		eJson, _ := json.Marshal(event)
		return fmt.Errorf("unknown event received: %s", string(eJson))
	}

	logger.InfoContext(ctx, "handling event", slog.String("source", event.Source), slog.String("detailType", event.DetailType))
	if err := handler.Handle(ctx, factory, event); err != nil {
		return fmt.Errorf("%s: %w", event.DetailType, err)
	}
	logger.InfoContext(ctx, "successfully handled event", slog.String("source", event.Source), slog.String("detailType", event.DetailType))

	return nil
}

func main() {
	ctx := context.Background()

	httpClient = &http.Client{Timeout: 30 * time.Second}

	logger = slog.New(telemetry.NewSlogHandler(slog.
		NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				switch a.Value.Kind() {
				case slog.KindAny:
					switch v := a.Value.Any().(type) {
					case *http.Request:
						return slog.Group(a.Key,
							slog.String("method", v.Method),
							slog.String("uri", v.URL.String()))
					}
				}

				return a
			},
		})).
		WithAttrs([]slog.Attr{
			slog.String("service_name", "opg-modernising-lpa/event-received"),
		}))

	var err error
	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to load default config", slog.Any("err", err))
		return
	}

	if len(awsBaseURL) > 0 {
		cfg.BaseEndpoint = aws.String(awsBaseURL)
	}

	var tp *trace.TracerProvider
	if xrayEnabled {
		tp, err = telemetry.SetupLambda(ctx, &cfg.APIOptions)
		if err != nil {
			logger.WarnContext(ctx, "error creating tracer provider", slog.Any("err", err))
		}
	}

	if tp != nil {
		httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

		defer func(ctx context.Context) {
			if err := tp.Shutdown(ctx); err != nil {
				logger.WarnContext(ctx, "error shutting down tracer provider", slog.Any("err", err))
			}
		}(ctx)

		lambda.Start(otellambda.InstrumentHandler(handler, xrayconfig.WithRecommendedOptions(tp)...))
	} else {
		lambda.Start(handler)
	}
}
