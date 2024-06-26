// Code generated by mockery v2.42.2. DO NOT EDIT.

package supporter

import (
	context "context"

	actor "github.com/ministryofjustice/opg-modernising-lpa/internal/actor"

	mock "github.com/stretchr/testify/mock"
)

// mockShareCodeStore is an autogenerated mock type for the ShareCodeStore type
type mockShareCodeStore struct {
	mock.Mock
}

type mockShareCodeStore_Expecter struct {
	mock *mock.Mock
}

func (_m *mockShareCodeStore) EXPECT() *mockShareCodeStore_Expecter {
	return &mockShareCodeStore_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: ctx, data
func (_m *mockShareCodeStore) Delete(ctx context.Context, data actor.ShareCodeData) error {
	ret := _m.Called(ctx, data)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, actor.ShareCodeData) error); ok {
		r0 = rf(ctx, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeStore_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type mockShareCodeStore_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - data actor.ShareCodeData
func (_e *mockShareCodeStore_Expecter) Delete(ctx interface{}, data interface{}) *mockShareCodeStore_Delete_Call {
	return &mockShareCodeStore_Delete_Call{Call: _e.mock.On("Delete", ctx, data)}
}

func (_c *mockShareCodeStore_Delete_Call) Run(run func(ctx context.Context, data actor.ShareCodeData)) *mockShareCodeStore_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(actor.ShareCodeData))
	})
	return _c
}

func (_c *mockShareCodeStore_Delete_Call) Return(_a0 error) *mockShareCodeStore_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeStore_Delete_Call) RunAndReturn(run func(context.Context, actor.ShareCodeData) error) *mockShareCodeStore_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// GetDonor provides a mock function with given fields: ctx
func (_m *mockShareCodeStore) GetDonor(ctx context.Context) (actor.ShareCodeData, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetDonor")
	}

	var r0 actor.ShareCodeData
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (actor.ShareCodeData, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) actor.ShareCodeData); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(actor.ShareCodeData)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockShareCodeStore_GetDonor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDonor'
type mockShareCodeStore_GetDonor_Call struct {
	*mock.Call
}

// GetDonor is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockShareCodeStore_Expecter) GetDonor(ctx interface{}) *mockShareCodeStore_GetDonor_Call {
	return &mockShareCodeStore_GetDonor_Call{Call: _e.mock.On("GetDonor", ctx)}
}

func (_c *mockShareCodeStore_GetDonor_Call) Run(run func(ctx context.Context)) *mockShareCodeStore_GetDonor_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockShareCodeStore_GetDonor_Call) Return(_a0 actor.ShareCodeData, _a1 error) *mockShareCodeStore_GetDonor_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockShareCodeStore_GetDonor_Call) RunAndReturn(run func(context.Context) (actor.ShareCodeData, error)) *mockShareCodeStore_GetDonor_Call {
	_c.Call.Return(run)
	return _c
}

// PutDonor provides a mock function with given fields: ctx, shareCode, data
func (_m *mockShareCodeStore) PutDonor(ctx context.Context, shareCode string, data actor.ShareCodeData) error {
	ret := _m.Called(ctx, shareCode, data)

	if len(ret) == 0 {
		panic("no return value specified for PutDonor")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, actor.ShareCodeData) error); ok {
		r0 = rf(ctx, shareCode, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeStore_PutDonor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PutDonor'
type mockShareCodeStore_PutDonor_Call struct {
	*mock.Call
}

// PutDonor is a helper method to define mock.On call
//   - ctx context.Context
//   - shareCode string
//   - data actor.ShareCodeData
func (_e *mockShareCodeStore_Expecter) PutDonor(ctx interface{}, shareCode interface{}, data interface{}) *mockShareCodeStore_PutDonor_Call {
	return &mockShareCodeStore_PutDonor_Call{Call: _e.mock.On("PutDonor", ctx, shareCode, data)}
}

func (_c *mockShareCodeStore_PutDonor_Call) Run(run func(ctx context.Context, shareCode string, data actor.ShareCodeData)) *mockShareCodeStore_PutDonor_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(actor.ShareCodeData))
	})
	return _c
}

func (_c *mockShareCodeStore_PutDonor_Call) Return(_a0 error) *mockShareCodeStore_PutDonor_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeStore_PutDonor_Call) RunAndReturn(run func(context.Context, string, actor.ShareCodeData) error) *mockShareCodeStore_PutDonor_Call {
	_c.Call.Return(run)
	return _c
}

// newMockShareCodeStore creates a new instance of mockShareCodeStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockShareCodeStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockShareCodeStore {
	mock := &mockShareCodeStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
