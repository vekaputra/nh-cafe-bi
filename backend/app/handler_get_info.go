package app

import "net/http"

const (
	getBranchQuery      = `SELECT id, branch_code, short_name, name, sharing_fee, created_at FROM branches WHERE branch_code NOT IN ('00000') AND is_active = 1 ORDER BY branch_code ASC;`
	getTxDateQuery      = `SELECT transaction_date FROM monthly_transaction_dates WHERE transaction_date >= '2023-01-01' ORDER BY transaction_date DESC LIMIT 12 OFFSET 1;`
	getReferralFeeQuery = `SELECT id, branch_id, referral_id, parent_id, code, display_code, sharing_fee, is_handle_tax, is_root_referral, assigned_at, created_at FROM referral_fees WHERE is_root_referral = 1 ORDER BY display_code ASC;`
)

func GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	db := GetDB()

	response := Info{}
	err := db.Select(&response.Branches, getBranchQuery)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.Select(&response.TransactionDates, getTxDateQuery)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.Select(&response.ReferralFees, getReferralFeeQuery)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ReturnJson(w, response, http.StatusOK)
}
