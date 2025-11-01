package api

import (
	"fmt"
)

type TimeSeries uint8

// TimeSeries specifies a frequency to query for stock data.
const (
	TimeSeriesDaily TimeSeries = iota
	TimeSeriesDailyAdjusted
	TimeSeriesWeekly
	TimeSeriesWeeklyAdjusted
	TimeSeriesMonthly
	TimeSeriesMonthlyAdjusted
)

const (
	timeSeriesErrorHeader = "unrecognized time series type"
)

func (t TimeSeries) Name() (string, error) {
	switch t {
	case TimeSeriesDaily:
		return "TimeSeriesDaily", nil
	case TimeSeriesDailyAdjusted:
		return "TimeSeriesDailyAdjusted", nil
	case TimeSeriesWeekly:
		return "TimeSeriesWeekly", nil
	case TimeSeriesWeeklyAdjusted:
		return "TimeSeriesWeeklyAdjusted", nil
	case TimeSeriesMonthly:
		return "TimeSeriesMonthly", nil
	case TimeSeriesMonthlyAdjusted:
		return "TimeSeriesMonthlyAdjusted", nil
	default:
		return "", fmt.Errorf("%s parsing name", timeSeriesErrorHeader)
	}
}

func (t TimeSeries) Function() (string, error) {
	switch t {
	case TimeSeriesDaily:
		return "TIME_SERIES_DAILY", nil
	case TimeSeriesDailyAdjusted:
		return "TIME_SERIES_DAILY_ADJUSTED", nil
	case TimeSeriesWeekly:
		return "TIME_SERIES_WEEKLY", nil
	case TimeSeriesWeeklyAdjusted:
		return "TIME_SERIES_WEEKLY_ADJUSTED", nil
	case TimeSeriesMonthly:
		return "TIME_SERIES_MONTHLY", nil
	case TimeSeriesMonthlyAdjusted:
		return "TIME_SERIES_MONTHLY_ADJUSTED", nil
	default:
		return "", fmt.Errorf("%s parsing function", timeSeriesErrorHeader)
	}
}

func (t TimeSeries) TimeSeriesKey() (string, error) {
	switch t {
	case TimeSeriesDaily, TimeSeriesDailyAdjusted:
		return "Daily Time Series", nil
	case TimeSeriesWeekly, TimeSeriesWeeklyAdjusted:
		return "Weekly Time Series", nil
	case TimeSeriesMonthly, TimeSeriesMonthlyAdjusted:
		return "Monthly Time Series", nil
	default:
		return "", fmt.Errorf("%s parsing time series key", timeSeriesErrorHeader)

	}
}