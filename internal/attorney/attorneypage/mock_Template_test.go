// Code generated by mockery. DO NOT EDIT.

package attorneypage

import (
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// mockTemplate is an autogenerated mock type for the Template type
type mockTemplate struct {
	mock.Mock
}

type mockTemplate_Expecter struct {
	mock *mock.Mock
}

func (_m *mockTemplate) EXPECT() *mockTemplate_Expecter {
	return &mockTemplate_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0, _a1
func (_m *mockTemplate) Execute(_a0 io.Writer, _a1 interface{}) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(io.Writer, interface{}) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockTemplate_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type mockTemplate_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 io.Writer
//   - _a1 interface{}
func (_e *mockTemplate_Expecter) Execute(_a0 interface{}, _a1 interface{}) *mockTemplate_Execute_Call {
	return &mockTemplate_Execute_Call{Call: _e.mock.On("Execute", _a0, _a1)}
}

func (_c *mockTemplate_Execute_Call) Run(run func(_a0 io.Writer, _a1 interface{})) *mockTemplate_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(io.Writer), args[1].(interface{}))
	})
	return _c
}

func (_c *mockTemplate_Execute_Call) Return(_a0 error) *mockTemplate_Execute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockTemplate_Execute_Call) RunAndReturn(run func(io.Writer, interface{}) error) *mockTemplate_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// newMockTemplate creates a new instance of mockTemplate. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockTemplate(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockTemplate {
	mock := &mockTemplate{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
