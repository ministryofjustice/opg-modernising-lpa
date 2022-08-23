package page

import (
	"net/http"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	app := App(&mockLogger{}, localize.Localizer{}, En, template.Templates{})

	assert.Implements(t, (*http.Handler)(nil), app)
}
