package app

import "net/http"

const (
	getNonApprovedUploadQuery = `SELECT id, file_name, file_hash, json, created_at FROM file_uploads WHERE approved_at IS NULL ORDER BY created_at ASC;`
)

func GetUploadHandler(w http.ResponseWriter, r *http.Request) {
	db := GetDB()

	var response []FileUpload
	err := db.Select(&response, getNonApprovedUploadQuery)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(response) == 0 {
		ReturnJson(w, []interface{}{}, http.StatusOK)
		return
	}

	ReturnJson(w, response, http.StatusOK)
}
