package core

import (
	"fmt"
	"log"
	"maps"
	"math"
	"slices"
	"time"

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
	Allocations []SimulationAllocation `json:"allocations"`
	MaxLookback time.Duration          `json:"maxlookback"` // what is this going to look like in api req?

	Iterations int   `json:"iterations"`
	Seed       int64 `json:"seed"`     // ^^
	DistType   int   `json:"disttype"` // standar normal, student t

	SimulationUnitOfTime int `json:"simulationunitoftime"` // daily, weekly, monthly, quarterly, yearly
	SimulationDuration   int `json:"simulationduration"`   // number of units of time to simulate
	DegreesOfFreedom     int `json:"degreesoffreedom"`     // degrees of freedom for student t distribution
}

type SeriesReturns struct {
	SimulationAllocation
	Returns             []float64
	Dates               []time.Time
	AnnualizationFactor int
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

func (sc *ServiceContext) RunEquityMonteCarloWithCovarianceMartix(request SimulationRequest) ([]*SimulationResult, error) {
	res := make([]*SimulationResult, request.Iterations)
	seriesReturns, err := sc.getSeriesReturns(request)
	if err != nil {
		return res, err
	}

	statisticalResources, err := GetStatisticalResources(request, seriesReturns)
	if err != nil {
		return res, err
	}

	nJobs := int(math.Ceil(float64(request.Iterations) / BatchSize / Workers))
	if nJobs == 0 && request.Iterations > 0 {
		nJobs = 1
	}

	log.Println("Starting monte carlo simulation:")
	log.Printf("\t Simulation duration: %v %s", request.SimulationDuration, convertFrequencyToString(request.SimulationUnitOfTime))
	log.Printf("\t Simulation paths: %v", request.Iterations)
	log.Printf("\t Simulation batch size: %v", BatchSize)
	log.Printf("\t Workers: %v", Workers)

	workerCount := ex.Min(nJobs, Workers)
	workerResources := make([]*WorkerResource, workerCount)
	for i := range workerCount {
		workerResources[i] = NewWorkerResources(statisticalResources, uint64(request.Seed), uint64(i))
	}

	jobs := make(chan job, nJobs) // TODO: if njobs is less than workers, take the minimum
	done := make(chan bool, ex.Min(nJobs, Workers))

	worker := func(wr *WorkerResource) {
		for j := range jobs { // this will loop over available jobs, and will reup if a job finishes and there are more jobs
			for sim := j.start; sim < j.end; sim++ { // this will loop over the iterations
				portfolioValue := 100.0
				pathValues := make([]float64, request.Iterations+1)
				pathValues[0] = portfolioValue

				for period := range request.Iterations {
					correlatedReturns := wr.GetCorrelatedReturns(request.SimulationUnitOfTime)
					portfolioReturn, err := ex.DotProduct(statisticalResources.AssetWeight, correlatedReturns)
					if err != nil {
						return // TODO: how to handle errors in channels?
					}

					portfolioValue *= math.Exp(portfolioReturn)
					pathValues[period+1] = portfolioValue
				}

				totalReturn := portfolioValue - 1.0
				fullDurationAnnualizationFactor := float64(request.SimulationDuration) / float64(request.SimulationUnitOfTime)
				annualizedReturn := math.Pow(portfolioValue, fullDurationAnnualizationFactor) - 1.0

				res[sim] = &SimulationResult{
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
	for i := range workerCount {
		go worker(workerResources[i])
	}

	// allocate the jobs and the respective dist index, start and end iteration indicies for result allocation
	for i := range nJobs {
		start := i * BatchSize
		end := ex.Min(start+BatchSize, request.Iterations)
		if start != end {
			jobs <- job{index: i, start: start, end: end}
		}
	}
	close(jobs) // close the job channel, there isnt anything else being added to it

	// this will loop until all of the jobs are complete
	for range workerCount {
		<-done
	}

	return res, nil
}

func (sc *ServiceContext) getSeriesReturns(request SimulationRequest) (res []*SeriesReturns, err error) {
	tickerLookup := make(map[int32]SimulationAllocation, len(request.Allocations))
	for _, allocation := range request.Allocations {
		tickerLookup[allocation.Id] = allocation
	}

	returns, err := sc.PostgresConnection.GetTimeSeriesReturns(sc.Context, slices.Collect(maps.Keys(tickerLookup)), request.MaxLookback)
	if err != nil {
		return res, fmt.Errorf("error getting time series returns: %v", err)
	}

	agg := make(map[int32]*SeriesReturns, len(request.Allocations))
	for _, ret := range returns {
		if agg[ret.Id] == nil {
			agg[ret.Id] = &SeriesReturns{
				SimulationAllocation: tickerLookup[ret.Id],
				Returns:              []float64{},
				Dates:                []time.Time{},
				AnnualizationFactor:  Weekly,
			}
		}

		agg[ret.Id].Returns = append(agg[ret.Id].Returns, ret.LogReturn)
		agg[ret.Id].Dates = append(agg[ret.Id].Dates, ret.Timestamp)
	}

	for _, tickerAgg := range agg {
		res = append(res, tickerAgg)
	}

	// sorts on source id for consistency, useful for testing
	slices.SortFunc(res, func(i, j *SeriesReturns) int {
		return int(i.Id - j.Id)
	})

	err = verifySeriesReturnIntegrity(res)
	return
}

func verifySeriesReturnIntegrity(data []*SeriesReturns) error {
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

func getTimeRange(data *SeriesReturns) (time.Time, time.Time, int) {
	// i dont think we need to keep the dates in the same order... but i dont want to find out
	dates := slices.Clone(data.Dates)
	slices.SortFunc(dates, func(i, j time.Time) int {
		return i.Compare(j)
	})

	first := dates[0]
	length := len(dates)
	last := dates[length-1]

	if last.Before(first) {
		log.Println("you dummy you missed the multipler in getTimeRange() sort compare")
	}

	return first, last, length
}
