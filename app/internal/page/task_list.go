package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type taskListData struct {
	App      AppData
	Sections []taskListSection
}

type taskListItem struct {
	Name      string
	Path      string
	Completed bool
	Disabled  bool
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &taskListData{
			App: appData,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{Name: "provideDonorDetails", Path: yourDetailsPath, Completed: lpa.You.Address.Line1 != ""},
						{Name: "chooseYourAttorneys", Path: chooseAttorneysPath, Completed: lpa.Attorney.Address.Line1 != ""},
						{Name: "chooseYourReplacementAttorneys"},
						{Name: "chooseWhenTheLpaCanBeUsed"},
						{Name: "addRestrictionsToTheLpa"},
						{Name: "chooseYourCertificateProvider"},
						{Name: "checkAndSendToYourCertificateProvider"},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{Name: "payForTheLpa"},
					},
				},
				{
					Heading: "confirmYourIdentity",
					Items: []taskListItem{
						{Name: "confirmYourIdentity"},
					},
				},
				{
					Heading: "signAndRegisterTheLpa",
					Items: []taskListItem{
						{Name: "signTheLpa", Disabled: true},
					},
				},
			},
		}

		return tmpl(w, data)
	}
}
