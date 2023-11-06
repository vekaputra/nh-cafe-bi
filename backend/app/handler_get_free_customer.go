package app

import "net/http"

const (
	getFreeCustomerQuery = `SELECT 
    c.id as customer_id,
    c.customer_code, 
    c.name as customer_name,
    b.id as branch_id,
    b.branch_code, 
    b.short_name as branch_short_name,
    MIN(mt.transaction_date) as transaction_date
FROM customers c
    JOIN monthly_transactions mt ON c.id = mt.customer_id
    JOIN branches b ON b.id = mt.branch_id
	LEFT JOIN (SELECT CONCAT(crm.id, "-",rf.id) as id, crm.customer_id, rf.branch_id, crm.assigned_at 
FROM customer_referral_mappings crm JOIN referral_fees rf ON crm.referral_fee_id = rf.id
WHERE rf.is_root_referral = 1) as crm_rf ON c.id = crm_rf.customer_id AND crm_rf.assigned_at <= mt.transaction_date AND mt.branch_id = crm_rf.branch_id
WHERE crm_rf.id IS NULL
GROUP BY c.id, c.customer_code, c.name, b.id, b.branch_code, b.short_name
ORDER BY c.customer_code ASC;`
)

func GetFreeCustomerHandler(w http.ResponseWriter, r *http.Request) {
	db := GetDB()

	freeCustomers := []FreeCustomer{}
	err := db.Select(&freeCustomers, getFreeCustomerQuery)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ReturnJson(w, freeCustomers, http.StatusOK)
}
