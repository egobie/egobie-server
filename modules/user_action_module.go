package modules

type ActionRequest struct {
	BaseRequest

	Data string `json:"data"`
}
