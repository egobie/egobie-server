package modules

import (
	"gopkg.in/guregu/null.v3"
)

type User struct {
	Id                int32       `json:"id"`
	Type              string      `json:"type"`
	Username          string      `json:"username"`
	Password          string      `json:"password"`
	Coupon            string      `json:"coupon"`
	Email             null.String `json:"email"`
	PhoneNumber       null.String `json:"phone_number"`
	FirstName         null.String `json:"first_name"`
	LastName          null.String `json:"last_name"`
	MiddleName        null.String `json:"middle_name"`
	HomeAddressState  null.String `json:"home_address_state"`
	HomeAddressZip    null.String `json:"home_address_zip"`
	HomeAddressCity   null.String `json:"home_address_city"`
	HomeAddressStreet null.String `json:"home_address_street"`
	WorkAddressState  null.String `json:"work_address_state"`
	WorkAddressZip    null.String `json:"work_address_zip"`
	WorkAddressCity   null.String `json:"work_address_city"`
	WorkAddressStreet null.String `json:"work_address_street"`
}

/**
{
	"id": 1,
	"token": "1234"
}
**/
type UserInfo struct {
	Id    int32  `json:"id"`
	Token string `json:"token"`
}

/**
{
	"user_id": 1,
	"user_token": "abcd",
}
**/
type UserRequest struct {
	BaseRequest
}

/**
{
	"user_id": 1,
	"user_token": "abcd",
	"first_name": "Bo",
	"last_name": "Huang",
	"middle_name": "X",
	"email": "abc@test.com",
	"phone_number": "1231231234"
}
**/
type UpdateUser struct {
	BaseRequest

	FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
	MiddleName  null.String `json:"middle_name"`
	Email       string      `json:"email"`
	PhoneNumber string      `json:"phone_number"`
}

/**
{
	"user_id": 1,
	"user_token": "abcd",
	"password": "123456",
	"new_password": "654321"
}
**/
type UpdatePassword struct {
	BaseRequest

	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

/**
{
	"user_id": 1,
	"user_token": "abcd",
	"state": "NJ",
	"zip": "07601",
	"city": "Hackensack",
	"street": "715 Adams Street"
}
**/
type UpdateAddress struct {
	BaseRequest

	State  string `json:"state"`
	Zip    string `json:"zip"`
	City   string `json:"city"`
	Street string `json:"street"`
}

type Feedback struct {
	UserId   int32  `json:"user_id"`
	Title    string `json:"title"`
	Feedback string `json:"feedback"`
}
