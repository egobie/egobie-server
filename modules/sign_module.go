package modules

/**
{
	"username": "test_user",
	"password": "a123456",
	"email": "test@test.com",
	"phone_number": "1234567890"
}
**/
type SignUp struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
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
