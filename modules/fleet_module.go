package modules

import (
	"gopkg.in/guregu/null.v3"
)

type FleetUserBasicInfo struct {
	SetUp     int32  `json:"fleet_setup"`
	FleetId   int32  `json:"fleet_id"`
	FleetName string `json:"fleet_name"`
	Token     string `json:"fleet_token"`
}

type FleetUserInfo struct {
	UserId int32 `json:"user_id"`

	UserContact
	UserWorkAddress

	FleetUserBasicInfo
}

type FleetUser struct {
	User
	FleetUserBasicInfo
}

type FleetOrderRequest struct {
	BaseRequest

	Note    string       `json:"note"`
	Opening int32        `json:"opening"`
	Day     string       `json:"day"`
	Hour    string       `json:"hour"`
	Orders  []FleetOrder `json:"orders"`
}

type FleetOrder struct {
	OrderId  int32          `json:"order_id"`
	CarCount int32          `json:"car_count"`
	Services []int32        `json:"service_ids"`
	Addons   []AddonRequest `json:"addons"`
}

type FleetService struct {
	Id               int32       `json:"id"`
	ReservationId    string      `json:"reservation_id"`
	Time             int32       `json:"time"`
	Price            float32     `json:"price"`
	Status           string      `json:"status"`
	ReserveStartTime string      `json:"reserve_start_time"`
	HowLong          int32       `json:"how_long"`
	Unit             string      `json:"unit"`
	ReserveTime      string      `json:"reserve_time"`
	Note             null.String `json:"note"`
	StartTime        null.String `json:"start_time"`
	EndTime          null.String `json:"end_time"`
}

type FleetReservationRequest struct {
	BaseRequest

	FleetServiceId int32 `json:"fleet_service_id"`
}

type FleetReservationDetail struct {
	CarCount int32                     `json:"car_count"`
	Services []FleetReservationService `json:"services"`
	Addons   []FleetReservationAddon   `json:"addons"`
}

type FleetReservationService struct {
	OrderId  int32 `json:"-"`
	CarCount int32 `json:"-"`

	Name string `json:"name"`
	Type string `json:"type"`
	Note string `json:"note"`
}

type FleetReservationAddon struct {
	OrderId  int32 `json:"-"`
	CarCount int32 `json:"-"`

	Name string `json:"name"`
	Note string `json:"note"`
}

type FleetHistory struct {
	Id             int32                     `json:"id"`
	FleetServiceId int32                     `json:"fleet_service_id"`
	Price          float32                   `json:"price"`
	Rating         float32                   `json:"rating"`
	Note           string                    `json:"note"`
	ReservationId  string                    `json:"reservation_id"`
	StartTime      string                    `json:"start_time"`
	EndTime        string                    `json:"end_time"`
}

type GetFleetUserRequest struct {
	BaseRequest

	Page int32 `json:"page"`
}

type FleetOpeningRequest struct {
	BaseRequest

	Orders []FleetOrder `json:"orders"`
}
