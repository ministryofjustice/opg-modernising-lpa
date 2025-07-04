// Code generated by mockery. DO NOT EDIT.

package supporterpage

import (
	context "context"

	accesscodedata "github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"

	mock "github.com/stretchr/testify/mock"

	supporterdata "github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
)

// mockMemberStore is an autogenerated mock type for the MemberStore type
type mockMemberStore struct {
	mock.Mock
}

type mockMemberStore_Expecter struct {
	mock *mock.Mock
}

func (_m *mockMemberStore) EXPECT() *mockMemberStore_Expecter {
	return &mockMemberStore_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, firstNames, lastName
func (_m *mockMemberStore) Create(ctx context.Context, firstNames string, lastName string) (*supporterdata.Member, error) {
	ret := _m.Called(ctx, firstNames, lastName)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *supporterdata.Member
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*supporterdata.Member, error)); ok {
		return rf(ctx, firstNames, lastName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *supporterdata.Member); ok {
		r0 = rf(ctx, firstNames, lastName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*supporterdata.Member)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, firstNames, lastName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMemberStore_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type mockMemberStore_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - firstNames string
//   - lastName string
func (_e *mockMemberStore_Expecter) Create(ctx interface{}, firstNames interface{}, lastName interface{}) *mockMemberStore_Create_Call {
	return &mockMemberStore_Create_Call{Call: _e.mock.On("Create", ctx, firstNames, lastName)}
}

func (_c *mockMemberStore_Create_Call) Run(run func(ctx context.Context, firstNames string, lastName string)) *mockMemberStore_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *mockMemberStore_Create_Call) Return(_a0 *supporterdata.Member, _a1 error) *mockMemberStore_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMemberStore_Create_Call) RunAndReturn(run func(context.Context, string, string) (*supporterdata.Member, error)) *mockMemberStore_Create_Call {
	_c.Call.Return(run)
	return _c
}

// CreateFromInvite provides a mock function with given fields: ctx, invite
func (_m *mockMemberStore) CreateFromInvite(ctx context.Context, invite *supporterdata.MemberInvite) error {
	ret := _m.Called(ctx, invite)

	if len(ret) == 0 {
		panic("no return value specified for CreateFromInvite")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *supporterdata.MemberInvite) error); ok {
		r0 = rf(ctx, invite)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockMemberStore_CreateFromInvite_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateFromInvite'
type mockMemberStore_CreateFromInvite_Call struct {
	*mock.Call
}

// CreateFromInvite is a helper method to define mock.On call
//   - ctx context.Context
//   - invite *supporterdata.MemberInvite
func (_e *mockMemberStore_Expecter) CreateFromInvite(ctx interface{}, invite interface{}) *mockMemberStore_CreateFromInvite_Call {
	return &mockMemberStore_CreateFromInvite_Call{Call: _e.mock.On("CreateFromInvite", ctx, invite)}
}

func (_c *mockMemberStore_CreateFromInvite_Call) Run(run func(ctx context.Context, invite *supporterdata.MemberInvite)) *mockMemberStore_CreateFromInvite_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*supporterdata.MemberInvite))
	})
	return _c
}

func (_c *mockMemberStore_CreateFromInvite_Call) Return(_a0 error) *mockMemberStore_CreateFromInvite_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockMemberStore_CreateFromInvite_Call) RunAndReturn(run func(context.Context, *supporterdata.MemberInvite) error) *mockMemberStore_CreateFromInvite_Call {
	_c.Call.Return(run)
	return _c
}

// CreateMemberInvite provides a mock function with given fields: ctx, organisation, firstNames, lastname, email, code, permission
func (_m *mockMemberStore) CreateMemberInvite(ctx context.Context, organisation *supporterdata.Organisation, firstNames string, lastname string, email string, code accesscodedata.Hashed, permission supporterdata.Permission) error {
	ret := _m.Called(ctx, organisation, firstNames, lastname, email, code, permission)

	if len(ret) == 0 {
		panic("no return value specified for CreateMemberInvite")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *supporterdata.Organisation, string, string, string, accesscodedata.Hashed, supporterdata.Permission) error); ok {
		r0 = rf(ctx, organisation, firstNames, lastname, email, code, permission)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockMemberStore_CreateMemberInvite_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateMemberInvite'
type mockMemberStore_CreateMemberInvite_Call struct {
	*mock.Call
}

// CreateMemberInvite is a helper method to define mock.On call
//   - ctx context.Context
//   - organisation *supporterdata.Organisation
//   - firstNames string
//   - lastname string
//   - email string
//   - code accesscodedata.Hashed
//   - permission supporterdata.Permission
func (_e *mockMemberStore_Expecter) CreateMemberInvite(ctx interface{}, organisation interface{}, firstNames interface{}, lastname interface{}, email interface{}, code interface{}, permission interface{}) *mockMemberStore_CreateMemberInvite_Call {
	return &mockMemberStore_CreateMemberInvite_Call{Call: _e.mock.On("CreateMemberInvite", ctx, organisation, firstNames, lastname, email, code, permission)}
}

func (_c *mockMemberStore_CreateMemberInvite_Call) Run(run func(ctx context.Context, organisation *supporterdata.Organisation, firstNames string, lastname string, email string, code accesscodedata.Hashed, permission supporterdata.Permission)) *mockMemberStore_CreateMemberInvite_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*supporterdata.Organisation), args[2].(string), args[3].(string), args[4].(string), args[5].(accesscodedata.Hashed), args[6].(supporterdata.Permission))
	})
	return _c
}

func (_c *mockMemberStore_CreateMemberInvite_Call) Return(_a0 error) *mockMemberStore_CreateMemberInvite_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockMemberStore_CreateMemberInvite_Call) RunAndReturn(run func(context.Context, *supporterdata.Organisation, string, string, string, accesscodedata.Hashed, supporterdata.Permission) error) *mockMemberStore_CreateMemberInvite_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteMemberInvite provides a mock function with given fields: ctx, organisationID, email
func (_m *mockMemberStore) DeleteMemberInvite(ctx context.Context, organisationID string, email string) error {
	ret := _m.Called(ctx, organisationID, email)

	if len(ret) == 0 {
		panic("no return value specified for DeleteMemberInvite")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, organisationID, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockMemberStore_DeleteMemberInvite_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteMemberInvite'
type mockMemberStore_DeleteMemberInvite_Call struct {
	*mock.Call
}

// DeleteMemberInvite is a helper method to define mock.On call
//   - ctx context.Context
//   - organisationID string
//   - email string
func (_e *mockMemberStore_Expecter) DeleteMemberInvite(ctx interface{}, organisationID interface{}, email interface{}) *mockMemberStore_DeleteMemberInvite_Call {
	return &mockMemberStore_DeleteMemberInvite_Call{Call: _e.mock.On("DeleteMemberInvite", ctx, organisationID, email)}
}

func (_c *mockMemberStore_DeleteMemberInvite_Call) Run(run func(ctx context.Context, organisationID string, email string)) *mockMemberStore_DeleteMemberInvite_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *mockMemberStore_DeleteMemberInvite_Call) Return(_a0 error) *mockMemberStore_DeleteMemberInvite_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockMemberStore_DeleteMemberInvite_Call) RunAndReturn(run func(context.Context, string, string) error) *mockMemberStore_DeleteMemberInvite_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx
func (_m *mockMemberStore) Get(ctx context.Context) (*supporterdata.Member, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *supporterdata.Member
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*supporterdata.Member, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *supporterdata.Member); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*supporterdata.Member)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMemberStore_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type mockMemberStore_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockMemberStore_Expecter) Get(ctx interface{}) *mockMemberStore_Get_Call {
	return &mockMemberStore_Get_Call{Call: _e.mock.On("Get", ctx)}
}

func (_c *mockMemberStore_Get_Call) Run(run func(ctx context.Context)) *mockMemberStore_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockMemberStore_Get_Call) Return(_a0 *supporterdata.Member, _a1 error) *mockMemberStore_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMemberStore_Get_Call) RunAndReturn(run func(context.Context) (*supporterdata.Member, error)) *mockMemberStore_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetAll provides a mock function with given fields: ctx
func (_m *mockMemberStore) GetAll(ctx context.Context) ([]*supporterdata.Member, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetAll")
	}

	var r0 []*supporterdata.Member
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*supporterdata.Member, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*supporterdata.Member); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*supporterdata.Member)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMemberStore_GetAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAll'
type mockMemberStore_GetAll_Call struct {
	*mock.Call
}

// GetAll is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockMemberStore_Expecter) GetAll(ctx interface{}) *mockMemberStore_GetAll_Call {
	return &mockMemberStore_GetAll_Call{Call: _e.mock.On("GetAll", ctx)}
}

func (_c *mockMemberStore_GetAll_Call) Run(run func(ctx context.Context)) *mockMemberStore_GetAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockMemberStore_GetAll_Call) Return(_a0 []*supporterdata.Member, _a1 error) *mockMemberStore_GetAll_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMemberStore_GetAll_Call) RunAndReturn(run func(context.Context) ([]*supporterdata.Member, error)) *mockMemberStore_GetAll_Call {
	_c.Call.Return(run)
	return _c
}

// GetAny provides a mock function with given fields: ctx
func (_m *mockMemberStore) GetAny(ctx context.Context) (*supporterdata.Member, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetAny")
	}

	var r0 *supporterdata.Member
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*supporterdata.Member, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *supporterdata.Member); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*supporterdata.Member)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMemberStore_GetAny_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAny'
type mockMemberStore_GetAny_Call struct {
	*mock.Call
}

// GetAny is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockMemberStore_Expecter) GetAny(ctx interface{}) *mockMemberStore_GetAny_Call {
	return &mockMemberStore_GetAny_Call{Call: _e.mock.On("GetAny", ctx)}
}

func (_c *mockMemberStore_GetAny_Call) Run(run func(ctx context.Context)) *mockMemberStore_GetAny_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockMemberStore_GetAny_Call) Return(_a0 *supporterdata.Member, _a1 error) *mockMemberStore_GetAny_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMemberStore_GetAny_Call) RunAndReturn(run func(context.Context) (*supporterdata.Member, error)) *mockMemberStore_GetAny_Call {
	_c.Call.Return(run)
	return _c
}

// GetByID provides a mock function with given fields: ctx, memberID
func (_m *mockMemberStore) GetByID(ctx context.Context, memberID string) (*supporterdata.Member, error) {
	ret := _m.Called(ctx, memberID)

	if len(ret) == 0 {
		panic("no return value specified for GetByID")
	}

	var r0 *supporterdata.Member
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*supporterdata.Member, error)); ok {
		return rf(ctx, memberID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *supporterdata.Member); ok {
		r0 = rf(ctx, memberID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*supporterdata.Member)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, memberID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMemberStore_GetByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByID'
type mockMemberStore_GetByID_Call struct {
	*mock.Call
}

// GetByID is a helper method to define mock.On call
//   - ctx context.Context
//   - memberID string
func (_e *mockMemberStore_Expecter) GetByID(ctx interface{}, memberID interface{}) *mockMemberStore_GetByID_Call {
	return &mockMemberStore_GetByID_Call{Call: _e.mock.On("GetByID", ctx, memberID)}
}

func (_c *mockMemberStore_GetByID_Call) Run(run func(ctx context.Context, memberID string)) *mockMemberStore_GetByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *mockMemberStore_GetByID_Call) Return(_a0 *supporterdata.Member, _a1 error) *mockMemberStore_GetByID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMemberStore_GetByID_Call) RunAndReturn(run func(context.Context, string) (*supporterdata.Member, error)) *mockMemberStore_GetByID_Call {
	_c.Call.Return(run)
	return _c
}

// InvitedMember provides a mock function with given fields: ctx
func (_m *mockMemberStore) InvitedMember(ctx context.Context) (*supporterdata.MemberInvite, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for InvitedMember")
	}

	var r0 *supporterdata.MemberInvite
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*supporterdata.MemberInvite, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *supporterdata.MemberInvite); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*supporterdata.MemberInvite)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMemberStore_InvitedMember_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InvitedMember'
type mockMemberStore_InvitedMember_Call struct {
	*mock.Call
}

// InvitedMember is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockMemberStore_Expecter) InvitedMember(ctx interface{}) *mockMemberStore_InvitedMember_Call {
	return &mockMemberStore_InvitedMember_Call{Call: _e.mock.On("InvitedMember", ctx)}
}

func (_c *mockMemberStore_InvitedMember_Call) Run(run func(ctx context.Context)) *mockMemberStore_InvitedMember_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockMemberStore_InvitedMember_Call) Return(_a0 *supporterdata.MemberInvite, _a1 error) *mockMemberStore_InvitedMember_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMemberStore_InvitedMember_Call) RunAndReturn(run func(context.Context) (*supporterdata.MemberInvite, error)) *mockMemberStore_InvitedMember_Call {
	_c.Call.Return(run)
	return _c
}

// InvitedMembers provides a mock function with given fields: ctx
func (_m *mockMemberStore) InvitedMembers(ctx context.Context) ([]*supporterdata.MemberInvite, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for InvitedMembers")
	}

	var r0 []*supporterdata.MemberInvite
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*supporterdata.MemberInvite, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*supporterdata.MemberInvite); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*supporterdata.MemberInvite)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMemberStore_InvitedMembers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InvitedMembers'
type mockMemberStore_InvitedMembers_Call struct {
	*mock.Call
}

// InvitedMembers is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockMemberStore_Expecter) InvitedMembers(ctx interface{}) *mockMemberStore_InvitedMembers_Call {
	return &mockMemberStore_InvitedMembers_Call{Call: _e.mock.On("InvitedMembers", ctx)}
}

func (_c *mockMemberStore_InvitedMembers_Call) Run(run func(ctx context.Context)) *mockMemberStore_InvitedMembers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockMemberStore_InvitedMembers_Call) Return(_a0 []*supporterdata.MemberInvite, _a1 error) *mockMemberStore_InvitedMembers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMemberStore_InvitedMembers_Call) RunAndReturn(run func(context.Context) ([]*supporterdata.MemberInvite, error)) *mockMemberStore_InvitedMembers_Call {
	_c.Call.Return(run)
	return _c
}

// InvitedMembersByEmail provides a mock function with given fields: ctx
func (_m *mockMemberStore) InvitedMembersByEmail(ctx context.Context) ([]*supporterdata.MemberInvite, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for InvitedMembersByEmail")
	}

	var r0 []*supporterdata.MemberInvite
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*supporterdata.MemberInvite, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*supporterdata.MemberInvite); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*supporterdata.MemberInvite)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockMemberStore_InvitedMembersByEmail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InvitedMembersByEmail'
type mockMemberStore_InvitedMembersByEmail_Call struct {
	*mock.Call
}

// InvitedMembersByEmail is a helper method to define mock.On call
//   - ctx context.Context
func (_e *mockMemberStore_Expecter) InvitedMembersByEmail(ctx interface{}) *mockMemberStore_InvitedMembersByEmail_Call {
	return &mockMemberStore_InvitedMembersByEmail_Call{Call: _e.mock.On("InvitedMembersByEmail", ctx)}
}

func (_c *mockMemberStore_InvitedMembersByEmail_Call) Run(run func(ctx context.Context)) *mockMemberStore_InvitedMembersByEmail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *mockMemberStore_InvitedMembersByEmail_Call) Return(_a0 []*supporterdata.MemberInvite, _a1 error) *mockMemberStore_InvitedMembersByEmail_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockMemberStore_InvitedMembersByEmail_Call) RunAndReturn(run func(context.Context) ([]*supporterdata.MemberInvite, error)) *mockMemberStore_InvitedMembersByEmail_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: ctx, member
func (_m *mockMemberStore) Put(ctx context.Context, member *supporterdata.Member) error {
	ret := _m.Called(ctx, member)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *supporterdata.Member) error); ok {
		r0 = rf(ctx, member)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockMemberStore_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type mockMemberStore_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - ctx context.Context
//   - member *supporterdata.Member
func (_e *mockMemberStore_Expecter) Put(ctx interface{}, member interface{}) *mockMemberStore_Put_Call {
	return &mockMemberStore_Put_Call{Call: _e.mock.On("Put", ctx, member)}
}

func (_c *mockMemberStore_Put_Call) Run(run func(ctx context.Context, member *supporterdata.Member)) *mockMemberStore_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*supporterdata.Member))
	})
	return _c
}

func (_c *mockMemberStore_Put_Call) Return(_a0 error) *mockMemberStore_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockMemberStore_Put_Call) RunAndReturn(run func(context.Context, *supporterdata.Member) error) *mockMemberStore_Put_Call {
	_c.Call.Return(run)
	return _c
}

// newMockMemberStore creates a new instance of mockMemberStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockMemberStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockMemberStore {
	mock := &mockMemberStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
