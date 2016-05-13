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
	Id            int32   `json:"id"`
	Rating        float32 `json:"rating"`
	Note          string  `json:"note"`
	UserServiceId int32   `json:"user_service_id"`
	ReservationId string  `json:"reservation_id"`
	UserPaymentId int32   `json:"user_payment_id"`
	UserCarId     int32   `json:"user_car_id"`
	Plate         string  `json:"plate"`
	Maker         string  `json:"maker"`
	Model         string  `json:"model"`
	Price         float32 `json:"price"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	Services      []int32 `json:"services"`
}

/**
{
	"user_id": 1,
	"page": 0
}
**/
type HistoryRequest struct {
	UserId int32 `json:"user_id"`
	Page   int32 `json:"page"`
}

/**
{
	"user_id": 1,
	"service_id": 1,
	"rating": 3.5,
	"note": "this is awesome"
}
**/
type RatingRequest struct {
	Id        int32   `json:"id"`
	UserId    int32   `json:"user_id"`
	ServiceId int32   `json:"service_id"`
	Rating    float32 `json:"rating"`
	Note      string  `json:"note"`
}
