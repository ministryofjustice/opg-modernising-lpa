package notify

import "github.com/ministryofjustice/opg-modernising-lpa/internal/localize"

type Email interface {
	emailID(localize.Lang) string
}

type InitialOriginalAttorneyEmail struct {
	DonorFullName             string
	LpaType                   string
	AttorneyFullName          string
	DonorFirstNames           string
	AttorneyStartPageURL      string
	AccessCode                string
	DonorFirstNamesPossessive string
	AttorneyOptOutURL         string
}

func (e InitialOriginalAttorneyEmail) emailID(_ localize.Lang) string {
	return "080071dc-0434-4b13-adb7-c4e5612c4b47"
}

type InitialReplacementAttorneyEmail struct {
	DonorFullName             string
	LpaType                   string
	AttorneyFullName          string
	DonorFirstNames           string
	AttorneyStartPageURL      string
	AccessCode                string
	DonorFirstNamesPossessive string
	AttorneyOptOutURL         string
}

func (e InitialReplacementAttorneyEmail) emailID(_ localize.Lang) string {
	return "8d335239-7002-4825-8393-cc00ad246648"
}

type CertificateProviderCertificateProvidedEmail struct {
	DonorFullNamePossessive     string
	LpaType                     string
	CertificateProviderFullName string
	CertificateProvidedDateTime string
	DonorFirstNamesPossessive   string
}

func (e CertificateProviderCertificateProvidedEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "3a52508e-b8f1-4192-b9f4-e912964db3e7"
	}

	return "64d7d56b-966b-464f-8084-1ac5d91c3d58"
}

type CertificateProviderInviteEmail struct {
	DonorFullName                string
	LpaType                      string
	CertificateProviderFullName  string
	DonorFirstNames              string
	DonorFirstNamesPossessive    string
	WhatLpaCovers                string
	CertificateProviderStartURL  string
	AccessCode                   string
	CertificateProviderOptOutURL string
}

func (e CertificateProviderInviteEmail) emailID(_ localize.Lang) string {
	return "08a8d89d-e5b7-4bb9-94d2-25139543e962"
}

type CertificateProviderProvideCertificatePromptEmail struct {
	DonorFullName               string
	DonorFullNamePossessive     string
	LpaType                     string
	CertificateProviderFullName string
	CertificateProviderStartURL string
	InvitedDate                 string
}

func (e CertificateProviderProvideCertificatePromptEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "73704800-f241-4b17-8e37-09ee2804a570"
	}

	return "eac04624-f058-411a-be48-854a77022ac8"
}

type CertificateProviderProvideCertificatePromptEmailAccessCodeUsed struct {
	DonorFullName               string
	DonorFullNamePossessive     string
	LpaType                     string
	CertificateProviderFullName string
	CertificateProviderStartURL string
}

func (e CertificateProviderProvideCertificatePromptEmailAccessCodeUsed) emailID(_ localize.Lang) string {
	return "eac04624-f058-411a-be48-854a77022ac8"
}

type OrganisationMemberInviteEmail struct {
	OrganisationName      string
	InviterFullName       string
	InviterEmail          string
	InviteCode            string
	JoinAnOrganisationURL string
}

func (e OrganisationMemberInviteEmail) emailID(_ localize.Lang) string {
	return "8433502f-7cbd-42de-a075-7f9343531167"
}

type DonorAccessEmail struct {
	SupporterFullName  string
	OrganisationName   string
	LpaType            string
	LpaReferenceNumber string
	DonorName          string
	URL                string
	AccessCode         string
}

func (e DonorAccessEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "cd0f3029-6d6e-4c00-9098-e20394982dc6"
	}

	return "4e7337cd-34aa-41ba-81e3-3c866e3daf4b"
}

type CertificateProviderOptedOutPreWitnessingEmail struct {
	Greeting                    string
	CertificateProviderFullName string
	DonorFullName               string
	LpaType                     string
	LpaReferenceNumber          string
	DonorStartPageURL           string
}

func (e CertificateProviderOptedOutPreWitnessingEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "28a72a2e-7dd7-4131-9ac1-616e8f453175"
	}

	return "1e85965d-4288-42ea-bdd5-f4a29020cf73"
}

type CertificateProviderOptedOutPostWitnessingEmail struct {
	Greeting                      string
	CertificateProviderFirstNames string
	CertificateProviderFullName   string
	DonorFullName                 string
	LpaType                       string
	LpaReferenceNumber            string
	DonorStartPageURL             string
}

func (e CertificateProviderOptedOutPostWitnessingEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "fb22a2fa-b3eb-42b1-9884-fb9398308bc4"
	}

	return "e284f26e-600a-44f8-b76a-95b93339a054"
}

type CertificateProviderFailedIdentityCheckEmail struct {
	Greeting                    string
	CertificateProviderFullName string
	LpaType                     string
	LpaReferenceNumber          string
	DonorStartPageURL           string
}

func (e CertificateProviderFailedIdentityCheckEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "c29f3c42-0cbf-42d0-9c5c-3ceea2095b07"
	}

	return "4020a281-8b64-45ec-85c6-19a89c08bcdb"
}

type PaymentConfirmationEmail struct {
	DonorFullNamesPossessive string
	LpaType                  string
	PaymentCardFullName      string
	LpaReferenceNumber       string
	PaymentReferenceID       string
	PaymentConfirmationDate  string
	AmountPaidWithCurrency   string
}

func (e PaymentConfirmationEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "b1d0a9d7-0886-4d5d-acc0-08fe69e492ad"
	}

	return "d0946a7d-d7fe-47cb-9b41-464f13727bf3"
}

type AttorneyOptedOutEmail struct {
	Greeting           string
	AttorneyFullName   string
	LpaType            string
	LpaReferenceNumber string
}

func (e AttorneyOptedOutEmail) emailID(_ localize.Lang) string {
	return "38bf7a04-b15c-4563-8214-bada37284744"
}

type DonorIdentityCheckExpiredEmail struct{}

func (e DonorIdentityCheckExpiredEmail) emailID(_ localize.Lang) string {
	return "c3c4a115-4d07-4e25-926d-a656dc33485a"
}

type VouchingAccessCodeEmail struct {
	AccessCode         string
	VoucherFullName    string
	DonorFullName      string
	LpaType            string
	LpaReferenceNumber string
}

func (s VouchingAccessCodeEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "93ee9148-b962-4398-ab94-d2625e39fbb1"
	}

	return "38e26a3f-d87d-4b0c-8985-8fb5bed79466"
}

type VoucherInviteEmail struct {
	VoucherFullName           string
	DonorFullName             string
	DonorFirstNamesPossessive string
	DonorFirstNames           string
	LpaType                   string
	VoucherStartPageURL       string
}

func (s VoucherInviteEmail) emailID(_ localize.Lang) string {
	return "36ad56ad-823b-4852-88a7-8acc4dfd1749"
}

type VouchingFailedAttemptEmail struct {
	Greeting           string
	VoucherFullName    string
	DonorStartPageURL  string
	LpaType            string
	LpaReferenceNumber string
}

func (e VouchingFailedAttemptEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "db45c036-1379-45d1-a521-919217f50e45"
	}

	return "f21ee857-8c3e-43ee-adf2-2d9f1ff1a1a8"
}

type VoucherHasConfirmedDonorIdentityEmail struct {
	VoucherFullName    string
	DonorFullName      string
	DonorStartPageURL  string
	LpaType            string
	LpaReferenceNumber string
}

func (e VoucherHasConfirmedDonorIdentityEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "d1c1bb6f-e9eb-44d6-802b-49590fb0d0fa"
	}

	return "67cd151e-6e7b-4fba-9457-f0252e75dfe2"
}

type VoucherHasConfirmedDonorIdentityOnSignedLpaEmail struct {
	VoucherFullName    string
	DonorFullName      string
	DonorStartPageURL  string
	LpaType            string
	LpaReferenceNumber string
}

func (e VoucherHasConfirmedDonorIdentityOnSignedLpaEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "3305cfe2-5622-4292-9838-78d6c152db23"
	}

	return "8df993ff-e4d9-43f2-b714-39053510c664"
}

type VoucherInformedTheyAreNoLongerNeededToVouchEmail struct {
	VoucherFullName string
	DonorFullName   string
}

func (e VoucherInformedTheyAreNoLongerNeededToVouchEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "TODO"
	}

	return "ca7c6a15-bdf3-47fe-ba01-d811ccdbc30d"
}

type AdviseCertificateProviderToSignOrOptOutEmail struct {
	DonorFullName                   string
	DonorFullNamePossessive         string
	LpaType                         string
	CertificateProviderFullName     string
	InvitedDate                     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
	CertificateProviderOptOutURL    string
}

func (e AdviseCertificateProviderToSignOrOptOutEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "083ace46-2e6d-41c3-9e82-84df8bc03faf"
	}

	return "fc01c541-28f3-4e04-921e-e8848f810278"
}

type AdviseCertificateProviderToSignOrOptOutEmailAccessCodeUsed struct {
	DonorFullName                   string
	DonorFullNamePossessive         string
	LpaType                         string
	CertificateProviderFullName     string
	InvitedDate                     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
	CertificateProviderOptOutURL    string
}

func (e AdviseCertificateProviderToSignOrOptOutEmailAccessCodeUsed) emailID(_ localize.Lang) string {
	return "c9ace5c4-01fc-43f3-998d-4df93231d32b"
}

type InformDonorCertificateProviderHasNotActedEmail struct {
	Greeting                        string
	CertificateProviderFullName     string
	LpaType                         string
	LpaReferenceNumber              string
	InvitedDate                     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
}

func (e InformDonorCertificateProviderHasNotActedEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "1b2c1c38-f2d6-4a61-88b2-5e28f80f9b6d"
	}

	return "b45e8f81-22da-45fa-a7ea-99430c749b61"
}

type AdviseCertificateProviderToConfirmIdentityEmail struct {
	DonorFullName                   string
	DonorFullNamePossessive         string
	LpaType                         string
	CertificateProviderFullName     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
}

func (e AdviseCertificateProviderToConfirmIdentityEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "75e8cad3-7e56-496d-aa2e-d5fc84834785"
	}

	return "9691f47f-b4eb-4089-8908-d921c0442781"
}

type InformDonorCertificateProviderHasNotConfirmedIdentityEmail struct {
	Greeting                        string
	LpaType                         string
	LpaReferenceNumber              string
	CertificateProviderFullName     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
}

func (e InformDonorCertificateProviderHasNotConfirmedIdentityEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "07e78d56-f2d4-45d7-8f2b-64a74ac704e2"
	}

	return "03f1d38a-6ab6-4d43-85fe-9d1fd00a9550"
}

type InformDonorAttorneyHasNotActedEmail struct {
	Greeting             string
	AttorneyFullName     string
	LpaType              string
	LpaReferenceNumber   string
	AttorneyStartPageURL string
	DeadlineDate         string
}

func (e InformDonorAttorneyHasNotActedEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "83317256-fa2a-4dd8-b8dc-64501d2b221c"
	}

	return "83f36e64-adb6-483c-ba60-cb70581af84d"
}

type InformDonorPaperAttorneyHasNotActedEmail struct {
	Greeting         string
	AttorneyFullName string
	LpaType          string
	DeadlineDate     string
	PostedDate       string
}

func (e InformDonorPaperAttorneyHasNotActedEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "2ae31d91-b10e-43fd-9732-2e274d8d83dd"
	}

	return "81c6c4b1-2d7c-4b74-b08c-1e646146a5bb"
}

type AdviseAttorneyToSignOrOptOutEmail struct {
	DonorFullName           string
	DonorFullNamePossessive string
	LpaType                 string
	AttorneyFullName        string
	InvitedDate             string
	DeadlineDate            string
	AttorneyStartPageURL    string
	AttorneyOptOutURL       string
	AccessCode              string
}

func (e AdviseAttorneyToSignOrOptOutEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "4c0e65c1-e490-475c-aa8e-a4c693864b7c"
	}

	return "1cef45e2-991c-4998-89d4-1f324a45bb25"
}

type DigitalDonorLpaSubmittedEmail struct {
	Greeting           string
	LpaType            string
	LpaReferenceNumber string
}

func (e DigitalDonorLpaSubmittedEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "5d1a3651-07f2-46e4-96f7-fd1cf2253f6d"
	}

	return "ce8a6d18-05ce-4028-8449-29c09bd1f958"
}

type DigitalDonorCertificateProvidedEmail struct {
	Greeting                    string
	CertificateProviderFullName string
	LpaType                     string
	LpaReferenceNumber          string
}

func (e DigitalDonorCertificateProvidedEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "2854a288-0807-4843-85f5-3f435ae660e3"
	}

	return "9db20b35-8728-434b-8090-52cf704bc5a9"
}

type InformDonorPaperCertificateProviderHasNotActedEmail struct {
	Greeting                    string
	CertificateProviderFullName string
	LpaType                     string
	PostedDate                  string
	DeadlineDate                string
}

func (e InformDonorPaperCertificateProviderHasNotActedEmail) emailID(_ localize.Lang) string {
	return "4ef85faf-68b6-47ba-84ba-3296910640c5"
}

type InformDonorPaperCertificateProviderHasNotConfirmedIdentityEmail struct {
	Greeting                    string
	CertificateProviderFullName string
	LpaType                     string
	LpaReferenceNumber          string
	PostedDate                  string
	DeadlineDate                string
}

func (e InformDonorPaperCertificateProviderHasNotConfirmedIdentityEmail) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "53cda24c-13d2-40a3-aefe-a2eab33f599c"
	}

	return "c43c88f5-c7a2-4a7b-abb4-90718388d3aa"
}

type VoucherLpaDeleted struct {
	DonorFullName           string
	DonorFullNamePossessive string
	InvitedDate             string
	LpaType                 string
	VoucherFullName         string
}

func (e VoucherLpaDeleted) emailID(_ localize.Lang) string {
	return "95cfb64d-6548-4319-b886-63abf8a79259"
}

type VoucherLpaRevoked struct {
	DonorFullName           string
	DonorFullNamePossessive string
	InvitedDate             string
	LpaType                 string
	VoucherFullName         string
}

func (e VoucherLpaRevoked) emailID(_ localize.Lang) string {
	return "b6c87143-4cd2-41bf-b7a1-06fd0eb950c0"
}

type AttorneyLpaRevoked struct {
	AttorneyFullName        string
	DonorFullName           string
	DonorFullNamePossessive string
	InvitedDate             string
	LpaType                 string
	AttorneyStartPageURL    string
}

func (e AttorneyLpaRevoked) emailID(_ localize.Lang) string {
	return "f248eeb7-9be5-43c9-87c1-e40710c10f10"
}

type InformCertificateProviderLPAHasBeenDeleted struct {
	DonorFullName                   string
	DonorFullNamePossessive         string
	LpaType                         string
	CertificateProviderFullName     string
	InvitedDate                     string
	CertificateProviderStartPageURL string
}

func (e InformCertificateProviderLPAHasBeenDeleted) emailID(_ localize.Lang) string {
	return "842ee10d-bf79-42e7-a500-dcbd03d69194"
}

type InformCertificateProviderLPAHasBeenRevoked struct {
	DonorFullName                   string
	DonorFullNamePossessive         string
	LpaType                         string
	CertificateProviderFullName     string
	InvitedDate                     string
	CertificateProviderStartPageURL string
}

func (e InformCertificateProviderLPAHasBeenRevoked) emailID(_ localize.Lang) string {
	return "d565b694-caea-42ff-9f9c-19ec7b79e229"
}

type InformDonorPaperCertificateProviderIdentityCheckFailed struct {
	Greeting                    string
	CertificateProviderFullName string
	LpaType                     string
	DonorStartPageURL           string
}

func (e InformDonorPaperCertificateProviderIdentityCheckFailed) emailID(lang localize.Lang) string {
	if lang.IsCy() {
		return "4ae408c2-51c0-4c5c-83ee-f850af50ca64"
	}

	return "ee349ded-8cfb-4a28-beac-5a4fb90aa823"
}

type CorrespondentInformedVouchingInProgress struct {
	CorrespondentFullName   string
	DonorFullName           string
	DonorFullNamePossessive string
	LpaType                 string
}

func (e CorrespondentInformedVouchingInProgress) emailID(_ localize.Lang) string {
	return "6ece1746-e263-4135-85d7-04c6a598ecf9"
}

type CertificateProviderRemoved struct {
	DonorFullName                  string
	CertificateProviderFullName    string
	CertificateProviderInvitedDate string
	LpaType                        string
	LpaUID                         string
	CertificateProviderStartURL    string
}

func (e CertificateProviderRemoved) emailID(_ localize.Lang) string {
	return "1ecd2e11-bcb3-41ae-a14a-fafec8781b32"
}
