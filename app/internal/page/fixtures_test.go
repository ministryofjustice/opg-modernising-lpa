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

	template := newMockTemplate(t)
	template.
		On("Execute", w, &fixtureData{
			App:  TestAppData,
			Form: &fixturesForm{},
		}).
		Return(nil)

	err := Fixtures(template.Execute)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostFixturesDonorFlow(t *testing.T) {
	form := url.Values{
		"journey":                      {"donor"},
		"donor-details":                {"withDonorDetails"},
		"choose-attorneys":             {"withAttorneys"},
		"choose-replacement-attorneys": {"withReplacementAttorneys"},
		"when-can-lpa-be-used":         {"whenCanBeUsedComplete"},
		"restrictions":                 {"withRestrictions"},
		"certificate-provider":         {"withCPDetails"},
		"people-to-notify":             {"withPeopleToNotify"},
		"ptn-count":                    {"5"},
		"check-and-send-to-cp":         {"lpaChecked"},
		"pay-for-lpa":                  {"paymentComplete"},
		"confirm-id-and-sign":          {"idConfirmedAndSigned"},
		"complete-all-sections":        {"completeLpa"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	template := newMockTemplate(t)

	err := Fixtures(template.Execute)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	expectedPath := "/testing-start?completeLpa=1&idConfirmedAndSigned=1&lpaChecked=1&paymentComplete=1&whenCanBeUsedComplete=1&withAttorneys=1&withCPDetails=1&withDonorDetails=1&withPeopleToNotify=5&withReplacementAttorneys=1&withRestrictions=1"
	assert.Equal(t, expectedPath, resp.Header.Get("Location"))
}

func TestPostFixturesCPFlow(t *testing.T) {
	testCases := map[string]struct {
		form         url.Values
		expectedPath string
	}{
		"Donor has paid": {
			form: url.Values{
				"journey":                {"certificate-provider"},
				"email":                  {"a@example.org"},
				"useTestShareCode":       {"1"},
				"cp-flow-has-donor-paid": {"startCpFlowDonorHasPaid"},
				"completeLpa":            {"1"},
			},
			expectedPath: "/testing-start?startCpFlowDonorHasPaid=1&useTestShareCode=1&withEmail=a%40example.org",
		},
		"Donor has not paid": {
			form: url.Values{
				"journey":                {"certificate-provider"},
				"useTestShareCode":       {"1"},
				"cp-flow-has-donor-paid": {"startCpFlowDonorHasNotPaid"},
				"completeLpa":            {"1"},
			},
			expectedPath: "/testing-start?startCpFlowDonorHasNotPaid=1&useTestShareCode=1",
		},
		"Donor has not paid and email": {
			form: url.Values{
				"journey":                {"certificate-provider"},
				"email":                  {"a@example.org"},
				"useTestShareCode":       {"1"},
				"cp-flow-has-donor-paid": {"startCpFlowDonorHasNotPaid"},
				"completeLpa":            {"1"},
			},
			expectedPath: "/testing-start?startCpFlowDonorHasNotPaid=1&useTestShareCode=1&withEmail=a%40example.org",
		},
		"Donor has not paid no email": {
			form: url.Values{
				"journey":                {"certificate-provider"},
				"useTestShareCode":       {"1"},
				"cp-flow-has-donor-paid": {"startCpFlowDonorHasNotPaid"},
				"completeLpa":            {"1"},
			},
			expectedPath: "/testing-start?startCpFlowDonorHasNotPaid=1&useTestShareCode=1",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", FormUrlEncoded)

			template := newMockTemplate(t)

			err := Fixtures(template.Execute)(TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedPath, resp.Header.Get("Location"))
		})
	}
}

func TestPostFixturesAttorneyFlow(t *testing.T) {
	testCases := map[string]struct {
		form         url.Values
		expectedPath string
	}{
		"with email and signed": {
			form: url.Values{
				"journey": {"attorney"},
				"email":   {"a@example.org"},
				"signed":  {"1"},
			},
			expectedPath: "/testing-start?completeLpa=1&howAttorneysAct=jointly-and-severally&howReplacementAttorneysAct=jointly&provideCertificate=1&redirect=%2Fattorney-start&sendAttorneyShare=1&signedByDonor=1&useTestShareCode=1&withAttorneys=1&withEmail=a%40example.org&withReplacementAttorneys=1&withRestrictions=1&withType=",
		},
		"without email": {
			form: url.Values{
				"journey": {"attorney"},
			},
			expectedPath: "/testing-start?completeLpa=1&howAttorneysAct=jointly-and-severally&howReplacementAttorneysAct=jointly&redirect=%2Fattorney-start&sendAttorneyShare=1&useTestShareCode=1&withAttorneys=1&withReplacementAttorneys=1&withRestrictions=1&withType=",
		},
		"for replacement attorney": {
			form: url.Values{
				"journey":                  {"attorney"},
				"for-replacement-attorney": {"1"},
			},
			expectedPath: "/testing-start?completeLpa=1&forReplacementAttorney=1&howAttorneysAct=jointly-and-severally&howReplacementAttorneysAct=jointly&redirect=%2Fattorney-start&sendAttorneyShare=1&useTestShareCode=1&withAttorneys=1&withReplacementAttorneys=1&withRestrictions=1&withType=",
		},
		"with type": {
			form: url.Values{
				"journey": {"attorney"},
				"type":    {"hw"},
			},
			expectedPath: "/testing-start?completeLpa=1&howAttorneysAct=jointly-and-severally&howReplacementAttorneysAct=jointly&redirect=%2Fattorney-start&sendAttorneyShare=1&useTestShareCode=1&withAttorneys=1&withReplacementAttorneys=1&withRestrictions=1&withType=hw",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", FormUrlEncoded)

			template := newMockTemplate(t)

			err := Fixtures(template.Execute)(TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedPath, resp.Header.Get("Location"))
		})
	}
}

func TestPostFixturesEverything(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("journey=everything"))
	r.Header.Add("Content-Type", FormUrlEncoded)

	template := newMockTemplate(t)

	err := Fixtures(template.Execute)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/testing-start?asAttorney=1&completeLpa=1&fresh=1&withCertificateProvider=1", resp.Header.Get("Location"))
}
