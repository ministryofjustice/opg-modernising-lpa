package scheduled

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
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
	SendActorEmail(ctx context.Context, to, lpaUID string, email notify.Email) error
}

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type Runner struct {
	logger       Logger
	store        ScheduledStore
	now          func() time.Time
	period       time.Duration
	donorStore   DonorStore
	notifyClient NotifyClient
	actions      map[Action]ActionFunc
}

func NewRunner(logger Logger, store ScheduledStore, donorStore DonorStore, notifyClient NotifyClient, period time.Duration) *Runner {
	r := &Runner{
		logger:       logger,
		store:        store,
		now:          time.Now,
		period:       period,
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}

	r.actions = map[Action]ActionFunc{
		ActionExpireDonorIdentity: r.stepCancelDonorIdentity,
	}

	return r
}

// Run the Runner, it is expected to be called in a Go routine.
func (r *Runner) Run(ctx context.Context) error {
	ticker := time.Tick(r.period)

	for {
		innerCtx, cancel := context.WithTimeout(ctx, r.period)
		defer cancel()

		r.logger.InfoContext(ctx, "runner step started")
		if err := r.step(innerCtx); err != nil {
			r.logger.ErrorContext(ctx, "runner step error", slog.Any("err", err))
		}
		r.logger.InfoContext(ctx, "runner step finished")

		select {
		case <-ctx.Done():
			return nil
		case <-ticker:
			continue
		}
	}
}

func (r *Runner) step(ctx context.Context) error {
	backoff := time.Second
	retryCount := 0

	for {
		row, err := r.store.Pop(ctx, r.now())
		if errors.Is(err, dynamo.NotFoundError{}) {
			return nil
		} else if errors.Is(err, dynamo.ConditionalCheckFailedError{}) {
			r.logger.InfoContext(ctx, "runner conditional check failed")
			retryCount++
			count := rand.IntN(retryCount) + 1
			time.Sleep(time.Duration(count) * backoff)
			if retryCount > 10 {
				return errors.New("runner deadlock")
			}
			continue
		} else if err != nil {
			return err
		}

		retryCount = 0
		r.logger.InfoContext(ctx, "runner action", slog.String("action", row.Action.String()))

		if fn, ok := r.actions[row.Action]; ok {
			if err := fn(ctx, row); err != nil {
				if errors.Is(err, errStepIgnored) {
					r.logger.InfoContext(ctx, "runner action ignored",
						slog.String("action", row.Action.String()),
						slog.String("target_pk", row.TargetLpaKey.PK()),
						slog.String("target_sk", row.TargetLpaOwnerKey.SK()))
				} else {
					r.logger.ErrorContext(ctx, "runner action error",
						slog.String("action", row.Action.String()),
						slog.String("target_pk", row.TargetLpaKey.PK()),
						slog.String("target_sk", row.TargetLpaOwnerKey.SK()),
						slog.Any("err", err))
				}
			} else {
				r.logger.InfoContext(ctx, "runner action success",
					slog.String("action", row.Action.String()),
					slog.String("target_pk", row.TargetLpaKey.PK()),
					slog.String("target_sk", row.TargetLpaOwnerKey.SK()))
			}
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			continue
		}
	}
}

func (r *Runner) stepCancelDonorIdentity(ctx context.Context, row *Event) error {
	provided, err := r.donorStore.One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey)
	if err != nil {
		return fmt.Errorf("error retrieving donor: %w", err)
	}

	if !provided.DonorIdentityUserData.Status.IsConfirmed() || !provided.SignedAt.IsZero() {
		return errStepIgnored
	}

	provided.DonorIdentityUserData = identity.UserData{Status: identity.StatusExpired}
	provided.Tasks.ConfirmYourIdentityAndSign = task.IdentityStateNotStarted

	if err := r.notifyClient.SendActorEmail(ctx, provided.CorrespondentEmail(), provided.LpaUID, notify.DonorIdentityCheckExpiredEmail{}); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	if err := r.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("error updating donor: %w", err)
	}

	return nil
}
