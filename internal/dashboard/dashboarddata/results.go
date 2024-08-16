package dashboarddata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type Results struct {
	Donor               []Actor
	CertificateProvider []Actor
	Attorney            []Actor
	Voucher             []Actor
}

type Actor struct {
	Lpa                 *lpadata.Lpa
	CertificateProvider *certificateproviderdata.Provided
	Attorney            *attorneydata.Provided
	Voucher             *voucherdata.Provided
}
