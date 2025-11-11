package repos

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	ex "mc.data/extensions"
	m "mc.data/models"
)

var id null.Int64

func Test_Base_CanGetConnectionAndPing(t *testing.T) {
	ctx := context.Background()
	pg := getConnection(t, ctx)
	err := pg.Ping(ctx)

	if err != nil {
		t.Errorf("error pinging postgres database: %s", err)
	}
}

func Test_TimeSeriesMetaDataRepo_CanInsert(t *testing.T) {
	symbol := "_TEST"
	
	testMetaData := m.TimeSeriesMetadata{
		Information:   null.StringFrom("TEST INFO"),
		Symbol:        null.StringFrom(symbol),
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
		TimeZone:      null.StringFrom("testtimezone"),
	}

	ctx := context.Background()
	pg := getConnection(t, ctx)

	idx, err := pg.InsertNewMetaData(ctx, &testMetaData)
	if err != nil {
		t.Fatalf("error inserting new meta data: %s", err)
	}

	id = null.NewInt(idx, true)
	res, err := pg.GetMetaDataBySymbol(ctx, symbol)

	if err != nil {
		t.Fatalf("error getting meta data by symbol")
	}

	ex.AssertAreEqual(t, "information", testMetaData.Information.String, res.Information.Ptr())	
}

func getConnection(t *testing.T, ctx context.Context) *Postgres {
	t.Helper()
	err := godotenv.Load("../.env");
	if err != nil {
		t.Fatalf("error loading environment: %s", err)
	}

	connectionString := os.Getenv("DATABASE_URL")
	res, err := GetPostgresConnection(ctx, connectionString)

	if err != nil {
		t.Fatalf("error getting postgres connection: %s", err)
	}

	// on test resolving, this will close connections, even if an error is thrown
	t.Cleanup(func() {
		res.deleteTestTimeSeriesData(t, ctx)
		res.Close()
	})

	return res
}

func (pg *Postgres) deleteTestTimeSeriesData(t *testing.T, ctx context.Context) {
	t.Helper()

	if id.Ptr() != nil {
		sql := `DELETE FROM time_series_data WHERE source_id = @source_id;
				DELETE FROM time_series_metadata WHERE id = @source_id`
		args := pgx.NamedArgs{"source_id": *id.Ptr()}
		
		pg.Execute(ctx, sql, args)
	}
}