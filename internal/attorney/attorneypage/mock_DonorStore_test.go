// Code generated by mockery v2.42.0. DO NOT EDIT.

package attorneypage

import (
	context "context"

	actor "github.com/ministryofjustice/opg-modernising-lpa/internal/actor"

	mock "github.com/stretchr/testify/mock"
)

// mockDonorStore is an autogenerated mock type for the DonorStore type
type mockDonorStore struct {
	mock.Mock
}

type mockDonorStore_Expecter struct {
	mock *mock.Mock
}

func (_m *mockDonorStore) EXPECT() *mockDonorStore_Expecter {
	return &mockDonorStore_Expecter{mock: &_m.Mock}
}

// GetAny provides a mock function with given fields: _a0
func (_m *mockDonorStore) GetAny(_a0 context.Context) (*actor.DonorProvidedDetails, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetAny")
	}

	var r0 *actor.DonorProvidedDetails
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*actor.DonorProvidedDetails, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *actor.DonorProvidedDetails); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*actor.DonorProvidedDetails)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDonorStore_GetAny_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAny'
type mockDonorStore_GetAny_Call struct {
	*mock.Call
}

// GetAny is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *mockDonorStore_Expecter) GetAny(_a0 interface{}) *mockDonorStore_GetAny_Call {
	return &mockDonorStore_GetAny_Call{Call: _e.mock.On("GetAny", _a0)}
}

func (_c *mockDonorStore_GetAny_Call) Run(run func(_a0 context.Context)) *mockDonorStore_GetAny_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockDonorStore_GetAny_Call) Return(_a0 *actor.DonorProvidedDetails, _a1 error) *mockDonorStore_GetAny_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDonorStore_GetAny_Call) RunAndReturn(run func(context.Context) (*actor.DonorProvidedDetails, error)) *mockDonorStore_GetAny_Call {
	_c.Call.Return(run)
	return _c
}

// newMockDonorStore creates a new instance of mockDonorStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockDonorStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockDonorStore {
	mock := &mockDonorStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}