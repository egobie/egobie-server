package modules

type BaseRequest struct {
	UserId    int32  `json:"user_id"`
	UserToken string `json:"user_token"`
}
