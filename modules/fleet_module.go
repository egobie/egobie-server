package modules

/**
{
	"user_id": 1,
	"user_token": "bc25",
	"fleet_name": "Paramus Car Dealer",
	"first_name": "Bo",
	"last_name": "Huang",
	"middle_name": "MJ",
	"email": "jkasdhjf@gmail.com",
	"phone": "2019120383"
}
**/
type NewFLeetUserRequest struct {
	BaseRequest

	FleetName  string `json:"fleet_name"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}

type FleetUser struct {
	FleetId    int32  `json:"fleet_id"`
	UserId     int32  `json:"user_id"`
	Token      string `json:"token"`
	FleetName  string `json:"fleet_name"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}
