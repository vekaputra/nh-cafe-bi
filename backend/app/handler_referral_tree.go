package app

import (
	"fmt"
	"net/http"
)

func GetReferralTreeHandler(w http.ResponseWriter, r *http.Request) {
	db := GetDB()

	var allFees []ReferralFee
	query := `
		SELECT 
			rf.id, rf.branch_id, rf.referral_id, rf.parent_id, rf.code, rf.display_code, 
			rf.sharing_fee, rf.is_handle_tax, rf.is_root_referral, rf.assigned_at, rf.created_at,
			b.short_name AS branch_name
		FROM referral_fees rf
		JOIN branches b ON rf.branch_id = b.id
		WHERE rf.code LIKE 'T_%' 
		ORDER BY rf.id ASC`
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

	// Link children to parents
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

	// Optional filtering by a specific referral ID
	referralIDStr := r.URL.Query().Get("referral_id")
	if referralIDStr != "" {
		var targetID int64
		_, err := fmt.Sscanf(referralIDStr, "%d", &targetID)
		if err == nil {
			// Find the tree that contains this ID
			var filteredRoots []*ReferralFee
			for _, root := range roots {
				if containsID(root, targetID) {
					filteredRoots = append(filteredRoots, root)
					break // Assuming one root contains one ID in a tree structure
				}
			}
			roots = filteredRoots
		}
	}

	ReturnJson(w, roots, http.StatusOK)
}

func containsID(node *ReferralFee, id int64) bool {
	if node.ID == id {
		return true
	}
	for _, child := range node.Children {
		if containsID(child, id) {
			return true
		}
	}
	return false
}
