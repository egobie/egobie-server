package controllers

import (
	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
)

func getFleetUserBasicInfo(userId int32) (info modules.FleetUserBasicInfo, err error) {
	query := `
		select f.id, f.setup, f.name, f.token
		from fleet f
		inner join user u on u.id = f.user_id
		where u.id = ?
	`

	err = config.DB.QueryRow(query, userId).Scan(
		&info.FleetId, &info.SetUp, &info.FleetName, &info.Token,
	)

	return
}

func getFleetUserInfoByUserId(userId int32) (info modules.FleetUserInfo, err error) {
	query := `
		select f.id, f.name, f.setup, f.token, u.id, u.first_name, u.last_name,
				u.middle_name, u.email, u.phone_number, u.work_address_street,
				u.work_address_city, u.work_address_state, u.work_address_zip
		from fleet f
		inner join user u on u.id = f.user_id
		where u.id = ?
	`

	err = config.DB.QueryRow(query, userId).Scan(
		&info.FleetId, &info.FleetName, &info.SetUp, &info.Token, &info.UserId,
		&info.FirstName, &info.LastName, &info.MiddleName, &info.Email,
		&info.PhoneNumber, &info.WorkAddressStreet, &info.WorkAddressCity,
		&info.WorkAddressState, &info.WorkAddressZip,
	)

	return
}

func getFleetUser(condition string, args ...interface{}) (user modules.FleetUser, err error) {
	var (
		u modules.User
		ui modules.FleetUserBasicInfo
	)

	if u, err = getUser(condition, args...); err != nil {
		return
	} else if ui, err = getFleetUserBasicInfo(u.Id); err == nil {
		user.User = u
		user.FleetUserBasicInfo = ui
	}

	return user, nil
}

func getFleetUserByUserId(userId int32) (user modules.FleetUser, err error) {
	return getFleetUser("id = ?", userId)
}

func getFleetUserByUsername(username string) (user modules.FleetUser, err error) {
	return getFleetUser("username = ?", username)
}
