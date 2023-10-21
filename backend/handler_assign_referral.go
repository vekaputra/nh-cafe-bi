package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	insertReferralMappingBaseQuery = `INSERT INTO customer_referral_mappings (customer_id, referral_fee_id, assigned_at) VALUES <()>`
)

func AssignReferralHandler(w http.ResponseWriter, r *http.Request) {
	var payload []AssignReferralReq

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	var newReferralArgs []interface{}
	insertReferralArgsQuery := ""
	count := 0
	for _, referral := range payload {
		if referral.ReferralFeeID == -1 {
			continue
		}
		newReferralArgs = append(newReferralArgs, referral.CustomerID, referral.ReferralFeeID, referral.TransactionDate)
		if count == 0 {
			insertReferralArgsQuery = "(?, ?, ?)"
			count++
			continue
		}
		insertReferralArgsQuery = fmt.Sprintf("%s, (?, ?, ?)", insertReferralArgsQuery)
		count++
	}

	db := GetDB()
	insertReferralMappingQuery := strings.ReplaceAll(insertReferralMappingBaseQuery, "<()>", insertReferralArgsQuery)
	_, err = db.Exec(insertReferralMappingQuery, newReferralArgs...)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	ReturnMessage(w, fmt.Sprintf("success assign %d customer to referral", count), http.StatusOK)
}
