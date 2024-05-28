// Package format provides a structure for metrics used in request/response.
// It contains the ID of the metric, the type of the metric (either gauge or counter),
// and the value of the metric which can be either Delta (for counter) or Value (for gauge).
package format

// Metric represents a structure for metrics used in request/response.
// It contains the ID of the metric, the type of the metric (either gauge or counter),
// and the value of the metric which can be either Delta (for counter) or Value (for gauge).
type Metric struct {
	ID    string   `json:"id"`              // ID is the name of the metric.
	MType string   `json:"type"`            // MType is the type of the metric, which can be either "gauge" or "counter".
	Delta *int64   `json:"delta,omitempty"` // Delta is the value of the metric if the type is "counter".
	Value *float64 `json:"value,omitempty"` // Value is the value of the metric if the type is "gauge".
}

const (
	// Gauge is a type of metric that represents a single numerical value that can arbitrarily go up and down.
	// It is used to measure values that can increase or decrease, such as temperature or current memory usage.
	Gauge = "gauge"

	// Counter is a type of metric that represents a cumulative value that only increases.
	// It is used to measure values that only go up, such as the number of requests received or errors encountered.
	Counter = "counter"
)
