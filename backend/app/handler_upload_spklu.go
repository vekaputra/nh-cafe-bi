package app

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/xuri/excelize/v2"
)

const (
	maxSPKLUUploadSize = 1 << 20  // matches nginx's default client_max_body_size
	maxSPKLUUnzipSize  = 32 << 20 // the container only has 64M

	upsertSPKLUStationQuery   = `INSERT INTO spklu_stations (name) VALUES (?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id);`
	upsertSPKLUEVSEQuery      = `INSERT INTO spklu_evses (station_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id);`
	upsertSPKLUConnectorQuery = `INSERT INTO spklu_connectors (evse_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id);`

	// reports overlap, so a re-upload refreshes sessions already loaded instead of duplicating them
	upsertSPKLUSessionQuery = `INSERT INTO spklu_charging_sessions (
		connector_id, started_at, stopped_at, kwh, electricity_cost, pln_ppj, hhe_ppj, platform_share, charger_share,
		land_share, sharing_revenue, pln_ppn, hhe_ppn, total_ppn, stamp_duty, pln_revenue, hhe_revenue, admin_fee,
		total_revenue
	) VALUES (
		:connector_id, :started_at, :stopped_at, :kwh, :electricity_cost, :pln_ppj, :hhe_ppj, :platform_share, :charger_share,
		:land_share, :sharing_revenue, :pln_ppn, :hhe_ppn, :total_ppn, :stamp_duty, :pln_revenue, :hhe_revenue, :admin_fee,
		:total_revenue
	) ON DUPLICATE KEY UPDATE
		stopped_at = VALUES(stopped_at),
		kwh = VALUES(kwh),
		electricity_cost = VALUES(electricity_cost),
		pln_ppj = VALUES(pln_ppj),
		hhe_ppj = VALUES(hhe_ppj),
		platform_share = VALUES(platform_share),
		charger_share = VALUES(charger_share),
		land_share = VALUES(land_share),
		sharing_revenue = VALUES(sharing_revenue),
		pln_ppn = VALUES(pln_ppn),
		hhe_ppn = VALUES(hhe_ppn),
		total_ppn = VALUES(total_ppn),
		stamp_duty = VALUES(stamp_duty),
		pln_revenue = VALUES(pln_revenue),
		hhe_revenue = VALUES(hhe_revenue),
		admin_fee = VALUES(admin_fee),
		total_revenue = VALUES(total_revenue);`
)

func UploadSPKLUHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSPKLUUploadSize)
	err := r.ParseMultipartForm(256 << 10) // max 256kb
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if filepath.Ext(handler.Filename) != ".xlsx" {
		ReturnMessage(w, "invalid file type", http.StatusBadRequest)
		return
	}

	workbook, err := excelize.OpenReader(file, excelize.Options{UnzipSizeLimit: maxSPKLUUnzipSize})
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer workbook.Close()

	var sessions []SPKLUSession
	skipped := 0
	for _, sheet := range workbook.GetSheetList() {
		// raw values, otherwise the cell number format would round the amounts
		rows, err := workbook.GetRows(sheet, excelize.Options{RawCellValue: true})
		if err != nil {
			ReturnMessage(w, err.Error(), http.StatusBadRequest)
			return
		}

		sheetSessions, sheetSkipped, err := ParseSPKLUSheet(rows)
		if err != nil {
			ReturnMessage(w, fmt.Sprintf("%s: %s", sheet, err.Error()), http.StatusBadRequest)
			return
		}
		sessions = append(sessions, sheetSessions...)
		skipped += sheetSkipped
	}

	inserted, updated, err := storeSPKLUSessions(sessions)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ReturnMessage(w, fmt.Sprintf("success: %d new, %d updated, %d unfinished skipped", inserted, updated, skipped), http.StatusOK)
}

// storeSPKLUSessions runs in one transaction so a failed upload leaves nothing half loaded
func storeSPKLUSessions(sessions []SPKLUSession) (inserted, updated int, err error) {
	tx, err := GetDB().Beginx()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	connectorIDs := map[[3]string]int64{}
	for _, session := range sessions {
		key := [3]string{session.StationName, session.EVSEName, session.ConnectorName}
		connectorID, ok := connectorIDs[key]
		if !ok {
			connectorID, err = upsertSPKLUConnector(tx, session)
			if err != nil {
				return 0, 0, err
			}
			connectorIDs[key] = connectorID
		}

		session.ConnectorID = connectorID
		result, err := tx.NamedExec(upsertSPKLUSessionQuery, session)
		if err != nil {
			return 0, 0, err
		}

		// mysql reports 1 for an insert, 2 for a changed row and 0 for an unchanged one
		affected, err := result.RowsAffected()
		if err != nil {
			return 0, 0, err
		}
		switch affected {
		case 1:
			inserted++
		case 2:
			updated++
		}
	}

	return inserted, updated, tx.Commit()
}

func upsertSPKLUConnector(tx *sqlx.Tx, session SPKLUSession) (int64, error) {
	stationID, err := upsertSPKLUID(tx, upsertSPKLUStationQuery, session.StationName)
	if err != nil {
		return 0, err
	}

	evseID, err := upsertSPKLUID(tx, upsertSPKLUEVSEQuery, stationID, session.EVSEName)
	if err != nil {
		return 0, err
	}

	return upsertSPKLUID(tx, upsertSPKLUConnectorQuery, evseID, session.ConnectorName)
}

func upsertSPKLUID(tx *sqlx.Tx, query string, args ...interface{}) (int64, error) {
	result, err := tx.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
