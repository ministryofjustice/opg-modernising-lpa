// Code generated by mockery. DO NOT EDIT.

package supporterpage

import (
	http "net/http"

	sesh "github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	mock "github.com/stretchr/testify/mock"
)

// mockSessionStore is an autogenerated mock type for the SessionStore type
type mockSessionStore struct {
	mock.Mock
}

type mockSessionStore_Expecter struct {
	mock *mock.Mock
}

func (_m *mockSessionStore) EXPECT() *mockSessionStore_Expecter {
	return &mockSessionStore_Expecter{mock: &_m.Mock}
}

// ClearLogin provides a mock function with given fields: r, w
func (_m *mockSessionStore) ClearLogin(r *http.Request, w http.ResponseWriter) error {
	ret := _m.Called(r, w)

	if len(ret) == 0 {
		panic("no return value specified for ClearLogin")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*http.Request, http.ResponseWriter) error); ok {
		r0 = rf(r, w)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockSessionStore_ClearLogin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClearLogin'
type mockSessionStore_ClearLogin_Call struct {
	*mock.Call
}

// ClearLogin is a helper method to define mock.On call
//   - r *http.Request
//   - w http.ResponseWriter
func (_e *mockSessionStore_Expecter) ClearLogin(r interface{}, w interface{}) *mockSessionStore_ClearLogin_Call {
	return &mockSessionStore_ClearLogin_Call{Call: _e.mock.On("ClearLogin", r, w)}
}

func (_c *mockSessionStore_ClearLogin_Call) Run(run func(r *http.Request, w http.ResponseWriter)) *mockSessionStore_ClearLogin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Request), args[1].(http.ResponseWriter))
	})
	return _c
}

func (_c *mockSessionStore_ClearLogin_Call) Return(_a0 error) *mockSessionStore_ClearLogin_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockSessionStore_ClearLogin_Call) RunAndReturn(run func(*http.Request, http.ResponseWriter) error) *mockSessionStore_ClearLogin_Call {
	_c.Call.Return(run)
	return _c
}

// Login provides a mock function with given fields: r
func (_m *mockSessionStore) Login(r *http.Request) (*sesh.LoginSession, error) {
	ret := _m.Called(r)

	if len(ret) == 0 {
		panic("no return value specified for Login")
	}

	var r0 *sesh.LoginSession
	var r1 error
	if rf, ok := ret.Get(0).(func(*http.Request) (*sesh.LoginSession, error)); ok {
		return rf(r)
	}
	if rf, ok := ret.Get(0).(func(*http.Request) *sesh.LoginSession); ok {
		r0 = rf(r)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sesh.LoginSession)
		}
	}

	if rf, ok := ret.Get(1).(func(*http.Request) error); ok {
		r1 = rf(r)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockSessionStore_Login_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Login'
type mockSessionStore_Login_Call struct {
	*mock.Call
}

// Login is a helper method to define mock.On call
//   - r *http.Request
func (_e *mockSessionStore_Expecter) Login(r interface{}) *mockSessionStore_Login_Call {
	return &mockSessionStore_Login_Call{Call: _e.mock.On("Login", r)}
}

func (_c *mockSessionStore_Login_Call) Run(run func(r *http.Request)) *mockSessionStore_Login_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Request))
	})
	return _c
}

func (_c *mockSessionStore_Login_Call) Return(_a0 *sesh.LoginSession, _a1 error) *mockSessionStore_Login_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockSessionStore_Login_Call) RunAndReturn(run func(*http.Request) (*sesh.LoginSession, error)) *mockSessionStore_Login_Call {
	_c.Call.Return(run)
	return _c
}

// OneLogin provides a mock function with given fields: r
func (_m *mockSessionStore) OneLogin(r *http.Request) (*sesh.OneLoginSession, error) {
	ret := _m.Called(r)

	if len(ret) == 0 {
		panic("no return value specified for OneLogin")
	}

	var r0 *sesh.OneLoginSession
	var r1 error
	if rf, ok := ret.Get(0).(func(*http.Request) (*sesh.OneLoginSession, error)); ok {
		return rf(r)
	}
	if rf, ok := ret.Get(0).(func(*http.Request) *sesh.OneLoginSession); ok {
		r0 = rf(r)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sesh.OneLoginSession)
		}
	}

	if rf, ok := ret.Get(1).(func(*http.Request) error); ok {
		r1 = rf(r)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockSessionStore_OneLogin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OneLogin'
type mockSessionStore_OneLogin_Call struct {
	*mock.Call
}

// OneLogin is a helper method to define mock.On call
//   - r *http.Request
func (_e *mockSessionStore_Expecter) OneLogin(r interface{}) *mockSessionStore_OneLogin_Call {
	return &mockSessionStore_OneLogin_Call{Call: _e.mock.On("OneLogin", r)}
}

func (_c *mockSessionStore_OneLogin_Call) Run(run func(r *http.Request)) *mockSessionStore_OneLogin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Request))
	})
	return _c
}

func (_c *mockSessionStore_OneLogin_Call) Return(_a0 *sesh.OneLoginSession, _a1 error) *mockSessionStore_OneLogin_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockSessionStore_OneLogin_Call) RunAndReturn(run func(*http.Request) (*sesh.OneLoginSession, error)) *mockSessionStore_OneLogin_Call {
	_c.Call.Return(run)
	return _c
}

// SetLogin provides a mock function with given fields: r, w, session
func (_m *mockSessionStore) SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error {
	ret := _m.Called(r, w, session)

	if len(ret) == 0 {
		panic("no return value specified for SetLogin")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*http.Request, http.ResponseWriter, *sesh.LoginSession) error); ok {
		r0 = rf(r, w, session)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockSessionStore_SetLogin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetLogin'
type mockSessionStore_SetLogin_Call struct {
	*mock.Call
}

// SetLogin is a helper method to define mock.On call
//   - r *http.Request
//   - w http.ResponseWriter
//   - session *sesh.LoginSession
func (_e *mockSessionStore_Expecter) SetLogin(r interface{}, w interface{}, session interface{}) *mockSessionStore_SetLogin_Call {
	return &mockSessionStore_SetLogin_Call{Call: _e.mock.On("SetLogin", r, w, session)}
}

func (_c *mockSessionStore_SetLogin_Call) Run(run func(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession)) *mockSessionStore_SetLogin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Request), args[1].(http.ResponseWriter), args[2].(*sesh.LoginSession))
	})
	return _c
}

func (_c *mockSessionStore_SetLogin_Call) Return(_a0 error) *mockSessionStore_SetLogin_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockSessionStore_SetLogin_Call) RunAndReturn(run func(*http.Request, http.ResponseWriter, *sesh.LoginSession) error) *mockSessionStore_SetLogin_Call {
	_c.Call.Return(run)
	return _c
}

// SetOneLogin provides a mock function with given fields: r, w, session
func (_m *mockSessionStore) SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error {
	ret := _m.Called(r, w, session)

	if len(ret) == 0 {
		panic("no return value specified for SetOneLogin")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*http.Request, http.ResponseWriter, *sesh.OneLoginSession) error); ok {
		r0 = rf(r, w, session)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockSessionStore_SetOneLogin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetOneLogin'
type mockSessionStore_SetOneLogin_Call struct {
	*mock.Call
}

// SetOneLogin is a helper method to define mock.On call
//   - r *http.Request
//   - w http.ResponseWriter
//   - session *sesh.OneLoginSession
func (_e *mockSessionStore_Expecter) SetOneLogin(r interface{}, w interface{}, session interface{}) *mockSessionStore_SetOneLogin_Call {
	return &mockSessionStore_SetOneLogin_Call{Call: _e.mock.On("SetOneLogin", r, w, session)}
}

func (_c *mockSessionStore_SetOneLogin_Call) Run(run func(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession)) *mockSessionStore_SetOneLogin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Request), args[1].(http.ResponseWriter), args[2].(*sesh.OneLoginSession))
	})
	return _c
}

func (_c *mockSessionStore_SetOneLogin_Call) Return(_a0 error) *mockSessionStore_SetOneLogin_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockSessionStore_SetOneLogin_Call) RunAndReturn(run func(*http.Request, http.ResponseWriter, *sesh.OneLoginSession) error) *mockSessionStore_SetOneLogin_Call {
	_c.Call.Return(run)
	return _c
}

// newMockSessionStore creates a new instance of mockSessionStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockSessionStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockSessionStore {
	mock := &mockSessionStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
