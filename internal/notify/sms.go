package notify

import "github.com/ministryofjustice/opg-modernising-lpa/internal/localize"

type SMS interface {
	smsID(bool, localize.Lang) string
}

type CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS struct {
	DonorFullNamePossessive string
	LpaType                 string
	LpaReferenceNumber      string
	DonorFirstNames         string
}

func (s CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS) smsID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "bacd1fe6-9259-48fd-a62e-f61bc3b95c19"
		}

		return "28873afc-f019-48c1-bd25-df88c27813e0"
	}

	if lang.IsCy() {
		return "792dc3d1-766a-4c5b-a9b6-59b5b47d22e7"
	}

	return "bcdc85a7-32b1-40a6-a61f-a552406e6ecc"
}

type CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS struct {
	DonorFullName string
	LpaType       string
}

func (s CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS) smsID(isProduction bool, _ localize.Lang) string {
	if isProduction {
		return "796990f2-cf49-48a4-9f04-fc12f4a9702b"
	}

	return "292cc508-811e-44fa-9962-3fdf10e2e8cd"
}

type CertificateProviderActingOnPaperDetailsChangedSMS struct {
	DonorFullName      string
	LpaReferenceNumber string
	DonorFirstNames    string
}

func (s CertificateProviderActingOnPaperDetailsChangedSMS) smsID(isProduction bool, _ localize.Lang) string {
	if isProduction {
		return "b3044df8-b58d-4eb0-bfc2-de6fa778a2c9"
	}

	return "dfa5e0d7-6327-4053-8f51-d2d7e60128dc"
}

type CertificateProviderActingOnPaperMeetingPromptSMS struct {
	DonorFullName                   string
	LpaType                         string
	DonorFirstNames                 string
	CertificateProviderStartPageURL string
}

func (s CertificateProviderActingOnPaperMeetingPromptSMS) smsID(isProduction bool, _ localize.Lang) string {
	if isProduction {
		return "45589f2d-c45e-420f-9d16-f2c7a3d64bfb"
	}

	return "aa76d354-200d-461b-a1ff-ba99fb9c4d9e"
}

type WitnessCodeSMS struct {
	WitnessCode   string
	DonorFullName string
	LpaType       string
}

func (s WitnessCodeSMS) smsID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "5ae6190d-0610-45a2-be4f-1cdcab6e579c"
		}

		return "e39849c0-ecab-4e16-87ec-6b22afb9d535"
	}

	if lang.IsCy() {
		return "482ee4ca-5934-42b0-b9eb-57de4aa58f5a"
	}

	return "dfa15e16-1f23-494a-bffb-a475513df6cc"
}

type VouchingShareCodeSMS struct {
	ShareCode                 string
	DonorFullNamePossessive   string
	LpaType                   string
	LpaReferenceNumber        string
	VoucherFullName           string
	DonorFirstNamesPossessive string
}

func (s VouchingShareCodeSMS) smsID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "ab573f2e-de40-42ad-a4cf-25ba3be1fe0c"
		}

		return "4864a99a-40a7-4aef-8c1d-7fdc0a4775b9"
	}

	if lang.IsCy() {
		return "ae5554c5-0c9c-4b39-9527-406c05167816"
	}

	return "84d70372-5c7a-4a88-a836-ee7c7dea203a"
}

type VoucherHasConfirmedDonorIdentitySMS struct {
	VoucherFullName    string
	DonorFullName      string
	DonorStartPageURL  string
	LpaType            string
	LpaReferenceNumber string
}

func (e VoucherHasConfirmedDonorIdentitySMS) smsID(isProduction bool, _ localize.Lang) string {
	if isProduction {
		return "ba3a4ae6-e68c-44e4-9923-f84d83c5f147"
	}

	return "aedb029f-fe6a-4e8a-a5a5-d38ba948fff4"
}

type VoucherHasConfirmedDonorIdentityOnSignedLpaSMS struct {
	VoucherFullName    string
	DonorStartPageURL  string
	DonorFullName      string
	LpaType            string
	LpaReferenceNumber string
}

func (e VoucherHasConfirmedDonorIdentityOnSignedLpaSMS) smsID(isProduction bool, _ localize.Lang) string {
	if isProduction {
		return "7067aa92-df60-4e80-b2bf-0c64a4256d68"
	}

	return "65072ed8-0d20-4e0d-9800-2f1407407d32"
}

type PaperDonorLpaSubmittedSMS struct {
	LpaType            string
	LpaReferenceNumber string
}

func (e PaperDonorLpaSubmittedSMS) smsID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "7656df1d-d84a-4c35-aef9-5b182c2d9199"
		}

		return "edd5d11d-e9e8-4e80-a4e1-daaa46efbe0f"
	}

	if lang.IsCy() {
		return "ad3cf00d-c564-454a-9a8f-0a01ec94e261"
	}

	return "e7476d24-6d37-4137-b4a0-de14d3a977ed"
}

type PaperDonorCertificateProvidedSMS struct {
	CertificateProviderFullName string
	LpaType                     string
	LpaReferenceNumber          string
}

func (e PaperDonorCertificateProvidedSMS) smsID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "fb5f82a6-2046-4242-ba5f-a9c55cb7318f"
		}

		return "6b3d9a6c-5103-4c16-8c09-6ebaaec58f93"
	}

	if lang.IsCy() {
		return "c483e0f0-1e70-4b40-a6b9-512b1f28b786"
	}

	return "ecdfef3e-cdc0-4393-add5-571bd9cd5c9f"
}
