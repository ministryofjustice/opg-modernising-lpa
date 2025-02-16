// Code generated by mockery. DO NOT EDIT.

package donordata

import mock "github.com/stretchr/testify/mock"

// mockLocalizer is an autogenerated mock type for the Localizer type
type mockLocalizer struct {
	mock.Mock
}

type mockLocalizer_Expecter struct {
	mock *mock.Mock
}

func (_m *mockLocalizer) EXPECT() *mockLocalizer_Expecter {
	return &mockLocalizer_Expecter{mock: &_m.Mock}
}

// Format provides a mock function with given fields: _a0, _a1
func (_m *mockLocalizer) Format(_a0 string, _a1 map[string]interface{}) string {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Format")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, map[string]interface{}) string); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// mockLocalizer_Format_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Format'
type mockLocalizer_Format_Call struct {
	*mock.Call
}

// Format is a helper method to define mock.On call
//   - _a0 string
//   - _a1 map[string]interface{}
func (_e *mockLocalizer_Expecter) Format(_a0 interface{}, _a1 interface{}) *mockLocalizer_Format_Call {
	return &mockLocalizer_Format_Call{Call: _e.mock.On("Format", _a0, _a1)}
}

func (_c *mockLocalizer_Format_Call) Run(run func(_a0 string, _a1 map[string]interface{})) *mockLocalizer_Format_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(map[string]interface{}))
	})
	return _c
}

func (_c *mockLocalizer_Format_Call) Return(_a0 string) *mockLocalizer_Format_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockLocalizer_Format_Call) RunAndReturn(run func(string, map[string]interface{}) string) *mockLocalizer_Format_Call {
	_c.Call.Return(run)
	return _c
}

// newMockLocalizer creates a new instance of mockLocalizer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockLocalizer(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockLocalizer {
	mock := &mockLocalizer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
