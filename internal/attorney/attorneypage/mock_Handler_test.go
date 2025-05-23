// Code generated by mockery. DO NOT EDIT.

package attorneypage

import (
	appcontext "github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	attorneydata "github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"

	http "net/http"

	lpadata "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"

	mock "github.com/stretchr/testify/mock"
)

// mockHandler is an autogenerated mock type for the Handler type
type mockHandler struct {
	mock.Mock
}

type mockHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *mockHandler) EXPECT() *mockHandler_Expecter {
	return &mockHandler_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: data, w, r, details, lpa
func (_m *mockHandler) Execute(data appcontext.Data, w http.ResponseWriter, r *http.Request, details *attorneydata.Provided, lpa *lpadata.Lpa) error {
	ret := _m.Called(data, w, r, details, lpa)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(appcontext.Data, http.ResponseWriter, *http.Request, *attorneydata.Provided, *lpadata.Lpa) error); ok {
		r0 = rf(data, w, r, details, lpa)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockHandler_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type mockHandler_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - data appcontext.Data
//   - w http.ResponseWriter
//   - r *http.Request
//   - details *attorneydata.Provided
//   - lpa *lpadata.Lpa
func (_e *mockHandler_Expecter) Execute(data interface{}, w interface{}, r interface{}, details interface{}, lpa interface{}) *mockHandler_Execute_Call {
	return &mockHandler_Execute_Call{Call: _e.mock.On("Execute", data, w, r, details, lpa)}
}

func (_c *mockHandler_Execute_Call) Run(run func(data appcontext.Data, w http.ResponseWriter, r *http.Request, details *attorneydata.Provided, lpa *lpadata.Lpa)) *mockHandler_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(appcontext.Data), args[1].(http.ResponseWriter), args[2].(*http.Request), args[3].(*attorneydata.Provided), args[4].(*lpadata.Lpa))
	})
	return _c
}

func (_c *mockHandler_Execute_Call) Return(_a0 error) *mockHandler_Execute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockHandler_Execute_Call) RunAndReturn(run func(appcontext.Data, http.ResponseWriter, *http.Request, *attorneydata.Provided, *lpadata.Lpa) error) *mockHandler_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// newMockHandler creates a new instance of mockHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockHandler {
	mock := &mockHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
