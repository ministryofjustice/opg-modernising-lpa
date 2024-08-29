package cron

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

const period = time.Hour

type CronStore interface {
	Pop(ctx context.Context, at time.Time) (Row, error)
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
	store        CronStore
	now          func() time.Time
	lastStep     time.Time
	donorStore   DonorStore
	notifyClient NotifyClient
	actions      map[Action]func(context.Context, Row) error
}

func NewRunner(logger Logger, store CronStore, donorStore DonorStore, notifyClient NotifyClient) *Runner {
	r := &Runner{
		logger:       logger,
		store:        store,
		now:          time.Now,
		lastStep:     time.Now().Add(-period),
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}

	r.actions = map[Action]func(context.Context, Row) error{
		ActionCancelDonorIdentity: r.stepCancelDonorIdentity,
	}

	return r
}

// Run the Runner, it is expected to be called in a Go routine.
func (r *Runner) Run(ctx context.Context) error {
	for {
		dur := period - r.now().Sub(r.lastStep)
		r.logger.InfoContext(ctx, "runner next step scheduled", slog.Duration("in", dur))

		select {
		case <-time.After(dur):
			r.logger.InfoContext(ctx, "runner step started")
			if err := r.step(ctx); err != nil {
				// log, or something
				r.logger.ErrorContext(ctx, "runner step error", slog.Any("err", err))
			}
			r.lastStep = r.now()
			r.logger.InfoContext(ctx, "runner step finished", slog.Time("lastStep", r.lastStep))
		case <-ctx.Done():
			return nil
		}
	}
}

func (r *Runner) step(ctx context.Context) error {
	for {
		row, err := r.store.Pop(ctx, r.now())
		if errors.Is(err, dynamo.NotFoundError{}) {
			return nil
		} else if err != nil {
			return err
		}

		r.logger.InfoContext(ctx, "runner action", slog.String("action", row.Action.String()))

		if fn, ok := r.actions[row.Action]; ok {
			if err := fn(ctx, row); err != nil {
				r.logger.ErrorContext(ctx, "runner action error",
					slog.String("action", row.Action.String()),
					slog.Any("err", err))
			}
		}
	}
}

func (r *Runner) stepCancelDonorIdentity(ctx context.Context, row Row) error {
	lpaKey, ok := row.TargetPK.Unwrap().(dynamo.LpaKeyType)
	if !ok {
		return fmt.Errorf("incorrect TargetPK: %v", row.TargetPK)
	}

	provided, err := r.donorStore.One(ctx, lpaKey, row.TargetSK.Unwrap())
	if err != nil {
		return fmt.Errorf("error retrieving donor: %w", err)
	}

	if !provided.DonorIdentityUserData.Status.IsConfirmed() || !provided.SignedAt.IsZero() {
		r.logger.InfoContext(ctx, "runner step ignored",
			slog.String("action", row.Action.String()),
			slog.String("target_pk", row.TargetPK.Unwrap().PK()),
			slog.String("target_sk", row.TargetSK.Unwrap().SK()))
		return nil
	}

	provided.DonorIdentityUserData = identity.UserData{Status: identity.StatusExpired}
	provided.Tasks.ConfirmYourIdentityAndSign = task.IdentityStateNotStarted

	if err := r.notifyClient.SendActorEmail(ctx, provided.CorrespondentEmail(), provided.LpaUID, notify.DonorIdentityCheckExpiredEmail{}); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	if err := r.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("error updating donor: %w", err)
	}

	r.logger.InfoContext(ctx, "runner step ran successfully",
		slog.String("action", row.Action.String()),
		slog.String("target_pk", row.TargetPK.Unwrap().PK()),
		slog.String("target_sk", row.TargetSK.Unwrap().SK()))

	return nil
}
