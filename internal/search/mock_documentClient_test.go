// Code generated by mockery. DO NOT EDIT.

package search

import (
	context "context"

	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	mock "github.com/stretchr/testify/mock"
)

// mockDocumentClient is an autogenerated mock type for the documentClient type
type mockDocumentClient struct {
	mock.Mock
}

type mockDocumentClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockDocumentClient) EXPECT() *mockDocumentClient_Expecter {
	return &mockDocumentClient_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: ctx, req
func (_m *mockDocumentClient) Delete(ctx context.Context, req opensearchapi.DocumentDeleteReq) (*opensearchapi.DocumentDeleteResp, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 *opensearchapi.DocumentDeleteResp
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, opensearchapi.DocumentDeleteReq) (*opensearchapi.DocumentDeleteResp, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, opensearchapi.DocumentDeleteReq) *opensearchapi.DocumentDeleteResp); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*opensearchapi.DocumentDeleteResp)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, opensearchapi.DocumentDeleteReq) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDocumentClient_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type mockDocumentClient_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - req opensearchapi.DocumentDeleteReq
func (_e *mockDocumentClient_Expecter) Delete(ctx interface{}, req interface{}) *mockDocumentClient_Delete_Call {
	return &mockDocumentClient_Delete_Call{Call: _e.mock.On("Delete", ctx, req)}
}

func (_c *mockDocumentClient_Delete_Call) Run(run func(ctx context.Context, req opensearchapi.DocumentDeleteReq)) *mockDocumentClient_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(opensearchapi.DocumentDeleteReq))
	})
	return _c
}

func (_c *mockDocumentClient_Delete_Call) Return(_a0 *opensearchapi.DocumentDeleteResp, _a1 error) *mockDocumentClient_Delete_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDocumentClient_Delete_Call) RunAndReturn(run func(context.Context, opensearchapi.DocumentDeleteReq) (*opensearchapi.DocumentDeleteResp, error)) *mockDocumentClient_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// newMockDocumentClient creates a new instance of mockDocumentClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockDocumentClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockDocumentClient {
	mock := &mockDocumentClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
