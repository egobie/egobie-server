package modules

import (
	"gopkg.in/guregu/null.v3"
)

type Task struct {
	PlaceTasks []PlaceTask `json:"placeTasks"`
}

type UserTask struct {
	Id         int32           `json:"id"`
	Status     string          `json:"status"`
	Start      string          `json:"start"`
	FirstName  string          `json:"first"`
	MiddleName string          `json:"middle"`
	LastName   string          `json:"last"`
	Phone      string          `json:"phone"`
	State      string          `json:"state"`
	Zip        string          `json:"zip"`
	City       string          `json:"city"`
	Street     string          `json:"street"`
	Plate      string          `json:"plate"`
	CarState   string          `json:"car_state"`
	Color      string          `json:"color"`
	Year       string          `json:"year"`
	Make       string          `json:"make"`
	Model      string          `json:"model"`
	Services   []SimpleService `json:"services"`
	Addons     []SimpleAddon   `json:"addons"`
}

type PlaceTask struct {
	Id        int32                `json:"id"`
	Status    string               `json:"status"`
	Price     float32              `json:"price"`
	Day       string               `json:"day"`
	PickUpBy  int32                `json:"pickUpBy"`
	FirstName string               `json:"firstName"`
	LastName  string               `json:"lastName"`
	Phone     string               `json:"phone"`
	Address   string               `json:"address"`
	Plate     string               `json:"plate"`
	State     string               `json:"state"`
	Color     string               `json:"color"`
	Year      string               `json:"year"`
	Make      string               `json:"make"`
	Model     string               `json:"model"`
	Services  []SimplePlaceService `json:"services"`
}

type FleetTask struct {
	Id        int32       `json:"id"`
	Status    string      `json:"status"`
	Start     string      `json:"start"`
	FleetName string      `json:"fleet_name"`
	FirstName string      `json:"first"`
	LastName  string      `json:"last"`
	Phone     string      `json:"phone"`
	State     string      `json:"state"`
	Zip       string      `json:"zip"`
	City      string      `json:"city"`
	Street    string      `json:"street"`
	Note      null.String `json:"note"`
}

type TaskRequest struct {
	BaseRequest

	PlaceIds []int32 `json:"placeIds"`
	Day      string  `json:"day"`
}

/**
{
	"userId": 1,
	"user_token": "abcd",
	"car_id": 1,
	"service_id": 1,
	"payment_id": 1
}
**/
type ChangeServiceStatus struct {
	BaseRequest

	ServiceId int32 `json:"serviceId"`
}

type TaskInfo struct {
	Status        string
	UserId        int32
	UserCarId     int32
	UserPaymentId int32
}
