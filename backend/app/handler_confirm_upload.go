package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	getUploadByIDQuery         = `SELECT id, file_name, file_hash, json, created_at FROM file_uploads WHERE id = ?;`
	getCustomerBaseQuery       = `SELECT id, customer_code, name, created_at FROM customers WHERE customer_code IN (<?>);`
	insertCustomerBaseQuery    = `INSERT INTO customers (customer_code, name) VALUES <()>;`
	insertTransactionBaseQuery = `INSERT INTO monthly_transactions (customer_id, branch_id, buy_amount, sell_amount, buy_fee, sell_fee, total_fee, transaction_date) VALUES <()>;`
	approveFileUploadQuery     = `UPDATE file_uploads SET approved_at = current_timestamp() WHERE id = ?;`
	deleteFileUploadQuery      = `DELETE FROM file_uploads WHERE id = ?;`

	actionApprove = "approve"
	actionDelete  = "delete"
)

func ConfirmUploadHandler(w http.ResponseWriter, r *http.Request) {
	var payload ConfirmUploadReq

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := GetDB()
	switch payload.Action {
	case actionApprove:
		// Store to DB
		var upload FileUpload
		err = db.Get(&upload, getUploadByIDQuery, payload.FileUploadID)
		if err != nil {
			ReturnMessage(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var csvPayload MonthlyCSV
		err = json.Unmarshal(upload.Json, &csvPayload)
		if err != nil {
			ReturnMessage(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var activeCustomers []interface{}
		questionStatement := ""
		for i, row := range csvPayload.Transactions {
			activeCustomers = append(activeCustomers, row.CustomerCode)
			if i == 0 {
				questionStatement = "?"
				continue
			}
			questionStatement = fmt.Sprintf("%s, ?", questionStatement)
		}

		customers := []Customer{}
		getCustomerQuery := strings.ReplaceAll(getCustomerBaseQuery, "<?>", questionStatement)
		err = db.Select(&customers, getCustomerQuery, activeCustomers...)
		if err != nil {
			ReturnMessage(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mapCustomers := map[string]Customer{}
		for _, customer := range customers {
			mapCustomers[customer.CustomerCode] = customer
		}

		var newCustomersArgs []interface{}
		insertCustomerArgsQuery := ""
		count := 0
		for _, row := range csvPayload.Transactions {
			if _, ok := mapCustomers[row.CustomerCode]; !ok {
				newCustomersArgs = append(newCustomersArgs, row.CustomerCode, row.CustomerName)
				if count == 0 {
					insertCustomerArgsQuery = "(?, ?)"
					count++
					continue
				}
				insertCustomerArgsQuery = fmt.Sprintf("%s, (?, ?)", insertCustomerArgsQuery)
				count++
			}
		}

		if len(newCustomersArgs) > 0 {
			insertCustomerQuery := strings.ReplaceAll(insertCustomerBaseQuery, "<()>", insertCustomerArgsQuery)
			_, err = db.Exec(insertCustomerQuery, newCustomersArgs...)
			if err != nil {
				ReturnMessage(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = db.Select(&customers, getCustomerQuery, activeCustomers...)
			if err != nil {
				ReturnMessage(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, customer := range customers {
				mapCustomers[customer.CustomerCode] = customer
			}
		}

		var newTransactionArgs []interface{}
		insertTransactionArgsQuery := ""
		for i, row := range csvPayload.Transactions {
			newTransactionArgs = append(
				newTransactionArgs,
				mapCustomers[row.CustomerCode].ID,
				csvPayload.BranchID,
				row.BuyAmount,
				row.SellAmount,
				row.BuyFee,
				row.SellFee,
				row.TotalFee,
				csvPayload.Date,
			)
			if i == 0 {
				insertTransactionArgsQuery = "(?, ?, ?, ?, ?, ?, ?, ?)"
				continue
			}
			insertTransactionArgsQuery = fmt.Sprintf("%s, (?, ?, ?, ?, ?, ?, ?, ?)", insertTransactionArgsQuery)
			continue
		}

		insertTransactionQuery := strings.ReplaceAll(insertTransactionBaseQuery, "<()>", insertTransactionArgsQuery)
		_, err = db.Exec(insertTransactionQuery, newTransactionArgs...)
		if err != nil {
			ReturnMessage(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(approveFileUploadQuery, payload.FileUploadID)
		if err != nil {
			ReturnMessage(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case actionDelete:
		fallthrough
	default:
		_, err = db.Exec(deleteFileUploadQuery, payload.FileUploadID)
		if err != nil {
			ReturnMessage(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	ReturnMessage(w, fmt.Sprintf("file %d, %s success", payload.FileUploadID, payload.Action), http.StatusOK)
}
