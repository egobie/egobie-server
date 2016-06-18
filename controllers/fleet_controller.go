package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
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

func getFleetUsersBySaleUserId(saleUserId int32, page int32) (
	all modules.AllFleetUser, err error,
) {
	size := 15
	query := `
		select f.id, f.name, f.setup, f.token, u.id, u.first_name, u.last_name,
				u.middle_name, u.email, u.phone_number, u.work_address_street,
				u.work_address_city, u.work_address_state, u.work_address_zip
		from fleet f
		inner join user u on u.id = f.user_id
		where f.sale_user_id = ? order by u.create_timestamp DESC
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

		all.Users = append(all.Users, info)
	}

	query = `
		select count(*) from fleet f
		inner join user u on u.id = f.user_id
		where f.sale_user_id = ?
	`

	if err = config.DB.QueryRow(query, saleUserId).Scan(
		&all.Total,
	); err != nil {
		return
	}

	return all, nil
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
		tx             *sql.Tx
		result         sql.Result
		data           []byte
		err            error
		assignee       int32
		time           int32
		gap            int32
		reserved       string
		types          string
		fleetServiceId int64
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

	time, types = calculateFleetOrderTimeAndTypes(
		request.Orders,
	)
	gap = calculateGap(time)

	if request.Opening != -1 {
		if reserved, err = calculateReservedTime(
			tx, request.Opening,
		); err != nil {
			return
		}

		if err = holdOpening(
			tx, request.Opening, request.Opening+gap, types,
		); err != nil {
			return
		}

		if assignee, err = assignService(
			tx, request.Opening, gap, types,
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
		query, request.UserId, request.Opening, reserved,
		gap, assignee, time, "WAITING", types,
	); err != nil {
		return
	} else if fleetServiceId, err = result.LastInsertId(); err != nil {
		return
	}

	// Insert fleet orders
	for _, order := range request.Orders {
		if err = insertFleetServiceList(
			tx, int32(fleetServiceId), order,
		); err != nil {
			return
		}

		if err = insertFleetAddonList(
			tx, int32(fleetServiceId), order,
		); err != nil {
			return
		}
	}
}

func insertFleetServiceList(tx *sql.Tx,
	fleetServiceId int32,
	order modules.FleetOrder,
) (err error) {
	var (
		result     sql.Result
		insertedId int64
	)

	query := `
		insert into fleet_service_list (fleet_service_id, order_id, car_count)
		values (?, ?, ?)
	`
	if result, err = tx.Exec(
		query, fleetServiceId, order.OrderId, order.CarCount,
	); err != nil {
		return
	} else if insertedId, err = result.LastInsertId(); err != nil {
		return
	}

	query = `
		insert into fleet_service_list_id (service_id, fleet_service_list_id)
		values (?, ?)
	`
	for _, id := range order.Services {
		if _, ok := cache.SERVICES_MAP[id]; ok {
			if _, err = tx.Exec(
				query, id, insertedId,
			); err != nil {
				return
			}
		}
	}

	return nil
}

func insertFleetAddonList(tx *sql.Tx,
	fleetServiceId int32,
	order modules.FleetOrder,
) (err error) {
	var (
		insertedId int64
		result     sql.Result
	)

	query := `
		insert into fleet_service_addon_list (fleet_service_id, order_id, car_count)
		values (?, ?, ?)
	`
	if result, err = tx.Exec(
		query, fleetServiceId, order.OrderId, order.CarCount,
	); err != nil {
		return
	} else if insertedId, err = result.LastInsertId(); err != nil {
		return
	}

	query = `
		insert into fleet_service_addon_list_id (service_addon_id,
			fleet_service_addon_list_id, amount
		) values (?, ?, ?)
	`
	for _, addon := range order.Addons {
		if _, ok := cache.FLEET_ADDONS_MAP[addon.Id]; ok {
			if _, err = tx.Exec(
				query, addon.Id, insertedId, addon.Amount,
			); err != nil {
				return
			}
		}
	}

	return nil
}

func CancelFleetOrder(c *gin.Context) {
	cancelFleet(c, false)
}

func ForceCancelFleetOrder(c *gin.Context) {
	cancelFleet(c, true)
}

func cancelFleet(c *gin.Context, force bool) {
	request := modules.CancelRequest{}
	temp := struct {
		Opening  int32
		Gap      int32
		Assignee int32
		Types    string
		Status   string
	}{}

	var (
		tx   *sql.Tx
		body []byte
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

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
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

	query := `
		select opening_id, gap, assignee, types, status
		from fleet_service
		where id = ? and user_id = ?
	`
	if err = tx.QueryRow(
		query, request.Id, request.UserId,
	).Scan(
		&temp.Opening, &temp.Gap, &temp.Assignee,
		&temp.Types, &temp.Status,
	); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("Reservation not found")
		}
		return
	}

	if !force && temp.Status != "WAITING" {
		err = errors.New("CANNOT")
		return
	}

	query = `
		update fleet_service set status = 'CANCEL', assignee = -1
		where id = ? and user_id = ?
	`
	if _, err = tx.Exec(
		query, request.Id, request.UserId,
	); err != nil {
		return
	}

	go cancelReservation(request.UserId)

	if err = releaseOpening(
		tx, temp.Opening, temp.Gap, temp.Types,
	); err != nil {
		return
	}

	if err = revokeUserOpening(
		tx, temp.Opening, temp.Gap, temp.Assignee,
	); err != nil {
		return
	}
}

func GetFleetOpening(c *gin.Context) {
	query := `
		select id, day, period
		from opening
		where day > DATE_FORMAT(CURDATE(), '%Y-%m-%d')
	`
	request := modules.FleetOpeningRequest{}
	var (
		totalTime int32
		body      []byte
		err       error
		openings  []modules.Opening
		types     string
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, openings)
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if len(request.Orders) == 0 {
		err = errors.New("Please choose services")
		return
	}

	totalTime, types = calculateFleetOrderTimeAndTypes(
		request.Orders,
	)

	if openings, err = loadOpening(query, types); err != nil {
		return
	}

	openings, err = filterOpening(
		calculateGap(totalTime), openings,
	)
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

	if fleetServices, err := getFleetServiceByFleetUser(
		request.UserId,
		"status = 'WAITING' or status = 'RESERVED' or status = 'IN_PROGRESS'",
	); err == nil {
		c.JSON(http.StatusOK, fleetServices)
	}
}

func GetFleetReservationDetail(c *gin.Context) {
	request := modules.FleetReservationRequest{}
	var (
		err      error
		data     []byte
		details  []modules.FleetReservationDetail
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		} else {
			c.JSON(http.StatusOK, details)
		}
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if details, err = getFleetReservationDetail(
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
		data        []byte
		err         error
		rows        *sql.Rows
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
	query := `
		select f.order_id, f.car_count, fi.service_id
		from fleet_service_list f
		inner join fleet_service_list_id fi on fi.fleet_service_list_id = f.id
		where f.fleet_service_id = ?
		order by f.order_id, f.id
	`
	var (
		rows          *sql.Rows
		tempServiceId int32
		ok            bool
		service       modules.Service
	)

	if rows, err = config.DB.Query(query, fleetServiceId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		s := modules.FleetReservationService{}

		if err = rows.Scan(
			&s.OrderId, &s.CarCount, &tempServiceId,
		); err != nil {
			return
		}

		if service, ok = cache.SERVICES_MAP[tempServiceId]; !ok {
			fmt.Println("Invalid ServiceId - fleet reservation - ", tempServiceId)
			continue
		}

		s.Name = service.Name
		s.Type = SERVICE_TYPES[service.Type]
		s.Note = service.Note

		services = append(services, s)
	}

	return services, nil
}

func getFleetReservationAddon(fleetServiceId int32) (
	addons []modules.FleetReservationAddon, err error,
) {
	query := `
		select f.order_id, f.car_count, fi.service_addon_id
		from fleet_service_addon_list f
		inner join fleet_service_addon_list_id fi on fi.fleet_service_addon_list_id = f.id
		where f.fleet_service_id = ?
		order by f.order_id, f.id
	`
	var (
		tempAddonId int32
		addon       modules.AddOn
		ok          bool
		rows        *sql.Rows
	)

	if rows, err = config.DB.Query(query, fleetServiceId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		a := modules.FleetReservationAddon{}

		if err = rows.Scan(
			&a.OrderId, &a.CarCount, &tempAddonId,
		); err != nil {
			return
		}

		if addon, ok = cache.FLEET_ADDONS_MAP[tempAddonId]; !ok {
			fmt.Println("Invalid AddonId - fleet reservation - ", tempAddonId)
			continue
		}

		a.Name = addon.Name
		a.Note = addon.Note

		addons = append(addons, a)
	}

	return addons, nil
}

func getFleetReservationDetail(fleetServiceId int32) (
	details  []modules.FleetReservationDetail, err error,
) {
	index := make(map[int32]int)
	var (
		addons   []modules.FleetReservationAddon
		services []modules.FleetReservationService
	)

	if services, err = getFleetReservationService(
		fleetServiceId,
	); err != nil {
		return
	}

	if addons, err = getFleetReservationAddon(
		fleetServiceId,
	); err != nil {
		return
	}

	for _, service := range services {
		if _, ok := index[service.OrderId]; !ok {
			index[service.OrderId] = len(details)
			detail := modules.FleetReservationDetail{}
			detail.CarCount = service.CarCount

			details = append(details, detail)
		}

		details[index[service.OrderId]].Services = append(
			details[index[service.OrderId]].Services,
			service,
		)
	}

	for _, addon := range addons {
		if _, ok := index[addon.OrderId]; ok {
			details[index[addon.OrderId]].Addons = append(
				details[index[addon.OrderId]].Addons,
				addon,
			)
		}
	}

	return details, nil
}

func getFleetServiceByFleetUser(userId int32, condition string) (
	fleetServices []modules.FleetService, err error,
) {
	return getFleetService(`
		select fs.id, fs.reservation_id, fs.estimated_time,
				fs.estimated_price, fs.reserved_start_timestamp,
				TIMESTAMPDIFF(MINUTE, CURRENT_TIMESTAMP(), fs.reserved_start_timestamp) as mins,
				fs.start_timestamp, fs.end_timestamp,
				fs.note, fs.status, fs.create_timestamp
		from fleet_service fs
		where fs.user_id = ? and (
	` + condition + ") order by fs.create_timestamp DESC", userId,
	)
}

func getFleetServiceBySaleUser(userId int32, condition string) (
	fleetServices []modules.FleetService, err error,
) {
	return getFleetService(`
		select fs.id, fs.reservation_id, fs.estimated_time,
				fs.estimated_price, fs.reserved_start_timestamp,
				TIMESTAMPDIFF(MINUTE, CURRENT_TIMESTAMP(), fs.reserved_start_timestamp) as mins,
				fs.start_timestamp, fs.end_timestamp,
				fs.note, fs.status, fs.create_timestamp
		from fleet_service fs
		inner join fleet f on f.user_id = fs.user_id
		where f.sale_user_id = ? and (
	` + condition + ") order by fs.create_timestamp DESC", userId,
	)
}

func getFleetService(query string, userId int32) (
	fleetServices []modules.FleetService, err error,
) {
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

func calculateFleetOrderTimeAndTypes(
	orders []modules.FleetOrder,
) (totalTime int32, types string) {
	var (
		tempTime int32
		wash     bool
		oil      bool
	)

	for _, order := range orders {
		// Service
		tempTime = 0

		for _, id := range order.Services {
			if s, ok := cache.SERVICES_MAP[id]; ok {
				tempTime += s.Time

				if s.Type == "CAR_WASH" {
					wash = true
				} else if s.Type == "OIL_CHANGE" {
					oil = true
				}
			}
		}

		// TODO totalTime += tempTime * order.CarCount
		totalTime += tempTime

		// Addon
		tempTime = 0

		for _, addon := range order.Addons {
			if a, ok := cache.FLEET_ADDONS_MAP[addon.Id]; ok {
				tempTime += a.Time
			}
		}

		// TODO totalTime += tempTime * addon.CarCount
		totalTime += tempTime
	}

	types = calculateOrderTypes(wash, oil)
	fmt.Println("Types - ", types)

	return
}

func createFleetHistory(tx *sql.Tx, serviceId int32) (err error) {
	query := `
		insert into fleet_history (fleet_service_id)
		values (?)
	`

	_, err = tx.Exec(query, serviceId)

	return
}
