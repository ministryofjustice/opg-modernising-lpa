package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type yourChosenIdentityOptionsData struct {
	App                   AppData
	Errors                map[string]string
	ArticleLabels, Labels map[IdentityOption]string
	Selected              []IdentityOption
	FirstChoice           IdentityOption
	SecondChoice          IdentityOption
	You                   Person
}

var identityOptionArticleLabels = map[IdentityOption]string{
	Passport:                 "aPassport",
	DrivingLicence:           "aDrivingLicence",
	GovernmentGatewayAccount: "aGovernmentGatewayAccount",
	DwpAccount:               "aDwpAccount",
	OnlineBankAccount:        "anOnlineBankAccount",
	UtilityBill:              "aUtilityBill",
	CouncilTaxBill:           "aCouncilTaxBill",
}

var identityOptionLabels = map[IdentityOption]string{
	Passport:                 "passport",
	DrivingLicence:           "drivingLicence",
	GovernmentGatewayAccount: "governmentGatewayAccount",
	DwpAccount:               "dwpAccount",
	OnlineBankAccount:        "onlineBankAccount",
	UtilityBill:              "utilityBill",
	CouncilTaxBill:           "councilTaxBill",
}

func YourChosenIdentityOptions(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			// will redirect to the correct ID method, for now just go to EasyID flow
			appData.Lang.Redirect(w, r, identityWithEasyIDPath, http.StatusFound)
			return nil
		}

		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		firstChoice, secondChoice := identityOptionsRanked(lpa.IdentityOptions)

		data := &yourChosenIdentityOptionsData{
			App:           appData,
			ArticleLabels: identityOptionArticleLabels,
			Labels:        identityOptionLabels,
			Selected:      lpa.IdentityOptions,
			FirstChoice:   firstChoice,
			SecondChoice:  secondChoice,
			You:           lpa.You,
		}

		return tmpl(w, data)
	}
}
