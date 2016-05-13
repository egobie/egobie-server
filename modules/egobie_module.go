package modules

type ChangeServiceStatus struct {
	UserId    int32 `json:"user_id"`
	CarId     int32 `json:"car_id"`
	ServiceId int32 `json:"service_id"`
	PaymentId int32 `json:"payment_id"`
}
