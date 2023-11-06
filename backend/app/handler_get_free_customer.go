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
    LEFT JOIN customer_referral_mappings crm ON c.id = crm.customer_id AND crm.assigned_at <= mt.transaction_date
	LEFT JOIN referral_fees rf ON crm.referral_fee_id = rf.id AND mt.branch_id = rf.branch_id AND rf.assigned_at <= mt.transaction_date
WHERE crm.id IS NULL OR rf.id IS NULL
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
