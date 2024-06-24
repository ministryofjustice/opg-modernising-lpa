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
