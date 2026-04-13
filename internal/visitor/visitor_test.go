package visitor

import (
	"testing"
	"time"
)

func TestParseDate_UsesUTC(t *testing.T) {
	parsed := parseDate("2026-01-02")
	if parsed.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %s", parsed.Location())
	}
}

func TestRecordIP_UsesUTCDayBoundary(t *testing.T) {
	state := runtimeState{
		byDay:    make(map[string]DayStat),
		todayUV:  make(map[string]struct{}),
		lastPVAt: make(map[string]time.Time),
	}

	// 2026-01-02 01:00:00 +0800 == 2026-01-01 17:00:00 UTC
	ev := recordEvent{
		ipHash: "ip1",
		at:     time.Date(2026, 1, 2, 1, 0, 0, 0, time.FixedZone("CST", 8*3600)),
	}
	recordIP(&state, ev)

	stat, ok := state.byDay["2026-01-01"]
	if !ok {
		t.Fatalf("expected UTC day bucket 2026-01-01, got %+v", state.byDay)
	}
	if stat.PV != 1 || stat.UV != 1 {
		t.Fatalf("expected pv=1 uv=1, got pv=%d uv=%d", stat.PV, stat.UV)
	}
}
