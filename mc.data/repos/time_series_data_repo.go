package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"mc.data/models"
)

func (pg *Postgres) GetTimeSeriesData(ctx context.Context, symbol string) ([]models.TimeSeriesData, error) {
	query := `SELECT timestamp, open, high, low, close, adjusted_close, volume, dividend_amount 
		FROM time_series_data 
		WHERE symbol = @symbol`
	
	args := pgx.NamedArgs{
		"symbol": symbol,
	}

	res, err := Query[models.TimeSeriesData](ctx, pg, query, args)
	if err != nil {
		return nil, fmt.Errorf("unable to query data by symbol (%s): %w", symbol, err)
	}
	return res, nil
}

func (pg *Postgres) InsertTimeSeriesData(ctx context.Context, data []models.TimeSeriesData, source_id int64) (int64, error) {
    columns := []string{
        "source_id", "timestamp", "open", "high", "low", 
        "close", "adjusted_close", "volume", "dividend_amount",
    }
    
    entries := make([][]any, len(data))
    for i, ent := range data {
        entries[i] = []any{
            source_id, ent.Timestamp, ent.Open, ent.High, ent.Low,
            ent.Close, ent.AdjustedClose, ent.Volume, ent.DividendAmount,
        }
    }

	return pg.BulkInsert(ctx, time_series_data, columns, entries)
}