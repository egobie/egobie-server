package modules

type ChangeServiceStatus struct {
	BaseRequest

	CarId     int32 `json:"car_id"`
	ServiceId int32 `json:"service_id"`
	PaymentId int32 `json:"payment_id"`
}
