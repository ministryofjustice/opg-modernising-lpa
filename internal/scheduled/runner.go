package scheduled

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
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

type NotifyClient interface {
	SendActorEmail(ctx context.Context, lang localize.Lang, to, lpaUID string, email notify.Email) error
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

type Runner struct {
	logger        Logger
	store         ScheduledStore
	now           func() time.Time
	since         func(time.Time) time.Duration
	donorStore    DonorStore
	notifyClient  NotifyClient
	actions       map[Action]ActionFunc
	waiter        Waiter
	metricsClient MetricsClient
	processed     float64
	ignored       float64
	errored       float64
	// TODO remove in MLPAB-2690
	metricsEnabled bool
}

func NewRunner(logger Logger, store ScheduledStore, donorStore DonorStore, notifyClient NotifyClient, metricsClient MetricsClient, metricsEnabled bool) *Runner {
	r := &Runner{
		logger:         logger,
		store:          store,
		now:            time.Now,
		since:          time.Since,
		donorStore:     donorStore,
		notifyClient:   notifyClient,
		waiter:         &waiter{backoff: time.Second, sleep: time.Sleep, maxRetries: 10},
		metricsClient:  metricsClient,
		metricsEnabled: metricsEnabled,
	}

	r.actions = map[Action]ActionFunc{
		ActionExpireDonorIdentity: r.stepCancelDonorIdentity,
	}

	return r
}

func (r *Runner) Processed() {
	r.processed++
}

func (r *Runner) Ignored() {
	r.ignored++
}

func (r *Runner) Errored() {
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
					r.logger.InfoContext(ctx, "runner action ignored",
						slog.String("action", row.Action.String()),
						slog.String("target_pk", row.TargetLpaKey.PK()),
						slog.String("target_sk", row.TargetLpaOwnerKey.SK()))

					r.Ignored()
				} else {
					r.logger.ErrorContext(ctx, "runner action error",
						slog.String("action", row.Action.String()),
						slog.String("target_pk", row.TargetLpaKey.PK()),
						slog.String("target_sk", row.TargetLpaOwnerKey.SK()),
						slog.Any("err", err))

					r.Errored()
				}
			} else {
				r.logger.InfoContext(ctx, "runner action success",
					slog.String("action", row.Action.String()),
					slog.String("target_pk", row.TargetLpaKey.PK()),
					slog.String("target_sk", row.TargetLpaOwnerKey.SK()))

				r.Processed()
			}
		}
	}
}

func (r *Runner) stepCancelDonorIdentity(ctx context.Context, row *Event) error {
	provided, err := r.donorStore.One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey)
	if err != nil {
		return fmt.Errorf("error retrieving donor: %w", err)
	}

	if !provided.IdentityUserData.Status.IsConfirmed() || !provided.SignedAt.IsZero() {
		return errStepIgnored
	}

	provided.IdentityUserData = identity.UserData{Status: identity.StatusExpired}
	provided.Tasks.ConfirmYourIdentity = task.IdentityStateNotStarted

	if err := r.notifyClient.SendActorEmail(ctx, provided.Donor.ContactLanguagePreference, provided.CorrespondentEmail(), provided.LpaUID, notify.DonorIdentityCheckExpiredEmail{}); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	if err := r.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("error updating donor: %w", err)
	}

	return nil
}
