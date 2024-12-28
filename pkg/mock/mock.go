package mock

import (
	"github.com/stretchr/testify/mock"
)

// Mock wraps testify/mock.Mock to provide a consistent mocking interface
type Mock struct {
	mock.Mock
}

// On wraps testify/mock.Mock.On
func (m *Mock) On(methodName string, arguments ...interface{}) *mock.Call {
	return m.Mock.On(methodName, arguments...)
}

// AssertExpectations wraps testify/mock.Mock.AssertExpectations
func (m *Mock) AssertExpectations(t mock.TestingT) bool {
	return m.Mock.AssertExpectations(t)
}

// MethodCalled wraps testify/mock.Mock.MethodCalled
func (m *Mock) MethodCalled(methodName string, arguments ...interface{}) mock.Arguments {
	return m.Mock.MethodCalled(methodName, arguments...)
}

// Called wraps testify/mock.Mock.Called
func (m *Mock) Called(arguments ...interface{}) mock.Arguments {
	return m.Mock.Called(arguments...)
}

// AssertCalled wraps testify/mock.Mock.AssertCalled
func (m *Mock) AssertCalled(t mock.TestingT, methodName string, arguments ...interface{}) bool {
	return m.Mock.AssertCalled(t, methodName, arguments...)
}

// AssertNotCalled wraps testify/mock.Mock.AssertNotCalled
func (m *Mock) AssertNotCalled(t mock.TestingT, methodName string, arguments ...interface{}) bool {
	return m.Mock.AssertNotCalled(t, methodName, arguments...)
}

// AssertNumberOfCalls wraps testify/mock.Mock.AssertNumberOfCalls
func (m *Mock) AssertNumberOfCalls(t mock.TestingT, methodName string, expectedCalls int) bool {
	return m.Mock.AssertNumberOfCalls(t, methodName, expectedCalls)
}

// MockMatchedBy wraps testify/mock.MatchedBy for type safety
func MockMatchedBy(fn interface{}) interface{} {
	return mock.MatchedBy(fn)
}
