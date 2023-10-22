package app

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

type AssignReferralReq struct {
	CustomerID      int64  `json:"customer_id"`
	TransactionDate string `json:"transaction_date"`
	ReferralFeeID   int64  `json:"referral_fee_id"`
}

type ConfirmUploadReq struct {
	FileUploadID int64  `json:"file_upload_id"`
	Action       string `json:"action"`
}

type AddPaymentReq struct {
	BranchID int64 `json:"branch_id"`
}

type MonthlyCSV struct {
	BranchID     int64         `json:"branch_id"`
	Date         string        `json:"date"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	CustomerCode string `json:"customer_code"`
	CustomerName string `json:"customer_name"`
	BuyAmount    int64  `json:"buy_amount"`
	SellAmount   int64  `json:"sell_amount"`
	BuyFee       int64  `json:"buy_fee"`
	SellFee      int64  `json:"sell_fee"`
	TotalFee     int64  `json:"total_fee"`
}

type Info struct {
	Branches         []Branch      `json:"branches"`
	TransactionDates []string      `json:"transaction_dates"`
	ReferralFees     []ReferralFee `json:"referral_fees"`
}

type Branch struct {
	ID         int64     `json:"id" db:"id"`
	BranchCode string    `json:"branch_code" db:"branch_code"`
	ShortName  string    `json:"short_name" db:"short_name"`
	Name       string    `json:"name" db:"name"`
	SharingFee float64   `json:"sharing_fee" db:"sharing_fee"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Customer struct {
	ID           int64     `db:"id"`
	CustomerCode string    `db:"customer_code"`
	Name         string    `db:"name"`
	CreatedAt    time.Time `db:"created_at"`
}

type CustomerReferralMapping struct {
	ID            int64     `db:"id"`
	CustomerID    int64     `db:"customer_id"`
	ReferralFeeID int64     `db:"referral_fee_id"`
	AssignedAt    string    `db:"date"`
	CreatedAt     time.Time `db:"created_at"`
}

type FileUpload struct {
	ID         int64          `json:"id" db:"id"`
	FileName   string         `json:"file_name" db:"file_name"`
	FileHash   string         `json:"file_hash" db:"file_hash"`
	Json       types.JSONText `json:"json" db:"json"`
	ApprovedAt *time.Time     `json:"approved_at" db:"approved_at"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}

type MonthlyPayment struct {
	ID          int64     `json:"id,omitempty" db:"id"`
	BranchID    int64     `json:"branch_id" db:"branch_id"`
	Amount      int64     `json:"amount" db:"amount"`
	PaymentDate string    `json:"payment_date" db:"payment_date"`
	CreatedAt   time.Time `json:"created_at,omitempty" db:"created_at"`
}

type MonthlyTransaction struct {
	ID              int64     `db:"id"`
	CustomerID      int64     `db:"customer_id"`
	BranchID        int64     `db:"branch_id"`
	BuyAmount       int64     `db:"buy_amount"`
	SellAmount      int64     `db:"sell_amount"`
	BuyFee          int64     `db:"buy_fee"`
	SellFee         int64     `db:"sell_fee"`
	TotalFee        int64     `db:"total_fee"`
	TransactionDate string    `db:"transaction_date"`
	CreatedAt       time.Time `db:"created_at"`
}

type ReferralFee struct {
	ID             int64     `json:"id" db:"id"`
	BranchID       int64     `json:"branch_id" db:"branch_id"`
	ReferralID     int64     `json:"referral_id" db:"referral_id"`
	ParentID       *int64    `json:"parent_id" db:"parent_id"`
	Code           string    `json:"code" db:"code"`
	DisplayCode    string    `json:"display_code" db:"display_code"`
	SharingFee     float64   `json:"sharing_fee" db:"sharing_fee"`
	IsHandleTax    bool      `json:"is_handle_tax" db:"is_handle_tax"`
	IsRootReferral bool      `json:"is_root_referral" db:"is_root_referral"`
	AssignedAt     string    `json:"assigned_at" db:"assigned_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type FreeCustomer struct {
	CustomerID      string `json:"customer_id" db:"customer_id"`
	CustomerCode    string `json:"customer_code" db:"customer_code"`
	CustomerName    string `json:"customer_name" db:"customer_name"`
	BranchID        string `json:"branch_id" db:"branch_id"`
	BranchCode      string `json:"branch_code" db:"branch_code"`
	BranchShortName string `json:"branch_short_name" db:"branch_short_name"`
	TransactionDate string `json:"transaction_date" db:"transaction_date"`
}
