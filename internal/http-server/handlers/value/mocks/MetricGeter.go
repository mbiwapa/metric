// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// MetricGeter is an autogenerated mock type for the MetricGeter type
type MetricGeter struct {
	mock.Mock
}

// GetMetric provides a mock function with given fields: typ, key
func (_m *MetricGeter) GetMetric(typ string, key string) (string, error) {
	ret := _m.Called(typ, key)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return rf(typ, key)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(typ, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(typ, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMetricGeter interface {
	mock.TestingT
	Cleanup(func())
}

// NewMetricGeter creates a new instance of MetricGeter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMetricGeter(t mockConstructorTestingTNewMetricGeter) *MetricGeter {
	mock := &MetricGeter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
