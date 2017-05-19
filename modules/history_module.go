package modules

/**
{
	"id": 1,
	"ratting": 3.5,
	"user_service_id": "11",
	"note": "This is awesome"
}
**/
type History struct {
	Id            int32           `json:"id"`
	UserServiceId int32           `json:"user_service_id"`
	Rating        float32         `json:"rating"`
	Note          string          `json:"note"`
	ReservationId string          `json:"reservation_id"`
	AccountName   string          `json:"account_name"`
	AccountNumber string          `json:"account_number"`
	AccountType   string          `json:"account_type"`
	Price         float32         `json:"price"`
	Plate         string          `json:"plate"`
	State         string          `json:"state"`
	Year          int32           `json:"year"`
	Color         string          `json:"color"`
	Make          string          `json:"make"`
	Model         string          `json:"model"`
	StartTime     string          `json:"start_time"`
	EndTime       string          `json:"end_time"`
	Services      []SimpleService `json:"services"`
	Addons        []SimpleAddon   `json:"addons"`
}

/**
{
	"page": 0
}
**/
type HistoryRequest struct {
	BaseRequest

	Page int32 `json:"page"`
}

/**
{
	"service_id": 1,
	"rating": 3.5,
	"note": "this is awesome"
}
**/
type RatingRequest struct {
	BaseRequest

	Id        int32   `json:"id"`
	ServiceId int32   `json:"service_id"`
	Rating    float32 `json:"rating"`
	Note      string  `json:"note"`
}
