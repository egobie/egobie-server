package modules

type Task struct {
	Id         int32           `json:"id"`
	Status     string          `json:"status"`
	Start      string          `json:"start"`
	FirstName  string          `json:"first"`
	MiddleName string          `json:"middle"`
	LastName   string          `json:"last"`
	Phone      string          `json:"phone"`
	State      string          `json:"state"`
	Zip        string          `json:"zip"`
	City       string          `json:"city"`
	Street     string          `json:"street"`
	Plate      string          `json:"plate"`
	CarState   string          `json:"car_state"`
	Color      string          `json:"color"`
	Maker      string          `json:"maker"`
	Model      string          `json:"model"`
	Services   []SimpleService `json:"services"`
	Addons     []SimpleAddon   `json:"addons"`
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

type TaskInfo struct {
	UserId        int32
	UserCarId     int32
	UserPaymentId int32
}
