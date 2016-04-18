package modules

import (
	"gopkg.in/guregu/null.v3"
)

/**
{
	"id": 1,
	"name": "Oil Change",
	"description": "Oil Change",
	"price": 99.99,
	"time": 120
}
**/
type Service struct {
	Id          int32       `json:"id"`
	Name        string      `json:"name"`
	Description null.String `json:"description"`
	Price       float64     `json:"price"`
	Time        int32       `json:"time"`
}

/**
{
	"id": 1,
	"user_id: 1,
	"car_id": 1,
	"payment_id": 1,
	"plate": "Y96EUV",
	"time": 120,
	"price": 99.99,
	"note": "some note written by user",
	"status": "IN_PROGRESS",
	"revserve_time": "2016-04-15 11:20:36"
	"start_time": "2016-04-15 11:20:36"
	"end_time": "2016-04-15 11:20:36"
	"services": [
		{
			"id": 1,
			"name": "Car Wash",
			"description": "Wash Car",
			"price": 45.99,
			"time": 30
		},
		{
			"id": 2,
			"name": "Oil Change",
			"description": "Change Oil",
			"price": 25.98,
			"time": 15
		}
	]
}
**/
type UserService struct {
	Id          int32       `json:"id"`
	UserId      int32       `json:"user_id"`
	CarId       int32       `json:"car_id"`
	CarPlate    string      `json:"plate"`
	PaymentId   int32       `json:"payment_id"`
	Time        int32       `json:"time"`
	Price       float32     `json:"price"`
	Note        null.String `json:"note"`
	Status      string      `json:"status"`
	ReserveTime null.String `json:"revserve_time"`
	StartTime   null.String `json:"start_time"`
	EndTime     null.String `json:"end_time"`
	ServiceList []Service   `json:"services"`
}

type OrderRequest struct {
	UserId    int32       `json:"user_id"`
	CarId     int32       `json:"car_id"`
	PaymentId int32       `json:"payment_id"`
	Services  []int32     `json:"services"`
	Note      string      `json:"note"`
	Time      int32       `json:"time"`
	Price     float32     `json:"price"`
	Type      string      `json:"type"`
	StartTime null.String `json:"start_time"`
	EndTime   null.String `json:"end_time"`
}
