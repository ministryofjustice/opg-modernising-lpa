package page

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"golang.org/x/exp/slices"
)

const (
	PayCookieName              = "pay"
	PayCookiePaymentIdValueKey = "paymentId"
	CostOfLpaPence             = 8200
)

type TaskState int

const (
	TaskNotStarted TaskState = iota
	TaskInProgress
	TaskCompleted
)

type IdentityOption string

func (o IdentityOption) String() string {
	return string(o)
}

const (
	IdentityOptionUnknown    = IdentityOption("")
	Yoti                     = IdentityOption("yoti")
	Passport                 = IdentityOption("passport")
	DrivingLicence           = IdentityOption("driving licence")
	GovernmentGatewayAccount = IdentityOption("government gateway account")
	DwpAccount               = IdentityOption("dwp account")
	OnlineBankAccount        = IdentityOption("online bank account")
	UtilityBill              = IdentityOption("utility bill")
	CouncilTaxBill           = IdentityOption("council tax bill")
)

func readIdentityOption(s string) IdentityOption {
	switch s {
	case "yoti":
		return Yoti
	case "passport":
		return Passport
	case "driving licence":
		return DrivingLicence
	case "government gateway account":
		return GovernmentGatewayAccount
	case "dwp account":
		return DwpAccount
	case "online bank account":
		return OnlineBankAccount
	case "utility bill":
		return UtilityBill
	case "council tax bill":
		return CouncilTaxBill
	default:
		return IdentityOptionUnknown
	}
}

func (o IdentityOption) ArticleLabel() string {
	switch o {
	case Yoti:
		return "theYoti"
	case Passport:
		return "aPassport"
	case DrivingLicence:
		return "aDrivingLicence"
	case GovernmentGatewayAccount:
		return "aGovernmentGatewayAccount"
	case DwpAccount:
		return "aDwpAccount"
	case OnlineBankAccount:
		return "anOnlineBankAccount"
	case UtilityBill:
		return "aUtilityBill"
	case CouncilTaxBill:
		return "aCouncilTaxBill"
	default:
		return ""
	}
}

func (o IdentityOption) Label() string {
	switch o {
	case Yoti:
		return "yoti"
	case Passport:
		return "passport"
	case DrivingLicence:
		return "drivingLicence"
	case GovernmentGatewayAccount:
		return "governmentGatewayAccount"
	case DwpAccount:
		return "dwpAccount"
	case OnlineBankAccount:
		return "onlineBankAccount"
	case UtilityBill:
		return "utilityBill"
	case CouncilTaxBill:
		return "councilTaxBill"
	default:
		return ""
	}
}

type Lpa struct {
	You                      Person
	Attorney                 Attorney
	CertificateProvider      CertificateProvider
	WhoFor                   string
	Contact                  []string
	Type                     string
	WantReplacementAttorneys string
	WhenCanTheLpaBeUsed      string
	Restrictions             string
	Tasks                    Tasks
	Checked                  bool
	HappyToShare             bool
	PaymentDetails           PaymentDetails
	CheckedAgain             bool
	ConfirmFreeWill          bool
	SignatureCode            string
	IdentityOptions          IdentityOptions
	YotiUserData             identity.UserData
}

type IdentityOptions struct {
	Selected []IdentityOption
	First    IdentityOption
	Second   IdentityOption
}

func (o IdentityOptions) NextPath(current IdentityOption) string {
	identityOptionPaths := map[IdentityOption]string{
		Yoti:                     identityWithYotiPath,
		Passport:                 identityWithPassportPath,
		DrivingLicence:           identityWithDrivingLicencePath,
		GovernmentGatewayAccount: identityWithGovernmentGatewayAccountPath,
		DwpAccount:               identityWithDwpAccountPath,
		OnlineBankAccount:        identityWithOnlineBankAccountPath,
		UtilityBill:              identityWithUtilityBillPath,
		CouncilTaxBill:           identityWithCouncilTaxBillPath,
	}

	if current == o.Second {
		return whatHappensWhenSigningPath
	}

	if current == o.First {
		return identityOptionPaths[o.Second]
	}

	return identityOptionPaths[o.First]
}

type PaymentDetails struct {
	PaymentReference string
	PaymentId        string
}

type Tasks struct {
	WhenCanTheLpaBeUsed        TaskState
	Restrictions               TaskState
	CertificateProvider        TaskState
	CheckYourLpa               TaskState
	PayForLpa                  TaskState
	ConfirmYourIdentityAndSign TaskState
}

type Person struct {
	FirstNames  string
	LastName    string
	Email       string
	OtherNames  string
	DateOfBirth time.Time
	Address     Address
}

type Attorney struct {
	FirstNames  string
	LastName    string
	Email       string
	DateOfBirth time.Time
	Address     Address
}

type CertificateProvider struct {
	FirstNames              string
	LastName                string
	Email                   string
	DateOfBirth             time.Time
	Relationship            []string
	RelationshipDescription string
	RelationshipLength      string
}

type Address struct {
	Line1      string
	Line2      string
	TownOrCity string
	Postcode   string
}

type AddressClient interface {
	LookupPostcode(string) ([]Address, error)
}

func (a Address) Encode() string {
	x, _ := json.Marshal(a)
	return string(x)
}

func DecodeAddress(s string) *Address {
	var v Address
	json.Unmarshal([]byte(s), &v)
	return &v
}

func (a Address) String() string {
	var parts []string

	if a.Line1 != "" {
		parts = append(parts, a.Line1)
	}
	if a.Line2 != "" {
		parts = append(parts, a.Line2)
	}
	if a.TownOrCity != "" {
		parts = append(parts, a.TownOrCity)
	}
	if a.Postcode != "" {
		parts = append(parts, a.Postcode)
	}

	return strings.Join(parts, ", ")
}

type Date struct {
	Day   string
	Month string
	Year  string
}

func readDate(t time.Time) Date {
	return Date{
		Day:   t.Format("2"),
		Month: t.Format("1"),
		Year:  t.Format("2006"),
	}
}

type rankedItem struct {
	item    IdentityOption
	rank    int
	subrank int
}

func identityOptionsRanked(options []IdentityOption) (firstChoice, secondChoice IdentityOption) {
	if len(options) == 0 {
		return IdentityOptionUnknown, IdentityOptionUnknown
	}

	table := map[IdentityOption]struct {
		rank    int
		subrank int
		not     []IdentityOption
	}{
		Yoti:                     {rank: 0, subrank: 0, not: []IdentityOption{Passport, DrivingLicence}},
		Passport:                 {rank: 1, subrank: 2, not: []IdentityOption{GovernmentGatewayAccount, OnlineBankAccount}},
		DrivingLicence:           {rank: 2, subrank: 3, not: []IdentityOption{}},
		DwpAccount:               {rank: 3, subrank: 5, not: []IdentityOption{GovernmentGatewayAccount}},
		OnlineBankAccount:        {rank: 4, subrank: 6, not: []IdentityOption{Passport}},
		GovernmentGatewayAccount: {rank: 6, subrank: 4, not: []IdentityOption{DwpAccount}},
		UtilityBill:              {rank: 7, subrank: 7, not: []IdentityOption{CouncilTaxBill}},
		CouncilTaxBill:           {rank: 8, subrank: 8, not: []IdentityOption{UtilityBill}},
	}

	rankedOptions := make([]rankedItem, len(options))
	for i, option := range options {
		rankedOptions[i] = rankedItem{item: option, rank: table[option].rank, subrank: table[option].subrank}
	}

	sort.Slice(rankedOptions, func(i, j int) bool {
		return rankedOptions[i].rank < rankedOptions[j].rank
	})

	firstChoice = rankedOptions[0].item

	var remainingOptions []rankedItem
	for _, option := range rankedOptions {
		if option.item != firstChoice && !slices.Contains(table[firstChoice].not, option.item) {
			remainingOptions = append(remainingOptions, option)
		}
	}

	sort.Slice(remainingOptions, func(i, j int) bool {
		return remainingOptions[i].subrank < remainingOptions[j].subrank
	})

	if len(remainingOptions) == 0 {
		return firstChoice, IdentityOptionUnknown
	}

	secondChoice = remainingOptions[0].item

	return firstChoice, secondChoice
}
