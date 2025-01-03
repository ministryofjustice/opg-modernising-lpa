// Code generated by mockery. DO NOT EDIT.

package donorpage

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// mockS3Client is an autogenerated mock type for the S3Client type
type mockS3Client struct {
	mock.Mock
}

type mockS3Client_Expecter struct {
	mock *mock.Mock
}

func (_m *mockS3Client) EXPECT() *mockS3Client_Expecter {
	return &mockS3Client_Expecter{mock: &_m.Mock}
}

// DeleteObject provides a mock function with given fields: _a0, _a1
func (_m *mockS3Client) DeleteObject(_a0 context.Context, _a1 string) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for DeleteObject")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockS3Client_DeleteObject_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteObject'
type mockS3Client_DeleteObject_Call struct {
	*mock.Call
}

// DeleteObject is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
func (_e *mockS3Client_Expecter) DeleteObject(_a0 interface{}, _a1 interface{}) *mockS3Client_DeleteObject_Call {
	return &mockS3Client_DeleteObject_Call{Call: _e.mock.On("DeleteObject", _a0, _a1)}
}

func (_c *mockS3Client_DeleteObject_Call) Run(run func(_a0 context.Context, _a1 string)) *mockS3Client_DeleteObject_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *mockS3Client_DeleteObject_Call) Return(_a0 error) *mockS3Client_DeleteObject_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockS3Client_DeleteObject_Call) RunAndReturn(run func(context.Context, string) error) *mockS3Client_DeleteObject_Call {
	_c.Call.Return(run)
	return _c
}

// PutObject provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockS3Client) PutObject(_a0 context.Context, _a1 string, _a2 []byte) error {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for PutObject")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockS3Client_PutObject_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PutObject'
type mockS3Client_PutObject_Call struct {
	*mock.Call
}

// PutObject is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
//   - _a2 []byte
func (_e *mockS3Client_Expecter) PutObject(_a0 interface{}, _a1 interface{}, _a2 interface{}) *mockS3Client_PutObject_Call {
	return &mockS3Client_PutObject_Call{Call: _e.mock.On("PutObject", _a0, _a1, _a2)}
}

func (_c *mockS3Client_PutObject_Call) Run(run func(_a0 context.Context, _a1 string, _a2 []byte)) *mockS3Client_PutObject_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]byte))
	})
	return _c
}

func (_c *mockS3Client_PutObject_Call) Return(_a0 error) *mockS3Client_PutObject_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockS3Client_PutObject_Call) RunAndReturn(run func(context.Context, string, []byte) error) *mockS3Client_PutObject_Call {
	_c.Call.Return(run)
	return _c
}

// newMockS3Client creates a new instance of mockS3Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockS3Client(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockS3Client {
	mock := &mockS3Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
