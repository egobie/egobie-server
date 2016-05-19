package modules

/**
{
	"id": 1,
	"user_id": 1,
	"account_name": "Test Test",
	"account_number": "1234",
	"account_type": "CREDIT",
	"account_zip": "07601"
	"card_type": "Visa",
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
	AccountZip    string `json:"account_zip"`
	CardType      string `json:"card_type"`
	Code          string `json:"code"`
	ExpireMonth   string `json:"expire_month"`
	ExpireYear    string `json:"expire_year"`
	Reserved      int32  `json:"reserved"`
}

/**
{
	"user_id": 1,
	"account_name": "Test Test",
	"account_number": "1234123412341234",
	"account_type": "CREDIT",
	"account_zip": "07601",
	"card_type": "Visa",
	"code": "333",
	"expire_month": "01",
	"expire_year": "12"
}
**/
type PaymentNew struct {
	BaseRequest

	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	AccountZip    string `json:"account_zip"`
	CardType      string `json:"card_type"`
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
	BaseRequest

	Id     int32 `json:"id"`
}

/**
{
	"id": 1,
	"user_id": 1,
	"account_name": "Test Test",
	"account_number": "1234123412341234",
	"account_type": "CREDIT",
	"account_zip": "07601",
	"card_type": "Visa",
	"code": "333",
	"expire_month": "01",
	"expire_year": "12"
}
**/
type UpdatePayment struct {
	BaseRequest

	Id            int32  `json:"id"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	AccountZip    string `json:"account_zip"`
	CardType      string `json:"card_type"`
	Code          string `json:"code"`
	ExpireMonth   string `json:"expire_month"`
	ExpireYear    string `json:"expire_year"`
}

/**
{
	"user_id": 1,
	"service_id": 1
}
**/
type ProcessRequest struct {
	BaseRequest

	PaymentId int32 `json:"payment_id"`
	ServiceId int32 `json:"service_id"`
}
