# Metrics

Metrics define success measurements and KPIs for roles.

## Overview

How do you know if an AI agent is performing well? Metrics provide:

- **Success measurements** - Quantifiable outcomes
- **Thresholds** - Warning and critical levels
- **Tracking** - Historical performance data
- **Accountability** - Clear expectations

Metrics are **data definitions only** - actual collection and storage happens in the runtime MetricsStore.

## MetricDefinition

```go
type MetricDefinition struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Description string        `json:"description,omitempty"`
    Type        MetricType    `json:"type"`
    Unit        string        `json:"unit,omitempty"`
    Target      *MetricTarget `json:"target,omitempty"`
    Labels      []string      `json:"labels,omitempty"`
    Buckets     []float64     `json:"buckets,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `ID` | string | Unique identifier |
| `Name` | string | Human-readable name |
| `Description` | string | What this metric measures |
| `Type` | MetricType | How the metric is measured |
| `Unit` | string | Measurement unit |
| `Target` | *MetricTarget | Success thresholds |
| `Labels` | []string | Additional dimensions |
| `Buckets` | []float64 | Histogram bucket boundaries |

## MetricType

```go
type MetricType string

const (
    MetricTypeCounter   MetricType = "counter"
    MetricTypeGauge     MetricType = "gauge"
    MetricTypeHistogram MetricType = "histogram"
    MetricTypeSummary   MetricType = "summary"
)
```

| Type | Description | Example |
|------|-------------|---------|
| `counter` | Cumulative count | Total meetings facilitated |
| `gauge` | Point-in-time value | Current active sessions |
| `histogram` | Distribution of values | Response time distribution |
| `summary` | Quantiles over time | 95th percentile latency |

## MetricTarget

Defines success thresholds.

```go
type MetricTarget struct {
    Value             float64 `json:"value"`
    Operator          string  `json:"operator,omitempty"`
    WarningThreshold  float64 `json:"warning_threshold,omitempty"`
    CriticalThreshold float64 `json:"critical_threshold,omitempty"`
    Period            string  `json:"period,omitempty"`
}
```

| Field | Description |
|-------|-------------|
| `Value` | Target value to achieve |
| `Operator` | Comparison: ">=", "<=", "==", ">", "<" |
| `WarningThreshold` | Triggers warning |
| `CriticalThreshold` | Triggers alert |
| `Period` | Time window: "1h", "24h", "7d" |

## Common Units

```go
const (
    UnitPercent      = "percent"
    UnitSeconds      = "seconds"
    UnitMilliseconds = "milliseconds"
    UnitCount        = "count"
    UnitBytes        = "bytes"
    UnitRequests     = "requests"
)
```

## Helper Functions

```go
// Create a counter metric
func NewCounterMetric(id, name, description string) MetricDefinition

// Create a gauge metric
func NewGaugeMetric(id, name, description, unit string) MetricDefinition

// Create a histogram metric with custom buckets
func NewHistogramMetric(id, name, description, unit string, buckets []float64) MetricDefinition
```

## Examples

### Meeting PM Metrics

```go
Metrics: []role.MetricDefinition{
    {
        ID:          "action-capture-rate",
        Name:        "Action Item Capture Rate",
        Description: "Percentage of action items captured vs discussed",
        Type:        role.MetricTypeGauge,
        Unit:        role.UnitPercent,
        Target: &role.MetricTarget{
            Value:             95.0,
            Operator:          ">=",
            WarningThreshold:  90.0,
            CriticalThreshold: 80.0,
            Period:            "7d",
        },
    },
    {
        ID:          "notes-publish-time",
        Name:        "Notes Publish Time",
        Description: "Time from meeting end to notes published",
        Type:        role.MetricTypeHistogram,
        Unit:        role.UnitSeconds,
        Buckets:     []float64{60, 300, 900, 1800, 3600}, // 1m, 5m, 15m, 30m, 1h
        Target: &role.MetricTarget{
            Value:             900, // 15 minutes
            Operator:          "<=",
            WarningThreshold:  1800, // 30 minutes
            CriticalThreshold: 3600, // 1 hour
        },
    },
    {
        ID:          "meetings-facilitated",
        Name:        "Meetings Facilitated",
        Description: "Total number of meetings facilitated",
        Type:        role.MetricTypeCounter,
        Unit:        role.UnitCount,
        Labels:      []string{"meeting_type", "outcome"},
    },
    {
        ID:          "participant-satisfaction",
        Name:        "Participant Satisfaction",
        Description: "Average satisfaction rating from participants",
        Type:        role.MetricTypeGauge,
        Unit:        "rating",
        Target: &role.MetricTarget{
            Value:             4.5, // out of 5
            Operator:          ">=",
            WarningThreshold:  4.0,
            CriticalThreshold: 3.5,
            Period:            "30d",
        },
    },
}
```

### Code Reviewer Metrics

```go
Metrics: []role.MetricDefinition{
    role.NewCounterMetric(
        "reviews-completed",
        "Reviews Completed",
        "Total pull requests reviewed",
    ),
    {
        ID:          "review-turnaround",
        Name:        "Review Turnaround Time",
        Description: "Time from PR opened to first review",
        Type:        role.MetricTypeHistogram,
        Unit:        role.UnitSeconds,
        Buckets:     []float64{300, 900, 1800, 3600, 7200, 14400}, // 5m to 4h
        Target: &role.MetricTarget{
            Value:    1800, // 30 minutes
            Operator: "<=",
        },
    },
    {
        ID:          "issues-found",
        Name:        "Issues Found per Review",
        Description: "Average number of issues identified per review",
        Type:        role.MetricTypeGauge,
        Unit:        role.UnitCount,
        Labels:      []string{"severity", "category"},
    },
    {
        ID:          "false-positive-rate",
        Name:        "False Positive Rate",
        Description: "Percentage of flagged issues that were dismissed",
        Type:        role.MetricTypeGauge,
        Unit:        role.UnitPercent,
        Target: &role.MetricTarget{
            Value:             10.0, // Less than 10%
            Operator:          "<=",
            WarningThreshold:  15.0,
            CriticalThreshold: 25.0,
            Period:            "7d",
        },
    },
}
```

### Support Agent Metrics

```go
Metrics: []role.MetricDefinition{
    {
        ID:          "resolution-rate",
        Name:        "First Contact Resolution Rate",
        Description: "Percentage of issues resolved on first contact",
        Type:        role.MetricTypeGauge,
        Unit:        role.UnitPercent,
        Target: &role.MetricTarget{
            Value:    80.0,
            Operator: ">=",
            Period:   "7d",
        },
    },
    {
        ID:          "response-time",
        Name:        "Average Response Time",
        Description: "Time to first response",
        Type:        role.MetricTypeHistogram,
        Unit:        role.UnitSeconds,
        Buckets:     []float64{30, 60, 120, 300, 600},
        Target: &role.MetricTarget{
            Value:    60, // 1 minute
            Operator: "<=",
        },
    },
    role.NewCounterMetric(
        "tickets-handled",
        "Tickets Handled",
        "Total support tickets processed",
    ),
    {
        ID:          "escalation-rate",
        Name:        "Escalation Rate",
        Description: "Percentage of tickets escalated to humans",
        Type:        role.MetricTypeGauge,
        Unit:        role.UnitPercent,
        Target: &role.MetricTarget{
            Value:             20.0, // Less than 20%
            Operator:          "<=",
            WarningThreshold:  25.0,
            CriticalThreshold: 35.0,
        },
    },
}
```

## Using Metrics in a Role

Implement `MetricsProvider`:

```go
type MeetingPMRole struct {
    role.BaseRole
}

func (r *MeetingPMRole) Metrics() []role.MetricDefinition {
    return []role.MetricDefinition{
        role.NewCounterMetric("meetings", "Meetings Facilitated", ""),
        role.NewGaugeMetric("capture-rate", "Action Capture Rate", "", "percent"),
    }
}
```

Or include in RoleSpec:

```go
func (r *MeetingPMRole) Spec() *role.RoleSpec {
    return &role.RoleSpec{
        ID:   "meeting-pm",
        Name: "Meeting Program Manager",
        Metrics: []role.MetricDefinition{
            // Metrics defined here
        },
    }
}
```

## Metric Categories

### Efficiency Metrics

| Metric | Type | Description |
|--------|------|-------------|
| Turnaround time | histogram | Time to complete tasks |
| Throughput | counter | Tasks completed per period |
| Utilization | gauge | Percentage of capacity used |

### Quality Metrics

| Metric | Type | Description |
|--------|------|-------------|
| Accuracy | gauge | Correctness of outputs |
| Error rate | gauge | Percentage of errors |
| False positive rate | gauge | Incorrect positive findings |

### Satisfaction Metrics

| Metric | Type | Description |
|--------|------|-------------|
| User rating | gauge | User satisfaction scores |
| Resolution rate | gauge | Issues resolved successfully |
| Escalation rate | gauge | Issues requiring human help |

## Best Practices

1. **Define clear targets** - What does success look like?
2. **Set appropriate thresholds** - Warning before critical
3. **Use meaningful labels** - Enable drill-down analysis
4. **Track over time** - Spot trends and regressions
5. **Balance metrics** - Don't optimize one at expense of others
6. **Review regularly** - Adjust targets as performance improves
