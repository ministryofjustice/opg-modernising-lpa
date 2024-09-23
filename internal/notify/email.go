package notify

type Email interface {
	emailID(bool) string
}

type InitialOriginalAttorneyEmail struct {
	DonorFullName             string
	LpaType                   string
	AttorneyFullName          string
	DonorFirstNames           string
	AttorneyStartPageURL      string
	ShareCode                 string
	DonorFirstNamesPossessive string
	AttorneyOptOutURL         string
}

func (e InitialOriginalAttorneyEmail) emailID(isProduction bool) string {
	if isProduction {
		return "080071dc-0434-4b13-adb7-c4e5612c4b47"
	}

	return "376d7ef2-7941-46c2-b372-bacca0e00c1d"
}

type InitialReplacementAttorneyEmail struct {
	DonorFullName             string
	LpaType                   string
	AttorneyFullName          string
	DonorFirstNames           string
	AttorneyStartPageURL      string
	ShareCode                 string
	DonorFirstNamesPossessive string
	AttorneyOptOutURL         string
}

func (e InitialReplacementAttorneyEmail) emailID(isProduction bool) string {
	if isProduction {
		return "8d335239-7002-4825-8393-cc00ad246648"
	}

	return "738d500f-b674-4e1e-8039-a7be53fce528"
}

type CertificateProviderCertificateProvidedEmail struct {
	DonorFullNamePossessive     string
	LpaType                     string
	CertificateProviderFullName string
	CertificateProvidedDateTime string
	DonorFirstNamesPossessive   string
}

func (e CertificateProviderCertificateProvidedEmail) emailID(isProduction bool) string {
	if isProduction {
		return "64d7d56b-966b-464f-8084-1ac5d91c3d58"
	}

	return "76f4370f-1a78-4488-9029-b00fbc292386"
}

type CertificateProviderInviteEmail struct {
	DonorFullName                string
	LpaType                      string
	CertificateProviderFullName  string
	DonorFirstNames              string
	DonorFirstNamesPossessive    string
	WhatLpaCovers                string
	CertificateProviderStartURL  string
	ShareCode                    string
	CertificateProviderOptOutURL string
}

func (e CertificateProviderInviteEmail) emailID(isProduction bool) string {
	if isProduction {
		return "08a8d89d-e5b7-4bb9-94d2-25139543e962"
	}

	return "5b4cb108-4eb1-479a-a83f-87f36799c284"
}

type CertificateProviderProvideCertificatePromptEmail struct {
	DonorFullName               string
	DonorFullNamePossessive     string
	LpaType                     string
	CertificateProviderFullName string
	CertificateProviderStartURL string
	ShareCode                   string
}

func (e CertificateProviderProvideCertificatePromptEmail) emailID(isProduction bool) string {
	if isProduction {
		return "eac04624-f058-411a-be48-854a77022ac8"
	}

	return "3ad5a806-6789-4687-8731-49ff7357372f"
}

type OrganisationMemberInviteEmail struct {
	OrganisationName      string
	InviterFullName       string
	InviterEmail          string
	InviteCode            string
	JoinAnOrganisationURL string
}

func (e OrganisationMemberInviteEmail) emailID(isProduction bool) string {
	if isProduction {
		return "8433502f-7cbd-42de-a075-7f9343531167"
	}

	return "eac6a25f-3055-4b72-be19-6067398551db"
}

type DonorAccessEmail struct {
	SupporterFullName string
	OrganisationName  string
	LpaType           string
	DonorName         string
	URL               string
	ShareCode         string
}

func (e DonorAccessEmail) emailID(isProduction bool) string {
	if isProduction {
		return "4e7337cd-34aa-41ba-81e3-3c866e3daf4b"
	}

	return "0d762056-570b-4fca-9871-1f6a69f9da47"
}

type CertificateProviderOptedOutPreWitnessingEmail struct {
	Greeting                    string
	CertificateProviderFullName string
	DonorFullName               string
	LpaType                     string
	LpaUID                      string
	DonorStartPageURL           string
}

func (e CertificateProviderOptedOutPreWitnessingEmail) emailID(isProduction bool) string {
	if isProduction {
		return "1e85965d-4288-42ea-bdd5-f4a29020cf73"
	}

	return "06691e59-899a-4b06-8337-68e4c93d5e29"
}

type CertificateProviderOptedOutPostWitnessingEmail struct {
	Greeting                      string
	CertificateProviderFirstNames string
	CertificateProviderFullName   string
	DonorFullName                 string
	LpaType                       string
	LpaUID                        string
	DonorStartPageURL             string
}

func (e CertificateProviderOptedOutPostWitnessingEmail) emailID(isProduction bool) string {
	if isProduction {
		return "e284f26e-600a-44f8-b76a-95b93339a054"
	}

	return "654332f4-4e53-4fa1-91d0-f480b577b3d9"
}

type CertificateProviderFailedIDCheckEmail struct {
	Greeting                    string
	DonorFullName               string
	CertificateProviderFullName string
	LpaType                     string
	DonorStartPageURL           string
}

func (e CertificateProviderFailedIDCheckEmail) emailID(isProduction bool) string {
	if isProduction {
		return "4020a281-8b64-45ec-85c6-19a89c08bcdb"
	}

	return "26d337be-eef3-405f-96ed-cb2ed76002b3"
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

func (e PaymentConfirmationEmail) emailID(isProduction bool) string {
	if isProduction {
		return "d0946a7d-d7fe-47cb-9b41-464f13727bf3"
	}

	return "ff757818-f066-4605-8751-af481afe8a2b"
}

type AttorneyOptedOutEmail struct {
	Greeting          string
	AttorneyFullName  string
	DonorFullName     string
	LpaType           string
	LpaUID            string
	DonorStartPageURL string
}

func (e AttorneyOptedOutEmail) emailID(isProduction bool) string {
	return "TODO"
}

type VoucherFailedIdentityCheckEmail struct {
	Greeting          string
	DonorFullName     string
	VoucherFullName   string
	LpaType           string
	DonorStartPageURL string
}

func (e VoucherFailedIdentityCheckEmail) emailID(isProduction bool) string {
	return "TODO"
}

type DonorIdentityCheckExpiredEmail struct{}

func (e DonorIdentityCheckExpiredEmail) emailID(isProduction bool) string {
	return "TODO"
}

type VouchingShareCodeEmail struct {
	ShareCode       string
	VoucherFullName string
	DonorFullName   string
	LpaType         string
}

func (s VouchingShareCodeEmail) emailID(isProduction bool) string {
	if isProduction {
		return "38e26a3f-d87d-4b0c-8985-8fb5bed79466"
	}

	return "881e25c4-4898-4525-bca3-722f51c5d6ee"
}

type VoucherInviteEmail struct {
	VoucherFullName           string
	DonorFullName             string
	DonorFirstNamesPossessive string
	DonorFirstNames           string
	LpaType                   string
	VoucherStartPageURL       string
}

func (s VoucherInviteEmail) emailID(isProduction bool) string {
	if isProduction {
		return "36ad56ad-823b-4852-88a7-8acc4dfd1749"
	}

	return "9af150b5-d9cd-4702-bf97-d3e6bfe81eec"
}

type VoucherFirstFailedVouchAttempt struct {
	Greeting          string
	VoucherFullName   string
	DonorStartPageURL string
}

func (e VoucherFirstFailedVouchAttempt) emailID(isProduction bool) string {
	if isProduction {
		return "f21ee857-8c3e-43ee-adf2-2d9f1ff1a1a8"
	}

	return "584412e6-f235-4227-aff9-6cb56ba48e31"
}

type VoucherSecondFailedVouchAttempt struct {
	Greeting          string
	VoucherFullName   string
	DonorStartPageURL string
}

func (e VoucherSecondFailedVouchAttempt) emailID(isProduction bool) string {
	if isProduction {
		return "44ffb252-ce34-4164-baa5-8b21036625ac"
	}

	return "3af1b3c4-35ce-4a23-abd2-bd0d019985c2"
}
