// Code generated by mockery v2.45.0. DO NOT EDIT.

package main

import (
	context "context"

	appcontext "github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"

	donordata "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"

	lpadata "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"

	mock "github.com/stretchr/testify/mock"

	sharecode "github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
)

// mockShareCodeSender is an autogenerated mock type for the ShareCodeSender type
type mockShareCodeSender struct {
	mock.Mock
}

type mockShareCodeSender_Expecter struct {
	mock *mock.Mock
}

func (_m *mockShareCodeSender) EXPECT() *mockShareCodeSender_Expecter {
	return &mockShareCodeSender_Expecter{mock: &_m.Mock}
}

// SendAttorneys provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockShareCodeSender) SendAttorneys(_a0 context.Context, _a1 appcontext.Data, _a2 *lpadata.Lpa) error {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for SendAttorneys")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, appcontext.Data, *lpadata.Lpa) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeSender_SendAttorneys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendAttorneys'
type mockShareCodeSender_SendAttorneys_Call struct {
	*mock.Call
}

// SendAttorneys is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 appcontext.Data
//   - _a2 *lpadata.Lpa
func (_e *mockShareCodeSender_Expecter) SendAttorneys(_a0 interface{}, _a1 interface{}, _a2 interface{}) *mockShareCodeSender_SendAttorneys_Call {
	return &mockShareCodeSender_SendAttorneys_Call{Call: _e.mock.On("SendAttorneys", _a0, _a1, _a2)}
}

func (_c *mockShareCodeSender_SendAttorneys_Call) Run(run func(_a0 context.Context, _a1 appcontext.Data, _a2 *lpadata.Lpa)) *mockShareCodeSender_SendAttorneys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(appcontext.Data), args[2].(*lpadata.Lpa))
	})
	return _c
}

func (_c *mockShareCodeSender_SendAttorneys_Call) Return(_a0 error) *mockShareCodeSender_SendAttorneys_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeSender_SendAttorneys_Call) RunAndReturn(run func(context.Context, appcontext.Data, *lpadata.Lpa) error) *mockShareCodeSender_SendAttorneys_Call {
	_c.Call.Return(run)
	return _c
}

// SendCertificateProviderInvite provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockShareCodeSender) SendCertificateProviderInvite(_a0 context.Context, _a1 appcontext.Data, _a2 sharecode.CertificateProviderInvite) error {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for SendCertificateProviderInvite")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, appcontext.Data, sharecode.CertificateProviderInvite) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeSender_SendCertificateProviderInvite_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendCertificateProviderInvite'
type mockShareCodeSender_SendCertificateProviderInvite_Call struct {
	*mock.Call
}

// SendCertificateProviderInvite is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 appcontext.Data
//   - _a2 sharecode.CertificateProviderInvite
func (_e *mockShareCodeSender_Expecter) SendCertificateProviderInvite(_a0 interface{}, _a1 interface{}, _a2 interface{}) *mockShareCodeSender_SendCertificateProviderInvite_Call {
	return &mockShareCodeSender_SendCertificateProviderInvite_Call{Call: _e.mock.On("SendCertificateProviderInvite", _a0, _a1, _a2)}
}

func (_c *mockShareCodeSender_SendCertificateProviderInvite_Call) Run(run func(_a0 context.Context, _a1 appcontext.Data, _a2 sharecode.CertificateProviderInvite)) *mockShareCodeSender_SendCertificateProviderInvite_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(appcontext.Data), args[2].(sharecode.CertificateProviderInvite))
	})
	return _c
}

func (_c *mockShareCodeSender_SendCertificateProviderInvite_Call) Return(_a0 error) *mockShareCodeSender_SendCertificateProviderInvite_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeSender_SendCertificateProviderInvite_Call) RunAndReturn(run func(context.Context, appcontext.Data, sharecode.CertificateProviderInvite) error) *mockShareCodeSender_SendCertificateProviderInvite_Call {
	_c.Call.Return(run)
	return _c
}

// SendCertificateProviderPrompt provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockShareCodeSender) SendCertificateProviderPrompt(_a0 context.Context, _a1 appcontext.Data, _a2 *donordata.Provided) error {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for SendCertificateProviderPrompt")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, appcontext.Data, *donordata.Provided) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeSender_SendCertificateProviderPrompt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendCertificateProviderPrompt'
type mockShareCodeSender_SendCertificateProviderPrompt_Call struct {
	*mock.Call
}

// SendCertificateProviderPrompt is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 appcontext.Data
//   - _a2 *donordata.Provided
func (_e *mockShareCodeSender_Expecter) SendCertificateProviderPrompt(_a0 interface{}, _a1 interface{}, _a2 interface{}) *mockShareCodeSender_SendCertificateProviderPrompt_Call {
	return &mockShareCodeSender_SendCertificateProviderPrompt_Call{Call: _e.mock.On("SendCertificateProviderPrompt", _a0, _a1, _a2)}
}

func (_c *mockShareCodeSender_SendCertificateProviderPrompt_Call) Run(run func(_a0 context.Context, _a1 appcontext.Data, _a2 *donordata.Provided)) *mockShareCodeSender_SendCertificateProviderPrompt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(appcontext.Data), args[2].(*donordata.Provided))
	})
	return _c
}

func (_c *mockShareCodeSender_SendCertificateProviderPrompt_Call) Return(_a0 error) *mockShareCodeSender_SendCertificateProviderPrompt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeSender_SendCertificateProviderPrompt_Call) RunAndReturn(run func(context.Context, appcontext.Data, *donordata.Provided) error) *mockShareCodeSender_SendCertificateProviderPrompt_Call {
	_c.Call.Return(run)
	return _c
}

// newMockShareCodeSender creates a new instance of mockShareCodeSender. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockShareCodeSender(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockShareCodeSender {
	mock := &mockShareCodeSender{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}