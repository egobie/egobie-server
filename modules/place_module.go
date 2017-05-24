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

type PlaceService struct {
	Id            int32     `json:"id"`
	ReservationId string    `json:"reservationId"`
	Location      string    `json:"location"`
	CarPlate      string    `json:"plate"`
	PickUpBy      int32     `json:"pickUpBy"`
	Day           string    `json:"day"`
	Price         float32   `json:"price"`
	Status        string    `json:"status"`
	ServiceList   []Service `json:"services"`
}
