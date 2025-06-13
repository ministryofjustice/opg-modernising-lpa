package donor

import (
	"context"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
)

type ReuseStore interface {
	PutCertificateProvider(ctx context.Context, certificateProvider donordata.CertificateProvider) error
	CertificateProviders(ctx context.Context) ([]donordata.CertificateProvider, error)
	DeleteCertificateProvider(ctx context.Context, certificateProvider donordata.CertificateProvider) error
	PutPersonToNotify(ctx context.Context, personToNotify donordata.PersonToNotify) error
	PutPeopleToNotify(ctx context.Context, peopleToNotify []donordata.PersonToNotify) error
	PeopleToNotify(ctx context.Context, provided *donordata.Provided) ([]donordata.PersonToNotify, error)
	DeletePersonToNotify(ctx context.Context, personToNotify donordata.PersonToNotify) error
}

type PutStore interface {
	Put(ctx context.Context, donor *donordata.Provided) error
}
