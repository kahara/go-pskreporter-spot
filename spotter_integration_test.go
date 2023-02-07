//go:build integration

package spot

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"testing"
	"time"
)

const (
	TimescaleDBConnectionString = "postgres://postgres:password@timescaledb"
	ReportsQuery                = "SELECT * FROM report;"
	ReceiverHostport            = "receiver:4739"
	FakespotCount               = 100
	FakespotCallsign            = "N0CALL"
	FakespotLocator             = "JJ00OG"
	FakespotAntennaInformation  = "Dipole"
	FakespotDecoderSoftware     = "fakespot v0"
	FakespotSpotKind            = SpotKind_CallsignFrequencySNRIMDModeSourceLocatorFlowstart
)

var (
	ctx    context.Context
	dbPool *pgxpool.Pool
)

type result struct {
	Time                 time.Time
	SenderCallsign       string
	ReceiverCallsign     string
	SenderLocator        string
	ReceiverLocator      string
	Frequency            uint64
	SNR                  int8
	IMD                  uint8
	DecoderSoftware      string
	AntennaInformation   string
	Mode                 string
	InformationSource    uint8
	PersistentIdentifier string
}

func init() {
	var (
		err error
	)

	ctx = context.Background()

	dbPool, err = pgxpool.New(ctx, TimescaleDBConnectionString)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func TestSpotter(t *testing.T) {
	var (
		err     error
		spotter *Spotter
		rows    pgx.Rows
		results []result
	)

	spotter = NewSpotter(ReceiverHostport, FakespotCallsign, FakespotLocator, FakespotAntennaInformation, FakespotDecoderSoftware, "", FakespotSpotKind)

	for i := 0; i < FakespotCount; i++ {
		// TODO make reports random
		spotter.Feed(NewSpot("N1CALL", "II00OG", 50313650+uint64(i), -23, 42, "FT8", 1, uint32(time.Now().UTC().Unix())))
		// TODO record (hashes of) reports' values
	}

	// TODO poll for expected number of rows
	time.Sleep(30 * time.Second)

	t.Logf("Connected to database pool %+v", dbPool)

	rows, err = dbPool.Query(ctx, ReportsQuery)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Rows: %+v", rows)

	for rows.Next() {
		var r result
		err = rows.Scan(&r.Time, &r.SenderCallsign, &r.ReceiverCallsign, &r.SenderLocator, &r.ReceiverLocator, &r.Frequency, &r.SNR, &r.IMD, &r.DecoderSoftware, &r.AntennaInformation, &r.Mode, &r.InformationSource, &r.PersistentIdentifier)

		if err != nil {
			t.Fatal(err)
		}

		// TODO remove result from recorded hashes
		results = append(results, r)
	}

	// TODO expect no hashes remaining

	t.Logf("%+v", results)
}
