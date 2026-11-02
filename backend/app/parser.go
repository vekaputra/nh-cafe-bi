package app

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	// columns are read by position, so a report with any other layout is rejected
	spkluHeaderRow1 = "TANGGAL START|TANGGAL STOP|SPKLU|EVSE|CONNECTOR|kWh|BIAYA LISTRIK|PPJ PLN|PPJ MITRA|Komposisi Revenue Sharing|||SHARING REVENUE|PPN Biaya Layanan||PPN TOTAL|MATERAI|Total Revenue||ADMIN|PENDAPATAN SPKLU|STATUS PAYMENT|TANGGAL BAYAR"
	spkluHeaderRow2 = "TANGGAL START|TANGGAL STOP|SPKLU|EVSE|CONNECTOR|kWh|BIAYA LISTRIK|PPJ PLN|PPJ MITRA|Platform\nPT PLN (PERSERO)|Charger\nPT HENING HIMPUN ENERGI|Lahan\nPT HENING HIMPUN ENERGI|SHARING REVENUE|PT PLN (PERSERO)|PT HENING HIMPUN ENERGI|PPN TOTAL|MATERAI|PT PLN (PERSERO)|PT HENING HIMPUN ENERGI|ADMIN|PENDAPATAN SPKLU|STATUS PAYMENT|TANGGAL BAYAR"

	spkluColumnCount    = 23
	spkluDataRowOffset  = 3
	spkluEmptyCell      = "-"
	spkluDateTimeLayout = "02 Jan 2006 15:04:05"
)

var (
	spkluMonthReplacer = strings.NewReplacer("MEI", "MAY", "AGT", "AUG", "AGS", "AUG", "AGU", "AUG", "OKT", "OCT", "DES", "DEC")
	// the report prints UTC+7, the mysql driver then writes these instants to DATETIME as UTC
	spkluLocation = time.FixedZone("UTC+7", 7*60*60)
)

func ParseSPKLUSheet(rows [][]string) (sessions []SPKLUSession, skipped int, err error) {
	if len(rows) < spkluDataRowOffset ||
		strings.Join(rows[0], "|") != spkluHeaderRow1 ||
		strings.Join(rows[1], "|") != spkluHeaderRow2 {
		return nil, 0, errors.New("unexpected SPKLU report layout")
	}

	for i, row := range rows[spkluDataRowOffset:] {
		if len(row) == 0 {
			continue
		}

		rowNumber := i + spkluDataRowOffset + 1
		if len(row) != spkluColumnCount {
			return nil, 0, fmt.Errorf("row %d: expected %d columns, got %d", rowNumber, spkluColumnCount, len(row))
		}

		// still charging or not calculated yet, a later report carries the final figures
		if row[1] == spkluEmptyCell || row[20] == spkluEmptyCell {
			skipped++
			continue
		}

		session, err := parseSPKLURow(row)
		if err != nil {
			return nil, 0, fmt.Errorf("row %d: %w", rowNumber, err)
		}
		sessions = append(sessions, session)
	}

	return sessions, skipped, nil
}

func parseSPKLURow(row []string) (session SPKLUSession, err error) {
	for _, value := range row[5:21] {
		if _, err = strconv.ParseFloat(value, 64); err != nil {
			return session, err
		}
	}

	startedAt, err := parseSPKLUTime(row[0])
	if err != nil {
		return session, err
	}
	stoppedAt, err := parseSPKLUTime(row[1])
	if err != nil {
		return session, err
	}

	return SPKLUSession{
		StationName:     row[2],
		EVSEName:        row[3],
		ConnectorName:   row[4],
		StartedAt:       startedAt,
		StoppedAt:       stoppedAt,
		Kwh:             row[5],
		ElectricityCost: row[6],
		PlnPpj:          row[7],
		HhePpj:          row[8],
		PlatformShare:   row[9],
		ChargerShare:    row[10],
		LandShare:       row[11],
		SharingRevenue:  row[12],
		PlnPpn:          row[13],
		HhePpn:          row[14],
		TotalPpn:        row[15],
		StampDuty:       row[16],
		PlnRevenue:      row[17],
		HheRevenue:      row[18],
		AdminFee:        row[19],
		TotalRevenue:    row[20],
	}, nil
}

func parseSPKLUTime(value string) (time.Time, error) {
	return time.ParseInLocation(spkluDateTimeLayout, spkluMonthReplacer.Replace(value), spkluLocation)
}

func ParseMonthlyCSV(data [][]string) (result []Transaction) {
	for _, row := range data {
		cleanRow := []string{}
		for _, col := range row {
			if col == "" {
				continue
			}
			cleanRow = append(cleanRow, col)
		}
		if len(cleanRow) == 8 {
			nameToken := strings.Split(cleanRow[1], " - ")
			result = append(result, Transaction{
				CustomerCode: nameToken[0],
				CustomerName: nameToken[1],
				BuyAmount:    clearDotToInt(cleanRow[2]),
				SellAmount:   clearDotToInt(cleanRow[3]),
				BuyFee:       clearDotToInt(cleanRow[4]),
				SellFee:      clearDotToInt(cleanRow[5]),
				TotalFee:     clearDotToInt(cleanRow[6]),
			})
		}
	}

	return result
}

func clearDotToInt(value string) int64 {
	if value == "." {
		return 0
	}

	result, err := strconv.ParseInt(strings.Split(value, ".")[0], 10, 64)
	if err != nil {
		log.Println(err)
		return -1
	}

	return result
}
