package app

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
)

const storeFileUploadQuery = `INSERT INTO file_uploads (file_name, file_hash, json, approved_at) VALUES (?, ?, ?, NULL);`

func UploadCSVHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(256 << 10) // max 256kb
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	date := r.FormValue("date")
	branchID, err := strconv.ParseInt(r.FormValue("branch_id"), 10, 64)
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

	if filepath.Ext(handler.Filename) != ".csv" {
		ReturnMessage(w, "invalid file type", http.StatusBadRequest)
		return
	}

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1
	csvContent, err := csvReader.ReadAll()
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := ParseMonthlyCSV(csvContent)
	response := MonthlyCSV{
		BranchID:     branchID,
		Date:         date,
		Transactions: result,
	}

	// Store to DB
	db := GetDB()

	csvByte, _ := json.Marshal(csvContent)
	h := sha256.New()
	h.Write(csvByte)
	h.Sum(nil)

	jsonByte, _ := json.Marshal(response)

	_, err = db.Exec(storeFileUploadQuery, handler.Filename, fmt.Sprintf("%x", h.Sum(nil)), jsonByte)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	ReturnMessage(w, "success upload file", http.StatusOK)
}
