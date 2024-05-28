package sender

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAllMetricGeter is a mock implementation of the AllMetricGeter interface
type MockAllMetricGeter struct {
	mock.Mock
}

func (m *MockAllMetricGeter) GetAllMetrics(ctx context.Context) ([][]string, [][]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([][]string), args.Get(1).([][]string), args.Error(2)
}

// MockMetricSender is a mock implementation of the MetricSender interface
type MockMetricSender struct {
	mock.Mock
}

func (m *MockMetricSender) Worker(jobs <-chan map[string][][]string, errorChanel chan<- error) {
	for job := range jobs {
		m.Called(job)
	}
}

func TestStartStopsOnContextCancel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockStor := new(MockAllMetricGeter)
	mockSender := new(MockMetricSender)
	errorChanel := make(chan error, 1)

	// Setup mock expectations
	mockStor.On("GetAllMetrics", mock.Anything).Return([][]string{}, [][]string{}, nil).Once()
	mockSender.On("Worker", mock.Anything).Return().Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the sender
	go Start(ctx, mockStor, mockSender, 1, logger, 1, errorChanel)

	// Allow some time for the sender to start
	time.Sleep(2 * time.Second)

	// Cancel the context to stop the sender
	cancel()

	// Allow some time for the sender to stop
	time.Sleep(2 * time.Second)

	// Assert that the mocks were called as expected
	mockStor.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestStartStopsOnContextCancelWithMultipleWorkers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockStor := new(MockAllMetricGeter)
	mockSender := new(MockMetricSender)
	errorChanel := make(chan error, 1)

	// Setup mock expectations
	mockStor.On("GetAllMetrics", mock.Anything).Return([][]string{}, [][]string{}, nil)
	mockSender.On("Worker", mock.Anything).Return()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the sender with multiple workers
	go Start(ctx, mockStor, mockSender, 1, logger, 3, errorChanel)

	// Allow some time for the sender to start
	time.Sleep(5 * time.Second)

	// Cancel the context to stop the sender
	cancel()

	// Allow some time for the sender to stop
	time.Sleep(2 * time.Second)

	// Assert that the mocks were called as expected
	mockStor.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestStartSendsMetricsPeriodically(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockStor := new(MockAllMetricGeter)
	mockSender := new(MockMetricSender)
	errorChanel := make(chan error, 1)

	// Setup mock expectations
	mockStor.On("GetAllMetrics", mock.Anything).Return([][]string{}, [][]string{}, nil).Twice()
	mockSender.On("Worker", mock.Anything).Return()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the sender
	go Start(ctx, mockStor, mockSender, 1, logger, 1, errorChanel)

	// Allow some time for the sender to send metrics twice
	time.Sleep(3 * time.Second)

	// Cancel the context to stop the sender
	cancel()

	// Allow some time for the sender to stop
	time.Sleep(2 * time.Second)

	// Assert that the mocks were called as expected
	mockStor.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}
