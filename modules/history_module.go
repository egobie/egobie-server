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
	UserServiceId int32   `json:"user_service_id"`
	UserPaymentId int32   `json:"user_payment_id"`
	UserCarId     int32   `json:"user_car_id"`
	Plate         string  `json:"plate"`
	Maker         string  `json:"maker"`
	Model         string  `json:"model"`
	Price         float32 `json:"price"`
	EndTime       string  `json:"end_time"`
}

type HistoryRequest struct {
	BaseRequest

	Page int `json:"page"`
}
