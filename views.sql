-- VIEW QUERY referral_fee_recursives

CREATE OR REPLACE VIEW referral_fee_recursives AS
  WITH RECURSIVE cte AS (
    SELECT id as root_id, id, referral_id, branch_id, code, display_code, parent_id, sharing_fee, is_handle_tax, is_root_referral, assigned_at
    FROM referral_fees
    UNION ALL
    SELECT cte.root_id, rf.id, rf.referral_id, rf.branch_id, rf.code, rf.display_code, rf.parent_id, rf.sharing_fee, rf.is_handle_tax, rf.is_root_referral, rf.assigned_at
    FROM referral_fees rf JOIN cte ON cte.parent_id = rf.id
  )
  SELECT 
    root_id,
    id as referral_fee_id,
    referral_id,
    branch_id,
    parent_id,
    is_root_referral,
    code,
    display_code,
    sharing_fee,
    (SELECT sum(c1.sharing_fee) FROM cte c1 WHERE c1.root_id = cte.root_id) as total_sharing_fee,
    CASE
        WHEN is_handle_tax = 1 THEN sharing_fee / (SELECT sum(c1.sharing_fee) FROM cte c1 WHERE c1.root_id = cte.root_id AND is_handle_tax = 1)
        ELSE 0
    END as sharing_tax,
    assigned_at
  FROM cte
  WHERE root_id IN (SELECT id FROM referral_fees rf WHERE rf.is_root_referral = 1)
  ORDER BY root_id, sharing_fee;

-- VIEW QUERY referral_transactions

CREATE OR REPLACE VIEW referral_transactions AS
WITH mtf AS (
    SELECT
        ROUND(mp.amount / (mp.nett_rate * SUM(mt.total_fee)), 6) as fee_modifier,
        ROUND(1-mp.nett_rate, 6) as tax_rate,
        mt.transaction_date,
        mt.branch_id
    FROM monthly_transactions mt
        LEFT JOIN monthly_payments mp ON mp.branch_id = mt.branch_id AND mp.payment_date = mt.transaction_date
    GROUP BY mt.transaction_date, mt.branch_id
)
SELECT
    rfr.root_id,
    rfr.referral_fee_id,
    rfr.parent_id,
    rfr.is_root_referral,
    rfr.branch_id,
    c.customer_code,
    c.name,
    b.short_name as branch,
    rfr.display_code,
    mt.buy_amount,
    mt.sell_amount,
    mt.total_fee,
    ROUND(mt.total_fee * mtf.fee_modifier) as gross_shared_fee,
    ROUND(mt.total_fee * mtf.fee_modifier * (rfr.sharing_fee / b.sharing_fee)) as gross_referral_fee,
    ROUND(mt.total_fee * mtf.fee_modifier * mtf.tax_rate) as shared_tax,
    ROUND(mt.total_fee * mtf.fee_modifier * mtf.tax_rate * rfr.sharing_tax) as tax,
    GREATEST(ROUND(mt.total_fee * mtf.fee_modifier * ((rfr.sharing_fee / b.sharing_fee) - (mtf.tax_rate * rfr.sharing_tax))), 0) as nett_referral_fee,
    mt.transaction_date
FROM monthly_transactions mt
         JOIN branches b ON mt.branch_id = b.id
         JOIN customers c ON mt.customer_id = c.id
         JOIN monthly_customer_referral_mappings mcrm ON mcrm.customer_id = mt.customer_id AND mcrm.transaction_date = mt.transaction_date AND mcrm.branch_id = mt.branch_id
         JOIN referral_fee_recursives rfr ON mcrm.referral_fee_id = rfr.root_id
         JOIN mtf ON mtf.branch_id = mt.branch_id AND mtf.transaction_date = mt.transaction_date
ORDER BY mt.transaction_date, c.customer_code ASC;

-- monthly_customer_referral_mappings

CREATE OR REPLACE VIEW monthly_customer_referral_mappings AS
SELECT DISTINCT mt.transaction_date, crm.customer_id, crm.referral_fee_id, crm.assigned_at, crm.branch_id
FROM monthly_transactions mt JOIN (
  SELECT c1.customer_id, c1.referral_fee_id, c1.assigned_at, r1.branch_id 
  FROM customer_referral_mappings c1 JOIN referral_fees r1 ON c1.referral_fee_id = r1.id) crm 
  	ON mt.customer_id = crm.customer_id AND mt.branch_id = crm.branch_id 
WHERE crm.assigned_at = (
        SELECT c2.assigned_at 
        FROM customer_referral_mappings c2 JOIN referral_fees r2 ON c2.referral_fee_id = r2.id
        WHERE c2.customer_id = mt.customer_id AND r2.branch_id = mt.branch_id AND c2.assigned_at <= mt.transaction_date
        ORDER BY c2.assigned_at DESC LIMIT 1)
ORDER BY transaction_date;

---

SELECT c2.assigned_at 
        FROM customer_referral_mappings c2 JOIN referral_fees r2 ON c2.referral_fee_id = r2.id
        WHERE c2.customer_id = 176 AND c2.assigned_at <= '2023-08-01' AND r2.branch_id = 5
        ORDER BY c2.assigned_at DESC;
---

SELECT (1 - (SUM(amount)/SUM(amount/tax_rate)))*100 as tax_rate FROM monthly_payments mp
WHERE mp.branch_id IN (SELECT id FROM branches WHERE branch_code IN ($branch_code))
  AND mp.payment_date IN ($transaction_date)