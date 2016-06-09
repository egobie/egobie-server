package modules

/**
{
	"user_id": 2,
	"user_token": "bc25",
	"fleet_name": "Paramus Car Dealer",
	"first_name": "Bo",
	"last_name": "Huang",
	"middle_name": "MJ",
	"email": "jkasdhjf@gmail.com",
	"phone": "2019120383",
	"street": "1 Hackensack Avenue",
	"city": "New York",
	"state": "NY",
	"zip": "10000"
}
**/
type NewFLeetUser struct {
	BaseRequest

	FleetName  string `json:"fleet_name"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        string `json:"zip"`
}
