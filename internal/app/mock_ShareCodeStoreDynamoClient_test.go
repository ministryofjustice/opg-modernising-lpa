// Code generated by mockery v2.42.2. DO NOT EDIT.

package app

import (
	context "context"

	dynamo "github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	mock "github.com/stretchr/testify/mock"
)

// mockShareCodeStoreDynamoClient is an autogenerated mock type for the ShareCodeStoreDynamoClient type
type mockShareCodeStoreDynamoClient struct {
	mock.Mock
}

type mockShareCodeStoreDynamoClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockShareCodeStoreDynamoClient) EXPECT() *mockShareCodeStoreDynamoClient_Expecter {
	return &mockShareCodeStoreDynamoClient_Expecter{mock: &_m.Mock}
}

// DeleteOne provides a mock function with given fields: ctx, pk, sk
func (_m *mockShareCodeStoreDynamoClient) DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error {
	ret := _m.Called(ctx, pk, sk)

	if len(ret) == 0 {
		panic("no return value specified for DeleteOne")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, dynamo.PK, dynamo.SK) error); ok {
		r0 = rf(ctx, pk, sk)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeStoreDynamoClient_DeleteOne_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteOne'
type mockShareCodeStoreDynamoClient_DeleteOne_Call struct {
	*mock.Call
}

// DeleteOne is a helper method to define mock.On call
//   - ctx context.Context
//   - pk dynamo.PK
//   - sk dynamo.SK
func (_e *mockShareCodeStoreDynamoClient_Expecter) DeleteOne(ctx interface{}, pk interface{}, sk interface{}) *mockShareCodeStoreDynamoClient_DeleteOne_Call {
	return &mockShareCodeStoreDynamoClient_DeleteOne_Call{Call: _e.mock.On("DeleteOne", ctx, pk, sk)}
}

func (_c *mockShareCodeStoreDynamoClient_DeleteOne_Call) Run(run func(ctx context.Context, pk dynamo.PK, sk dynamo.SK)) *mockShareCodeStoreDynamoClient_DeleteOne_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dynamo.PK), args[2].(dynamo.SK))
	})
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_DeleteOne_Call) Return(_a0 error) *mockShareCodeStoreDynamoClient_DeleteOne_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_DeleteOne_Call) RunAndReturn(run func(context.Context, dynamo.PK, dynamo.SK) error) *mockShareCodeStoreDynamoClient_DeleteOne_Call {
	_c.Call.Return(run)
	return _c
}

// One provides a mock function with given fields: ctx, pk, sk, v
func (_m *mockShareCodeStoreDynamoClient) One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
	ret := _m.Called(ctx, pk, sk, v)

	if len(ret) == 0 {
		panic("no return value specified for One")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, dynamo.PK, dynamo.SK, interface{}) error); ok {
		r0 = rf(ctx, pk, sk, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeStoreDynamoClient_One_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'One'
type mockShareCodeStoreDynamoClient_One_Call struct {
	*mock.Call
}

// One is a helper method to define mock.On call
//   - ctx context.Context
//   - pk dynamo.PK
//   - sk dynamo.SK
//   - v interface{}
func (_e *mockShareCodeStoreDynamoClient_Expecter) One(ctx interface{}, pk interface{}, sk interface{}, v interface{}) *mockShareCodeStoreDynamoClient_One_Call {
	return &mockShareCodeStoreDynamoClient_One_Call{Call: _e.mock.On("One", ctx, pk, sk, v)}
}

func (_c *mockShareCodeStoreDynamoClient_One_Call) Run(run func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{})) *mockShareCodeStoreDynamoClient_One_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dynamo.PK), args[2].(dynamo.SK), args[3].(interface{}))
	})
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_One_Call) Return(_a0 error) *mockShareCodeStoreDynamoClient_One_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_One_Call) RunAndReturn(run func(context.Context, dynamo.PK, dynamo.SK, interface{}) error) *mockShareCodeStoreDynamoClient_One_Call {
	_c.Call.Return(run)
	return _c
}

// OneByPK provides a mock function with given fields: ctx, pk, v
func (_m *mockShareCodeStoreDynamoClient) OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error {
	ret := _m.Called(ctx, pk, v)

	if len(ret) == 0 {
		panic("no return value specified for OneByPK")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, dynamo.PK, interface{}) error); ok {
		r0 = rf(ctx, pk, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeStoreDynamoClient_OneByPK_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OneByPK'
type mockShareCodeStoreDynamoClient_OneByPK_Call struct {
	*mock.Call
}

// OneByPK is a helper method to define mock.On call
//   - ctx context.Context
//   - pk dynamo.PK
//   - v interface{}
func (_e *mockShareCodeStoreDynamoClient_Expecter) OneByPK(ctx interface{}, pk interface{}, v interface{}) *mockShareCodeStoreDynamoClient_OneByPK_Call {
	return &mockShareCodeStoreDynamoClient_OneByPK_Call{Call: _e.mock.On("OneByPK", ctx, pk, v)}
}

func (_c *mockShareCodeStoreDynamoClient_OneByPK_Call) Run(run func(ctx context.Context, pk dynamo.PK, v interface{})) *mockShareCodeStoreDynamoClient_OneByPK_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dynamo.PK), args[2].(interface{}))
	})
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_OneByPK_Call) Return(_a0 error) *mockShareCodeStoreDynamoClient_OneByPK_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_OneByPK_Call) RunAndReturn(run func(context.Context, dynamo.PK, interface{}) error) *mockShareCodeStoreDynamoClient_OneByPK_Call {
	_c.Call.Return(run)
	return _c
}

// OneBySK provides a mock function with given fields: ctx, sk, v
func (_m *mockShareCodeStoreDynamoClient) OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error {
	ret := _m.Called(ctx, sk, v)

	if len(ret) == 0 {
		panic("no return value specified for OneBySK")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, dynamo.SK, interface{}) error); ok {
		r0 = rf(ctx, sk, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeStoreDynamoClient_OneBySK_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OneBySK'
type mockShareCodeStoreDynamoClient_OneBySK_Call struct {
	*mock.Call
}

// OneBySK is a helper method to define mock.On call
//   - ctx context.Context
//   - sk dynamo.SK
//   - v interface{}
func (_e *mockShareCodeStoreDynamoClient_Expecter) OneBySK(ctx interface{}, sk interface{}, v interface{}) *mockShareCodeStoreDynamoClient_OneBySK_Call {
	return &mockShareCodeStoreDynamoClient_OneBySK_Call{Call: _e.mock.On("OneBySK", ctx, sk, v)}
}

func (_c *mockShareCodeStoreDynamoClient_OneBySK_Call) Run(run func(ctx context.Context, sk dynamo.SK, v interface{})) *mockShareCodeStoreDynamoClient_OneBySK_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dynamo.SK), args[2].(interface{}))
	})
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_OneBySK_Call) Return(_a0 error) *mockShareCodeStoreDynamoClient_OneBySK_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_OneBySK_Call) RunAndReturn(run func(context.Context, dynamo.SK, interface{}) error) *mockShareCodeStoreDynamoClient_OneBySK_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: ctx, v
func (_m *mockShareCodeStoreDynamoClient) Put(ctx context.Context, v interface{}) error {
	ret := _m.Called(ctx, v)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) error); ok {
		r0 = rf(ctx, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockShareCodeStoreDynamoClient_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type mockShareCodeStoreDynamoClient_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - ctx context.Context
//   - v interface{}
func (_e *mockShareCodeStoreDynamoClient_Expecter) Put(ctx interface{}, v interface{}) *mockShareCodeStoreDynamoClient_Put_Call {
	return &mockShareCodeStoreDynamoClient_Put_Call{Call: _e.mock.On("Put", ctx, v)}
}

func (_c *mockShareCodeStoreDynamoClient_Put_Call) Run(run func(ctx context.Context, v interface{})) *mockShareCodeStoreDynamoClient_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}))
	})
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_Put_Call) Return(_a0 error) *mockShareCodeStoreDynamoClient_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockShareCodeStoreDynamoClient_Put_Call) RunAndReturn(run func(context.Context, interface{}) error) *mockShareCodeStoreDynamoClient_Put_Call {
	_c.Call.Return(run)
	return _c
}

// newMockShareCodeStoreDynamoClient creates a new instance of mockShareCodeStoreDynamoClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockShareCodeStoreDynamoClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockShareCodeStoreDynamoClient {
	mock := &mockShareCodeStoreDynamoClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
