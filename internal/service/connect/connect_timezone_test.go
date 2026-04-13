package service

import (
	"testing"
	"time"
)

func TestSiteMetricsTimezone_IsFixedUTC(t *testing.T) {
	prevLocal := time.Local
	time.Local = time.FixedZone("Asia/Shanghai", 8*3600)
	t.Cleanup(func() {
		time.Local = prevLocal
	})

	if siteMetricsTimezone != "UTC" {
		t.Fatalf("expected site metrics timezone UTC, got %s", siteMetricsTimezone)
	}
}
