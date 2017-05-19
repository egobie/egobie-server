package modules

type BaseRequest struct {
	UserId    int32  `json:"userId"`
	UserToken string `json:"userToken"`
}
