// Code generated by mockery. DO NOT EDIT.

package supporterpage

import (
	context "context"

	accesscodedata "github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"

	donordata "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"

	dynamo "github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"

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

// DeleteDonorAccess provides a mock function with given fields: ctx, link
func (_m *mockDonorStore) DeleteDonorAccess(ctx context.Context, link accesscodedata.Link) error {
	ret := _m.Called(ctx, link)

	if len(ret) == 0 {
		panic("no return value specified for DeleteDonorAccess")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, accesscodedata.Link) error); ok {
		r0 = rf(ctx, link)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDonorStore_DeleteDonorAccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteDonorAccess'
type mockDonorStore_DeleteDonorAccess_Call struct {
	*mock.Call
}

// DeleteDonorAccess is a helper method to define mock.On call
//   - ctx context.Context
//   - link accesscodedata.Link
func (_e *mockDonorStore_Expecter) DeleteDonorAccess(ctx interface{}, link interface{}) *mockDonorStore_DeleteDonorAccess_Call {
	return &mockDonorStore_DeleteDonorAccess_Call{Call: _e.mock.On("DeleteDonorAccess", ctx, link)}
}

func (_c *mockDonorStore_DeleteDonorAccess_Call) Run(run func(ctx context.Context, link accesscodedata.Link)) *mockDonorStore_DeleteDonorAccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(accesscodedata.Link))
	})
	return _c
}

func (_c *mockDonorStore_DeleteDonorAccess_Call) Return(_a0 error) *mockDonorStore_DeleteDonorAccess_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDonorStore_DeleteDonorAccess_Call) RunAndReturn(run func(context.Context, accesscodedata.Link) error) *mockDonorStore_DeleteDonorAccess_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx
func (_m *mockDonorStore) Get(ctx context.Context) (*donordata.Provided, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *donordata.Provided
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*donordata.Provided, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *donordata.Provided); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*donordata.Provided)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDonorStore_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type mockDonorStore_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockDonorStore_Expecter) Get(ctx interface{}) *mockDonorStore_Get_Call {
	return &mockDonorStore_Get_Call{Call: _e.mock.On("Get", ctx)}
}

func (_c *mockDonorStore_Get_Call) Run(run func(ctx context.Context)) *mockDonorStore_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockDonorStore_Get_Call) Return(_a0 *donordata.Provided, _a1 error) *mockDonorStore_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDonorStore_Get_Call) RunAndReturn(run func(context.Context) (*donordata.Provided, error)) *mockDonorStore_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetByKeys provides a mock function with given fields: ctx, keys
func (_m *mockDonorStore) GetByKeys(ctx context.Context, keys []dynamo.Keys) ([]donordata.Provided, error) {
	ret := _m.Called(ctx, keys)

	if len(ret) == 0 {
		panic("no return value specified for GetByKeys")
	}

	var r0 []donordata.Provided
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []dynamo.Keys) ([]donordata.Provided, error)); ok {
		return rf(ctx, keys)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []dynamo.Keys) []donordata.Provided); ok {
		r0 = rf(ctx, keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]donordata.Provided)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []dynamo.Keys) error); ok {
		r1 = rf(ctx, keys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDonorStore_GetByKeys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByKeys'
type mockDonorStore_GetByKeys_Call struct {
	*mock.Call
}

// GetByKeys is a helper method to define mock.On call
//   - ctx context.Context
//   - keys []dynamo.Keys
func (_e *mockDonorStore_Expecter) GetByKeys(ctx interface{}, keys interface{}) *mockDonorStore_GetByKeys_Call {
	return &mockDonorStore_GetByKeys_Call{Call: _e.mock.On("GetByKeys", ctx, keys)}
}

func (_c *mockDonorStore_GetByKeys_Call) Run(run func(ctx context.Context, keys []dynamo.Keys)) *mockDonorStore_GetByKeys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]dynamo.Keys))
	})
	return _c
}

func (_c *mockDonorStore_GetByKeys_Call) Return(_a0 []donordata.Provided, _a1 error) *mockDonorStore_GetByKeys_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDonorStore_GetByKeys_Call) RunAndReturn(run func(context.Context, []dynamo.Keys) ([]donordata.Provided, error)) *mockDonorStore_GetByKeys_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: ctx, donor
func (_m *mockDonorStore) Put(ctx context.Context, donor *donordata.Provided) error {
	ret := _m.Called(ctx, donor)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *donordata.Provided) error); ok {
		r0 = rf(ctx, donor)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDonorStore_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type mockDonorStore_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - ctx context.Context
//   - donor *donordata.Provided
func (_e *mockDonorStore_Expecter) Put(ctx interface{}, donor interface{}) *mockDonorStore_Put_Call {
	return &mockDonorStore_Put_Call{Call: _e.mock.On("Put", ctx, donor)}
}

func (_c *mockDonorStore_Put_Call) Run(run func(ctx context.Context, donor *donordata.Provided)) *mockDonorStore_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*donordata.Provided))
	})
	return _c
}

func (_c *mockDonorStore_Put_Call) Return(_a0 error) *mockDonorStore_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDonorStore_Put_Call) RunAndReturn(run func(context.Context, *donordata.Provided) error) *mockDonorStore_Put_Call {
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
