// Code generated by mockery. DO NOT EDIT.

package main

import (
	context "context"

	event "github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	mock "github.com/stretchr/testify/mock"
)

// mockEventClient is an autogenerated mock type for the EventClient type
type mockEventClient struct {
	mock.Mock
}

type mockEventClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockEventClient) EXPECT() *mockEventClient_Expecter {
	return &mockEventClient_Expecter{mock: &_m.Mock}
}

// SendApplicationUpdated provides a mock function with given fields: ctx, _a1
func (_m *mockEventClient) SendApplicationUpdated(ctx context.Context, _a1 event.ApplicationUpdated) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SendApplicationUpdated")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, event.ApplicationUpdated) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEventClient_SendApplicationUpdated_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendApplicationUpdated'
type mockEventClient_SendApplicationUpdated_Call struct {
	*mock.Call
}

// SendApplicationUpdated is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 event.ApplicationUpdated
func (_e *mockEventClient_Expecter) SendApplicationUpdated(ctx interface{}, _a1 interface{}) *mockEventClient_SendApplicationUpdated_Call {
	return &mockEventClient_SendApplicationUpdated_Call{Call: _e.mock.On("SendApplicationUpdated", ctx, _a1)}
}

func (_c *mockEventClient_SendApplicationUpdated_Call) Run(run func(ctx context.Context, _a1 event.ApplicationUpdated)) *mockEventClient_SendApplicationUpdated_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(event.ApplicationUpdated))
	})
	return _c
}

func (_c *mockEventClient_SendApplicationUpdated_Call) Return(_a0 error) *mockEventClient_SendApplicationUpdated_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEventClient_SendApplicationUpdated_Call) RunAndReturn(run func(context.Context, event.ApplicationUpdated) error) *mockEventClient_SendApplicationUpdated_Call {
	_c.Call.Return(run)
	return _c
}

// SendAttorneyStarted provides a mock function with given fields: ctx, _a1
func (_m *mockEventClient) SendAttorneyStarted(ctx context.Context, _a1 event.AttorneyStarted) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SendAttorneyStarted")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, event.AttorneyStarted) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEventClient_SendAttorneyStarted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendAttorneyStarted'
type mockEventClient_SendAttorneyStarted_Call struct {
	*mock.Call
}

// SendAttorneyStarted is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 event.AttorneyStarted
func (_e *mockEventClient_Expecter) SendAttorneyStarted(ctx interface{}, _a1 interface{}) *mockEventClient_SendAttorneyStarted_Call {
	return &mockEventClient_SendAttorneyStarted_Call{Call: _e.mock.On("SendAttorneyStarted", ctx, _a1)}
}

func (_c *mockEventClient_SendAttorneyStarted_Call) Run(run func(ctx context.Context, _a1 event.AttorneyStarted)) *mockEventClient_SendAttorneyStarted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(event.AttorneyStarted))
	})
	return _c
}

func (_c *mockEventClient_SendAttorneyStarted_Call) Return(_a0 error) *mockEventClient_SendAttorneyStarted_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEventClient_SendAttorneyStarted_Call) RunAndReturn(run func(context.Context, event.AttorneyStarted) error) *mockEventClient_SendAttorneyStarted_Call {
	_c.Call.Return(run)
	return _c
}

// SendCertificateProviderStarted provides a mock function with given fields: ctx, _a1
func (_m *mockEventClient) SendCertificateProviderStarted(ctx context.Context, _a1 event.CertificateProviderStarted) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SendCertificateProviderStarted")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, event.CertificateProviderStarted) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEventClient_SendCertificateProviderStarted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendCertificateProviderStarted'
type mockEventClient_SendCertificateProviderStarted_Call struct {
	*mock.Call
}

// SendCertificateProviderStarted is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 event.CertificateProviderStarted
func (_e *mockEventClient_Expecter) SendCertificateProviderStarted(ctx interface{}, _a1 interface{}) *mockEventClient_SendCertificateProviderStarted_Call {
	return &mockEventClient_SendCertificateProviderStarted_Call{Call: _e.mock.On("SendCertificateProviderStarted", ctx, _a1)}
}

func (_c *mockEventClient_SendCertificateProviderStarted_Call) Run(run func(ctx context.Context, _a1 event.CertificateProviderStarted)) *mockEventClient_SendCertificateProviderStarted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(event.CertificateProviderStarted))
	})
	return _c
}

func (_c *mockEventClient_SendCertificateProviderStarted_Call) Return(_a0 error) *mockEventClient_SendCertificateProviderStarted_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEventClient_SendCertificateProviderStarted_Call) RunAndReturn(run func(context.Context, event.CertificateProviderStarted) error) *mockEventClient_SendCertificateProviderStarted_Call {
	_c.Call.Return(run)
	return _c
}

// SendLpaAccessGranted provides a mock function with given fields: ctx, _a1
func (_m *mockEventClient) SendLpaAccessGranted(ctx context.Context, _a1 event.LpaAccessGranted) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SendLpaAccessGranted")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, event.LpaAccessGranted) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEventClient_SendLpaAccessGranted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendLpaAccessGranted'
type mockEventClient_SendLpaAccessGranted_Call struct {
	*mock.Call
}

// SendLpaAccessGranted is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 event.LpaAccessGranted
func (_e *mockEventClient_Expecter) SendLpaAccessGranted(ctx interface{}, _a1 interface{}) *mockEventClient_SendLpaAccessGranted_Call {
	return &mockEventClient_SendLpaAccessGranted_Call{Call: _e.mock.On("SendLpaAccessGranted", ctx, _a1)}
}

func (_c *mockEventClient_SendLpaAccessGranted_Call) Run(run func(ctx context.Context, _a1 event.LpaAccessGranted)) *mockEventClient_SendLpaAccessGranted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(event.LpaAccessGranted))
	})
	return _c
}

func (_c *mockEventClient_SendLpaAccessGranted_Call) Return(_a0 error) *mockEventClient_SendLpaAccessGranted_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEventClient_SendLpaAccessGranted_Call) RunAndReturn(run func(context.Context, event.LpaAccessGranted) error) *mockEventClient_SendLpaAccessGranted_Call {
	_c.Call.Return(run)
	return _c
}

// SendNotificationSent provides a mock function with given fields: ctx, notificationSentEvent
func (_m *mockEventClient) SendNotificationSent(ctx context.Context, notificationSentEvent event.NotificationSent) error {
	ret := _m.Called(ctx, notificationSentEvent)

	if len(ret) == 0 {
		panic("no return value specified for SendNotificationSent")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, event.NotificationSent) error); ok {
		r0 = rf(ctx, notificationSentEvent)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEventClient_SendNotificationSent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendNotificationSent'
type mockEventClient_SendNotificationSent_Call struct {
	*mock.Call
}

// SendNotificationSent is a helper method to define mock.On call
//   - ctx context.Context
//   - notificationSentEvent event.NotificationSent
func (_e *mockEventClient_Expecter) SendNotificationSent(ctx interface{}, notificationSentEvent interface{}) *mockEventClient_SendNotificationSent_Call {
	return &mockEventClient_SendNotificationSent_Call{Call: _e.mock.On("SendNotificationSent", ctx, notificationSentEvent)}
}

func (_c *mockEventClient_SendNotificationSent_Call) Run(run func(ctx context.Context, notificationSentEvent event.NotificationSent)) *mockEventClient_SendNotificationSent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(event.NotificationSent))
	})
	return _c
}

func (_c *mockEventClient_SendNotificationSent_Call) Return(_a0 error) *mockEventClient_SendNotificationSent_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEventClient_SendNotificationSent_Call) RunAndReturn(run func(context.Context, event.NotificationSent) error) *mockEventClient_SendNotificationSent_Call {
	_c.Call.Return(run)
	return _c
}

// SendPaperFormRequested provides a mock function with given fields: ctx, paperFormRequestedEvent
func (_m *mockEventClient) SendPaperFormRequested(ctx context.Context, paperFormRequestedEvent event.PaperFormRequested) error {
	ret := _m.Called(ctx, paperFormRequestedEvent)

	if len(ret) == 0 {
		panic("no return value specified for SendPaperFormRequested")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, event.PaperFormRequested) error); ok {
		r0 = rf(ctx, paperFormRequestedEvent)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEventClient_SendPaperFormRequested_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendPaperFormRequested'
type mockEventClient_SendPaperFormRequested_Call struct {
	*mock.Call
}

// SendPaperFormRequested is a helper method to define mock.On call
//   - ctx context.Context
//   - paperFormRequestedEvent event.PaperFormRequested
func (_e *mockEventClient_Expecter) SendPaperFormRequested(ctx interface{}, paperFormRequestedEvent interface{}) *mockEventClient_SendPaperFormRequested_Call {
	return &mockEventClient_SendPaperFormRequested_Call{Call: _e.mock.On("SendPaperFormRequested", ctx, paperFormRequestedEvent)}
}

func (_c *mockEventClient_SendPaperFormRequested_Call) Run(run func(ctx context.Context, paperFormRequestedEvent event.PaperFormRequested)) *mockEventClient_SendPaperFormRequested_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(event.PaperFormRequested))
	})
	return _c
}

func (_c *mockEventClient_SendPaperFormRequested_Call) Return(_a0 error) *mockEventClient_SendPaperFormRequested_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEventClient_SendPaperFormRequested_Call) RunAndReturn(run func(context.Context, event.PaperFormRequested) error) *mockEventClient_SendPaperFormRequested_Call {
	_c.Call.Return(run)
	return _c
}

// newMockEventClient creates a new instance of mockEventClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockEventClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockEventClient {
	mock := &mockEventClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
