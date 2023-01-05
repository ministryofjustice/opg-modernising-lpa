package page

type IdentityOption string

const (
	IdentityOptionUnknown    = IdentityOption("")
	OneLogin                 = IdentityOption("one login")
	EasyID                   = IdentityOption("easy id")
	Passport                 = IdentityOption("passport")
	BiometricResidencePermit = IdentityOption("biometric residence permit")
	DrivingLicencePhotocard  = IdentityOption("driving licence photocard")
	DrivingLicencePaper      = IdentityOption("driving licence paper")
	OnlineBankAccount        = IdentityOption("online bank account")
)

func readIdentityOption(s string) IdentityOption {
	switch s {
	case "one login":
		return OneLogin
	case "easy id":
		return EasyID
	case "passport":
		return Passport
	case "biometric residence permit":
		return BiometricResidencePermit
	case "driving licence photocard":
		return DrivingLicencePhotocard
	case "driving licence paper":
		return DrivingLicencePaper
	case "online bank account":
		return OnlineBankAccount
	default:
		return IdentityOptionUnknown
	}
}

func (o IdentityOption) ArticleLabel() string {
	switch o {
	case OneLogin:
		return "yourOneLogin"
	case EasyID:
		return "postOfficeEasyID"
	case Passport:
		return "passport"
	case BiometricResidencePermit:
		return "biometricResidencePermit"
	case DrivingLicencePhotocard:
		return "drivingLicencePhotocard"
	case DrivingLicencePaper:
		return "drivingLicencePaper"
	case OnlineBankAccount:
		return "yourBankAccount"
	default:
		return ""
	}
}
