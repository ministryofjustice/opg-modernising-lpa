// Code generated by mockery v2.20.0. DO NOT EDIT.

package page

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// mockDataStore is an autogenerated mock type for the DataStore type
type mockDataStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, pk, sk, v
func (_m *mockDataStore) Get(ctx context.Context, pk string, sk string, v interface{}) error {
	ret := _m.Called(ctx, pk, sk, v)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}) error); ok {
		r0 = rf(ctx, pk, sk, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllByGsi provides a mock function with given fields: ctx, gsi, sk, v
func (_m *mockDataStore) GetAllByGsi(ctx context.Context, gsi string, sk string, v interface{}) error {
	ret := _m.Called(ctx, gsi, sk, v)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}) error); ok {
		r0 = rf(ctx, gsi, sk, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetOneByPartialSK provides a mock function with given fields: ctx, pk, partialSk, v
func (_m *mockDataStore) GetOneByPartialSK(ctx context.Context, pk string, partialSk string, v interface{}) error {
	ret := _m.Called(ctx, pk, partialSk, v)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}) error); ok {
		r0 = rf(ctx, pk, partialSk, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Put provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *mockDataStore) Put(_a0 context.Context, _a1 string, _a2 string, _a3 interface{}) error {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}) error); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTnewMockDataStore interface {
	mock.TestingT
	Cleanup(func())
}

// newMockDataStore creates a new instance of mockDataStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockDataStore(t mockConstructorTestingTnewMockDataStore) *mockDataStore {
	mock := &mockDataStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
