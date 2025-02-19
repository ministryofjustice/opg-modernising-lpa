package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReadYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{Donor: donordata.Donor{LpaLanguagePreference: localize.Cy}}

	localizer := newMockLocalizer(t)

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.Cy).
		Return(localizer)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &readYourLpaData{
			App: testAppData,
			LpaLanguageApp: appcontext.Data{
				SessionID:         "session-id",
				LpaID:             "lpa-id",
				Lang:              localize.Cy,
				Localizer:         localizer,
				LoginSessionEmail: "logged-in@example.com",
			},
			Donor: donor,
		}).
		Return(nil)

	err := ReadYourLpa(template.Execute, bundle)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
