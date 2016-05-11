package modules

type ChangeServiceStatus struct {
	UserId    int32 `json:"user_id"`
	ServiceId int32 `json:"service_id"`
}
