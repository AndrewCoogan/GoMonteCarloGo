package core

import "fmt"

func (sc *ServiceContext) SyncData(symbol string) (bool, error) {
	// what do i need to return???

	md, err := sc.PostgresConnection.GetMetaDataBySymbol(sc.Context, symbol) 

	if err != nil {
		return false, fmt.Errorf("error determining if meta data exists in sync data: %w", err)
	}

	if md == nil { // we do not have this ticker saved, nor do we have any information
		ts, err := sc.AlphaVantageClient.GetStockWeeklyAdjustedMetrics(symbol)
	}



	return true, nil
}