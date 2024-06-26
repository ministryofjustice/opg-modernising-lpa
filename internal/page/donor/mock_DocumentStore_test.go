// Code generated by mockery v2.42.2. DO NOT EDIT.

package donor

import (
	context "context"

	actor "github.com/ministryofjustice/opg-modernising-lpa/internal/actor"

	mock "github.com/stretchr/testify/mock"

	page "github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

// mockDocumentStore is an autogenerated mock type for the DocumentStore type
type mockDocumentStore struct {
	mock.Mock
}

type mockDocumentStore_Expecter struct {
	mock *mock.Mock
}

func (_m *mockDocumentStore) EXPECT() *mockDocumentStore_Expecter {
	return &mockDocumentStore_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *mockDocumentStore) Create(_a0 context.Context, _a1 *actor.DonorProvidedDetails, _a2 string, _a3 []byte) (page.Document, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 page.Document
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *actor.DonorProvidedDetails, string, []byte) (page.Document, error)); ok {
		return rf(_a0, _a1, _a2, _a3)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *actor.DonorProvidedDetails, string, []byte) page.Document); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(page.Document)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *actor.DonorProvidedDetails, string, []byte) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDocumentStore_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type mockDocumentStore_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *actor.DonorProvidedDetails
//   - _a2 string
//   - _a3 []byte
func (_e *mockDocumentStore_Expecter) Create(_a0 interface{}, _a1 interface{}, _a2 interface{}, _a3 interface{}) *mockDocumentStore_Create_Call {
	return &mockDocumentStore_Create_Call{Call: _e.mock.On("Create", _a0, _a1, _a2, _a3)}
}

func (_c *mockDocumentStore_Create_Call) Run(run func(_a0 context.Context, _a1 *actor.DonorProvidedDetails, _a2 string, _a3 []byte)) *mockDocumentStore_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*actor.DonorProvidedDetails), args[2].(string), args[3].([]byte))
	})
	return _c
}

func (_c *mockDocumentStore_Create_Call) Return(_a0 page.Document, _a1 error) *mockDocumentStore_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDocumentStore_Create_Call) RunAndReturn(run func(context.Context, *actor.DonorProvidedDetails, string, []byte) (page.Document, error)) *mockDocumentStore_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: _a0, _a1
func (_m *mockDocumentStore) Delete(_a0 context.Context, _a1 page.Document) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, page.Document) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDocumentStore_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type mockDocumentStore_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 page.Document
func (_e *mockDocumentStore_Expecter) Delete(_a0 interface{}, _a1 interface{}) *mockDocumentStore_Delete_Call {
	return &mockDocumentStore_Delete_Call{Call: _e.mock.On("Delete", _a0, _a1)}
}

func (_c *mockDocumentStore_Delete_Call) Run(run func(_a0 context.Context, _a1 page.Document)) *mockDocumentStore_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(page.Document))
	})
	return _c
}

func (_c *mockDocumentStore_Delete_Call) Return(_a0 error) *mockDocumentStore_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDocumentStore_Delete_Call) RunAndReturn(run func(context.Context, page.Document) error) *mockDocumentStore_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteInfectedDocuments provides a mock function with given fields: _a0, _a1
func (_m *mockDocumentStore) DeleteInfectedDocuments(_a0 context.Context, _a1 page.Documents) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for DeleteInfectedDocuments")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, page.Documents) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDocumentStore_DeleteInfectedDocuments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteInfectedDocuments'
type mockDocumentStore_DeleteInfectedDocuments_Call struct {
	*mock.Call
}

// DeleteInfectedDocuments is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 page.Documents
func (_e *mockDocumentStore_Expecter) DeleteInfectedDocuments(_a0 interface{}, _a1 interface{}) *mockDocumentStore_DeleteInfectedDocuments_Call {
	return &mockDocumentStore_DeleteInfectedDocuments_Call{Call: _e.mock.On("DeleteInfectedDocuments", _a0, _a1)}
}

func (_c *mockDocumentStore_DeleteInfectedDocuments_Call) Run(run func(_a0 context.Context, _a1 page.Documents)) *mockDocumentStore_DeleteInfectedDocuments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(page.Documents))
	})
	return _c
}

func (_c *mockDocumentStore_DeleteInfectedDocuments_Call) Return(_a0 error) *mockDocumentStore_DeleteInfectedDocuments_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDocumentStore_DeleteInfectedDocuments_Call) RunAndReturn(run func(context.Context, page.Documents) error) *mockDocumentStore_DeleteInfectedDocuments_Call {
	_c.Call.Return(run)
	return _c
}

// GetAll provides a mock function with given fields: _a0
func (_m *mockDocumentStore) GetAll(_a0 context.Context) (page.Documents, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetAll")
	}

	var r0 page.Documents
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (page.Documents, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) page.Documents); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(page.Documents)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDocumentStore_GetAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAll'
type mockDocumentStore_GetAll_Call struct {
	*mock.Call
}

// GetAll is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *mockDocumentStore_Expecter) GetAll(_a0 interface{}) *mockDocumentStore_GetAll_Call {
	return &mockDocumentStore_GetAll_Call{Call: _e.mock.On("GetAll", _a0)}
}

func (_c *mockDocumentStore_GetAll_Call) Run(run func(_a0 context.Context)) *mockDocumentStore_GetAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockDocumentStore_GetAll_Call) Return(_a0 page.Documents, _a1 error) *mockDocumentStore_GetAll_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDocumentStore_GetAll_Call) RunAndReturn(run func(context.Context) (page.Documents, error)) *mockDocumentStore_GetAll_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: _a0, _a1
func (_m *mockDocumentStore) Put(_a0 context.Context, _a1 page.Document) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, page.Document) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDocumentStore_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type mockDocumentStore_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 page.Document
func (_e *mockDocumentStore_Expecter) Put(_a0 interface{}, _a1 interface{}) *mockDocumentStore_Put_Call {
	return &mockDocumentStore_Put_Call{Call: _e.mock.On("Put", _a0, _a1)}
}

func (_c *mockDocumentStore_Put_Call) Run(run func(_a0 context.Context, _a1 page.Document)) *mockDocumentStore_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(page.Document))
	})
	return _c
}

func (_c *mockDocumentStore_Put_Call) Return(_a0 error) *mockDocumentStore_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDocumentStore_Put_Call) RunAndReturn(run func(context.Context, page.Document) error) *mockDocumentStore_Put_Call {
	_c.Call.Return(run)
	return _c
}

// Submit provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockDocumentStore) Submit(_a0 context.Context, _a1 *actor.DonorProvidedDetails, _a2 page.Documents) error {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for Submit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *actor.DonorProvidedDetails, page.Documents) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDocumentStore_Submit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Submit'
type mockDocumentStore_Submit_Call struct {
	*mock.Call
}

// Submit is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *actor.DonorProvidedDetails
//   - _a2 page.Documents
func (_e *mockDocumentStore_Expecter) Submit(_a0 interface{}, _a1 interface{}, _a2 interface{}) *mockDocumentStore_Submit_Call {
	return &mockDocumentStore_Submit_Call{Call: _e.mock.On("Submit", _a0, _a1, _a2)}
}

func (_c *mockDocumentStore_Submit_Call) Run(run func(_a0 context.Context, _a1 *actor.DonorProvidedDetails, _a2 page.Documents)) *mockDocumentStore_Submit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*actor.DonorProvidedDetails), args[2].(page.Documents))
	})
	return _c
}

func (_c *mockDocumentStore_Submit_Call) Return(_a0 error) *mockDocumentStore_Submit_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDocumentStore_Submit_Call) RunAndReturn(run func(context.Context, *actor.DonorProvidedDetails, page.Documents) error) *mockDocumentStore_Submit_Call {
	_c.Call.Return(run)
	return _c
}

// newMockDocumentStore creates a new instance of mockDocumentStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockDocumentStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockDocumentStore {
	mock := &mockDocumentStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
