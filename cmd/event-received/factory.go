package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type LambdaClient interface {
	Do(*http.Request) (*http.Response, error)
}

type LpaStoreClient interface {
	SendLpa(ctx context.Context, uid string, body lpastore.CreateLpa) error
	Lpa(ctx context.Context, uid string) (*lpadata.Lpa, error)
}

type SecretsClient interface {
	Secret(ctx context.Context, name string) (string, error)
}

type ShareCodeSender interface {
	SendCertificateProviderInvite(context.Context, appcontext.Data, sharecode.CertificateProviderInvite, notify.ToEmail) error
	SendCertificateProviderPrompt(context.Context, appcontext.Data, *donordata.Provided) error
	SendAttorneys(context.Context, appcontext.Data, *lpadata.Lpa) error
}

type UidStore interface {
	Set(ctx context.Context, provided *donordata.Provided, uid string) error
}

type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (string, error)
}

type Factory struct {
	logger                *slog.Logger
	now                   func() time.Time
	uuidString            func() string
	cfg                   aws.Config
	dynamoClient          dynamodbClient
	appPublicURL          string
	lpaStoreBaseURL       string
	lpaStoreSecretARN     string
	uidBaseURL            string
	notifyBaseURL         string
	notifyIsProduction    bool
	eventBusName          string
	searchEndpoint        string
	searchIndexName       string
	searchIndexingEnabled bool
	eventClient           EventClient
	httpClient            *http.Client

	// previously constructed values
	appData         *appcontext.Data
	bundle          Bundle
	lambdaClient    LambdaClient
	secretsClient   SecretsClient
	shareCodeSender ShareCodeSender
	lpaStoreClient  LpaStoreClient
	uidStore        UidStore
	uidClient       UidClient
	scheduledStore  ScheduledStore
	notifyClient    NotifyClient
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

func (f *Factory) ShareCodeSender(ctx context.Context) (ShareCodeSender, error) {
	if f.shareCodeSender == nil {
		notifyClient, err := f.NotifyClient(ctx)
		if err != nil {
			return nil, err
		}

		f.shareCodeSender = sharecode.NewSender(sharecode.NewStore(f.dynamoClient), notifyClient, f.appPublicURL, event.NewClient(f.cfg, f.eventBusName), certificateprovider.NewStore(f.dynamoClient), scheduled.NewStore(f.dynamoClient))
	}

	return f.shareCodeSender, nil
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
		f.eventClient = event.NewClient(f.cfg, f.eventBusName)
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

		notifyClient, err := notify.New(f.logger, f.notifyIsProduction, f.notifyBaseURL, notifyApiKey, f.httpClient, event.NewClient(f.cfg, f.eventBusName), bundle)
		if err != nil {
			return nil, err
		}

		f.notifyClient = notifyClient
	}

	return f.notifyClient, nil
}
