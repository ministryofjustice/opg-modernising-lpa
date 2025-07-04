// Code generated by mockery. DO NOT EDIT.

package accesscode

import (
	context "context"

	notify "github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	mock "github.com/stretchr/testify/mock"
)

// mockNotifyClient is an autogenerated mock type for the NotifyClient type
type mockNotifyClient struct {
	mock.Mock
}

type mockNotifyClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockNotifyClient) EXPECT() *mockNotifyClient_Expecter {
	return &mockNotifyClient_Expecter{mock: &_m.Mock}
}

// SendActorEmail provides a mock function with given fields: _a0, to, lpaUID, email
func (_m *mockNotifyClient) SendActorEmail(_a0 context.Context, to notify.ToEmail, lpaUID string, email notify.Email) error {
	ret := _m.Called(_a0, to, lpaUID, email)

	if len(ret) == 0 {
		panic("no return value specified for SendActorEmail")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, notify.ToEmail, string, notify.Email) error); ok {
		r0 = rf(_a0, to, lpaUID, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockNotifyClient_SendActorEmail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendActorEmail'
type mockNotifyClient_SendActorEmail_Call struct {
	*mock.Call
}

// SendActorEmail is a helper method to define mock.On call
//   - _a0 context.Context
//   - to notify.ToEmail
//   - lpaUID string
//   - email notify.Email
func (_e *mockNotifyClient_Expecter) SendActorEmail(_a0 interface{}, to interface{}, lpaUID interface{}, email interface{}) *mockNotifyClient_SendActorEmail_Call {
	return &mockNotifyClient_SendActorEmail_Call{Call: _e.mock.On("SendActorEmail", _a0, to, lpaUID, email)}
}

func (_c *mockNotifyClient_SendActorEmail_Call) Run(run func(_a0 context.Context, to notify.ToEmail, lpaUID string, email notify.Email)) *mockNotifyClient_SendActorEmail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(notify.ToEmail), args[2].(string), args[3].(notify.Email))
	})
	return _c
}

func (_c *mockNotifyClient_SendActorEmail_Call) Return(_a0 error) *mockNotifyClient_SendActorEmail_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockNotifyClient_SendActorEmail_Call) RunAndReturn(run func(context.Context, notify.ToEmail, string, notify.Email) error) *mockNotifyClient_SendActorEmail_Call {
	_c.Call.Return(run)
	return _c
}

// SendActorSMS provides a mock function with given fields: _a0, to, lpaUID, sms
func (_m *mockNotifyClient) SendActorSMS(_a0 context.Context, to notify.ToMobile, lpaUID string, sms notify.SMS) error {
	ret := _m.Called(_a0, to, lpaUID, sms)

	if len(ret) == 0 {
		panic("no return value specified for SendActorSMS")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, notify.ToMobile, string, notify.SMS) error); ok {
		r0 = rf(_a0, to, lpaUID, sms)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockNotifyClient_SendActorSMS_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendActorSMS'
type mockNotifyClient_SendActorSMS_Call struct {
	*mock.Call
}

// SendActorSMS is a helper method to define mock.On call
//   - _a0 context.Context
//   - to notify.ToMobile
//   - lpaUID string
//   - sms notify.SMS
func (_e *mockNotifyClient_Expecter) SendActorSMS(_a0 interface{}, to interface{}, lpaUID interface{}, sms interface{}) *mockNotifyClient_SendActorSMS_Call {
	return &mockNotifyClient_SendActorSMS_Call{Call: _e.mock.On("SendActorSMS", _a0, to, lpaUID, sms)}
}

func (_c *mockNotifyClient_SendActorSMS_Call) Run(run func(_a0 context.Context, to notify.ToMobile, lpaUID string, sms notify.SMS)) *mockNotifyClient_SendActorSMS_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(notify.ToMobile), args[2].(string), args[3].(notify.SMS))
	})
	return _c
}

func (_c *mockNotifyClient_SendActorSMS_Call) Return(_a0 error) *mockNotifyClient_SendActorSMS_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockNotifyClient_SendActorSMS_Call) RunAndReturn(run func(context.Context, notify.ToMobile, string, notify.SMS) error) *mockNotifyClient_SendActorSMS_Call {
	_c.Call.Return(run)
	return _c
}

// newMockNotifyClient creates a new instance of mockNotifyClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockNotifyClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockNotifyClient {
	mock := &mockNotifyClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
