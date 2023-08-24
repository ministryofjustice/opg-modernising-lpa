package identity

type Option string

const (
	UnknownOption            = Option("")
	OneLogin                 = Option("one login")
	EasyID                   = Option("easy id")
	Passport                 = Option("passport")
	BiometricResidencePermit = Option("biometric residence permit")
	DrivingLicencePhotocard  = Option("driving licence photocard")
	DrivingLicencePaper      = Option("driving licence paper")
	OnlineBankAccount        = Option("online bank account")
)

func ReadOption(s string) Option {
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
		return UnknownOption
	}
}

func (o Option) ArticleLabel() string {
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
