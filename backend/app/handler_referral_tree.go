package app

import (
	"net/http"
)

func GetReferralTreeHandler(w http.ResponseWriter, r *http.Request) {
	db := GetDB()

	var allFees []ReferralFee
	query := `SELECT id, branch_id, referral_id, parent_id, code, display_code, sharing_fee, is_handle_tax, is_root_referral, assigned_at, created_at FROM referral_fees ORDER BY id ASC`
	err := db.Select(&allFees, query)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map to quickly find nodes by ID
	feeMap := make(map[int64]*ReferralFee)
	for i := range allFees {
		feeMap[allFees[i].ID] = &allFees[i]
	}

	var roots []*ReferralFee
	for i := range allFees {
		fee := &allFees[i]
		if fee.ParentID == nil || *fee.ParentID == 0 {
			roots = append(roots, fee)
		} else {
			if parent, ok := feeMap[*fee.ParentID]; ok {
				parent.Children = append(parent.Children, fee)
			} else {
				// If parent not found, treat as root or handle error
				roots = append(roots, fee)
			}
		}
	}

	ReturnJson(w, roots, http.StatusOK)
}
