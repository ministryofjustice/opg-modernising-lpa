package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemovePersonToNotify(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}

	personToNotify := PersonToNotify{
		ID: "123",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	template := &mockTemplate{}
	template.
		On("Func", w, &removePersonToNotifyData{
			App:            appData,
			PersonToNotify: personToNotify,
			Errors:         nil,
			Form:           removePersonToNotifyForm{},
		}).
		Return(nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := RemovePersonToNotify(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetRemovePersonToNotifyErrorOnStore(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", "error getting lpa from store: err").
		Return(nil)

	template := &mockTemplate{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := RemovePersonToNotify(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, logger)
}

func TestGetRemovePersonToNotifyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}

	template := &mockTemplate{}

	personToNotify := PersonToNotify{
		ID: "123",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	err := RemovePersonToNotify(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRemovePersonToNotify(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	template := &mockTemplate{}

	personToNotifyWithAddress := PersonToNotify{
		ID: "with-address",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := PersonToNotify{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{PeopleToNotify: []PersonToNotify{personToNotifyWithAddress}}).
		Return(nil)

	form := url.Values{
		"remove-person-to-notify": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemovePersonToNotify(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemovePersonToNotifyWithFormValueNo(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	template := &mockTemplate{}

	personToNotifyWithAddress := PersonToNotify{
		ID: "with-address",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := PersonToNotify{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}}, nil)

	form := url.Values{
		"remove-person-to-notify": {"no"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemovePersonToNotify(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostRemovePersonToNotifyErrorOnPutStore(t *testing.T) {
	w := httptest.NewRecorder()

	template := &mockTemplate{}

	logger := &mockLogger{}
	logger.
		On("Print", "error removing PersonToNotify from LPA: err").
		Return(nil)

	personToNotifyWithAddress := PersonToNotify{
		ID: "with-address",
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := PersonToNotify{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{PeopleToNotify: []PersonToNotify{personToNotifyWithAddress}}).
		Return(expectedError)

	form := url.Values{
		"remove-person-to-notify": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemovePersonToNotify(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template, logger)
}

func TestRemovePersonToNotifyFormValidation(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotifyWithoutAddress := PersonToNotify{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotifyWithoutAddress}}, nil)

	validationError := map[string]string{
		"remove-person-to-notify": "selectRemovePersonToNotify",
	}

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *removePersonToNotifyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"remove-person-to-notify": {""},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemovePersonToNotify(nil, template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestRemovePersonToNotifyRemoveLastPersonRedirectsToChoosePeopleToNotify(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	template := &mockTemplate{}

	personToNotifyWithoutAddress := PersonToNotify{
		ID:      "without-address",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{personToNotifyWithoutAddress},
			Tasks:          Tasks{PeopleToNotify: TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			PeopleToNotify: []PersonToNotify{},
			Tasks:          Tasks{PeopleToNotify: TaskNotStarted},
		}).
		Return(nil)

	form := url.Values{
		"remove-person-to-notify": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=without-address", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := RemovePersonToNotify(logger, template.Func, lpaStore)(appData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.DoYouWantToNotifyPeople, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestRemovePersonToNotifyFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *removePersonToNotifyForm
		errors map[string]string
	}{
		"valid - yes": {
			form: &removePersonToNotifyForm{
				RemovePersonToNotify: "yes",
			},
			errors: map[string]string{},
		},
		"valid - no": {
			form: &removePersonToNotifyForm{
				RemovePersonToNotify: "no",
			},
			errors: map[string]string{},
		},
		"missing-value": {
			form: &removePersonToNotifyForm{
				RemovePersonToNotify: "",
			},
			errors: map[string]string{
				"remove-person-to-notify": "selectRemovePersonToNotify",
			},
		},
		"unexpected-value": {
			form: &removePersonToNotifyForm{
				RemovePersonToNotify: "not expected",
			},
			errors: map[string]string{
				"remove-person-to-notify": "selectRemovePersonToNotify",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
