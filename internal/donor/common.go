package donor

import (
	"context"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
)

type ReuseStore interface {
	PutAttorney(ctx context.Context, attorney donordata.Attorney) error
	PutAttorneys(ctx context.Context, attorneys []donordata.Attorney) error
	Attorneys(ctx context.Context, provided *donordata.Provided) ([]donordata.Attorney, error)
	DeleteAttorney(ctx context.Context, attorney donordata.Attorney) error
	PutCertificateProvider(ctx context.Context, certificateProvider donordata.CertificateProvider) error
	CertificateProviders(ctx context.Context) ([]donordata.CertificateProvider, error)
	DeleteCertificateProvider(ctx context.Context, certificateProvider donordata.CertificateProvider) error
	PutCorrespondent(ctx context.Context, correspondent donordata.Correspondent) error
	Correspondents(ctx context.Context) ([]donordata.Correspondent, error)
	DeleteCorrespondent(ctx context.Context, correspondent donordata.Correspondent) error
	PutPersonToNotify(ctx context.Context, personToNotify donordata.PersonToNotify) error
	PutPeopleToNotify(ctx context.Context, peopleToNotify []donordata.PersonToNotify) error
	PeopleToNotify(ctx context.Context, provided *donordata.Provided) ([]donordata.PersonToNotify, error)
	DeletePersonToNotify(ctx context.Context, personToNotify donordata.PersonToNotify) error
	PutTrustCorporation(ctx context.Context, trustCorporation donordata.TrustCorporation) error
	TrustCorporations(ctx context.Context, provided *donordata.Provided) ([]donordata.TrustCorporation, error)
	DeleteTrustCorporation(ctx context.Context, trustCorporation donordata.TrustCorporation) error
}

type PutStore interface {
	Put(ctx context.Context, donor *donordata.Provided) error
}
