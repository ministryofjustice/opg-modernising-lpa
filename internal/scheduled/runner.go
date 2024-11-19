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

type Runner struct {
	logger       Logger
	store        ScheduledStore
	now          func() time.Time
	since        func(time.Time) time.Duration
	donorStore   DonorStore
	notifyClient NotifyClient
	actions      map[Action]ActionFunc
	waiter       Waiter
	metrics      map[string]cloudwatch.PutMetricDataInput
}

func NewRunner(logger Logger, store ScheduledStore, donorStore DonorStore, notifyClient NotifyClient) *Runner {
	r := &Runner{
		logger:       logger,
		store:        store,
		now:          time.Now,
		since:        time.Since,
		donorStore:   donorStore,
		notifyClient: notifyClient,
		waiter:       &waiter{backoff: time.Second, sleep: time.Sleep, maxRetries: 10},
	}

	r.actions = map[Action]ActionFunc{
		ActionExpireDonorIdentity: r.stepCancelDonorIdentity,
	}

	return r
}

func (r *Runner) Run(ctx context.Context) (cloudwatch.PutMetricDataInput, error) {
	r.waiter.Reset()

	processedTasks := float64(0)
	ignoredTasks := float64(0)
	errorTasks := float64(0)
	start := r.now()

	for {
		row, err := r.store.Pop(ctx, r.now())

		if errors.Is(err, dynamo.NotFoundError{}) {
			r.logger.InfoContext(ctx, "no scheduled tasks to process")

			var metrics cloudwatch.PutMetricDataInput

			if processedTasks > 0 || errorTasks > 0 || ignoredTasks > 0 {
				elapsed := r.since(start)

				metrics = cloudwatch.PutMetricDataInput{
					Namespace: aws.String("schedule-runner"),
					MetricData: []types.MetricDatum{
						{
							MetricName: aws.String("TasksProcessed"),
							Unit:       types.StandardUnitCount,
							Value:      aws.Float64(processedTasks),
						},
						{
							MetricName: aws.String("TasksIgnored"),
							Unit:       types.StandardUnitCount,
							Value:      aws.Float64(ignoredTasks),
						},
						{
							MetricName: aws.String("Errors"),
							Unit:       types.StandardUnitCount,
							Value:      aws.Float64(errorTasks),
						},
						{
							MetricName: aws.String("ProcessingTime"),
							Unit:       types.StandardUnitMilliseconds,
							Value:      aws.Float64(float64(elapsed.Milliseconds())),
						},
					},
				}
			}

			return metrics, nil
		} else if err != nil {
			r.logger.ErrorContext(ctx, "error getting scheduled task", slog.Any("err", err))

			if err := r.waiter.Wait(); err != nil {
				return cloudwatch.PutMetricDataInput{}, err
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

					ignoredTasks++
				} else {
					r.logger.ErrorContext(ctx, "runner action error",
						slog.String("action", row.Action.String()),
						slog.String("target_pk", row.TargetLpaKey.PK()),
						slog.String("target_sk", row.TargetLpaOwnerKey.SK()),
						slog.Any("err", err))

					errorTasks++
				}
			} else {
				r.logger.InfoContext(ctx, "runner action success",
					slog.String("action", row.Action.String()),
					slog.String("target_pk", row.TargetLpaKey.PK()),
					slog.String("target_sk", row.TargetLpaOwnerKey.SK()))

				processedTasks++
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
