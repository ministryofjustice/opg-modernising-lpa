// Code generated by mockery. DO NOT EDIT.

package telemetry

import (
	context "context"

	cloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	mock "github.com/stretchr/testify/mock"
)

// mockCloudwatchClient is an autogenerated mock type for the CloudwatchClient type
type mockCloudwatchClient struct {
	mock.Mock
}

type mockCloudwatchClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockCloudwatchClient) EXPECT() *mockCloudwatchClient_Expecter {
	return &mockCloudwatchClient_Expecter{mock: &_m.Mock}
}

// PutMetricData provides a mock function with given fields: ctx, params, optFns
func (_m *mockCloudwatchClient) PutMetricData(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for PutMetricData")
	}

	var r0 *cloudwatch.PutMetricDataOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *cloudwatch.PutMetricDataInput, ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error)); ok {
		return rf(ctx, params, optFns...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *cloudwatch.PutMetricDataInput, ...func(*cloudwatch.Options)) *cloudwatch.PutMetricDataOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cloudwatch.PutMetricDataOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *cloudwatch.PutMetricDataInput, ...func(*cloudwatch.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockCloudwatchClient_PutMetricData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PutMetricData'
type mockCloudwatchClient_PutMetricData_Call struct {
	*mock.Call
}

// PutMetricData is a helper method to define mock.On call
//   - ctx context.Context
//   - params *cloudwatch.PutMetricDataInput
//   - optFns ...func(*cloudwatch.Options)
func (_e *mockCloudwatchClient_Expecter) PutMetricData(ctx interface{}, params interface{}, optFns ...interface{}) *mockCloudwatchClient_PutMetricData_Call {
	return &mockCloudwatchClient_PutMetricData_Call{Call: _e.mock.On("PutMetricData",
		append([]interface{}{ctx, params}, optFns...)...)}
}

func (_c *mockCloudwatchClient_PutMetricData_Call) Run(run func(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options))) *mockCloudwatchClient_PutMetricData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*cloudwatch.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*cloudwatch.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*cloudwatch.PutMetricDataInput), variadicArgs...)
	})
	return _c
}

func (_c *mockCloudwatchClient_PutMetricData_Call) Return(_a0 *cloudwatch.PutMetricDataOutput, _a1 error) *mockCloudwatchClient_PutMetricData_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockCloudwatchClient_PutMetricData_Call) RunAndReturn(run func(context.Context, *cloudwatch.PutMetricDataInput, ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error)) *mockCloudwatchClient_PutMetricData_Call {
	_c.Call.Return(run)
	return _c
}

// newMockCloudwatchClient creates a new instance of mockCloudwatchClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockCloudwatchClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockCloudwatchClient {
	mock := &mockCloudwatchClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}