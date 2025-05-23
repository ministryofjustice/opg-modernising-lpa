// Code generated by mockery. DO NOT EDIT.

package main

import (
	context "context"

	appcontext "github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"

	donordata "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"

	dynamo "github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"

	lpadata "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"

	mock "github.com/stretchr/testify/mock"
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

// SendAttorneys provides a mock function with given fields: ctx, appData, lpa
func (_m *mockShareCodeSender) SendAttorneys(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa) error {
	ret := _m.Called(ctx, appData, lpa)

	if len(ret) == 0 {
		panic("no return value specified for SendAttorneys")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, appcontext.Data, *lpadata.Lpa) error); ok {
		r0 = rf(ctx, appData, lpa)
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
//   - ctx context.Context
//   - appData appcontext.Data
//   - lpa *lpadata.Lpa
func (_e *mockShareCodeSender_Expecter) SendAttorneys(ctx interface{}, appData interface{}, lpa interface{}) *mockShareCodeSender_SendAttorneys_Call {
	return &mockShareCodeSender_SendAttorneys_Call{Call: _e.mock.On("SendAttorneys", ctx, appData, lpa)}
}

func (_c *mockShareCodeSender_SendAttorneys_Call) Run(run func(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa)) *mockShareCodeSender_SendAttorneys_Call {
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

// SendCertificateProviderPrompt provides a mock function with given fields: ctx, appData, provided
func (_m *mockShareCodeSender) SendCertificateProviderPrompt(ctx context.Context, appData appcontext.Data, provided *donordata.Provided) error {
	ret := _m.Called(ctx, appData, provided)

	if len(ret) == 0 {
		panic("no return value specified for SendCertificateProviderPrompt")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, appcontext.Data, *donordata.Provided) error); ok {
		r0 = rf(ctx, appData, provided)
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
//   - ctx context.Context
//   - appData appcontext.Data
//   - provided *donordata.Provided
func (_e *mockShareCodeSender_Expecter) SendCertificateProviderPrompt(ctx interface{}, appData interface{}, provided interface{}) *mockShareCodeSender_SendCertificateProviderPrompt_Call {
	return &mockShareCodeSender_SendCertificateProviderPrompt_Call{Call: _e.mock.On("SendCertificateProviderPrompt", ctx, appData, provided)}
}

func (_c *mockShareCodeSender_SendCertificateProviderPrompt_Call) Run(run func(ctx context.Context, appData appcontext.Data, provided *donordata.Provided)) *mockShareCodeSender_SendCertificateProviderPrompt_Call {
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

// SendLpaCertificateProviderPrompt provides a mock function with given fields: ctx, appData, lpaKey, lpaOwnerKey, lpa
func (_m *mockShareCodeSender) SendLpaCertificateProviderPrompt(ctx context.Context, appData appcontext.Data, lpaKey dynamo.LpaKeyType, lpaOwnerKey dynamo.LpaOwnerKeyType, lpa *lpadata.Lpa) error {
	ret := _m.Called(ctx, appData, lpaKey, lpaOwnerKey, lpa)

	if len(ret) == 0 {
		panic("no return value specified for SendLpaCertificateProviderPrompt")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, appcontext.Data, dynamo.LpaKeyType, dynamo.LpaOwnerKeyType, *lpadata.Lpa) error); ok {
		r0 = rf(ctx, appData, lpaKey, lpaOwnerKey, lpa)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeSender_SendLpaCertificateProviderPrompt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendLpaCertificateProviderPrompt'
type mockShareCodeSender_SendLpaCertificateProviderPrompt_Call struct {
	*mock.Call
}

// SendLpaCertificateProviderPrompt is a helper method to define mock.On call
//   - ctx context.Context
//   - appData appcontext.Data
//   - lpaKey dynamo.LpaKeyType
//   - lpaOwnerKey dynamo.LpaOwnerKeyType
//   - lpa *lpadata.Lpa
func (_e *mockShareCodeSender_Expecter) SendLpaCertificateProviderPrompt(ctx interface{}, appData interface{}, lpaKey interface{}, lpaOwnerKey interface{}, lpa interface{}) *mockShareCodeSender_SendLpaCertificateProviderPrompt_Call {
	return &mockShareCodeSender_SendLpaCertificateProviderPrompt_Call{Call: _e.mock.On("SendLpaCertificateProviderPrompt", ctx, appData, lpaKey, lpaOwnerKey, lpa)}
}

func (_c *mockShareCodeSender_SendLpaCertificateProviderPrompt_Call) Run(run func(ctx context.Context, appData appcontext.Data, lpaKey dynamo.LpaKeyType, lpaOwnerKey dynamo.LpaOwnerKeyType, lpa *lpadata.Lpa)) *mockShareCodeSender_SendLpaCertificateProviderPrompt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(appcontext.Data), args[2].(dynamo.LpaKeyType), args[3].(dynamo.LpaOwnerKeyType), args[4].(*lpadata.Lpa))
	})
	return _c
}

func (_c *mockShareCodeSender_SendLpaCertificateProviderPrompt_Call) Return(_a0 error) *mockShareCodeSender_SendLpaCertificateProviderPrompt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeSender_SendLpaCertificateProviderPrompt_Call) RunAndReturn(run func(context.Context, appcontext.Data, dynamo.LpaKeyType, dynamo.LpaOwnerKeyType, *lpadata.Lpa) error) *mockShareCodeSender_SendLpaCertificateProviderPrompt_Call {
	_c.Call.Return(run)
	return _c
}

// SendVoucherAccessCode provides a mock function with given fields: ctx, provided, appData
func (_m *mockShareCodeSender) SendVoucherAccessCode(ctx context.Context, provided *donordata.Provided, appData appcontext.Data) error {
	ret := _m.Called(ctx, provided, appData)

	if len(ret) == 0 {
		panic("no return value specified for SendVoucherAccessCode")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *donordata.Provided, appcontext.Data) error); ok {
		r0 = rf(ctx, provided, appData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeSender_SendVoucherAccessCode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendVoucherAccessCode'
type mockShareCodeSender_SendVoucherAccessCode_Call struct {
	*mock.Call
}

// SendVoucherAccessCode is a helper method to define mock.On call
//   - ctx context.Context
//   - provided *donordata.Provided
//   - appData appcontext.Data
func (_e *mockShareCodeSender_Expecter) SendVoucherAccessCode(ctx interface{}, provided interface{}, appData interface{}) *mockShareCodeSender_SendVoucherAccessCode_Call {
	return &mockShareCodeSender_SendVoucherAccessCode_Call{Call: _e.mock.On("SendVoucherAccessCode", ctx, provided, appData)}
}

func (_c *mockShareCodeSender_SendVoucherAccessCode_Call) Run(run func(ctx context.Context, provided *donordata.Provided, appData appcontext.Data)) *mockShareCodeSender_SendVoucherAccessCode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*donordata.Provided), args[2].(appcontext.Data))
	})
	return _c
}

func (_c *mockShareCodeSender_SendVoucherAccessCode_Call) Return(_a0 error) *mockShareCodeSender_SendVoucherAccessCode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeSender_SendVoucherAccessCode_Call) RunAndReturn(run func(context.Context, *donordata.Provided, appcontext.Data) error) *mockShareCodeSender_SendVoucherAccessCode_Call {
	_c.Call.Return(run)
	return _c
}

// SendVoucherInvite provides a mock function with given fields: ctx, provided, appData
func (_m *mockShareCodeSender) SendVoucherInvite(ctx context.Context, provided *donordata.Provided, appData appcontext.Data) error {
	ret := _m.Called(ctx, provided, appData)

	if len(ret) == 0 {
		panic("no return value specified for SendVoucherInvite")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *donordata.Provided, appcontext.Data) error); ok {
		r0 = rf(ctx, provided, appData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeSender_SendVoucherInvite_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendVoucherInvite'
type mockShareCodeSender_SendVoucherInvite_Call struct {
	*mock.Call
}

// SendVoucherInvite is a helper method to define mock.On call
//   - ctx context.Context
//   - provided *donordata.Provided
//   - appData appcontext.Data
func (_e *mockShareCodeSender_Expecter) SendVoucherInvite(ctx interface{}, provided interface{}, appData interface{}) *mockShareCodeSender_SendVoucherInvite_Call {
	return &mockShareCodeSender_SendVoucherInvite_Call{Call: _e.mock.On("SendVoucherInvite", ctx, provided, appData)}
}

func (_c *mockShareCodeSender_SendVoucherInvite_Call) Run(run func(ctx context.Context, provided *donordata.Provided, appData appcontext.Data)) *mockShareCodeSender_SendVoucherInvite_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*donordata.Provided), args[2].(appcontext.Data))
	})
	return _c
}

func (_c *mockShareCodeSender_SendVoucherInvite_Call) Return(_a0 error) *mockShareCodeSender_SendVoucherInvite_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeSender_SendVoucherInvite_Call) RunAndReturn(run func(context.Context, *donordata.Provided, appcontext.Data) error) *mockShareCodeSender_SendVoucherInvite_Call {
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
