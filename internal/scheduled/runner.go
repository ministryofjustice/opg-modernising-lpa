package scheduled

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
)

// errStepIgnored is returned by steps when they don't require processing
var errStepIgnored = errors.New("step ignored")

type ActionFunc func(ctx context.Context, row *Event) error

type ScheduledStore interface {
	Pop(ctx context.Context, at time.Time) (*Event, error)
}

type DonorStore interface {
	One(ctx context.Context, pk dynamo.LpaKeyType, sk dynamo.SK) (*donordata.Provided, error)
	Put(ctx context.Context, provided *donordata.Provided) error
}

type CertificateProviderStore interface {
	One(ctx context.Context, pk dynamo.LpaKeyType) (*certificateproviderdata.Provided, error)
}

type AttorneyStore interface {
	All(ctx context.Context, pk dynamo.LpaKeyType) ([]*attorneydata.Provided, error)
}

type NotifyClient interface {
	EmailGreeting(lpa *lpadata.Lpa) string
	SendActorEmail(ctx context.Context, to notify.ToEmail, lpaUID string, email notify.Email) error
}

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type Waiter interface {
	Reset()
	Wait() error
}

type MetricsClient interface {
	PutMetrics(ctx context.Context, input *cloudwatch.PutMetricDataInput) error
}

type Localizer interface {
	localize.Localizer
}

type Bundle interface {
	For(lang localize.Lang) localize.Localizer
}

type EventClient interface {
	SendLetterRequested(ctx context.Context, event event.LetterRequested) error
}

type LpaStoreResolvingService interface {
	Resolve(ctx context.Context, provided *donordata.Provided) (*lpadata.Lpa, error)
}

type Runner struct {
	logger                       Logger
	store                        ScheduledStore
	now                          func() time.Time
	since                        func(time.Time) time.Duration
	donorStore                   DonorStore
	certificateProviderStore     CertificateProviderStore
	attorneyStore                AttorneyStore
	lpaStoreResolvingService     LpaStoreResolvingService
	notifyClient                 NotifyClient
	eventClient                  EventClient
	bundle                       Bundle
	actions                      map[scheduleddata.Action]ActionFunc
	waiter                       Waiter
	metricsClient                MetricsClient
	certificateProviderStartURL  string
	certificateProviderOptOutURL string
	attorneyStartURL             string
	attorneyOptOutURL            string
	// TODO remove in MLPAB-2690
	metricsEnabled bool

	processed float64
	ignored   float64
	errored   float64
}

func NewRunner(
	logger Logger,
	store ScheduledStore,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	lpaStoreResolvingService LpaStoreResolvingService,
	notifyClient NotifyClient,
	eventClient EventClient,
	bundle Bundle,
	metricsClient MetricsClient,
	metricsEnabled bool,
	certificateProviderStartURL string,
	attorneyStartURL string,
	appPublicURL string,
) *Runner {
	r := &Runner{
		logger:                       logger,
		store:                        store,
		now:                          time.Now,
		since:                        time.Since,
		donorStore:                   donorStore,
		certificateProviderStore:     certificateProviderStore,
		attorneyStore:                attorneyStore,
		lpaStoreResolvingService:     lpaStoreResolvingService,
		notifyClient:                 notifyClient,
		eventClient:                  eventClient,
		bundle:                       bundle,
		waiter:                       &waiter{backoff: time.Second, sleep: time.Sleep, maxRetries: 10},
		metricsClient:                metricsClient,
		metricsEnabled:               metricsEnabled,
		certificateProviderStartURL:  certificateProviderStartURL,
		certificateProviderOptOutURL: appPublicURL + page.PathCertificateProviderEnterAccessCodeOptOut.Format(),
		attorneyStartURL:             attorneyStartURL,
		attorneyOptOutURL:            appPublicURL + page.PathAttorneyEnterAccessCodeOptOut.Format(),
	}

	r.actions = map[scheduleddata.Action]ActionFunc{
		scheduleddata.ActionExpireDonorIdentity:                        r.stepCancelDonorIdentity,
		scheduleddata.ActionRemindCertificateProviderToComplete:        r.stepRemindCertificateProviderToComplete,
		scheduleddata.ActionRemindCertificateProviderToConfirmIdentity: r.stepRemindCertificateProviderToConfirmIdentity,
		scheduleddata.ActionRemindAttorneyToComplete:                   r.stepRemindAttorneyToComplete,
	}

	return r
}

func (r *Runner) Processed(ctx context.Context, row *Event) {
	r.logger.InfoContext(ctx, "runner action success",
		slog.String("action", row.Action.String()),
		slog.String("target_pk", row.TargetLpaKey.PK()),
		slog.String("target_sk", row.TargetLpaOwnerKey.SK()))

	r.processed++
}

func (r *Runner) Ignored(ctx context.Context, row *Event) {
	r.logger.InfoContext(ctx, "runner action ignored",
		slog.String("action", row.Action.String()),
		slog.String("target_pk", row.TargetLpaKey.PK()),
		slog.String("target_sk", row.TargetLpaOwnerKey.SK()))

	r.ignored++
}

func (r *Runner) Errored(ctx context.Context, row *Event, err error) {
	r.logger.ErrorContext(ctx, "runner action error",
		slog.String("action", row.Action.String()),
		slog.String("target_pk", row.TargetLpaKey.PK()),
		slog.String("target_sk", row.TargetLpaOwnerKey.SK()),
		slog.Any("err", err))

	r.errored++
}

func (r *Runner) Metrics(processingTime time.Duration) cloudwatch.PutMetricDataInput {
	return cloudwatch.PutMetricDataInput{
		Namespace: aws.String("schedule-runner"),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("TasksProcessed"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(r.processed),
			},
			{
				MetricName: aws.String("TasksIgnored"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(r.ignored),
			},
			{
				MetricName: aws.String("Errors"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(r.errored),
			},
			{
				MetricName: aws.String("ProcessingTime"),
				Unit:       types.StandardUnitMilliseconds,
				Value:      aws.Float64(float64(processingTime.Milliseconds())),
			},
		},
	}
}

func (r *Runner) Run(ctx context.Context) error {
	r.waiter.Reset()

	start := r.now()

	for {
		row, err := r.store.Pop(ctx, r.now())

		if errors.Is(err, dynamo.NotFoundError{}) {
			r.logger.InfoContext(ctx, "no scheduled tasks to process")

			if (r.processed > 0 || r.ignored > 0 || r.errored > 0) && r.metricsEnabled {
				metrics := r.Metrics(r.since(start))

				if err = r.metricsClient.PutMetrics(ctx, &metrics); err != nil {
					r.logger.ErrorContext(ctx, "error putting metrics", slog.Any("err", err))
					return err
				}
			}

			return nil
		} else if err != nil {
			r.logger.ErrorContext(ctx, "error getting scheduled task", slog.Any("err", err))

			if err := r.waiter.Wait(); err != nil {
				return err
			}
			continue
		}

		r.waiter.Reset()
		r.logger.InfoContext(ctx,
			"runner action",
			slog.String("action", row.Action.String()),
		)

		if fn, ok := r.actions[row.Action]; ok {
			if err := fn(ctx, row); err != nil {
				if errors.Is(err, errStepIgnored) {
					r.Ignored(ctx, row)
				} else {
					r.Errored(ctx, row, err)
				}
			} else {
				r.Processed(ctx, row)
			}
		}
	}
}
