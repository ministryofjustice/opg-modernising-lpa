package donor

import (
	"context"
	"errors"
	"io"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/mock"
)

const formUrlEncoded = "application/x-www-form-urlencoded"

var (
	expectedError = errors.New("err")
	appData       = page.AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     page.Paths,
	}
)

type mockLpaStore struct {
	mock.Mock
}

func (m *mockLpaStore) Create(ctx context.Context) (*page.Lpa, error) {
	args := m.Called(ctx)

	return args.Get(0).(*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) GetAll(ctx context.Context) ([]*page.Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) Get(ctx context.Context) (*page.Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).(*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) Put(ctx context.Context, v *page.Lpa) error {
	return m.Called(ctx, v).Error(0)
}

type mockTemplate struct {
	mock.Mock
}

func (m *mockTemplate) Func(w io.Writer, data interface{}) error {
	args := m.Called(w, data)
	return args.Error(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Print(v ...interface{}) {
	m.Called(v...)
}
