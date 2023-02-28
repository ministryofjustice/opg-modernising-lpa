package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetEnterYourName(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	lpa := &page.Lpa{
		CertificateProvider: actor.CertificateProvider{FirstNames: "Bob", LastName: "Smith"},
	}

	data := checkYourNameData{
		App:  testAppData,
		Form: &checkYourNameForm{},
		Lpa:  lpa,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	err := CheckYourName(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
