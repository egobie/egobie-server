package modules

type Task struct {
	Id         int32         `json:"id"`
	Start      string        `json:"start"`
	FirstName  string        `json:"first"`
	MiddleName string        `json:"middle"`
	LastName   string        `json:"last"`
	Phone      string        `json:"phone"`
	State      string        `json:"state"`
	Zip        string        `json:"zip"`
	City       string        `json:"city"`
	Street     string        `json:"street"`
	Plate      string        `json:"plate"`
	CarState   string        `json:"car_state"`
	Color      string        `json:"color"`
	Maker      string        `json:"maker"`
	Model      string        `json:"model"`
	Services   []TaskService `json:"services"`
	Addons     []TaskAddon   `json:"addons"`
}

type TaskService struct {
	Id            int32  `json:"id"`
	Name          string `json:"name"`
	Note          string `json:"note"`
	Type          string `json:"type"`
	UserServiceId int32  `json:"user_service_id"`
}

type TaskAddon struct {
	Id            int32  `json:"id"`
	Name          string `json:"name"`
	Note          string `json:"note"`
	Amount        int32  `json:"amount"`
	Unit          string `json:"unit"`
	UserServiceId int32  `json:"user_service_id"`
}

type TaskRequest struct {
	BaseRequest
}

type ChangeServiceStatus struct {
	BaseRequest

	CarId     int32 `json:"car_id"`
	ServiceId int32 `json:"service_id"`
	PaymentId int32 `json:"payment_id"`
}
