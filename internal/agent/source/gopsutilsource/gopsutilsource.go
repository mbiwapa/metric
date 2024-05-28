// Package gopsutilsource provides a source that retrieves metrics using the gopsutil library.
// It provides functions to retrieve the value of a metric by key and to retrieve a list of observable metrics.
package gopsutilsource

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// MetricsRepo is a wrapper structure for the runtime package to implement agent interfaces.
type MetricsRepo struct {
}

// New returns an instance of the source.
// It initializes and returns a new MetricsRepo instance.
func New() (*MetricsRepo, error) {
	var storage MetricsRepo
	return &storage, nil
}

// MetricGet returns the value of a metric by key.
// It takes the metric name and source type as parameters and returns the metric value as a float64 and an error if any.
// Supported metrics are "TotalMemory", "FreeMemory", and CPU utilization metrics in the format "CPUutilizationN" where N is the CPU core number.
func (s *MetricsRepo) MetricGet(metricName string, sourceType string) (float64, error) {

	switch metricName {
	case "TotalMemory":
		vmStat, err := mem.VirtualMemory()
		if err != nil {
			return 0, err
		}
		return float64(vmStat.Total), nil
	case "FreeMemory":
		vmStat, err := mem.VirtualMemory()
		if err != nil {
			return 0, err
		}
		return float64(vmStat.Free), nil
	default:
		if sourceType == "cpu" {

			cpuPercent, err := cpu.Percent(0, true)
			if err != nil {
				return 0, err
			}

			numCPU := strings.ReplaceAll(metricName, "CPUutilization", "")
			i, err := strconv.Atoi(numCPU)
			if err != nil {
				return 0, err
			}

			return cpuPercent[i-1], nil
		}
	}

	return 0, fmt.Errorf("undefined metric: %s", metricName)
}

// GetObservableMetrics returns a list of metrics to be monitored by the agent.
// It returns a map where the keys are metric names and the values are their source types (e.g., "memory" or "cpu").
// The function also handles errors that may occur while fetching the number of CPU cores.
func (s *MetricsRepo) GetObservableMetrics() (map[string]string, error) {

	observableMetrics := map[string]string{
		"TotalMemory": "memory",
		"FreeMemory":  "memory",
	}

	cores, err := cpu.Counts(false)
	if err != nil {
		return observableMetrics, err
	}

	for i := 0; i < cores; i++ {
		observableMetrics["CPUutilization"+strconv.Itoa(i+1)] = "cpu"
	}

	return observableMetrics, nil
}
