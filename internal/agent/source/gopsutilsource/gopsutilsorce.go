package gopsutilsource

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// MetricsRepo структура обертка пакета рантайм для имплементации интерфейсов агента
type MetricsRepo struct {
}

// New возвращает инстанс источника
func New() (*MetricsRepo, error) {
	var storage MetricsRepo
	return &storage, nil
}

// MetricGet возвращает значение метрики по ключу
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

// GetObservableMetrics возвращает список метрик для отслеживание агентом
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
