package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type Localizer interface {
	localize.Localizer
}

type LambdaClient interface {
	Do(*http.Request) (*http.Response, error)
}

type LpaStoreClient interface {
	Lpa(ctx context.Context, uid string) (*lpadata.Lpa, error)
	SendCertificateProviderConfirmIdentity(ctx context.Context, lpaUID string, certificateProvider *certificateproviderdata.Provided) error
	SendDonorConfirmIdentity(ctx context.Context, donor *donordata.Provided) error
	SendLpa(ctx context.Context, uid string, body lpastore.CreateLpa) error
	SendPaperCertificateProviderAccessOnline(ctx context.Context, lpa *lpadata.Lpa, certificateProviderEmail string) error
}

type SecretsClient interface {
	Secret(ctx context.Context, name string) (string, error)
}

type AccessCodeSender interface {
	SendAttorneys(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa) error
	SendCertificateProviderPrompt(ctx context.Context, appData appcontext.Data, provided *donordata.Provided) error
	SendLpaCertificateProviderPrompt(ctx context.Context, appData appcontext.Data, lpaKey dynamo.LpaKeyType, lpaOwnerKey dynamo.LpaOwnerKeyType, lpa *lpadata.Lpa) error
	SendVoucherAccessCode(ctx context.Context, provided *donordata.Provided, appData appcontext.Data) error
	SendVoucherInvite(ctx context.Context, provided *donordata.Provided, appData appcontext.Data) error
}

type UidStore interface {
	Set(ctx context.Context, provided *donordata.Provided, uid string) error
}

type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (string, error)
}

type CertificateProviderStore interface {
	Delete(ctx context.Context) error
	OneByUID(ctx context.Context, uid string) (*certificateproviderdata.Provided, error)
	Put(ctx context.Context, certificateProvider *certificateproviderdata.Provided) error
}

type Factory struct {
	logger                      *slog.Logger
	now                         func() time.Time
	uuidString                  func() string
	cfg                         aws.Config
	dynamoClient                dynamodbClient
	appPublicURL                string
	donorStartURL               string
	certificateProviderStartURL string
	attorneyStartURL            string
	lpaStoreBaseURL             string
	lpaStoreSecretARN           string
	uidBaseURL                  string
	notifyBaseURL               string
	eventBusName                string
	searchEndpoint              string
	searchIndexName             string
	searchIndexingEnabled       bool
	eventClient                 EventClient
	httpClient                  *http.Client
	environment                 string

	// previously constructed values
	appData                  *appcontext.Data
	bundle                   Bundle
	certificateProviderStore CertificateProviderStore
	lambdaClient             LambdaClient
	lpaStoreClient           LpaStoreClient
	notifyClient             NotifyClient
	secretsClient            SecretsClient
	accessCodeSender         AccessCodeSender
	scheduledStore           ScheduledStore
	uidStore                 UidStore
	uidClient                UidClient
}

func (f *Factory) Now() func() time.Time {
	return f.now
}

func (f *Factory) DynamoClient() dynamodbClient {
	return f.dynamoClient
}

func (f *Factory) UuidString() func() string {
	return f.uuidString
}

func (f *Factory) Bundle() (Bundle, error) {
	if f.bundle == nil {
		bundle, err := localize.NewBundle("./lang/en.json", "./lang/cy.json")
		if err != nil {
			return nil, err
		}

		f.bundle = bundle
	}

	return f.bundle, nil
}

func (f *Factory) AppData() (appcontext.Data, error) {
	if f.appData == nil {
		bundle, err := f.Bundle()
		if err != nil {
			return appcontext.Data{}, err
		}

		//TODO do this in handleFeeApproved when/if we save lang preference in LPA
		f.appData = &appcontext.Data{Localizer: bundle.For(localize.En)}
	}

	return *f.appData, nil
}

func (f *Factory) LambdaClient() LambdaClient {
	if f.lambdaClient == nil {
		f.lambdaClient = lambda.New(f.cfg, v4.NewSigner(), f.httpClient, time.Now)
	}

	return f.lambdaClient
}

func (f *Factory) SecretsClient() (SecretsClient, error) {
	if f.secretsClient == nil {
		client, err := secrets.NewClient(f.cfg, time.Hour)
		if err != nil {
			return nil, err
		}

		f.secretsClient = client
	}

	return f.secretsClient, nil
}

func (f *Factory) AccessCodeSender(ctx context.Context) (AccessCodeSender, error) {
	if f.accessCodeSender == nil {
		notifyClient, err := f.NotifyClient(ctx)
		if err != nil {
			return nil, err
		}

		f.accessCodeSender = accesscode.NewSender(
			accesscode.NewStore(f.dynamoClient),
			notifyClient,
			f.appPublicURL,
			f.certificateProviderStartURL,
			f.attorneyStartURL,
			f.EventClient(),
			certificateprovider.NewStore(f.dynamoClient),
			scheduled.NewStore(f.dynamoClient),
		)
	}

	return f.accessCodeSender, nil
}

func (f *Factory) LpaStoreClient() (LpaStoreClient, error) {
	if f.lpaStoreClient == nil {
		secretsClient, err := f.SecretsClient()
		if err != nil {
			return nil, err
		}

		f.lpaStoreClient = lpastore.New(f.lpaStoreBaseURL, secretsClient, f.lpaStoreSecretARN, f.LambdaClient())
	}

	return f.lpaStoreClient, nil
}

func (f *Factory) UidStore() (UidStore, error) {
	if f.uidStore == nil {
		searchClient, err := search.NewClient(f.cfg, f.searchEndpoint, f.searchIndexName, f.searchIndexingEnabled)
		if err != nil {
			return nil, err
		}

		f.uidStore = app.NewUidStore(f.dynamoClient, searchClient, f.now)
	}

	return f.uidStore, nil
}

func (f *Factory) UidClient() UidClient {
	if f.uidClient == nil {
		f.uidClient = uid.New(f.uidBaseURL, f.LambdaClient())
	}

	return f.uidClient
}

func (f *Factory) EventClient() EventClient {
	if f.eventClient == nil {
		f.eventClient = event.NewClient(f.cfg, f.environment, f.eventBusName)
	}

	return f.eventClient
}

func (f *Factory) ScheduledStore() ScheduledStore {
	if f.scheduledStore == nil {
		f.scheduledStore = scheduled.NewStore(f.dynamoClient)
	}

	return f.scheduledStore
}

func (f *Factory) NotifyClient(ctx context.Context) (NotifyClient, error) {
	if f.notifyClient == nil {
		bundle, err := f.Bundle()
		if err != nil {
			return nil, err
		}

		secretsClient, err := f.SecretsClient()
		if err != nil {
			return nil, err
		}

		notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
		if err != nil {
			return nil, fmt.Errorf("failed to get notify API secret: %w", err)
		}

		notifyClient, err := notify.New(f.logger, f.notifyBaseURL, notifyApiKey, f.httpClient, f.EventClient(), bundle)
		if err != nil {
			return nil, err
		}

		f.notifyClient = notifyClient
	}

	return f.notifyClient, nil
}

func (f *Factory) CertificateProviderStore() CertificateProviderStore {
	if f.certificateProviderStore == nil {
		f.certificateProviderStore = certificateprovider.NewStore(f.dynamoClient)
	}

	return f.certificateProviderStore
}

func (f *Factory) AppPublicURL() string {
	return f.appPublicURL
}

func (f *Factory) DonorStartURL() string {
	return f.donorStartURL
}
