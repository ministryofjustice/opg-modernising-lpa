package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetFixtures(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &fixtureData{
			App:                     TestAppData,
			Form:                    &fixturesForm{},
			CPStartLpaNotSignedPath: "/testing-start?redirect=/certificate-provider-start&withCP=1&withDonorDetails=1&startCpFlowWithoutId=1",
			CPStartLpaSignedPath:    "/testing-start?redirect=/certificate-provider-start&completeLpa=1&startCpFlowWithId=1",
		}).
		Return(nil)

	err := Fixtures(template.Func)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostFixtures(t *testing.T) {
	form := url.Values{
		"donor-details":                {"withDonorDetails"},
		"choose-attorneys":             {"withAttorneys"},
		"choose-replacement-attorneys": {"withReplacementAttorneys"},
		"when-can-lpa-be-used":         {"whenCanBeUsedComplete"},
		"restrictions":                 {"withRestrictions"},
		"certificate-provider":         {"withCP"},
		"people-to-notify":             {"withPeopleToNotify"},
		"check-and-send-to-cp":         {"lpaChecked"},
		"pay-for-lpa":                  {"paymentComplete"},
		"confirm-id-and-sign":          {"idConfirmedAndSigned"},
		"complete-all-sections":        {"completeLpa"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	template := &MockTemplate{}
	template.
		On("Func", w, &fixtureData{
			App:                     TestAppData,
			Form:                    &fixturesForm{},
			CPStartLpaNotSignedPath: "/testing-start?redirect=/certificate-provider-start&withCP=1&withDonorDetails=1&startCpFlowWithoutId=1",
			CPStartLpaSignedPath:    "/testing-start?redirect=/certificate-provider-start&completeLpa=1&startCpFlowWithId=1",
		}).
		Return(nil)

	err := Fixtures(template.Func)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	expectedPath := "/testing-start?completeLpa=1&idConfirmedAndSigned=1&lpaChecked=1&paymentComplete=1&whenCanBeUsedComplete=1&withAttorneys=1&withCP=1&withDonorDetails=1&withPeopleToNotify=1&withReplacementAttorneys=1&withRestrictions=1"
	assert.Equal(t, expectedPath, resp.Header.Get("Location"))
}
