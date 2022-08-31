package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApp(t *testing.T) {
	app := App(&mockLogger{}, localize.Localizer{}, En, template.Templates{}, nil)

	assert.Implements(t, (*http.Handler)(nil), app)
}

func TestLangRedirect(t *testing.T) {
	testCases := map[Lang]string{
		En: "/somewhere",
		Cy: "/cy/somewhere",
	}

	for lang, url := range testCases {
		t.Run("En", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			lang.Redirect(w, r, "/somewhere", http.StatusFound)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestFakeAddressClient(t *testing.T) {
	addresses, _ := fakeAddressClient{}.LookupPostcode("xyz")

	assert.Equal(t, []Address{
		{Line1: "123 Fake Street", TownOrCity: "Someville", Postcode: "xyz"},
		{Line1: "456 Fake Street", TownOrCity: "Someville", Postcode: "xyz"},
	}, addresses)
}

func TestFakeDataStore(t *testing.T) {
	logger := &mockLogger{}
	logger.
		On("Print", "null")

	fakeDataStore{logger: logger}.Save(nil)

	mock.AssertExpectationsForObjects(t, logger)
}
