// Code generated by mockery. DO NOT EDIT.

package donorpage

import (
	context "context"

	donordata "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	mock "github.com/stretchr/testify/mock"
)

// mockGetDonorStore is an autogenerated mock type for the GetDonorStore type
type mockGetDonorStore struct {
	mock.Mock
}

type mockGetDonorStore_Expecter struct {
	mock *mock.Mock
}

func (_m *mockGetDonorStore) EXPECT() *mockGetDonorStore_Expecter {
	return &mockGetDonorStore_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: _a0
func (_m *mockGetDonorStore) Get(_a0 context.Context) (*donordata.Provided, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *donordata.Provided
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*donordata.Provided, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *donordata.Provided); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*donordata.Provided)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockGetDonorStore_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type mockGetDonorStore_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *mockGetDonorStore_Expecter) Get(_a0 interface{}) *mockGetDonorStore_Get_Call {
	return &mockGetDonorStore_Get_Call{Call: _e.mock.On("Get", _a0)}
}

func (_c *mockGetDonorStore_Get_Call) Run(run func(_a0 context.Context)) *mockGetDonorStore_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockGetDonorStore_Get_Call) Return(_a0 *donordata.Provided, _a1 error) *mockGetDonorStore_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockGetDonorStore_Get_Call) RunAndReturn(run func(context.Context) (*donordata.Provided, error)) *mockGetDonorStore_Get_Call {
	_c.Call.Return(run)
	return _c
}

// newMockGetDonorStore creates a new instance of mockGetDonorStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockGetDonorStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockGetDonorStore {
	mock := &mockGetDonorStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
