package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotify(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	provided := &donordata.Provided{}
	personToNotifys := []donordata.PersonToNotify{{FirstNames: "John"}}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), provided).
		Return(personToNotifys, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &choosePeopleToNotifyData{
			App:            testAppData,
			Form:           &choosePeopleToNotifyForm{},
			Donor:          &donordata.Provided{},
			PeopleToNotify: personToNotifys,
		}).
		Return(nil)

	err := ChoosePeopleToNotify(template.Execute, service)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyWhenNoReusablePeopleToNotify(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?addAnother=1", nil)

	provided := &donordata.Provided{LpaID: "lpa-id"}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), provided).
		Return(nil, nil)

	err := ChoosePeopleToNotify(nil, service)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterPersonToNotify.FormatQuery("lpa-id", url.Values{"addAnother": {"1"}}), resp.Header.Get("Location"))
}

func TestGetChoosePeopleToNotifyWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := ChoosePeopleToNotify(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestGetChoosePeopleToNotifyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), mock.Anything).
		Return([]donordata.PersonToNotify{{FirstNames: "John"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ChoosePeopleToNotify(template.Execute, service)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotify(t *testing.T) {
	form := url.Values{
		"option": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifys := []donordata.PersonToNotify{{FirstNames: "John"}, {FirstNames: "Dave", Address: place.Address{Line1: "123"}}}
	provided := &donordata.Provided{LpaID: "lpa-id"}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), mock.Anything).
		Return(personToNotifys, nil)
	service.EXPECT().
		PutMany(r.Context(), provided, []donordata.PersonToNotify{personToNotifys[1]}).
		Return(nil)

	err := ChoosePeopleToNotify(nil, service)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChoosePeopleToNotifyWhenNoneSelected(t *testing.T) {
	form := url.Values{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?addAnother=hello", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), mock.Anything).
		Return([]donordata.PersonToNotify{{}}, nil)

	err := ChoosePeopleToNotify(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEnterPersonToNotify.FormatQuery("lpa-id", url.Values{"addAnother": {"hello"}}), resp.Header.Get("Location"))
}

func TestPostChoosePeopleToNotifyWhenReuseStoreError(t *testing.T) {
	form := url.Values{
		"option": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), mock.Anything).
		Return([]donordata.PersonToNotify{{}}, nil)
	service.EXPECT().
		PutMany(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChoosePeopleToNotify(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestReadChoosePeopleToNotifyForm(t *testing.T) {
	form := url.Values{
		"option": {"1", "6"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChoosePeopleToNotifyForm(r)

	assert.Equal(t, []int{1, 6}, result.Indices)
}

func TestChoosePeopleToNotifyFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *choosePeopleToNotifyForm
		errors validation.List
	}{
		"none": {
			form: &choosePeopleToNotifyForm{},
		},
		"some": {
			form: &choosePeopleToNotifyForm{Indices: []int{1, 4, 6}},
		},
		"too many": {
			form:   &choosePeopleToNotifyForm{Indices: []int{1, 4, 6, 7}},
			errors: validation.With("option", validation.CustomError{Label: "youCannotSelectMoreThanFivePeopleToNotify"}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(2))
		})
	}
}

func TestChoosePeopleToNotifyFormSelected(t *testing.T) {
	form := &choosePeopleToNotifyForm{Indices: []int{2}}

	assert.True(t, form.Selected(2))
	assert.False(t, form.Selected(3))
}
