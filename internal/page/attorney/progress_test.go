package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	lpaSignedAt := time.Now()
	attorneySignedAt := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		attorney        *actor.AttorneyProvidedDetails
		attorneys       []*actor.AttorneyProvidedDetails
		signed          bool
		attorneysSigned bool
	}{
		"unsigned": {
			attorney: &actor.AttorneyProvidedDetails{},
		},
		"attorney signed": {
			attorney: &actor.AttorneyProvidedDetails{
				LpaSignedAt: lpaSignedAt,
				Confirmed:   attorneySignedAt,
			},
			attorneys: []*actor.AttorneyProvidedDetails{{}},
			signed:    true,
		},
		"all signed": {
			attorney: &actor.AttorneyProvidedDetails{},
			attorneys: []*actor.AttorneyProvidedDetails{{
				LpaSignedAt: lpaSignedAt,
				Confirmed:   attorneySignedAt,
			}},
			attorneysSigned: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			donor := &actor.DonorProvidedDetails{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				GetAny(r.Context()).
				Return(donor, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				GetAny(r.Context()).
				Return(tc.attorneys, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &progressData{App: testAppData, Donor: donor, Signed: tc.signed, AttorneysSigned: tc.attorneysSigned}).
				Return(nil)

			err := Progress(template.Execute, attorneyStore, donorStore)(testAppData, w, r, tc.attorney)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestProgressWhenAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return(nil, expectedError)

	err := Progress(nil, attorneyStore, donorStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestProgressWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(donor, expectedError)

	err := Progress(nil, nil, donorStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestProgressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &progressData{App: testAppData, Donor: &actor.DonorProvidedDetails{}}).
		Return(expectedError)

	err := Progress(template.Execute, attorneyStore, donorStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}
