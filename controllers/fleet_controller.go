package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/cache"
	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
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
		u  modules.User
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

func PlaceFleetOrder(c *gin.Context) {
	request := modules.FleetOrderRequest{}
	var (
		tx       *sql.Tx
		result   sql.Result
		data     []byte
		err      error
		assignee int32
		time     int32
		gap      int32
		reserved string

		fleetServiceId int64
		serviceListId  int64
		addonListId    int64
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if tx, err = config.DB.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				fmt.Println("Error - Rollback - ", err1.Error())
			}
		} else {
			err = tx.Commit()
		}
	}()

	time = calculateServiceTime(request.Services) +
		calculateAddonTime(request.Addons)
	gap = calculateGap(time)

	if request.Opening != -1 {
		if reserved, err = calculateReservedTime(
			tx, request.Opening,
		); err != nil {
			return
		}

		if assignee, err = assignService(
			tx, request.Opening, gap, request.Types,
		); err != nil {
			return
		}
	} else {
		assignee = -1
		reserved = request.Day + " " + request.Hour
	}

	query := `
		insert into fleet_service (user_id, opening_id, reserved_start_timestamp,
			gap, assignee, estimated_time, status, types
		) values (?, ?, ?, ?, ?, ?, ?, ?)
	`

	if result, err = tx.Exec(
		query, request.UserId, request.Opening, reserved, gap,
		assignee, time, "RESERVED", request.Types,
	); err != nil {
		return
	} else if fleetServiceId, err = result.LastInsertId(); err != nil {
		return
	}

	queryService := `
		insert into fleet_service_list (fleet_service_id, car_count)
		values (?, ?)
	`
	queryServiceIds := `
		insert into fleet_service_list_id (service_id, fleet_service_list_id)
		values (?, ?)
	`

	for _, service := range request.Services {
		if result, err = tx.Exec(
			queryService, fleetServiceId, service.CarCount,
		); err != nil {
			return
		} else if serviceListId, err = result.LastInsertId(); err != nil {
			return
		}

		for _, id := range service.ServicesIds {
			if _, ok := cache.SERVICES_MAP[id]; ok {
				if _, err = tx.Exec(
					queryServiceIds, id, serviceListId,
				); err != nil {
					return
				}
			}
		}
	}

	queryAddon := `
		insert into fleet_service_addon_list (fleet_service_id, car_count)
	`
	queryAddonIds := `
		insert into fleet_service_addon_list_id (service_addon_id,
			fleet_service_addon_list_id, amount
		) values (?, ?, ?)
	`

	for _, addon := range request.Addons {
		if result, err = tx.Exec(
			queryAddon, fleetServiceId, addon.CarCount,
		); err != nil {
			return
		} else if addonListId, err = result.LastInsertId(); err != nil {
			return
		}

		for _, info := range addon.AddonInfos {
			if _, ok := cache.FLEET_ADDONS_MAP[info.Id]; ok {
				if _, err = tx.Exec(
					queryAddonIds, info.Id, addonListId, info.Amount,
				); err != nil {
					return
				}
			}
		}
	}
}

func calculateServiceTime(services []modules.FleetServiceRequest) (totalTime int32) {
	var tempTime int32 = 0

	for _, service := range services {
		for _, id := range service.ServicesIds {
			if s, ok := cache.SERVICES_MAP[id]; ok {
				tempTime += s.Time
			}
		}

		totalTime += tempTime * service.CarCount
		tempTime = 0
	}

	return
}

func calculateAddonTime(addons []modules.FleetAddonRequest) (totalTime int32) {
	var tempTime int32 = 0

	for _, addon := range addons {
		for _, info := range addon.AddonInfos {
			if t, ok := cache.FLEET_ADDONS_MAP[info.Id]; ok {
				tempTime += t.Time
			}
		}

		totalTime += tempTime * addon.CarCount
		tempTime = 0
	}

	return
}
