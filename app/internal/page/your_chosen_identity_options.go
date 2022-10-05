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
	Yoti:                     "theYoti",
	Passport:                 "aPassport",
	DrivingLicence:           "aDrivingLicence",
	GovernmentGatewayAccount: "aGovernmentGatewayAccount",
	DwpAccount:               "aDwpAccount",
	OnlineBankAccount:        "anOnlineBankAccount",
	UtilityBill:              "aUtilityBill",
	CouncilTaxBill:           "aCouncilTaxBill",
}

var identityOptionLabels = map[IdentityOption]string{
	Yoti:                     "yoti",
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
			appData.Lang.Redirect(w, r, identityOptionRedirectPath, http.StatusFound)
			return nil
		}

		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &yourChosenIdentityOptionsData{
			App:           appData,
			ArticleLabels: identityOptionArticleLabels,
			Labels:        identityOptionLabels,
			Selected:      lpa.IdentityOptions.Selected,
			FirstChoice:   lpa.IdentityOptions.First,
			SecondChoice:  lpa.IdentityOptions.Second,
			You:           lpa.You,
		}

		return tmpl(w, data)
	}
}
