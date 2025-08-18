package notify

import "github.com/ministryofjustice/opg-modernising-lpa/internal/localize"

type SMS interface {
	smsID(localize.Lang) string
}

type CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS struct {
	DonorFullNamePossessive string
	LpaType                 string
	LpaReferenceNumber      string
	DonorFirstNames         string
}

func (s CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS) smsID(lang localize.Lang) string {
	if lang.IsCy() {
		return "bacd1fe6-9259-48fd-a62e-f61bc3b95c19"
	}

	return "28873afc-f019-48c1-bd25-df88c27813e0"
}

type CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS struct {
	DonorFullName string
	LpaType       string
}

func (s CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS) smsID(_ localize.Lang) string {
	return "796990f2-cf49-48a4-9f04-fc12f4a9702b"
}

type CertificateProviderActingOnPaperDetailsChangedSMS struct {
	DonorFullName      string
	LpaReferenceNumber string
	DonorFirstNames    string
}

func (s CertificateProviderActingOnPaperDetailsChangedSMS) smsID(_ localize.Lang) string {
	return "b3044df8-b58d-4eb0-bfc2-de6fa778a2c9"
}

type CertificateProviderActingOnPaperMeetingPromptSMS struct {
	DonorFullName                   string
	LpaType                         string
	DonorFirstNames                 string
	CertificateProviderStartPageURL string
}

func (s CertificateProviderActingOnPaperMeetingPromptSMS) smsID(_ localize.Lang) string {
	return "45589f2d-c45e-420f-9d16-f2c7a3d64bfb"
}

type WitnessCodeSMS struct {
	WitnessCode   string
	DonorFullName string
	LpaType       string
}

func (s WitnessCodeSMS) smsID(lang localize.Lang) string {
	if lang.IsCy() {
		return "5ae6190d-0610-45a2-be4f-1cdcab6e579c"
	}

	return "e39849c0-ecab-4e16-87ec-6b22afb9d535"
}

type VouchingAccessCodeSMS struct {
	AccessCode                string
	DonorFullNamePossessive   string
	LpaType                   string
	LpaReferenceNumber        string
	VoucherFullName           string
	DonorFirstNamesPossessive string
}

func (s VouchingAccessCodeSMS) smsID(lang localize.Lang) string {
	if lang.IsCy() {
		return "ab573f2e-de40-42ad-a4cf-25ba3be1fe0c"
	}

	return "4864a99a-40a7-4aef-8c1d-7fdc0a4775b9"
}

type VoucherHasConfirmedDonorIdentitySMS struct {
	VoucherFullName    string
	DonorFullName      string
	DonorStartPageURL  string
	LpaType            string
	LpaReferenceNumber string
}

func (e VoucherHasConfirmedDonorIdentitySMS) smsID(_ localize.Lang) string {
	return "ba3a4ae6-e68c-44e4-9923-f84d83c5f147"
}

type VoucherHasConfirmedDonorIdentityOnSignedLpaSMS struct {
	VoucherFullName    string
	DonorStartPageURL  string
	DonorFullName      string
	LpaType            string
	LpaReferenceNumber string
}

func (e VoucherHasConfirmedDonorIdentityOnSignedLpaSMS) smsID(_ localize.Lang) string {
	return "7067aa92-df60-4e80-b2bf-0c64a4256d68"
}

type PaperDonorLpaSubmittedSMS struct {
	LpaType            string
	LpaReferenceNumber string
}

func (e PaperDonorLpaSubmittedSMS) smsID(lang localize.Lang) string {
	if lang.IsCy() {
		return "7656df1d-d84a-4c35-aef9-5b182c2d9199"
	}

	return "edd5d11d-e9e8-4e80-a4e1-daaa46efbe0f"
}

type PaperDonorCertificateProvidedSMS struct {
	CertificateProviderFullName string
	LpaType                     string
	LpaReferenceNumber          string
}

func (e PaperDonorCertificateProvidedSMS) smsID(lang localize.Lang) string {
	if lang.IsCy() {
		return "fb5f82a6-2046-4242-ba5f-a9c55cb7318f"
	}

	return "6b3d9a6c-5103-4c16-8c09-6ebaaec58f93"
}

type OnlineDonorLPASubmissionConfirmation struct {
	LpaType            string
	LpaReferenceNumber string
}

func (e OnlineDonorLPASubmissionConfirmation) smsID(lang localize.Lang) string {
	if lang.IsCy() {
		return "87dc8630-0248-41e1-af0e-2963b77e0cf6"
	}

	return "7e0d73ff-9ee6-444d-b85d-f72ffbe9e7b4"
}
