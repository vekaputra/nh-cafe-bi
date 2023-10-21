package app

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const storePaymentQuery = `INSERT INTO monthly_payments (branch_id, amount, payment_date) VALUES (?, ?, ?);`

func AddPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var payload MonthlyPayment

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := GetDB()
	_, err = db.Exec(storePaymentQuery, payload.BranchID, payload.Amount, payload.PaymentDate)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ReturnMessage(w, fmt.Sprintf("success add payment %d at %s", payload.BranchID, payload.PaymentDate), http.StatusOK)
}
