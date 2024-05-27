// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Backuper is an autogenerated mock type for the Backuper type
type Backuper struct {
	mock.Mock
}

// IsSyncMode provides a mock function with given fields:
func (_m *Backuper) IsSyncMode() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SaveToFile provides a mock function with given fields:
func (_m *Backuper) SaveToFile() {
	_m.Called()
}

// SaveToStruct provides a mock function with given fields: typ, name, value
func (_m *Backuper) SaveToStruct(typ string, name string, value string) error {
	ret := _m.Called(typ, name, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(typ, name, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewBackuper interface {
	mock.TestingT
	Cleanup(func())
}

// NewBackuper creates a new instance of Backuper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewBackuper(t mockConstructorTestingTNewBackuper) *Backuper {
	mock := &Backuper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
