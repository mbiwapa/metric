// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Updater is an autogenerated mock type for the Updater type
type Updater struct {
	mock.Mock
}

// CounterUpdate provides a mock function with given fields: key, value
func (_m *Updater) CounterUpdate(key string, value int64) error {
	ret := _m.Called(key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GaugeUpdate provides a mock function with given fields: key, value
func (_m *Updater) GaugeUpdate(key string, value float64) error {
	ret := _m.Called(key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, float64) error); ok {
		r0 = rf(key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewUpdater interface {
	mock.TestingT
	Cleanup(func())
}

// NewUpdater creates a new instance of Updater. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUpdater(t mockConstructorTestingTNewUpdater) *Updater {
	mock := &Updater{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
