package api

import "fmt"

// TimeInterval specifies a frequency to query for intraday stock data.
type TimeInterval uint8

const (
	TimeIntervalOneMinute TimeInterval = iota
	TimeIntervalFiveMinute
	TimeIntervalFifteenMinute
	TimeIntervalThirtyMinute
	TimeIntervalSixtyMinute
)

const (
	timeSeriesIntervalErrorHeader = "Unrecognized time series interval type"
)

func (t TimeInterval) Name() (string, error) {
	switch t {
	case TimeIntervalOneMinute:
		return "TimeIntervalOneMinute", nil
	case TimeIntervalFiveMinute:
		return "TimeIntervalFiveMinute", nil
	case TimeIntervalFifteenMinute:
		return "TimeIntervalFifteenMinute", nil
	case TimeIntervalThirtyMinute:
		return "TimeIntervalThirtyMinute", nil
	case TimeIntervalSixtyMinute:
		return "TimeIntervalSixtyMinute", nil
	default:
		return "", fmt.Errorf("%s parsing name.", timeSeriesIntervalErrorHeader)
	}
}

func (t TimeInterval) Interval() (string, error) {
	switch t {
	case TimeIntervalOneMinute:
		return "1min", nil
	case TimeIntervalFiveMinute:
		return "5min", nil
	case TimeIntervalFifteenMinute:
		return "15min", nil
	case TimeIntervalThirtyMinute:
		return "30min", nil
	case TimeIntervalSixtyMinute:
		return "60min", nil
	default:
		return "", fmt.Errorf("%s parsing interval", timeSeriesIntervalErrorHeader)
	}
}
