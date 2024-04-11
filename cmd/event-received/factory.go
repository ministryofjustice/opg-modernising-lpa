package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type LambdaClient interface {
	Do(*http.Request) (*http.Response, error)
}

type LpaStoreClient interface {
	SendLpa(ctx context.Context, donor *actor.DonorProvidedDetails) error
	Lpa(ctx context.Context, uid string) (*lpastore.Lpa, error)
}

type SecretsClient interface {
	Secret(ctx context.Context, name string) (string, error)
}

type ShareCodeSender interface {
	SendCertificateProviderInvite(context.Context, page.AppData, page.CertificateProviderInvite) error
	SendCertificateProviderPrompt(context.Context, page.AppData, *actor.DonorProvidedDetails) error
	SendAttorneys(context.Context, page.AppData, *lpastore.Lpa) error
}

type UidStore interface {
	Set(ctx context.Context, lpaID, sessionID, organisationID, uid string) error
}

type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (string, error)
}

type Factory struct {
	now                   func() time.Time
	cfg                   aws.Config
	dynamoClient          dynamodbClient
	appPublicURL          string
	lpaStoreBaseURL       string
	uidBaseURL            string
	notifyBaseURL         string
	notifyIsProduction    bool
	eventBusName          string
	searchEndpoint        string
	searchIndexingEnabled bool

	// previously constructed values
	appData         *page.AppData
	lambdaClient    LambdaClient
	secretsClient   SecretsClient
	shareCodeSender ShareCodeSender
	lpaStoreClient  LpaStoreClient
	uidStore        UidStore
	uidClient       UidClient
}

func (f *Factory) AppData() (page.AppData, error) {
	if f.appData == nil {
		bundle, err := localize.NewBundle("./lang/en.json", "./lang/cy.json")
		if err != nil {
			return page.AppData{}, err
		}

		//TODO do this in handleFeeApproved when/if we save lang preference in LPA
		f.appData = &page.AppData{Localizer: bundle.For(localize.En)}
	}

	return *f.appData, nil
}

func (f *Factory) LambdaClient() LambdaClient {
	if f.lambdaClient == nil {
		f.lambdaClient = lambda.New(f.cfg, v4.NewSigner(), &http.Client{Timeout: 10 * time.Second}, time.Now)
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
		secretsClient, err := f.SecretsClient()
		if err != nil {
			return nil, err
		}

		notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
		if err != nil {
			return nil, fmt.Errorf("failed to get notify API secret: %w", err)
		}

		notifyClient, err := notify.New(f.notifyIsProduction, f.notifyBaseURL, notifyApiKey, http.DefaultClient, event.NewClient(f.cfg, f.eventBusName))
		if err != nil {
			return nil, err
		}

		f.shareCodeSender = page.NewShareCodeSender(app.NewShareCodeStore(f.dynamoClient), notifyClient, f.appPublicURL, random.String, event.NewClient(f.cfg, f.eventBusName))
	}

	return f.shareCodeSender, nil
}

func (f *Factory) LpaStoreClient() (LpaStoreClient, error) {
	if f.lpaStoreClient == nil {
		secretsClient, err := f.SecretsClient()
		if err != nil {
			return nil, err
		}

		f.lpaStoreClient = lpastore.New(f.lpaStoreBaseURL, secretsClient, f.LambdaClient())
	}

	return f.lpaStoreClient, nil
}

func (f *Factory) UidStore() (UidStore, error) {
	if f.uidStore == nil {
		searchClient, err := search.NewClient(f.cfg, f.searchEndpoint, f.searchIndexingEnabled)
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
