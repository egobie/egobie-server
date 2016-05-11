package modules

import (
	"gopkg.in/guregu/null.v3"
)

/**
{
	"id": 1,
	"user_id": 1,
	"report_id": 1,
	"plate": "Y96EUV",
	"state": "NJ",
	"year": 2012,
	"color": "GRAY",
	"maker": "Honda",
	"model": "Accord",
	"maker_id": 13,
	"model_id": 251
}
**/
type Car struct {
	Id       int32    `json:"id"`
	UserId   int32    `json:"user_id"`
	ReportId null.Int `json:"report_id"`
	Plate    string   `json:"plate"`
	State    string   `json:"state"`
	Year     int32    `json:"year"`
	Color    string   `json:"color"`
	Maker    string   `json:"maker"`
	Model    string   `json:"model"`
	MakerId  int32    `json:"maker_id"`
	ModelId  int32    `json:"model_id"`
	Reserved int32    `json:"reserved"`
}

/**
{
	"id": 1,
	"title": "Honda"
}
**/
type CarMaker struct {
	Id    int32  `json:"id"`
	Title string `json:"title"`
}

/**
{
	"id": 1,
	"maker_id": 2,
	"title": "Accord"
}
**/
type CarModel struct {
	Id      int32  `json:"id"`
	MakerId int32  `json:"maker_id"`
	Title   string `json:"title"`
}

/**
{
	"id": 1,
	"user_id": 1
}
**/
type CarRequst struct {
	Id     int32 `json:"id"`
	UserId int32 `json:"user_id"`
}

/**
{
	"user_id": 1
}
**/
type CarRequstForUser struct {
	UserId int32 `json:"user_id"`
}

/**
{
	"user_id": 1,
	"plate": "Y96EUV",
	"state": "NJ",
	"year": 2012,
	"color": "GRAY",
	"maker": 13,
	"model": 256
}
**/
type CarNew struct {
	UserId int32  `json:"user_id"`
	Plate  string `json:"plate"`
	State  string `json:"state"`
	Year   int32  `json:"year"`
	Color  string `json:"color"`
	Maker  int32  `json:"maker"`
	Model  int32  `json:"model"`
}

/**
{
	"id": 1,
	"user_id": 1,
	"plate": "Y96EUV",
	"state": "NJ",
	"year": 2012,
	"color": "GRAY",
	"maker": 13,
	"model": 256
}
**/
type UpdateCar struct {
	Id     int32  `json:"id"`
	UserId int32  `json:"user_id"`
	Plate  string `json:"plate"`
	State  string `json:"state"`
	Year   int32  `json:"year"`
	Color  string `json:"color"`
	Maker  int32  `json:"maker"`
	Model  int32  `json:"model"`
}
