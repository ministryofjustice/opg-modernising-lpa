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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
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

type CertificateProviderStore interface {
	One(ctx context.Context, pk dynamo.LpaKeyType) (*certificateproviderdata.Provided, error)
}

type NotifyClient interface {
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

type LpaStoreClient interface {
	Lpa(ctx context.Context, lpaUID string) (*lpadata.Lpa, error)
}

type Bundle interface {
	For(lang localize.Lang) *localize.Localizer
}

type EventClient interface {
	SendLetterRequested(ctx context.Context, event event.LetterRequested) error
}

type Runner struct {
	logger                   Logger
	store                    ScheduledStore
	now                      func() time.Time
	since                    func(time.Time) time.Duration
	donorStore               DonorStore
	certificateProviderStore CertificateProviderStore
	notifyClient             NotifyClient
	eventClient              EventClient
	bundle                   Bundle
	actions                  map[Action]ActionFunc
	waiter                   Waiter
	metricsClient            MetricsClient
	processed                float64
	ignored                  float64
	errored                  float64
	// TODO remove in MLPAB-2690
	metricsEnabled bool
}

func NewRunner(logger Logger, store ScheduledStore, donorStore DonorStore, certificateProviderStore CertificateProviderStore, notifyClient NotifyClient, bundle Bundle, metricsClient MetricsClient, metricsEnabled bool) *Runner {
	r := &Runner{
		logger:                   logger,
		store:                    store,
		now:                      time.Now,
		since:                    time.Since,
		donorStore:               donorStore,
		certificateProviderStore: certificateProviderStore,
		notifyClient:             notifyClient,
		bundle:                   bundle,
		waiter:                   &waiter{backoff: time.Second, sleep: time.Sleep, maxRetries: 10},
		metricsClient:            metricsClient,
		metricsEnabled:           metricsEnabled,
	}

	r.actions = map[Action]ActionFunc{
		ActionExpireDonorIdentity: r.stepCancelDonorIdentity,
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

	if err := r.notifyClient.SendActorEmail(ctx, notify.ToDonor(provided), provided.LpaUID, notify.DonorIdentityCheckExpiredEmail{}); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	if err := r.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("error updating donor: %w", err)
	}

	return nil
}

func (r *Runner) stepRemindCertificateProviderToComplete(ctx context.Context, row *Event) error {
	certificateProvider, err := r.certificateProviderStore.One(ctx, row.TargetLpaKey)
	if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return fmt.Errorf("error retrieving certificate provider: %w", err)
	}

	if certificateProvider != nil && certificateProvider.Tasks.ProvideTheCertificate.IsCompleted() {
		return errStepIgnored
	}

	provided, err := r.donorStore.One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey)
	if err != nil {
		return fmt.Errorf("error retrieving donor: %w", err)
	}

	beforeExpiry := provided.ExpiresAt().AddDate(0, -3, 0)
	afterInvite := provided.CertificateProviderInvitedAt.AddDate(0, 3, 0)

	if r.now().Before(afterInvite) || r.now().Before(beforeExpiry) {
		return errStepIgnored
	}

	emailTo := notify.ToCertificateProvider(provided.CertificateProvider)
	if certificateProvider != nil {
		emailTo = notify.ToProvidedCertificateProvider(certificateProvider, provided.CertificateProvider)
	}

	if provided.CertificateProvider.CarryOutBy.IsPaper() {
		if err := r.eventClient.SendLetterRequested(ctx, event.LetterRequested{
			UID:        provided.LpaUID,
			LetterType: "ADVISE_CERTIFICATE_PROVIDER_TO_SIGN_OR_OPT_OUT",
			ActorType:  actor.TypeCertificateProvider,
			ActorUID:   provided.CertificateProvider.UID,
		}); err != nil {
			return fmt.Errorf("could not send certificate provider letter request: %w", err)
		}
	} else {
		var localizer *localize.Localizer
		if certificateProvider != nil && !certificateProvider.ContactLanguagePreference.Empty() {
			localizer = r.bundle.For(certificateProvider.ContactLanguagePreference)
		} else {
			localizer = r.bundle.For(localize.En)
		}

		if err := r.notifyClient.SendActorEmail(ctx, emailTo, provided.LpaUID, notify.AdviseCertificateProviderToSignOrOptOutEmail{
			DonorFullName:               provided.Donor.FullName(),
			LpaType:                     localizer.T(provided.Type.String()),
			CertificateProviderFullName: provided.CertificateProvider.FullName(),
			InvitedDate:                 localizer.FormatDate(provided.CertificateProviderInvitedAt),
			DeadlineDate:                localizer.FormatDate(provided.ExpiresAt()),
		}); err != nil {
			return fmt.Errorf("could not send certificate provider email: %w", err)
		}
	}

	if provided.Donor.Channel.IsPaper() {
		letterRequest := event.LetterRequested{
			UID:        provided.LpaUID,
			LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_ACTED",
			ActorType:  actor.TypeDonor,
			ActorUID:   provided.Donor.UID,
		}

		if provided.Correspondent.Address.Line1 != "" {
			letterRequest.ActorType = actor.TypeCorrespondent
			letterRequest.ActorUID = provided.Correspondent.UID
		}

		if err := r.eventClient.SendLetterRequested(ctx, letterRequest); err != nil {
			return fmt.Errorf("could not send certificate provider letter request: %w", err)
		}
	} else {
		localizer := r.bundle.For(provided.Donor.ContactLanguagePreference)

		if err := r.notifyClient.SendActorEmail(ctx, notify.ToDonor(provided), provided.LpaUID, notify.InformDonorCertificateProviderHasNotActedEmail{
			CertificateProviderFullName: provided.CertificateProvider.FullName(),
			LpaType:                     localizer.T(provided.Type.String()),
			DonorFullName:               provided.Donor.FullName(),
			InvitedDate:                 localizer.FormatDate(provided.CertificateProviderInvitedAt),
			DeadlineDate:                localizer.FormatDate(provided.ExpiresAt()),
		}); err != nil {
			return fmt.Errorf("could not send donor email: %w", err)
		}
	}

	return nil
}
