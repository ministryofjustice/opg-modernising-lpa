package notify

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
)

func TestToDonor(t *testing.T) {
	to := ToDonor(&donordata.Provided{
		Donor: donordata.Donor{Mobile: "0777", Email: "a@b.c", ContactLanguagePreference: localize.Cy},
	})

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.Cy, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0777", mobile)
	assert.Equal(t, localize.Cy, lang)

	assert.False(t, to.ignore())
}

func TestToDonorWhenCorrespondent(t *testing.T) {
	to := ToDonor(&donordata.Provided{
		Donor:         donordata.Donor{Mobile: "0777", Email: "a@b.c", ContactLanguagePreference: localize.Cy},
		Correspondent: donordata.Correspondent{Phone: "0779", Email: "d@e.f"},
	})

	email, lang := to.toEmail()
	assert.Equal(t, "d@e.f", email)
	assert.Equal(t, localize.Cy, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0779", mobile)
	assert.Equal(t, localize.Cy, lang)

	assert.False(t, to.ignore())
}

func TestToLpaDonor(t *testing.T) {
	to := ToLpaDonor(&lpadata.Lpa{
		Donor: lpadata.Donor{Mobile: "0777", Email: "a@b.c", ContactLanguagePreference: localize.Cy},
	})

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.Cy, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0777", mobile)
	assert.Equal(t, localize.Cy, lang)

	assert.False(t, to.ignore())
}

func TestToLpaDonorWhenCorrespondent(t *testing.T) {
	to := ToLpaDonor(&lpadata.Lpa{
		Donor:         lpadata.Donor{Mobile: "0777", Email: "a@b.c", ContactLanguagePreference: localize.Cy},
		Correspondent: lpadata.Correspondent{Phone: "0779", Email: "d@e.f"},
	})

	email, lang := to.toEmail()
	assert.Equal(t, "d@e.f", email)
	assert.Equal(t, localize.Cy, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0779", mobile)
	assert.Equal(t, localize.Cy, lang)

	assert.False(t, to.ignore())
}

func TestToCertificateProvider(t *testing.T) {
	to := ToCertificateProvider(donordata.CertificateProvider{
		Mobile: "0777",
		Email:  "a@b.c",
	})

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.En, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0777", mobile)
	assert.Equal(t, localize.En, lang)

	assert.False(t, to.ignore())
}

func TestToProvidedCertificateProvider(t *testing.T) {
	to := ToProvidedCertificateProvider(&certificateproviderdata.Provided{
		Email:                     "d@e.f",
		ContactLanguagePreference: localize.Cy,
	}, donordata.CertificateProvider{
		Mobile: "0777",
		Email:  "a@b.c",
	})

	email, lang := to.toEmail()
	assert.Equal(t, "d@e.f", email)
	assert.Equal(t, localize.Cy, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0777", mobile)
	assert.Equal(t, localize.Cy, lang)

	assert.False(t, to.ignore())
}

func TestToLpaCertificateProvider(t *testing.T) {
	to := ToLpaCertificateProvider(&certificateproviderdata.Provided{
		Email:                     "d@e.f",
		ContactLanguagePreference: localize.Cy,
	}, &lpadata.Lpa{
		CertificateProvider: lpadata.CertificateProvider{
			Phone: "0777",
			Email: "a@b.c",
		},
	})

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.Cy, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0777", mobile)
	assert.Equal(t, localize.Cy, lang)

	assert.False(t, to.ignore())
}

func TestToLpaAttorney(t *testing.T) {
	to := ToLpaAttorney(lpadata.Attorney{
		Mobile:                    "0777",
		Email:                     "a@b.c",
		ContactLanguagePreference: localize.Cy,
		Removed:                   true,
	})

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.Cy, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0777", mobile)
	assert.Equal(t, localize.Cy, lang)

	assert.True(t, to.ignore())
}

func TestToLpaTrustCorporation(t *testing.T) {
	to := ToLpaTrustCorporation(lpadata.TrustCorporation{
		Mobile:                    "0777",
		Email:                     "a@b.c",
		ContactLanguagePreference: localize.Cy,
		Removed:                   true,
	})

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.Cy, lang)

	mobile, lang := to.toMobile()
	assert.Equal(t, "0777", mobile)
	assert.Equal(t, localize.Cy, lang)

	assert.True(t, to.ignore())
}

func TestToIndependentWitness(t *testing.T) {
	to := ToIndependentWitness(donordata.IndependentWitness{
		Mobile: "0777",
	})

	mobile, lang := to.toMobile()
	assert.Equal(t, "0777", mobile)
	assert.Equal(t, localize.En, lang)

	assert.False(t, to.ignore())
}

func TestToVoucher(t *testing.T) {
	to := ToVoucher(donordata.Voucher{
		Email: "a@b.c",
	})

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.En, lang)

	assert.False(t, to.ignore())
}

func TestToPayee(t *testing.T) {
	to := ToPayee(pay.GetPaymentResponse{
		Email: "a@b.c",
	})

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.En, lang)

	assert.False(t, to.ignore())
}

func TestToCustomEmail(t *testing.T) {
	to := ToCustomEmail(localize.Cy, "a@b.c")

	email, lang := to.toEmail()
	assert.Equal(t, "a@b.c", email)
	assert.Equal(t, localize.Cy, lang)

	assert.False(t, to.ignore())
}
