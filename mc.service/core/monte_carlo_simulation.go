package core

type SimulationAllocation struct {
	Id     int32 `json:"id"`
	Ticker string `json:"ticker"`
	Weight float64 `json:"weight"`
}

type SimulationRequest struct {
	Allocations []SimulationAllocation `json:"allocations"`
	Iterations  int `json:"iterations"`
	// other stuff?
}

func (sc *ServiceContext) RunMonteCarloSimulation(request SimulationRequest) error {
	return nil
}