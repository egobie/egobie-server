package modules

type FleetUserBasicInfo struct {
	SetUp     int32  `json:"setup"`
	FleetId   int32  `json:"fleet_id"`
	FleetName string `json:"fleet_name"`
	Token     string `json:"token"`
}

type FleetUserInfo struct {
	UserId int32 `json:"user_id"`

	UserContact
	UserWorkAddress

	FleetUserBasicInfo
}

type FleetUser struct {
	User
	FleetUserBasicInfo
}

type FleetOrderRequest struct {
	BaseRequest

	Note     string                `json:"note"`
	Types    string                `json:"types"`
	Opening  int32                 `json:"opening"`
	Day      string                `json:"day"`
	Hour     string                `json:"Hour"`
	Services []FleetServiceRequest `json:"services"`
	Addons   []FleetAddonRequest   `json:"addons"`
}

type FleetServiceRequest struct {
	CarCount    int32   `json:"car_count"`
	ServicesIds []int32 `json:"service_ids"`
}

type FleetAddonRequest struct {
	CarCount   int32          `json:"car_count"`
	AddonInfos []AddonRequest `json:"addons"`
}
