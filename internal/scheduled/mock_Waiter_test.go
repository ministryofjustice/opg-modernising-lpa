// Code generated by mockery. DO NOT EDIT.

package scheduled

import mock "github.com/stretchr/testify/mock"

// mockWaiter is an autogenerated mock type for the Waiter type
type mockWaiter struct {
	mock.Mock
}

type mockWaiter_Expecter struct {
	mock *mock.Mock
}

func (_m *mockWaiter) EXPECT() *mockWaiter_Expecter {
	return &mockWaiter_Expecter{mock: &_m.Mock}
}

// Reset provides a mock function with no fields
func (_m *mockWaiter) Reset() {
	_m.Called()
}

// mockWaiter_Reset_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reset'
type mockWaiter_Reset_Call struct {
	*mock.Call
}

// Reset is a helper method to define mock.On call
func (_e *mockWaiter_Expecter) Reset() *mockWaiter_Reset_Call {
	return &mockWaiter_Reset_Call{Call: _e.mock.On("Reset")}
}

func (_c *mockWaiter_Reset_Call) Run(run func()) *mockWaiter_Reset_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockWaiter_Reset_Call) Return() *mockWaiter_Reset_Call {
	_c.Call.Return()
	return _c
}

func (_c *mockWaiter_Reset_Call) RunAndReturn(run func()) *mockWaiter_Reset_Call {
	_c.Run(run)
	return _c
}

// Wait provides a mock function with no fields
func (_m *mockWaiter) Wait() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Wait")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockWaiter_Wait_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Wait'
type mockWaiter_Wait_Call struct {
	*mock.Call
}

// Wait is a helper method to define mock.On call
func (_e *mockWaiter_Expecter) Wait() *mockWaiter_Wait_Call {
	return &mockWaiter_Wait_Call{Call: _e.mock.On("Wait")}
}

func (_c *mockWaiter_Wait_Call) Run(run func()) *mockWaiter_Wait_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockWaiter_Wait_Call) Return(_a0 error) *mockWaiter_Wait_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockWaiter_Wait_Call) RunAndReturn(run func() error) *mockWaiter_Wait_Call {
	_c.Call.Return(run)
	return _c
}

// newMockWaiter creates a new instance of mockWaiter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockWaiter(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockWaiter {
	mock := &mockWaiter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
