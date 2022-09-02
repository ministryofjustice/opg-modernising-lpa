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

func TaskList(logger Logger, tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) {
		data := &taskListData{
			App: appData,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{Name: "provideDonorDetails", Path: donorDetailsPath, Completed: true},
						{Name: "chooseYourContactPreferences", Path: howWouldYouLikeToBeContactedPath, Completed: true},
						{Name: "chooseYourAttorneys"},
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

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}
