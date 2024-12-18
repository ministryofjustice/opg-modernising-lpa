package notify

import "github.com/ministryofjustice/opg-modernising-lpa/internal/localize"

type Email interface {
	emailID(bool, localize.Lang) string
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

func (e InitialOriginalAttorneyEmail) emailID(isProduction bool, _ localize.Lang) string {
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

func (e InitialReplacementAttorneyEmail) emailID(isProduction bool, _ localize.Lang) string {
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

func (e CertificateProviderCertificateProvidedEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "3a52508e-b8f1-4192-b9f4-e912964db3e7"
		}

		return "64d7d56b-966b-464f-8084-1ac5d91c3d58"
	}

	if lang.IsCy() {
		return "ef87ab25-1d4a-4f2d-aaa6-8ef3200c6643"
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

func (e CertificateProviderInviteEmail) emailID(isProduction bool, _ localize.Lang) string {
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

func (e CertificateProviderProvideCertificatePromptEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "73704800-f241-4b17-8e37-09ee2804a570"
		}

		return "eac04624-f058-411a-be48-854a77022ac8"
	}

	if lang.IsCy() {
		return "675b29ff-63ec-4257-923d-8fcd0db057f7"
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

func (e OrganisationMemberInviteEmail) emailID(isProduction bool, _ localize.Lang) string {
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

func (e DonorAccessEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "cd0f3029-6d6e-4c00-9098-e20394982dc6"
		}

		return "4e7337cd-34aa-41ba-81e3-3c866e3daf4b"
	}

	if lang.IsCy() {
		return "12c7981d-9db2-4b5c-8922-0d437cf997e1"
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

func (e CertificateProviderOptedOutPreWitnessingEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "28a72a2e-7dd7-4131-9ac1-616e8f453175"
		}

		return "1e85965d-4288-42ea-bdd5-f4a29020cf73"
	}

	if lang.IsCy() {
		return "3ad13ea7-aed6-4814-90b1-186ac3b78a14"
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

func (e CertificateProviderOptedOutPostWitnessingEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "fb22a2fa-b3eb-42b1-9884-fb9398308bc4"
		}

		return "e284f26e-600a-44f8-b76a-95b93339a054"
	}

	if lang.IsCy() {
		return "ce8f18dd-4edf-4289-b98d-a016283217fc"
	}

	return "654332f4-4e53-4fa1-91d0-f480b577b3d9"
}

type CertificateProviderFailedIdentityCheckEmail struct {
	Greeting                    string
	DonorFullName               string
	CertificateProviderFullName string
	LpaType                     string
	DonorStartPageURL           string
}

func (e CertificateProviderFailedIdentityCheckEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "c29f3c42-0cbf-42d0-9c5c-3ceea2095b07"
		}

		return "4020a281-8b64-45ec-85c6-19a89c08bcdb"
	}

	if lang.IsCy() {
		return "de242ce4-ff15-4467-a207-dd2d8f7e2ae5"
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

func (e PaymentConfirmationEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "b1d0a9d7-0886-4d5d-acc0-08fe69e492ad"
		}

		return "d0946a7d-d7fe-47cb-9b41-464f13727bf3"
	}

	if lang.IsCy() {
		return "662548fd-4382-42f3-88eb-7426bd868a23"
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

func (e AttorneyOptedOutEmail) emailID(isProduction bool, _ localize.Lang) string {
	return "TODO"
}

type DonorIdentityCheckExpiredEmail struct{}

func (e DonorIdentityCheckExpiredEmail) emailID(isProduction bool, _ localize.Lang) string {
	if isProduction {
		return "c3c4a115-4d07-4e25-926d-a656dc33485a"
	}

	return "26509ca9-83d0-4417-ab5d-a3844916519e"
}

type VouchingShareCodeEmail struct {
	ShareCode       string
	VoucherFullName string
	DonorFullName   string
	LpaType         string
}

func (s VouchingShareCodeEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "93ee9148-b962-4398-ab94-d2625e39fbb1"
		}

		return "38e26a3f-d87d-4b0c-8985-8fb5bed79466"
	}

	if lang.IsCy() {
		return "a5d9bc34-a7f4-44c4-8473-376845c5b0b9"
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

func (s VoucherInviteEmail) emailID(isProduction bool, _ localize.Lang) string {
	if isProduction {
		return "36ad56ad-823b-4852-88a7-8acc4dfd1749"
	}

	return "9af150b5-d9cd-4702-bf97-d3e6bfe81eec"
}

type VouchingFailedAttemptEmail struct {
	Greeting          string
	VoucherFullName   string
	DonorStartPageURL string
}

func (e VouchingFailedAttemptEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "db45c036-1379-45d1-a521-919217f50e45"
		}

		return "f21ee857-8c3e-43ee-adf2-2d9f1ff1a1a8"
	}

	if lang.IsCy() {
		return "6f647fa8-587b-4f49-be75-d85a44f167b2"
	}

	return "584412e6-f235-4227-aff9-6cb56ba48e31"
}

type VoucherHasConfirmedDonorIdentityEmail struct {
	VoucherFullName   string
	DonorFullName     string
	DonorStartPageURL string
}

func (e VoucherHasConfirmedDonorIdentityEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "d1c1bb6f-e9eb-44d6-802b-49590fb0d0fa"
		}

		return "67cd151e-6e7b-4fba-9457-f0252e75dfe2"
	}

	if lang.IsCy() {
		return "d81b27c2-07e3-47c3-baa6-a8114673c32d"
	}

	return "86e6d479-6bef-428c-a09d-02a325b97972"
}

type VoucherHasConfirmedDonorIdentityOnSignedLpaEmail struct {
	VoucherFullName   string
	DonorFullName     string
	DonorStartPageURL string
}

func (e VoucherHasConfirmedDonorIdentityOnSignedLpaEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "3305cfe2-5622-4292-9838-78d6c152db23"
		}

		return "8df993ff-e4d9-43f2-b714-39053510c664"
	}

	if lang.IsCy() {
		return "3884d77c-09c2-4396-b6df-cd32e76a4bb0"
	}

	return "efa0ef78-9e65-4edf-88c8-70d3da7a4b0e"
}

type VoucherInformedTheyAreNoLongerNeededToVouchEmail struct {
	VoucherFullName string
	DonorFullName   string
}

func (e VoucherInformedTheyAreNoLongerNeededToVouchEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "TODO"
		}

		return "ca7c6a15-bdf3-47fe-ba01-d811ccdbc30d"
	}

	if lang.IsCy() {
		return "TODO"
	}

	return "00ad14c6-f6df-4d7f-ae44-d7e27f6a9187"
}

type AdviseCertificateProviderToSignOrOptOutEmail struct {
	DonorFullName                   string
	DonorFullNamePossessive         string
	LpaType                         string
	CertificateProviderFullName     string
	InvitedDate                     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
}

func (e AdviseCertificateProviderToSignOrOptOutEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "083ace46-2e6d-41c3-9e82-84df8bc03faf"
		}

		return "fc01c541-28f3-4e04-921e-e8848f810278"
	}

	if lang.IsCy() {
		return "22a19484-cd44-4476-a7b6-7826af5932ae"
	}

	return "d9b3e36a-5814-4e6b-84b1-baf763c49220"
}

type InformDonorCertificateProviderHasNotActedEmail struct {
	Greeting                        string
	CertificateProviderFullName     string
	LpaType                         string
	InvitedDate                     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
}

func (e InformDonorCertificateProviderHasNotActedEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "688d8284-ae13-4e87-97be-3d6b21767755"
		}

		return "b45e8f81-22da-45fa-a7ea-99430c749b61"
	}

	if lang.IsCy() {
		return "4fc578f0-5cce-4082-a926-957aebb824bd"
	}

	return "0f7cbfed-1ffa-43d7-92c0-8d162aadc0ea"
}

type AdviseCertificateProviderToConfirmIdentityEmail struct {
	DonorFullName                   string
	DonorFullNamePossessive         string
	LpaType                         string
	CertificateProviderFullName     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
}

func (e AdviseCertificateProviderToConfirmIdentityEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "75e8cad3-7e56-496d-aa2e-d5fc84834785"
		}

		return "fc01c541-28f3-4e04-921e-e8848f810278"
	}

	if lang.IsCy() {
		return "2ad9669e-72ad-486f-bf8f-0a422870e6ee"
	}

	return "5e4b67ce-4175-4d5d-baf9-1c81a1ebc213"
}

type InformDonorCertificateProviderHasNotConfirmedIdentityEmail struct {
	Greeting                        string
	LpaType                         string
	CertificateProviderFullName     string
	DeadlineDate                    string
	CertificateProviderStartPageURL string
}

func (e InformDonorCertificateProviderHasNotConfirmedIdentityEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "07e78d56-f2d4-45d7-8f2b-64a74ac704e2"
		}

		return "03f1d38a-6ab6-4d43-85fe-9d1fd00a9550"
	}

	if lang.IsCy() {
		return "36e61fbc-df51-4bf5-bc1f-877ec559de8f"
	}

	return "3a6bf17f-f690-4ee6-b815-b5bfe2f70c55"
}

type InformDonorAttorneyHasNotActedEmail struct {
	Greeting             string
	AttorneyFullName     string
	LpaType              string
	AttorneyStartPageURL string
	DeadlineDate         string
	InvitedDate          string
}

func (e InformDonorAttorneyHasNotActedEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "83317256-fa2a-4dd8-b8dc-64501d2b221c"
		}

		return "83f36e64-adb6-483c-ba60-cb70581af84d"
	}

	if lang.IsCy() {
		return "2ade25e7-5864-45dc-953a-22b2d956f9b5"
	}

	return "efc93b6f-d2f3-487d-afef-c6961a0abaed"
}

type AdviseAttorneyToSignOrOptOutEmail struct {
	DonorFullName           string
	DonorFullNamePossessive string
	LpaType                 string
	AttorneyFullName        string
	InvitedDate             string
	DeadlineDate            string
	AttorneyStartPageURL    string
}

func (e AdviseAttorneyToSignOrOptOutEmail) emailID(isProduction bool, lang localize.Lang) string {
	if isProduction {
		if lang.IsCy() {
			return "4c0e65c1-e490-475c-aa8e-a4c693864b7c"
		}

		return "1cef45e2-991c-4998-89d4-1f324a45bb25"
	}

	if lang.IsCy() {
		return "9df92f3d-4070-4000-bad2-c25ca9daa68e"
	}

	return "3ddfd30a-02b6-4625-8fbf-5785f5b33864"
}
