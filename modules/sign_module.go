package modules

/**
{
	"password": "a123456",
	"email": "test@test.com",
	"phoneNumber": "1234567890",
	"firstName": "First",
	"lastName": "Last",
	"coupon": "A1B2C"
}
**/
type SignUp struct {
	Password    string `json:"password"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Coupon      string `json:"coupon"`
}

/**
{
	"email": "test@test.com",
	"token": "747AD",
	"password": "12345678"
}
**/
type SignUpFleet struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

/**
{
	"email": "test@example.com",
	"password": "a123456"
}
**/
type SignIn struct {
	Email    string `json:"email"`
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
	Email string `json:"email"`
}

type ResetPasswordStep2 struct {
	UserId int32  `json:"userId"`
	Token  string `json:"token"`
}

type ResetPasswordStep3 struct {
	UserId   int32  `json:"userId"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

type ResetPasswordResend struct {
	UserId int32 `json:"userId"`
}
