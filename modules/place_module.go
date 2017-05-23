package modules

type Place struct {
	Id        int32   `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longtitude"`
}

type PlaceOpeningRequest struct {
	BaseRequest

	Id        int32   `json:"id"`
	Date      string  `json:"date"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longtitude"`
}
