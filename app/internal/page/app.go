package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type Lang int

const (
	En Lang = iota
	Cy
)

func App(logger *logging.Logger, localizer localize.Localizer, lang Lang, tmpls template.Templates) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", Start(logger, localizer, lang, tmpls.Get("start.gohtml")))

	return mux
}
