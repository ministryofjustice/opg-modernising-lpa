// Code generated by mockery v2.42.2. DO NOT EDIT.

package sharecode

import (
	context "context"

	dynamo "github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	mock "github.com/stretchr/testify/mock"
)

// mockDynamoClient is an autogenerated mock type for the DynamoClient type
type mockDynamoClient struct {
	mock.Mock
}

type mockDynamoClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockDynamoClient) EXPECT() *mockDynamoClient_Expecter {
	return &mockDynamoClient_Expecter{mock: &_m.Mock}
}

// DeleteOne provides a mock function with given fields: ctx, pk, sk
func (_m *mockDynamoClient) DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error {
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

// mockDynamoClient_DeleteOne_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteOne'
type mockDynamoClient_DeleteOne_Call struct {
	*mock.Call
}

// DeleteOne is a helper method to define mock.On call
//   - ctx context.Context
//   - pk dynamo.PK
//   - sk dynamo.SK
func (_e *mockDynamoClient_Expecter) DeleteOne(ctx interface{}, pk interface{}, sk interface{}) *mockDynamoClient_DeleteOne_Call {
	return &mockDynamoClient_DeleteOne_Call{Call: _e.mock.On("DeleteOne", ctx, pk, sk)}
}

func (_c *mockDynamoClient_DeleteOne_Call) Run(run func(ctx context.Context, pk dynamo.PK, sk dynamo.SK)) *mockDynamoClient_DeleteOne_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dynamo.PK), args[2].(dynamo.SK))
	})
	return _c
}

func (_c *mockDynamoClient_DeleteOne_Call) Return(_a0 error) *mockDynamoClient_DeleteOne_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDynamoClient_DeleteOne_Call) RunAndReturn(run func(context.Context, dynamo.PK, dynamo.SK) error) *mockDynamoClient_DeleteOne_Call {
	_c.Call.Return(run)
	return _c
}

// One provides a mock function with given fields: ctx, pk, sk, v
func (_m *mockDynamoClient) One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
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

// mockDynamoClient_One_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'One'
type mockDynamoClient_One_Call struct {
	*mock.Call
}

// One is a helper method to define mock.On call
//   - ctx context.Context
//   - pk dynamo.PK
//   - sk dynamo.SK
//   - v interface{}
func (_e *mockDynamoClient_Expecter) One(ctx interface{}, pk interface{}, sk interface{}, v interface{}) *mockDynamoClient_One_Call {
	return &mockDynamoClient_One_Call{Call: _e.mock.On("One", ctx, pk, sk, v)}
}

func (_c *mockDynamoClient_One_Call) Run(run func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{})) *mockDynamoClient_One_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dynamo.PK), args[2].(dynamo.SK), args[3].(interface{}))
	})
	return _c
}

func (_c *mockDynamoClient_One_Call) Return(_a0 error) *mockDynamoClient_One_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDynamoClient_One_Call) RunAndReturn(run func(context.Context, dynamo.PK, dynamo.SK, interface{}) error) *mockDynamoClient_One_Call {
	_c.Call.Return(run)
	return _c
}

// OneByPK provides a mock function with given fields: ctx, pk, v
func (_m *mockDynamoClient) OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error {
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

// mockDynamoClient_OneByPK_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OneByPK'
type mockDynamoClient_OneByPK_Call struct {
	*mock.Call
}

// OneByPK is a helper method to define mock.On call
//   - ctx context.Context
//   - pk dynamo.PK
//   - v interface{}
func (_e *mockDynamoClient_Expecter) OneByPK(ctx interface{}, pk interface{}, v interface{}) *mockDynamoClient_OneByPK_Call {
	return &mockDynamoClient_OneByPK_Call{Call: _e.mock.On("OneByPK", ctx, pk, v)}
}

func (_c *mockDynamoClient_OneByPK_Call) Run(run func(ctx context.Context, pk dynamo.PK, v interface{})) *mockDynamoClient_OneByPK_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dynamo.PK), args[2].(interface{}))
	})
	return _c
}

func (_c *mockDynamoClient_OneByPK_Call) Return(_a0 error) *mockDynamoClient_OneByPK_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDynamoClient_OneByPK_Call) RunAndReturn(run func(context.Context, dynamo.PK, interface{}) error) *mockDynamoClient_OneByPK_Call {
	_c.Call.Return(run)
	return _c
}

// OneBySK provides a mock function with given fields: ctx, sk, v
func (_m *mockDynamoClient) OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error {
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

// mockDynamoClient_OneBySK_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OneBySK'
type mockDynamoClient_OneBySK_Call struct {
	*mock.Call
}

// OneBySK is a helper method to define mock.On call
//   - ctx context.Context
//   - sk dynamo.SK
//   - v interface{}
func (_e *mockDynamoClient_Expecter) OneBySK(ctx interface{}, sk interface{}, v interface{}) *mockDynamoClient_OneBySK_Call {
	return &mockDynamoClient_OneBySK_Call{Call: _e.mock.On("OneBySK", ctx, sk, v)}
}

func (_c *mockDynamoClient_OneBySK_Call) Run(run func(ctx context.Context, sk dynamo.SK, v interface{})) *mockDynamoClient_OneBySK_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dynamo.SK), args[2].(interface{}))
	})
	return _c
}

func (_c *mockDynamoClient_OneBySK_Call) Return(_a0 error) *mockDynamoClient_OneBySK_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDynamoClient_OneBySK_Call) RunAndReturn(run func(context.Context, dynamo.SK, interface{}) error) *mockDynamoClient_OneBySK_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: ctx, v
func (_m *mockDynamoClient) Put(ctx context.Context, v interface{}) error {
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

// mockDynamoClient_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type mockDynamoClient_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - ctx context.Context
//   - v interface{}
func (_e *mockDynamoClient_Expecter) Put(ctx interface{}, v interface{}) *mockDynamoClient_Put_Call {
	return &mockDynamoClient_Put_Call{Call: _e.mock.On("Put", ctx, v)}
}

func (_c *mockDynamoClient_Put_Call) Run(run func(ctx context.Context, v interface{})) *mockDynamoClient_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}))
	})
	return _c
}

func (_c *mockDynamoClient_Put_Call) Return(_a0 error) *mockDynamoClient_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDynamoClient_Put_Call) RunAndReturn(run func(context.Context, interface{}) error) *mockDynamoClient_Put_Call {
	_c.Call.Return(run)
	return _c
}

// newMockDynamoClient creates a new instance of mockDynamoClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockDynamoClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockDynamoClient {
	mock := &mockDynamoClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}