package certificateprovider

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/mock"
)

var (
	testAddress = place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "e",
	}
	expectedError = errors.New("err")
	testAppData   = page.AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     page.Paths,
	}
)

type mockSessionsStore struct {
	mock.Mock
}

func (m *mockSessionsStore) New(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *mockSessionsStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *mockSessionsStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	args := m.Called(r, w, session)
	return args.Error(0)
}
