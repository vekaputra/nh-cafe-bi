package app

import (
	"net/http"
)

func GetReferralsHandler(w http.ResponseWriter, r *http.Request) {
	db := GetDB()

	var referrals []Referral
	query := `SELECT id, name, bank_account, bank_name, created_at FROM referrals ORDER BY name ASC`
	err := db.Select(&referrals, query)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ReturnJson(w, referrals, http.StatusOK)
}
