package modules

import (
	"gopkg.in/guregu/null.v3"
)

/**
{
	"id": 1,
	"ratting": 3.5,
	"user_service_id": "11",
	"note": "This is awesome"
}
**/
type History struct {
	Id            int32     `json:"id"`
	Ratting       float32   `json:"ratting"`
	UserServiceId int32     `json:"user_service_id"`
	Price         float32   `json:"price"`
	StartTime     null.Time `json:"start_time"`
	EndTime       null.Time `json:"end_time"`
	Note          string    `json:"note"`
}

type HistoryRequest struct {
	BaseRequest

	Page int `json:"page"`
}
