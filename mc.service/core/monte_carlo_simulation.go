package core

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"time"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"

	ex "mc.data/extensions"
)

const (
	Workers   = 8
	BatchSize = 10_000
)

type SimulationAllocation struct {
	Id     int32   `json:"id"`
	Ticker string  `json:"ticker"`
	Weight float64 `json:"weight"`
}

type SimulationRequest struct {
	Allocations        []SimulationAllocation `json:"allocations"`
	Iterations         int                    `json:"iterations"`
	MaxLookback        time.Duration          `json:"maxlookback"`        // what is this going to look like in api req?
	SimulationDuration time.Duration          `json:"simulationduration"` // ^^
	Seed               int64                  `json:"seed"`               // ^^
}

type SeriesReturns struct {
	SourceID   int32
	Ticker     string
	Weight     float64
	Returns    []float64
	Dates      []time.Time
	MeanReturn float64
	StdDev     float64
}

type SimulationResult struct {
	FinalValue       float64
	TotalReturn      float64
	AnnualizedReturn float64
	PathValues       []float64
}

type job struct {
	index, start, end int
}

func (sr SimulationRequest) Validate() error {
	// make sure the total weight is 100%
	weightSum := 0.0
	for _, w := range sr.Allocations {
		weightSum += w.Weight
	}
	if math.Abs(weightSum-1.0) > 1e-6 {
		return fmt.Errorf("weights must sum to 1.0, got %.6f", weightSum)
	}

	// make sure assets allocated to are unique
	v := make(map[int32]bool, len(sr.Allocations))
	for _, a := range sr.Allocations {
		v[a.Id] = true
	}

	if len(sr.Allocations) != len(slices.Collect(maps.Keys(v))) {
		return fmt.Errorf("did not recieve a unique asset list")
	}

	// anything else we want to validate before kicking off a simulation?

	return nil
}

func (sc *ServiceContext) RunEquityMonteCarloWithCovarianceMartix(request SimulationRequest) error {
	/*
		seriesReturns, err := sc.getSeriesReturns(request)

		if err != nil {
			return err
		}

		returns := make([][]float64, len(seriesReturns))
		for i, r := range seriesReturns {
			j, wr := range r. {

			}
		}

		sr := GetStatisticalResources()


		jobs := make(chan job, nJobs)
		done := make(chan bool, Workers)

		numPeriods := int(math.Ceil(request.SimulationDuration.Hours() / (24 * 7)))

		res := make([]SimulationResult, request.Iterations)

		worker := func() {
			for j := range jobs { // this will loop over available jobs, and will reup if a job finishes and there are more jobs
				for sim := j.start; sim < j.end; sim++ {
					portfolioValue := 100.0
					pathValues := make([]float64, numPeriods+1)
					pathValues[0] = portfolioValue

					for period := range numPeriods {
						correlatedReturns := generateCorrelatedReturns(&L, meanReturns, &dists[j.index])
						portfolioReturn, _ := DotProduct(weights, correlatedReturns)
						portfolioValue *= math.Exp(portfolioReturn)
						pathValues[period+1] = portfolioValue
					}

					totalReturn := portfolioValue - 1.0
					annualizedReturn := math.Pow(portfolioValue, 52.0/float64(numPeriods)) - 1.0

					res[sim] = SimulationResult{
						FinalValue:       portfolioValue,
						TotalReturn:      totalReturn,
						AnnualizedReturn: annualizedReturn,
						PathValues:       pathValues,
					}
				}
			}
			done <- true
		}

		// starts the workers
		for range Workers {
			go worker()
		}

		// allocate the jobs and the respective dist index, start and end iteration indicies for result allocation
		for i := range nJobs {
			start := i * BatchSize
			end := int(math.Min(float64(start+BatchSize), float64(request.Iterations)))
			if start != end {
				jobs <- job{index: i, start: start, end: end}
			}
		}
		close(jobs) // close the job channel, there isnt anything else being added to it

		// this will loop until all of the jobs are complete
		for range Workers {
			<-done
		}

		//return results, nil
	*/

	return nil
}

func generateCorrelatedReturns(L *mat.TriDense, means []float64, normalDist *distuv.Normal) []float64 {
	n := len(means)

	z := make([]float64, n)
	for i := range n {
		z[i] = normalDist.Rand()
	}

	// transform to correlated returns: y = L * z + mean
	zVec := mat.NewVecDense(n, z)
	yVec := mat.NewVecDense(n, nil)
	yVec.MulVec(L, zVec)

	// add the asset mean returns
	correlatedReturns := make([]float64, n)
	for i := range n {
		correlatedReturns[i] = yVec.AtVec(i) + means[i] // TODO: figure out annualization
	}

	return correlatedReturns
}

func (sc *ServiceContext) getSeriesReturns(request SimulationRequest) (res []SeriesReturns, err error) {
	tickerLookup := make(map[int32]SimulationAllocation, len(request.Allocations))
	for _, allocation := range request.Allocations {
		tickerLookup[allocation.Id] = SimulationAllocation{
			Ticker: allocation.Ticker,
			Weight: allocation.Weight,
		}
	}

	returns, err := sc.PostgresConnection.GetTimeSeriesReturns(sc.Context, slices.Collect(maps.Keys(tickerLookup)), request.MaxLookback)
	if err != nil {
		return res, fmt.Errorf("error getting time series returns: %v", err)
	}

	agg := make(map[int32]*SeriesReturns, len(request.Allocations))
	for _, ret := range returns {
		if agg[ret.Id] == nil {
			agg[ret.Id] = &SeriesReturns{
				SourceID: ret.Id,
				Ticker:   tickerLookup[ret.Id].Ticker,
				Weight:   tickerLookup[ret.Id].Weight,
				Returns:  []float64{},
				Dates:    []time.Time{},
			}
		}

		agg[ret.Id].Returns = append(agg[ret.Id].Returns, ret.LogReturn)
		agg[ret.Id].Dates = append(agg[ret.Id].Dates, ret.Timestamp)
	}

	for _, tickerAgg := range agg {
		res = append(res, *tickerAgg)
	}

	// sorts on source id for consistency, useful for testing
	slices.SortFunc(res, func(i, j SeriesReturns) int {
		return int(i.SourceID - j.SourceID)
	})

	for i, r := range res {
		res[i].MeanReturn = stat.Mean(r.Returns, nil)
		res[i].StdDev = stat.StdDev(r.Returns, nil)
	}

	if err = verifyDataIntegrity(res); err != nil {
		return
	}

	return
}

func verifyDataIntegrity(data []SeriesReturns) error {
	firstDates := make([]time.Time, len(data))
	lastDates := make([]time.Time, len(data))
	lengths := make([]int, len(data))
	for _, v := range data {
		first, last, length := getTimeRange(v)
		firstDates = append(firstDates, first)
		lastDates = append(lastDates, last)
		lengths = append(lengths, length)
	}

	if ex.AreAllEqual(firstDates) {
		return fmt.Errorf("data validation failed, first dates in range do not align")
	}

	if ex.AreAllEqual(lastDates) {
		return fmt.Errorf("data validation failed, last dates in range do not align")
	}

	if ex.AreAllEqual(lengths) {
		return fmt.Errorf("data validation failed, length of dates in range do not align")
	}

	return nil
}

func getTimeRange(data SeriesReturns) (first, last time.Time, length int) {
	first = time.Date(3000, time.December, 31, 0, 0, 0, 0, time.UTC)
	for _, v := range data.Dates {
		if v.Before(first) {
			first = v
		}
		if v.After(last) {
			last = v
		}
		length++
	}
	return
}
