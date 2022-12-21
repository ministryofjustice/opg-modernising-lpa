package page

import (
	"sort"

	"golang.org/x/exp/slices"
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

type IdentityOptions struct {
	Selected []IdentityOption
	First    IdentityOption
	Second   IdentityOption
}

func (o IdentityOptions) NextPath(current IdentityOption, paths AppPaths) string {
	identityOptionPaths := map[IdentityOption]string{
		Yoti:                     paths.IdentityWithYoti,
		Passport:                 paths.IdentityWithPassport,
		DrivingLicence:           paths.IdentityWithDrivingLicence,
		GovernmentGatewayAccount: paths.IdentityWithGovernmentGatewayAccount,
		DwpAccount:               paths.IdentityWithDwpAccount,
		OnlineBankAccount:        paths.IdentityWithOnlineBankAccount,
		UtilityBill:              paths.IdentityWithUtilityBill,
		CouncilTaxBill:           paths.IdentityWithCouncilTaxBill,
	}

	if current == o.Second {
		return paths.WitnessingYourSignature
	}

	if current == o.First {
		return identityOptionPaths[o.Second]
	}

	return identityOptionPaths[o.First]
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
