// Code generated by mockery v2.46.1. DO NOT EDIT.

package voucherpage

import (
	context "context"

	lpadata "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	mock "github.com/stretchr/testify/mock"

	voucherdata "github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

// mockVouchFailer is an autogenerated mock type for the vouchFailer type
type mockVouchFailer struct {
	mock.Mock
}

type mockVouchFailer_Expecter struct {
	mock *mock.Mock
}

func (_m *mockVouchFailer) EXPECT() *mockVouchFailer_Expecter {
	return &mockVouchFailer_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: ctx, provided, lpa
func (_m *mockVouchFailer) Execute(ctx context.Context, provided *voucherdata.Provided, lpa *lpadata.Lpa) error {
	ret := _m.Called(ctx, provided, lpa)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *voucherdata.Provided, *lpadata.Lpa) error); ok {
		r0 = rf(ctx, provided, lpa)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockVouchFailer_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type mockVouchFailer_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
//   - provided *voucherdata.Provided
//   - lpa *lpadata.Lpa
func (_e *mockVouchFailer_Expecter) Execute(ctx interface{}, provided interface{}, lpa interface{}) *mockVouchFailer_Execute_Call {
	return &mockVouchFailer_Execute_Call{Call: _e.mock.On("Execute", ctx, provided, lpa)}
}

func (_c *mockVouchFailer_Execute_Call) Run(run func(ctx context.Context, provided *voucherdata.Provided, lpa *lpadata.Lpa)) *mockVouchFailer_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*voucherdata.Provided), args[2].(*lpadata.Lpa))
	})
	return _c
}

func (_c *mockVouchFailer_Execute_Call) Return(_a0 error) *mockVouchFailer_Execute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockVouchFailer_Execute_Call) RunAndReturn(run func(context.Context, *voucherdata.Provided, *lpadata.Lpa) error) *mockVouchFailer_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// newMockVouchFailer creates a new instance of mockVouchFailer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockVouchFailer(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockVouchFailer {
	mock := &mockVouchFailer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}