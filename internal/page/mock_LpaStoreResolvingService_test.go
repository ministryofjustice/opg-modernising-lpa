// Code generated by mockery. DO NOT EDIT.

package page

import (
	context "context"

	lpadata "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	mock "github.com/stretchr/testify/mock"
)

// mockLpaStoreResolvingService is an autogenerated mock type for the LpaStoreResolvingService type
type mockLpaStoreResolvingService struct {
	mock.Mock
}

type mockLpaStoreResolvingService_Expecter struct {
	mock *mock.Mock
}

func (_m *mockLpaStoreResolvingService) EXPECT() *mockLpaStoreResolvingService_Expecter {
	return &mockLpaStoreResolvingService_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: ctx
func (_m *mockLpaStoreResolvingService) Get(ctx context.Context) (*lpadata.Lpa, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *lpadata.Lpa
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*lpadata.Lpa, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *lpadata.Lpa); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*lpadata.Lpa)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockLpaStoreResolvingService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type mockLpaStoreResolvingService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockLpaStoreResolvingService_Expecter) Get(ctx interface{}) *mockLpaStoreResolvingService_Get_Call {
	return &mockLpaStoreResolvingService_Get_Call{Call: _e.mock.On("Get", ctx)}
}

func (_c *mockLpaStoreResolvingService_Get_Call) Run(run func(ctx context.Context)) *mockLpaStoreResolvingService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockLpaStoreResolvingService_Get_Call) Return(_a0 *lpadata.Lpa, _a1 error) *mockLpaStoreResolvingService_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockLpaStoreResolvingService_Get_Call) RunAndReturn(run func(context.Context) (*lpadata.Lpa, error)) *mockLpaStoreResolvingService_Get_Call {
	_c.Call.Return(run)
	return _c
}

// newMockLpaStoreResolvingService creates a new instance of mockLpaStoreResolvingService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockLpaStoreResolvingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockLpaStoreResolvingService {
	mock := &mockLpaStoreResolvingService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
