package modules

/**
{
	"id": 1,
	"user_id": 1,
	"account_name": "Test Test",
	"account_number": "1234",
	"account_type": "CREDIT",
	"expire_month": "01",
	"expire_year": "12"
}
**/
type Payment struct {
	Id            int32  `json:"id"`
	UserId        int32  `json:"user_id"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	ExpireMonth   string `json:"expire_month"`
	ExpireYear    string `json:"expire_year"`
}

/**
{
	"user_id": 1,
	"account_name": "Test Test",
	"account_number": "1234123412341234",
	"account_type": "CREDIT",
	"code": "333",
	"expire_month": "01",
	"expire_year": "12"
}
**/
type PaymentNew struct {
	UserId        int32  `json:"user_id"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	Code          string `json:"code"`
	ExpireMonth   string `json:"expire_month"`
	ExpireYear    string `json:"expire_year"`
}

/**
{
	"id": 1,
	"user_id": 1
}
**/
type PaymentRequest struct {
	Id     int32 `json:"id"`
	UserId int32 `json:"user_id"`
}

/**
{
	"user_id": 1
}
**/
type PaymentRequestForUser struct {
	UserId int32 `json:"user_id"`
}

/**
{
	"id": 1,
	"user_id": 1,
	"account_name": "Test Test",
	"account_number": "1234123412341234",
	"account_type": "CREDIT",
	"code": "333",
	"expire_month": "01",
	"expire_year": "12"
}
**/
type UpdatePayment struct {
	Id            int32  `json:"id"`
	UserId        int32  `json:"user_id"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	Code          string `json:"code"`
	ExpireMonth   string `json:"expire_month"`
	ExpireYear    string `json:"expire_year"`
}
