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

func getFleetUserInfoBySaleUserId(saleUserId int32, page int32) (
	infoList []modules.FleetUserInfo, err error,
) {
	size := 15
	query := `
		select f.id, f.name, f.setup, f.token, u.id, u.first_name, u.last_name,
				u.middle_name, u.email, u.phone_number, u.work_address_street,
				u.work_address_city, u.work_address_state, u.work_address_zip
		from fleet f
		inner join user u on u.id = f.user_id
		where f.sale_user_id = ?
		limit ?, ?
	`
	var (
		rows *sql.Rows
	)

	if rows, err = config.DB.Query(
		query, saleUserId, page*int32(size), size,
	); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		info := modules.FleetUserInfo{}

		if err = rows.Scan(
			&info.FleetId, &info.FleetName, &info.SetUp, &info.Token,
			&info.UserId, &info.FirstName, &info.LastName, &info.MiddleName,
			&info.Email, &info.PhoneNumber, &info.WorkAddressStreet,
			&info.WorkAddressCity, &info.WorkAddressState, &info.WorkAddressZip,
		); err != nil {
			return
		}

		info.Token = ""
		info.UserId = 0

		infoList = append(infoList, info)
	}

	return infoList, nil
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

func GetFleetService(c *gin.Context) {
	c.JSON(http.StatusOK, cache.SERVICES_ARRAY)
}

func GetFleetAddon(c *gin.Context) {
	c.JSON(http.StatusOK, cache.FLEET_ADDONS_ARRAY)
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

func GetFleetReservation(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		err  error
		body []byte
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if fleetServices, err := getFleetService(
		request.UserId,
		"status = 'WAITING' or status = 'RESERVED' or status = 'IN_PROGRESS'",
	); err == nil {
		c.JSON(http.StatusOK, fleetServices)
	}
}

func GetFleetReservationDetail(c *gin.Context) {
	request := modules.FleetReservationRequest{}
	var (
		err    error
		data   []byte
		detail modules.FleetReservationDetail
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		} else {
			c.JSON(http.StatusOK, detail)
		}
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if detail.Services, err = getFleetReservationService(
		request.FleetServiceId,
	); err != nil {
		return
	}

	if detail.Addons, err = getFleetReservationAddon(
		request.FleetServiceId,
	); err != nil {
		return
	}
}

func GetFleetHistory(c *gin.Context) {
	size := 6
	query := `
		select fh.id, fs.id, fs.reservation_id, fs.estimated_price,
				fs.start_timestamp, fs.end_timestamp, fh.rating, fh.note
		from fleet_service fs
		inner join fleet_history fh on fh.fleet_service_id = fs.id and fs.status = 'DONE'
		where fs.user_id = ?
		order by fh.create_timestamp DESC
		limit ?, ?
	`
	request := modules.HistoryRequest{}
	var (
		data    []byte
		err     error
		rows    *sql.Rows
		historyList []modules.FleetHistory
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, historyList)
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if rows, err = config.DB.Query(
		query, request.UserId, request.Page*int32(size), size,
	); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		history := modules.FleetHistory{}

		if err = rows.Scan(
			&history.Id, &history.FleetServiceId, &history.ReservationId,
			&history.Price, &history.StartTime, &history.EndTime,
			&history.Rating, &history.Note,
		); err != nil {
			return
		}

		if history.Services, err = getFleetReservationService(
			history.FleetServiceId,
		); err != nil {
			return
		}

		if history.Addons, err = getFleetReservationAddon(
			history.FleetServiceId,
		); err != nil {
			return
		}

		historyList = append(historyList, history)
	}
}

func RatingFleet(c *gin.Context) {
	query := `
		update fleet_history set rating = ?, note = ?
		where id = ? and fleet_service_id = ?
	`

	request := modules.RatingRequest{}
	var (
		data []byte
		err  error
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, "OK")
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if _, err = config.DB.Exec(
		query, request.Rating, request.Note, request.Id, request.ServiceId,
	); err != nil {
		return
	}
}

func getFleetReservationService(fleetServiceId int32) (
	services []modules.FleetReservationService, err error,
) {
	var (
		rows          *sql.Rows
		prevId        int32
		tempId        int32
		tempServiceId int32
		tempCarCount  int32
		ok            bool
		service       modules.Service
	)

	prevId = -1
	query := `
		select f.id, f.car_count, fi.service_id
		from fleet_service_list f
		inner join fleet_service_list_id fi on fi.fleet_service_list_id = f.id
		where f.fleet_service_id = ?
		order by f.id
	`

	if rows, err = config.DB.Query(query, fleetServiceId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&tempId, &tempCarCount, &tempServiceId,
		); err != nil {
			return
		}

		if service, ok = cache.SERVICES_MAP[tempServiceId]; !ok {
			fmt.Println("Invalid ServiceId - fleet reservation - ", tempServiceId)
			continue
		}

		if tempId != prevId {
			s := modules.FleetReservationService{}
			s.CarCount = tempCarCount
			s.Info = append(
				s.Info, modules.FleetReservationServiceInfo{
					Name: service.Name,
					Type: service.Type,
					Note: service.Note,
				},
			)
			services = append(services, s)
		} else {
			s := services[len(services)-1]
			s.Info = append(
				s.Info, modules.FleetReservationServiceInfo{
					Name: service.Name,
					Type: service.Type,
					Note: service.Note,
				},
			)
		}

		prevId = tempId
	}

	return services, nil
}

func getFleetReservationAddon(fleetServiceId int32) (
	addons []modules.FleetReservationAddon, err error,
) {
	var (
		prevId       int32
		tempId       int32
		tempCarCount int32
		tempAddonId  int32
		addon        modules.AddOn
		ok           bool
		rows         *sql.Rows
	)

	prevId = -1
	query := `
		select f.id, f.car_count, fi.service_addon_id
		from fleet_service_addon_list f
		inner join fleet_service_addon_list_id fi on fi.fleet_service_addon_list_id = f.id
		where f.fleet_service_id = ?
		order by f.id
	`

	if rows, err = config.DB.Query(query, fleetServiceId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&tempId, &tempCarCount, &tempAddonId,
		); err != nil {
			return
		}

		if addon, ok = cache.FLEET_ADDONS_MAP[tempAddonId]; !ok {
			fmt.Println("Invalid AddonId - fleet reservation - ", tempAddonId)
			continue
		}

		if tempId != prevId {
			a := modules.FleetReservationAddon{}
			a.CarCount = tempCarCount
			a.Info = append(
				a.Info, modules.FleetReservationAddonInfo{
					Name: addon.Name,
					Note: addon.Note,
				},
			)
			addons = append(addons, a)
		} else {
			a := addons[len(addons)-1]
			a.Info = append(
				a.Info, modules.FleetReservationAddonInfo{
					Name: addon.Name,
					Note: addon.Note,
				},
			)
		}

		prevId = tempId
	}

	return addons, nil
}

func getFleetService(userId int32, condition string) (fleetServices []modules.FleetService, err error) {
	query := `
		select fs.id, fs.reservation_id, fs.estimated_time,
				fs.estimated_price, fs.reserved_start_timestamp,
				TIMESTAMPDIFF(MINUTE, CURRENT_TIMESTAMP(), fs.reserved_start_timestamp) as mins,
				fs.start_timestamp, fs.end_timestamp,
				fs.note, fs.status, fs.create_timestamp
		from fleet_service fs
		where fs.user_id = ? and (
	` + condition + ") order by fs.id"

	var (
		rows *sql.Rows
		mins int32
	)

	if rows, err = config.DB.Query(query, userId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		fleetService := modules.FleetService{}

		if err = rows.Scan(
			&fleetService.Id, &fleetService.ReservationId, &fleetService.Time,
			&fleetService.Price, &fleetService.ReserveStartTime, &mins,
			&fleetService.StartTime, &fleetService.EndTime, &fleetService.Note,
			&fleetService.Status, &fleetService.ReserveTime,
		); err != nil {
			return
		}

		fleetService.HowLong, fleetService.Unit = calculateHowLong(mins)
		fleetServices = append(fleetServices, fleetService)
	}

	return fleetServices, nil
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
