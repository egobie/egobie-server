package modules

type Place struct {
	Id        int32   `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longtitude"`
}

type PlaceOpening struct {
	Id  int32  `json:"id"`
	Day string `json:"day"`
}

type PlaceOpeningRequest struct {
	BaseRequest

	Id        int32   `json:"id"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longtitude"`
}

type PlaceOpeningTodayRequest struct {
	BaseRequest

	Id        int32   `json:"id"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longtitude"`
}
