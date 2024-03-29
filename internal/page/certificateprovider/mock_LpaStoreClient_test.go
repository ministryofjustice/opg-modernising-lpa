// Code generated by mockery v2.42.0. DO NOT EDIT.

package certificateprovider

import (
	context "context"

	actor "github.com/ministryofjustice/opg-modernising-lpa/internal/actor"

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

// SendCertificateProvider provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockLpaStoreClient) SendCertificateProvider(_a0 context.Context, _a1 string, _a2 *actor.CertificateProviderProvidedDetails) error {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for SendCertificateProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *actor.CertificateProviderProvidedDetails) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockLpaStoreClient_SendCertificateProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendCertificateProvider'
type mockLpaStoreClient_SendCertificateProvider_Call struct {
	*mock.Call
}

// SendCertificateProvider is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
//   - _a2 *actor.CertificateProviderProvidedDetails
func (_e *mockLpaStoreClient_Expecter) SendCertificateProvider(_a0 interface{}, _a1 interface{}, _a2 interface{}) *mockLpaStoreClient_SendCertificateProvider_Call {
	return &mockLpaStoreClient_SendCertificateProvider_Call{Call: _e.mock.On("SendCertificateProvider", _a0, _a1, _a2)}
}

func (_c *mockLpaStoreClient_SendCertificateProvider_Call) Run(run func(_a0 context.Context, _a1 string, _a2 *actor.CertificateProviderProvidedDetails)) *mockLpaStoreClient_SendCertificateProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*actor.CertificateProviderProvidedDetails))
	})
	return _c
}

func (_c *mockLpaStoreClient_SendCertificateProvider_Call) Return(_a0 error) *mockLpaStoreClient_SendCertificateProvider_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockLpaStoreClient_SendCertificateProvider_Call) RunAndReturn(run func(context.Context, string, *actor.CertificateProviderProvidedDetails) error) *mockLpaStoreClient_SendCertificateProvider_Call {
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
