package notify

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
)

type To interface {
	ToEmail
	ToMobile
}

type ToEmail interface {
	toEmail() (string, localize.Lang)
	ignore() bool
}

type ToMobile interface {
	toMobile() (string, localize.Lang)
	ignore() bool
}

type to struct {
	email   string
	mobile  string
	lang    localize.Lang
	ignored bool
}

func (t to) toEmail() (string, localize.Lang)  { return t.email, t.lang }
func (t to) toMobile() (string, localize.Lang) { return t.mobile, t.lang }
func (t to) ignore() bool                      { return t.ignored }

func ToDonor(donor *donordata.Provided) To {
	to := to{
		mobile: donor.Donor.Mobile,
		email:  donor.Donor.Email,
		lang:   donor.Donor.ContactLanguagePreference,
	}

	if donor.Correspondent.Email != "" {
		to.email = donor.Correspondent.Email
	}
	if donor.Correspondent.Phone != "" {
		to.mobile = donor.Correspondent.Phone
	}

	return to
}

func ToLpaDonor(lpa *lpadata.Lpa) To {
	to := to{
		mobile: lpa.Donor.Mobile,
		email:  lpa.Donor.Email,
		lang:   lpa.Donor.ContactLanguagePreference,
	}

	if lpa.Correspondent.Email != "" {
		to.email = lpa.Correspondent.Email
	}
	if lpa.Correspondent.Phone != "" {
		to.mobile = lpa.Correspondent.Phone
	}

	return to
}

// ToCertificateProvider should only be used for the initial communication with
// the certificate provider, after that it may be possible to use the data they
// have entered so only use this as a fallback.
func ToCertificateProvider(certificateProvider donordata.CertificateProvider) To {
	return to{
		mobile: certificateProvider.Mobile,
		email:  certificateProvider.Email,
		lang:   localize.En,
	}
}

func ToProvidedCertificateProvider(provided *certificateproviderdata.Provided, certificateProvider donordata.CertificateProvider) To {
	return to{
		mobile: certificateProvider.Mobile,
		email:  provided.Email,
		lang:   provided.ContactLanguagePreference,
	}
}

func ToLpaCertificateProvider(provided *certificateproviderdata.Provided, lpa *lpadata.Lpa) To {
	return to{
		mobile: lpa.CertificateProvider.Phone,
		email:  lpa.CertificateProvider.Email,
		lang:   provided.ContactLanguagePreference,
	}
}

func ToLpaAttorney(attorney lpadata.Attorney) To {
	return to{
		mobile:  attorney.Mobile,
		email:   attorney.Email,
		lang:    attorney.ContactLanguagePreference,
		ignored: attorney.Removed,
	}
}

func ToLpaTrustCorporation(trustCorporation lpadata.TrustCorporation) To {
	return to{
		mobile: trustCorporation.Mobile,
		email:  trustCorporation.Email,
		lang:   trustCorporation.ContactLanguagePreference,
	}
}

func ToIndependentWitness(independentWitness donordata.IndependentWitness) ToMobile {
	return to{
		mobile: independentWitness.Mobile,
		lang:   localize.En,
	}
}

func ToVoucher(voucher donordata.Voucher) ToEmail {
	return to{
		email: voucher.Email,
		lang:  localize.En,
	}
}

func ToPayee(resp pay.GetPaymentResponse) ToEmail {
	return to{
		email: resp.Email,
		lang:  localize.En,
	}
}

func ToCustomEmail(lang localize.Lang, email string) ToEmail {
	return to{
		lang:  lang,
		email: email,
	}
}
