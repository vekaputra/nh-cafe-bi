package app

import (
	"strings"
	"testing"
	"time"
)

func spkluTestRow(start, stop, totalRevenue string) []string {
	return []string{
		start, stop, "SPKLU A", "1. EVSE", "A. CCS2", "39.664", "57302.5808", "1216.193476", "1719.077424",
		"810.8", "27567.05", "12161.93", "40540", "0", "0", "0", "0", "2027", "98751", "0", totalRevenue, "Selesai", "-",
	}
}

func spkluTestSheet(rows ...[]string) [][]string {
	return append([][]string{
		strings.Split(spkluHeaderRow1, "|"),
		strings.Split(spkluHeaderRow2, "|"),
		{},
	}, rows...)
}

func TestParseSPKLUSheet(t *testing.T) {
	sessions, skipped, err := ParseSPKLUSheet(spkluTestSheet(
		spkluTestRow("05 MEI 2026 23:50:04", "06 MEI 2026 00:38:38", "100778"),
		spkluTestRow("14 AGU 2026 15:50:04", "14 AGU 2026 16:00:00", "100778"),
		spkluTestRow("02 NOV 2026 08:55:59", "-", "-"),
		spkluTestRow("02 DES 2026 08:55:59", "02 DES 2026 09:00:00", "-"),
	))
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 2 || len(sessions) != 2 {
		t.Fatalf("got %d sessions, %d skipped", len(sessions), skipped)
	}

	got := sessions[0]
	if want := time.Date(2026, time.May, 5, 16, 50, 4, 0, time.UTC); !got.StartedAt.Equal(want) {
		t.Errorf("StartedAt = %v", got.StartedAt)
	}
	if got.Kwh != "39.664" || got.LandShare != "12161.93" || got.PlnRevenue != "2027" || got.HheRevenue != "98751" || got.TotalRevenue != "100778" {
		t.Errorf("columns mapped wrong: %+v", got)
	}
	if sessions[1].StartedAt.Month() != time.August {
		t.Errorf("StartedAt = %v", sessions[1].StartedAt)
	}
}

func TestParseSPKLUSheetRejectsBadInput(t *testing.T) {
	badAmount := spkluTestRow("14 OKT 2026 15:50:04", "14 OKT 2026 16:38:38", "100778")
	badAmount[5] = "abc"

	tests := map[string][][]string{
		"other layout": {{"TANGGAL START"}, {}, {}},
		"bad amount":   spkluTestSheet(badAmount),
		"bad date":     spkluTestSheet(spkluTestRow("14 XXX 2026 15:50:04", "14 OKT 2026 16:38:38", "100778")),
		"short row":    spkluTestSheet(badAmount[:10]),
	}
	for name, rows := range tests {
		t.Run(name, func(t *testing.T) {
			if _, _, err := ParseSPKLUSheet(rows); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
