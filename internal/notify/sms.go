package notify

type SMS interface {
	smsID(bool) string
}

type CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS struct {
	DonorFullNamePossessive string
	LpaType                 string
	LpaUID                  string
	DonorFirstNames         string
}

func (s CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS) smsID(isProduction bool) string {
	if isProduction {
		return "28873afc-f019-48c1-bd25-df88c27813e0"
	}

	return "bcdc85a7-32b1-40a6-a61f-a552406e6ecc"
}

type CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS struct {
	DonorFullName string
	LpaType       string
}

func (s CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS) smsID(isProduction bool) string {
	if isProduction {
		return "796990f2-cf49-48a4-9f04-fc12f4a9702b"
	}

	return "292cc508-811e-44fa-9962-3fdf10e2e8cd"
}

type CertificateProviderActingOnPaperDetailsChangedSMS struct {
	DonorFullName   string
	LpaUID          string
	DonorFirstNames string
}

func (s CertificateProviderActingOnPaperDetailsChangedSMS) smsID(isProduction bool) string {
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

func (s CertificateProviderActingOnPaperMeetingPromptSMS) smsID(isProduction bool) string {
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

func (s WitnessCodeSMS) smsID(isProduction bool) string {
	if isProduction {
		return "e39849c0-ecab-4e16-87ec-6b22afb9d535"
	}

	return "dfa15e16-1f23-494a-bffb-a475513df6cc"
}

type VouchingShareCodeSMS struct {
	ShareCode                 string
	DonorFullNamePossessive   string
	LpaType                   string
	VoucherFullName           string
	DonorFirstNamesPossessive string
}

func (s VouchingShareCodeSMS) smsID(isProduction bool) string {
	if isProduction {
		return "4864a99a-40a7-4aef-8c1d-7fdc0a4775b9"
	}

	return "84d70372-5c7a-4a88-a836-ee7c7dea203a"
}
