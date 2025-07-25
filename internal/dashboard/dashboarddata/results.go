package dashboarddata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type Results struct {
	Donor               []Actor
	CertificateProvider []Actor
	Attorney            []Actor
	Voucher             []Actor
}

func (r Results) Empty() bool {
	return len(r.Donor) == 0 && len(r.CertificateProvider) == 0 && len(r.Attorney) == 0 && len(r.Voucher) == 0
}

func (r Results) ByActorType(actorType actor.Type) []Actor {
	switch actorType {
	case actor.TypeDonor:
		return r.Donor
	case actor.TypeAttorney, actor.TypeReplacementAttorney, actor.TypeTrustCorporation, actor.TypeReplacementTrustCorporation:
		return r.Attorney
	case actor.TypeCertificateProvider:
		return r.CertificateProvider
	case actor.TypeVoucher:
		return r.Voucher
	default:
		return []Actor{}
	}
}

type Actor struct {
	Lpa                 *lpadata.Lpa
	LpaAttorney         *lpadata.Attorney
	LpaTrustCorporation *lpadata.TrustCorporation
	Donor               *donordata.Provided
	CertificateProvider *certificateproviderdata.Provided
	Attorney            *attorneydata.Provided
	Voucher             *voucherdata.Provided
}
