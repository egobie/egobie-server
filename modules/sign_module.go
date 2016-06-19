package modules

/**
{
	"username": "test_user",
	"password": "a123456",
	"email": "test@test.com",
	"phone_number": "1234567890",
	"coupon": "A1B2C"
}
**/
type SignUp struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Coupon      string `json:"coupon"`
}

/**
{
	"email": "test@test.com",
	"token": "747AD",
	"username": "fleet-1",
	"password": "12345678"
}
**/
type SignUpFleet struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

/**
{
	"username": "test_user",
	"password": "a123456"
}
**/
type SignIn struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

/**
{
	"value": "email_address"
}
**/
type Check struct {
	Value string `json:"value"`
}

type ResetPasswordStep1 struct {
	Username string `json:"username"`
}

type ResetPasswordStep2 struct {
	UserId int32  `json:"user_id"`
	Token  string `json:"token"`
}

type ResetPasswordStep3 struct {
	UserId   int32  `json:"user_id"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

type ResetPasswordResend struct {
	UserId   int32  `json:"user_id"`
	Username string `json:"username"`
}
