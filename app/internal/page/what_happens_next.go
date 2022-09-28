package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type whatHappensNextData struct {
	App      AppData
	Errors   map[string]string
	Continue string
}

func WhatHappensNext(tmpl template.Template) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &whatHappensNextData{
			App:      appData,
			Continue: taskListPath,
		}

		return tmpl(w, data)
	}
}
