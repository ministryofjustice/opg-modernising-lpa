package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type whatHappensNextData struct {
	App      AppData
	Errors   map[string]string
	Continue string
	Lpa      Lpa
}

func WhatHappensNext(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &whatHappensNextData{
			App:      appData,
			Continue: taskListPath,
		}

		if err := dataStore.Get(r.Context(), appData.SessionID, &data.Lpa); err != nil {
			return err
		}

		return tmpl(w, data)
	}
}
