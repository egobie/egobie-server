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
