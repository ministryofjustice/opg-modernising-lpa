package donorpage

import (
	html "html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/templatefn"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReadYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{Donor: donordata.Donor{LpaLanguagePreference: localize.En}}

	bundle, err := localize.NewBundle("../../../lang/en.json")
	if err != nil {
		t.Fatal(err)
	}

	localizer := bundle.For(localize.En)

	layouts, err := parseLayoutTemplates("../../../web/template/layout", templatefn.All(&templatefn.Globals{
		ActorTypes: actor.TypeValues,
	}))
	if err != nil {
		t.Fatal(err)
	}

	tmpls, err := parseTemplates("../../../web/template/donor", layouts)
	if err != nil {
		t.Fatal(err)
	}

	testAppData.Localizer = localizer

	err = ReadYourLpa(tmpls.Get("read_your_lpa.gohtml"), bundle)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	testhelper.RenderHTMLWithCSS(t, resp)
}

func TestReadYourLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
		Return(newMockLocalizer(t))

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ReadYourLpa(template.Execute, bundle)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func parseLayoutTemplates(layoutDir string, funcs html.FuncMap) (*html.Template, error) {
	return html.New("").Funcs(funcs).ParseGlob(filepath.Join(layoutDir, "*.*"))
}

func parseTemplates(templateDir string, layouts *html.Template) (template.Templates, error) {
	files, err := filepath.Glob(filepath.Join(templateDir, "*.*"))
	if err != nil {
		return nil, err
	}

	tmpls := map[string]*html.Template{}
	for _, file := range files {
		clone, err := layouts.Clone()
		if err != nil {
			return nil, err
		}

		tmpl, err := clone.ParseFiles(file)
		if err != nil {
			return nil, err
		}

		tmpls[filepath.Base(file)] = tmpl
	}

	return tmpls, nil
}
