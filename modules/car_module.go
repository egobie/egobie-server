package modules

import (
	"gopkg.in/guregu/null.v3"
)

/**
{
	"id": 1,
	"userId": 1,
	"report_id": 1,
	"plate": "Y96EUV",
	"state": "NJ",
	"year": 2012,
	"color": "GRAY",
	"make": "Honda",
	"model": "Accord",
	"makeId": 13,
	"modelId": 251
}
**/
type Car struct {
	Id       int32    `json:"id"`
	UserId   int32    `json:"userId"`
	ReportId null.Int `json:"report_id"`
	Plate    string   `json:"plate"`
	State    string   `json:"state"`
	Year     int32    `json:"year"`
	Color    string   `json:"color"`
	Make     string   `json:"make"`
	Model    string   `json:"model"`
	MakeId   int32    `json:"makeId"`
	ModelId  int32    `json:"modelId"`
	Reserved int32    `json:"reserved"`
}

/**
{
	"id": 1,
	"title": "Honda"
}
**/
type CarMake struct {
	Id    int32  `json:"id"`
	Title string `json:"title"`
}

/**
{
	"id": 1,
	"makeId": 2,
	"title": "Accord"
}
**/
type CarModel struct {
	Id     int32  `json:"id"`
	MakeId int32  `json:"makeId"`
	Title  string `json:"title"`
}

/**
{
	"id": 1
}
**/
type CarRequst struct {
	BaseRequest

	Id int32 `json:"id"`
}

/**
{
	"userId": 1,
	"plate": "Y96EUV",
	"state": "NJ",
	"year": 2012,
	"color": "GRAY",
	"make": 13,
	"model": 256
}
**/
type CarNew struct {
	BaseRequest

	Plate string `json:"plate"`
	State string `json:"state"`
	Year  int32  `json:"year"`
	Color string `json:"color"`
	Make  int32  `json:"make"`
	Model int32  `json:"model"`
}

/**
{
	"id": 1,
	"userId": 1,
	"plate": "Y96EUV",
	"state": "NJ",
	"year": 2012,
	"color": "GRAY",
	"make": 13,
	"model": 256
}
**/
type UpdateCar struct {
	BaseRequest

	Id    int32  `json:"id"`
	Plate string `json:"plate"`
	State string `json:"state"`
	Year  int32  `json:"year"`
	Color string `json:"color"`
	Make  int32  `json:"make"`
	Model int32  `json:"model"`
}
