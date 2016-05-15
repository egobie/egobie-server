package modules

import (
	"gopkg.in/guregu/null.v3"
)

/**
{
	"id": 1,
	"name": "Premium (Full)",
	"type": "CAR_WASH",
	"items": [
		"Full Exterior Hand Wash",
		"Tire Shine & Rim Cleaning",
		"Undercarriage Rinse",
		"Total Interior Wipe-down",
		"Interior Vacuum",
		"Trunk Vacuum"
	],
	"description": "Wash Car",
	"price": 25,
	"time": 30,
	"addons": [{
		"id": 1,
		"service_id": 1,
		"name": "Extra Conventional Oil",
		"note": "",
		"price": 4,
		"time": 0,
		"max": 30,
		"unit": "quart"
	}]
}
**/
type Service struct {
	Id          int32       `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Items       []string    `json:"items"`
	Description null.String `json:"description"`
	Note        string      `json:"note"`
	Price       float64     `json:"price"`
	Time        int32       `json:"time"`
	AddOns      []AddOn     `json:"addons"`
}

/**
{
	"id": 1,
	"service_id": 1,
	"name": "Extra Conventional Oil",
	"note": "",
	"price": 4,
	"time": 0,
	"max": 30,
	"unit": "quart"
}
**/
type AddOn struct {
	Id        int32   `json:"id"`
	ServiceId int32   `json:"service_id"`
	Name      string  `json:"name"`
	Note      string  `json:"note"`
	Price     float32 `json:"price"`
	Time      int32   `json:"time"`
	Max       int32   `json:"max"`
	Unit      string  `json:"unit"`
}

/**
{
	"id": 1,
	"user_id": 1,
	"car_id": 1,
	"plate": "Y96EUV",
	"payment_id": 1,
	"time": 100,
	"price": 99.89,
	"note": null,
	"status": "RESERVED",
	"revserve_time": "2016-04-21 11:01:37",
	"start_time": null,
	"end_time": null,
	"services": [
		{
			"id": 1,
			"name": "Premium (Full)",
			"type": "CAR_WASH",
			"items": [
				"Full Exterior Hand Wash",
				"Tire Shine & Rim Cleaning",
				"Undercarriage Rinse",
				"Total Interior Wipe-down",
				"Interior Vacuum",
				"Trunk Vacuum"
			],
			"description": "Wash Car",
			"price": 25,
			"time": 30,
			"addons": false
		},
		{
			"id": 7,
			"name": "Exterior",
			"type": "DETAILING",
			"items": [
				"Full Exterior Hand Wash",
				"Tire Shine & Rim Cleaning",
				"Undercarriage Rinse",
				"Paint Protection",
				"Windshield Protectant",
				"Hand wax",
				"Engine cleaning",
				"Headlight restoration",
				"Compressed air detailing in tight spaces",
				"Multi-layer wax + polish",
				"Multi-layer paint protectant"
			],
			"description": "Detailing",
			"price": 175,
			"time": 90,
			"addons": false
		}
	]
}
**/
type UserService struct {
	Id               int32       `json:"id"`
	ReservationId    string      `json:"reservation_id"`
	UserId           int32       `json:"user_id"`
	CarId            int32       `json:"car_id"`
	CarPlate         string      `json:"plate"`
	PaymentId        int32       `json:"payment_id"`
	Time             int32       `json:"time"`
	Price            float32     `json:"price"`
	Note             null.String `json:"note"`
	Status           string      `json:"status"`
	ReserveStartTime string      `json:"reserve_start_time"`
	HowLong          int32       `json:"how_long"`
	Unit             string      `json:"unit"`
	ReserveTime      null.String `json:"reserve_time"`
	StartTime        null.String `json:"start_time"`
	EndTime          null.String `json:"end_time"`
	ServiceList      []Service   `json:"services"`
}

type Period struct {
	Id    int32   `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type Opening struct {
	Day   string   `json:"day"`
	Range []Period `json:"range"`
}

type ServiceInfo struct {
	Type   string  `json:"type"`
	Count  int32   `json:"count"`
	Price  float32 `json:"price"`
	Time   int32   `json:"time"`
	AddOns bool    `json:"addons"`
}

/**
{
	"user_id": 1,
	"car_id": 1,
	"payment_id": 1,
	"services": [1,2,3],
	"note": "This is awesome!",
	"Opening": 1
}
**/
type OrderRequest struct {
	BaseRequest

	CarId     int32   `json:"car_id"`
	PaymentId int32   `json:"payment_id"`
	Services  []int32 `json:"services"`
	Note      string  `json:"note"`
	Opening   int32   `json:"opening"`
}

type OpeningRequest struct {
	BaseRequest

	Services []int32 `json:"services"`
}

/**
{
	"id": 1,
	"user_id": 1
}
**/
type CancelRequest struct {
	BaseRequest

	Id int32 `json:"id"`
}

type ServiceDemandRequest struct {
	BaseRequest

	Services []int32 `json:"services"`
}
