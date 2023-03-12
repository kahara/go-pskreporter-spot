//go:build integration

package spot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"testing"
	"time"
)

const (
	TimescaleDBConnectionString = "postgres://postgres:password@timescaledb"
	CountQuery                  = "SELECT COUNT(*) FROM report;"
	ReportsQuery                = "SELECT * FROM report;"
	ReceiverHostport            = "receiver:4739"
	FakespotCount               = 2250
	FakespotCallsign            = "N0CALL"
	FakespotLocator             = "JJ00OG"
	FakespotAntennaInformation  = "Dipole"
	FakespotDecoderSoftware     = "fakespot v0"
	FakespotSpotKind            = SpotKind_CallsignFrequencySNRIMDModeSourceLocatorFlowstart
)

var (
	ctx    context.Context
	dbPool *pgxpool.Pool
	wiggle uint64 = 0
)

type count struct {
	Count uint64
}

type report struct {
	Time                 time.Time
	SenderCallsign       pgtype.Text
	ReceiverCallsign     pgtype.Text
	SenderLocator        pgtype.Text
	ReceiverLocator      pgtype.Text
	Frequency            uint64
	SNR                  int8
	IMD                  uint8
	DecoderSoftware      pgtype.Text
	AntennaInformation   pgtype.Text
	Mode                 pgtype.Text
	InformationSource    uint8
	PersistentIdentifier pgtype.Text
}

func fakespot() *Spot {
	// TODO make reports random
	wiggle += 1
	return NewSpot(
		"N1CALL", "II00OG", 50313650+wiggle, 23, 42, "FT8", 1, uint32(time.Now().UTC().Unix()),
	)
}

func init() {
	var err error
	ctx = context.Background()
	dbPool, err = pgxpool.New(ctx, TimescaleDBConnectionString)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func TestSpotterIntegration(t *testing.T) {
	var (
		err     error
		spotter *Spotter
		hashes  [][32]byte
		rows    pgx.Rows
	)

	spotter = NewSpotter(ReceiverHostport, FakespotCallsign, FakespotLocator, FakespotAntennaInformation, FakespotDecoderSoftware, "", FakespotSpotKind)

	for i := 0; i < FakespotCount; i++ {
		// TODO make reports random
		spot := fakespot()
		hashes = append(hashes, sha256.Sum256([]byte(fmt.Sprintf("%x", spot))))
		spotter.Feed(spot)
	}

	t.Logf("Connected to database pool %+v", dbPool)

	// Poll until all spots have supposedly landed in the table
	ticker := time.NewTicker(1 * time.Second)
Poll:
	for {
		// FIXME timeout if no spots are appearing
		select {
		case <-ticker.C:
			t.Logf("polling")
			rows, err = dbPool.Query(ctx, CountQuery)
			if err != nil {
				t.Fatal(err)
			}
			rows.Next()
			var c count
			err = rows.Scan(&c.Count)
			if err != nil {
				t.Error(err)
				continue
			}
			t.Logf("count is %d", c.Count)
			if c.Count == FakespotCount {
				break Poll
			}
		}
	}

	// Read all spots
	rows, err = dbPool.Query(ctx, ReportsQuery)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var r report
		err = rows.Scan(&r.Time, &r.SenderCallsign, &r.ReceiverCallsign, &r.SenderLocator, &r.ReceiverLocator, &r.Frequency, &r.SNR, &r.IMD, &r.DecoderSoftware, &r.AntennaInformation, &r.Mode, &r.InformationSource, &r.PersistentIdentifier)
		if err != nil {
			t.Fatal(err)
		}

		// TODO check that data made the roundtrip intact
		t.Logf("%+v", r)

		// TODO remove result from recorded hashes
	}

	// TODO expect no hashes remaining
}
