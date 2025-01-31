// Code generated by mockery. DO NOT EDIT.

package lambda

import (
	context "context"

	aws "github.com/aws/aws-sdk-go-v2/aws"

	http "net/http"

	mock "github.com/stretchr/testify/mock"

	time "time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

// mockSigner is an autogenerated mock type for the Signer type
type mockSigner struct {
	mock.Mock
}

type mockSigner_Expecter struct {
	mock *mock.Mock
}

func (_m *mockSigner) EXPECT() *mockSigner_Expecter {
	return &mockSigner_Expecter{mock: &_m.Mock}
}

// SignHTTP provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7
func (_m *mockSigner) SignHTTP(_a0 context.Context, _a1 aws.Credentials, _a2 *http.Request, _a3 string, _a4 string, _a5 string, _a6 time.Time, _a7 ...func(*v4.SignerOptions)) error {
	_va := make([]interface{}, len(_a7))
	for _i := range _a7 {
		_va[_i] = _a7[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1, _a2, _a3, _a4, _a5, _a6)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for SignHTTP")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, aws.Credentials, *http.Request, string, string, string, time.Time, ...func(*v4.SignerOptions)) error); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockSigner_SignHTTP_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SignHTTP'
type mockSigner_SignHTTP_Call struct {
	*mock.Call
}

// SignHTTP is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 aws.Credentials
//   - _a2 *http.Request
//   - _a3 string
//   - _a4 string
//   - _a5 string
//   - _a6 time.Time
//   - _a7 ...func(*v4.SignerOptions)
func (_e *mockSigner_Expecter) SignHTTP(_a0 interface{}, _a1 interface{}, _a2 interface{}, _a3 interface{}, _a4 interface{}, _a5 interface{}, _a6 interface{}, _a7 ...interface{}) *mockSigner_SignHTTP_Call {
	return &mockSigner_SignHTTP_Call{Call: _e.mock.On("SignHTTP",
		append([]interface{}{_a0, _a1, _a2, _a3, _a4, _a5, _a6}, _a7...)...)}
}

func (_c *mockSigner_SignHTTP_Call) Run(run func(_a0 context.Context, _a1 aws.Credentials, _a2 *http.Request, _a3 string, _a4 string, _a5 string, _a6 time.Time, _a7 ...func(*v4.SignerOptions))) *mockSigner_SignHTTP_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*v4.SignerOptions), len(args)-7)
		for i, a := range args[7:] {
			if a != nil {
				variadicArgs[i] = a.(func(*v4.SignerOptions))
			}
		}
		run(args[0].(context.Context), args[1].(aws.Credentials), args[2].(*http.Request), args[3].(string), args[4].(string), args[5].(string), args[6].(time.Time), variadicArgs...)
	})
	return _c
}

func (_c *mockSigner_SignHTTP_Call) Return(_a0 error) *mockSigner_SignHTTP_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockSigner_SignHTTP_Call) RunAndReturn(run func(context.Context, aws.Credentials, *http.Request, string, string, string, time.Time, ...func(*v4.SignerOptions)) error) *mockSigner_SignHTTP_Call {
	_c.Call.Return(run)
	return _c
}

// newMockSigner creates a new instance of mockSigner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockSigner(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockSigner {
	mock := &mockSigner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
