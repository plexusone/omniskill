// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package role

// MetricType categorizes metrics by how they're measured.
type MetricType string

const (
	// MetricTypeCounter counts cumulative occurrences.
	MetricTypeCounter MetricType = "counter"

	// MetricTypeGauge measures a point-in-time value.
	MetricTypeGauge MetricType = "gauge"

	// MetricTypeHistogram measures distribution of values.
	MetricTypeHistogram MetricType = "histogram"

	// MetricTypeSummary provides quantiles over time.
	MetricTypeSummary MetricType = "summary"
)

// MetricDefinition defines a success metric or KPI for the role.
//
// MetricDefinition is data only - actual collection and storage
// happens in the runtime MetricsStore.
type MetricDefinition struct {
	// ID is a unique identifier for this metric.
	ID string `json:"id"`

	// Name is a human-readable name (e.g., "Action Capture Rate").
	Name string `json:"name"`

	// Description explains what this metric measures.
	Description string `json:"description,omitempty"`

	// Type specifies how this metric is measured.
	Type MetricType `json:"type"`

	// Unit is the measurement unit (e.g., "percent", "seconds", "count").
	Unit string `json:"unit,omitempty"`

	// Target defines success thresholds.
	Target *MetricTarget `json:"target,omitempty"`

	// Labels are additional dimensions for this metric.
	Labels []string `json:"labels,omitempty"`

	// Buckets are histogram bucket boundaries (for histogram type).
	Buckets []float64 `json:"buckets,omitempty"`
}

// MetricTarget defines success thresholds for a metric.
type MetricTarget struct {
	// Value is the target value to achieve.
	Value float64 `json:"value"`

	// Operator specifies how to compare (e.g., ">=", "<=", "==").
	Operator string `json:"operator,omitempty"`

	// WarningThreshold triggers a warning when crossed.
	WarningThreshold float64 `json:"warning_threshold,omitempty"`

	// CriticalThreshold triggers an alert when crossed.
	CriticalThreshold float64 `json:"critical_threshold,omitempty"`

	// Period is the time window for evaluation (e.g., "1h", "24h", "7d").
	Period string `json:"period,omitempty"`
}

// Target comparison operators.
const (
	OperatorGreaterThanOrEqual = ">="
	OperatorLessThanOrEqual    = "<="
	OperatorEqual              = "=="
	OperatorGreaterThan        = ">"
	OperatorLessThan           = "<"
)

// Common metric units.
const (
	UnitPercent      = "percent"
	UnitSeconds      = "seconds"
	UnitMilliseconds = "milliseconds"
	UnitCount        = "count"
	UnitBytes        = "bytes"
	UnitRequests     = "requests"
)

// NewCounterMetric creates a counter MetricDefinition.
func NewCounterMetric(id, name, description string) MetricDefinition {
	return MetricDefinition{
		ID:          id,
		Name:        name,
		Description: description,
		Type:        MetricTypeCounter,
		Unit:        UnitCount,
	}
}

// NewGaugeMetric creates a gauge MetricDefinition.
func NewGaugeMetric(id, name, description, unit string) MetricDefinition {
	return MetricDefinition{
		ID:          id,
		Name:        name,
		Description: description,
		Type:        MetricTypeGauge,
		Unit:        unit,
	}
}

// NewHistogramMetric creates a histogram MetricDefinition with custom buckets.
func NewHistogramMetric(id, name, description, unit string, buckets []float64) MetricDefinition {
	return MetricDefinition{
		ID:          id,
		Name:        name,
		Description: description,
		Type:        MetricTypeHistogram,
		Unit:        unit,
		Buckets:     buckets,
	}
}
