package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func MultiTransactionHandler(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")
	download := r.URL.Query().Get("download")

	if month == "" || year == "" {
		ReturnMessage(w, "month and year are required", http.StatusBadRequest)
		return
	}

	// target_date is the first day of the selected month and year
	targetDateStr := fmt.Sprintf("%s-%s-01", year, month)
	_, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		ReturnMessage(w, "invalid month or year", http.StatusBadRequest)
		return
	}

	// For the query, we want to replace NOW() with targetDate
	// In MySQL, we can use a string literal that represents the date.
	// We need to be careful about where NOW() is used.

	// DATE_FORMAT(DATE_ADD(?, INTERVAL 7 HOUR), '%Y%m%d') -> we can just pass the date
	// TIMESTAMPDIFF(MONTH, '2025-08-01', ?) -> replace NOW() with targetDate

	// transaction_date IN ($transaction_date)
	// We need to find all transaction dates in that month.
	// Actually, let's see how transaction_date is stored.
	// Based on handler_get_info.go: SELECT transaction_date FROM monthly_transaction_dates

	db := GetDB()
	var transactionDates []string
	queryDates := `SELECT transaction_date FROM monthly_transaction_dates WHERE transaction_date LIKE ?`
	err = db.Select(&transactionDates, queryDates, fmt.Sprintf("%s-%s-%%", year, month))
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(transactionDates) == 0 {
		ReturnMessage(w, "no transactions found for the selected month and year", http.StatusNotFound)
		return
	}

	quotedDates := make([]string, len(transactionDates))
	for i, d := range transactionDates {
		quotedDates[i] = fmt.Sprintf("'%s'", d)
	}
	transactionDateList := strings.Join(quotedDates, ",")

	sqlQuery := `
WITH RECURSIVE
-- Initialize row number counter
init_row_num AS (
SELECT 0 AS rn
),
-- Define the core commission data (COMM)
comm AS (
SELECT
b.transfer_type AS trf_type,
r.bank_account AS credited_acc,
CASE r.bank_name
WHEN 'BCA' THEN SUM(rt.nett_referral_fee)
ELSE SUM(rt.nett_referral_fee) - 2900 -- Apply fee reduction for non-BCA
END AS amount,
DATE_FORMAT(DATE_ADD(?, INTERVAL 7 HOUR), '%Y%m%d') AS eff_date,
b.bank_code AS bank_code,
b.bank_name AS bank_name,
r.name AS rec_name,
b.cust_type AS cust_type,
b.cust_residence AS cust_residence,
b.transaction_code AS trx_code
FROM
referral_transactions rt
JOIN referral_fees rf ON rt.referral_fee_id = rf.id
JOIN referrals r ON r.id = rf.referral_id
JOIN banks b ON b.name = r.bank_name
WHERE
rt.transaction_date IN (%s)
AND rt.branch_id NOT IN (20, 22, 24)
AND r.id NOT IN (1, 2, 49, 50, 60)
AND r.bank_account IS NOT NULL
GROUP BY
r.bank_account,
b.transfer_type,
b.bank_code,
b.bank_name,
r.name,
b.cust_type,
b.cust_residence,
b.transaction_code
HAVING
(b.bank_code = '' AND SUM(rt.nett_referral_fee) > 0) OR
(b.bank_code != '' AND SUM(rt.nett_referral_fee) - 2900 > 0)
ORDER BY
r.name ASC -- Ordered for reliable row numbering
),
-- 2. Generate the detailed transaction lines with row numbers
transaction_lines AS (
SELECT
ROW_NUMBER() OVER (ORDER BY T.rec_name, T.credited_acc) AS row_num,
CONCAT(
'1|00000NHCAFE',
DATE_FORMAT(?, '%%y%%m'),
LPAD(ROW_NUMBER() OVER (ORDER BY T.rec_name, T.credited_acc), 3, '0'),
'|',
T.trf_type,
'|||',
T.credited_acc,
'|',
T.amount,
'.00||||||komisi nh||',
T.bank_code,
'|',
T.bank_name,
'|',
T.rec_name,
'|',
T.cust_type,
'|',
T.cust_residence,
'|',
T.trx_code,
'|'
) AS line
FROM
comm T
),
-- 3. Calculate the total count of transaction lines
line_count AS (
SELECT
COUNT(1) AS total_lines
FROM
transaction_lines
)
-- 4. Final Union of Header and Data
SELECT
-- Generate the Header Line
CONCAT(
'0|FT|MD|kbbstefanp|',
LPAD((100 + (TIMESTAMPDIFF(MONTH, '2025-08-01', ?) * 5)), 8, '0'),
'|',
DATE_FORMAT(?, '%%Y%%m%%d'),
'||0878777878|OUR|0878777878|',
LPAD(LC.total_lines, 5, '0'), -- Use the count from the dedicated CTE
'|IDR|B|09|komisi nh|'
) AS line
FROM
line_count LC
UNION ALL
-- Append all Transaction Lines
SELECT
line
FROM
transaction_lines;
`
	// The sample result has 20260109 in the header, which is YYYYMMDD.
	// But the query in md has DATE_FORMAT(NOW(), '%Y%m') which is YYYYMM.
	// Looking at line 114: 0|FT|MD|kbbstefanp|00000125|20260109||0878777878|OUR|0878777878|00045|IDR|B|09|komisi nh|
	// 20260109 is definitely YYYYMMDD.
	// I will use targetDate as the date.

	finalSQL := fmt.Sprintf(sqlQuery, transactionDateList)

	var lines []string
	// Parameters: 1. eff_date, 2. DATE_FORMAT, 3. TIMESTAMPDIFF, 4. Header DATE_FORMAT
	err = db.Select(&lines, finalSQL, targetDateStr, targetDateStr, targetDateStr, targetDateStr, targetDateStr)
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := strings.Join(lines, "\n")

	if download == "true" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=multi_transaction_%s_%s.txt", year, month))
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(result))
		return
	}

	ReturnJson(w, map[string]string{"result": result}, http.StatusOK)
}
