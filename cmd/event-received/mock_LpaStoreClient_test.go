// Code generated by mockery. DO NOT EDIT.

package main

import (
	context "context"

	lpastore "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	lpadata "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"

	mock "github.com/stretchr/testify/mock"
)

// mockLpaStoreClient is an autogenerated mock type for the LpaStoreClient type
type mockLpaStoreClient struct {
	mock.Mock
}

type mockLpaStoreClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockLpaStoreClient) EXPECT() *mockLpaStoreClient_Expecter {
	return &mockLpaStoreClient_Expecter{mock: &_m.Mock}
}

// Lpa provides a mock function with given fields: ctx, uid
func (_m *mockLpaStoreClient) Lpa(ctx context.Context, uid string) (*lpadata.Lpa, error) {
	ret := _m.Called(ctx, uid)

	if len(ret) == 0 {
		panic("no return value specified for Lpa")
	}

	var r0 *lpadata.Lpa
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*lpadata.Lpa, error)); ok {
		return rf(ctx, uid)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *lpadata.Lpa); ok {
		r0 = rf(ctx, uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*lpadata.Lpa)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, uid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockLpaStoreClient_Lpa_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Lpa'
type mockLpaStoreClient_Lpa_Call struct {
	*mock.Call
}

// Lpa is a helper method to define mock.On call
//   - ctx context.Context
//   - uid string
func (_e *mockLpaStoreClient_Expecter) Lpa(ctx interface{}, uid interface{}) *mockLpaStoreClient_Lpa_Call {
	return &mockLpaStoreClient_Lpa_Call{Call: _e.mock.On("Lpa", ctx, uid)}
}

func (_c *mockLpaStoreClient_Lpa_Call) Run(run func(ctx context.Context, uid string)) *mockLpaStoreClient_Lpa_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *mockLpaStoreClient_Lpa_Call) Return(_a0 *lpadata.Lpa, _a1 error) *mockLpaStoreClient_Lpa_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockLpaStoreClient_Lpa_Call) RunAndReturn(run func(context.Context, string) (*lpadata.Lpa, error)) *mockLpaStoreClient_Lpa_Call {
	_c.Call.Return(run)
	return _c
}

// SendLpa provides a mock function with given fields: ctx, uid, body
func (_m *mockLpaStoreClient) SendLpa(ctx context.Context, uid string, body lpastore.CreateLpa) error {
	ret := _m.Called(ctx, uid, body)

	if len(ret) == 0 {
		panic("no return value specified for SendLpa")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, lpastore.CreateLpa) error); ok {
		r0 = rf(ctx, uid, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockLpaStoreClient_SendLpa_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendLpa'
type mockLpaStoreClient_SendLpa_Call struct {
	*mock.Call
}

// SendLpa is a helper method to define mock.On call
//   - ctx context.Context
//   - uid string
//   - body lpastore.CreateLpa
func (_e *mockLpaStoreClient_Expecter) SendLpa(ctx interface{}, uid interface{}, body interface{}) *mockLpaStoreClient_SendLpa_Call {
	return &mockLpaStoreClient_SendLpa_Call{Call: _e.mock.On("SendLpa", ctx, uid, body)}
}

func (_c *mockLpaStoreClient_SendLpa_Call) Run(run func(ctx context.Context, uid string, body lpastore.CreateLpa)) *mockLpaStoreClient_SendLpa_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(lpastore.CreateLpa))
	})
	return _c
}

func (_c *mockLpaStoreClient_SendLpa_Call) Return(_a0 error) *mockLpaStoreClient_SendLpa_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockLpaStoreClient_SendLpa_Call) RunAndReturn(run func(context.Context, string, lpastore.CreateLpa) error) *mockLpaStoreClient_SendLpa_Call {
	_c.Call.Return(run)
	return _c
}

// newMockLpaStoreClient creates a new instance of mockLpaStoreClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockLpaStoreClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockLpaStoreClient {
	mock := &mockLpaStoreClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
