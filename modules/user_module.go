package modules

import (
	"gopkg.in/guregu/null.v3"
)

type UserContact struct {
	FirstName   null.String `json:"firstName"`
	LastName    null.String `json:"lastName"`
	MiddleName  null.String `json:"middleName"`
	Email       null.String `json:"email"`
	PhoneNumber null.String `json:"phoneNumber"`
}

type UserHomeAddress struct {
	HomeAddressState  null.String `json:"homeAddressState"`
	HomeAddressZip    null.String `json:"homeAddressZip"`
	HomeAddressCity   null.String `json:"homeAddressCity"`
	HomeAddressStreet null.String `json:"homeAddressStreet"`
}

type UserWorkAddress struct {
	WorkAddressState  null.String `json:"workAddressState"`
	WorkAddressZip    null.String `json:"workAddressZip"`
	WorkAddressCity   null.String `json:"workAddressCity"`
	WorkAddressStreet null.String `json:"workAddressStreet"`
}

type User struct {
	Id             int32  `json:"id"`
	Type           string `json:"type"`
	Password       string `json:"password"`
	Coupon         string `json:"coupon"`
	Discount       int32  `json:"discount"`
	FirstTime      int32  `json:"firstTime"`
	CouponDiscount int32  `json:"couponDiscount"`

	UserContact
	UserHomeAddress
	UserWorkAddress
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
	"userId": 1,
	"userToken": "abcd",
}
**/
type UserRequest struct {
	BaseRequest
}

/**
{
	"userId": 1,
	"userToken": "abcd",
	"firstName": "Bo",
	"lastName": "Huang",
	"email": "abc@test.com",
	"phoneNumber": "1231231234"
}
**/
type UpdateUser struct {
	BaseRequest

	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

/**
{
	"userId": 1,
	"userToken": "abcd",
	"password": "123456",
	"newPassword": "654321"
}
**/
type UpdatePassword struct {
	BaseRequest

	Password    string `json:"password"`
	NewPassword string `json:"newPassword"`
}

/**
{
	"userId": 1,
	"userToken": "abcd",
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

type ApplyCouponRequest struct {
	BaseRequest

	Coupon string `json:"code"`
}

type Coupon struct {
	Id       int32
	Discount float32
	Percent  int32
}

type Feedback struct {
	BaseRequest

	Title    string `json:"title"`
	Feedback string `json:"feedback"`
}

const USER_EGOBIE_TOKEN int32 = 4
const USER_SALE_TOKEN int32 = 4
const USER_RESIDENTIAL_TOKEN int32 = 6
const USER_FLEET_TOKEN int32 = 8
const USER_BUSINESS_TOKEN int32 = 10

const USER_BUSINESS string = "BUSINESS"
const USER_RESIDENTIAL string = "RESIDENTIAL"
const USER_EGOBIE string = "EGOBIE"
const USER_FLEET string = "FLEET"
const USER_SALE string = "SALE"

type CheckUserFunc func(string) bool

func IsResidential(userType string) bool {
	return userType == USER_RESIDENTIAL
}

func IsBusiness(userType string) bool {
	return userType == USER_BUSINESS
}

func IsEgobie(userType string) bool {
	return userType == USER_EGOBIE
}

func IsFleet(userType string) bool {
	return userType == USER_FLEET
}

func IsSale(userType string) bool {
	return userType == USER_SALE
}
