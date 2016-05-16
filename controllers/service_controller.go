package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

var OPENING_GAP = 0.5
var OPENING_BASE = 8.0

func getUserService(userId int32, condition string) (userServices []modules.UserService, err error) {
	query := `
		select us.id, us.reservation_id, us.user_id, us.user_car_id, uc.plate,
				us.user_payment_id, us.estimated_time, us.estimated_price,
				us.reserved_start_timestamp,
				TIMESTAMPDIFF(MINUTE, CURRENT_TIMESTAMP(), us.reserved_start_timestamp) as mins,
				us.start_timestamp, us.end_timestamp,
				us.note, us.status, us.create_timestamp,
				s.id, s.name, s.type, s.items, s.description,
				s.estimated_time, s.estimated_price
		from user_service us
		inner join user_car uc on uc.id = us.user_car_id
		inner join user_service_list usl on usl.user_service_id = us.id
		inner join service s on s.id = usl.service_id
		where us.user_id = ? and (
	` + condition + ") order by us.id"

	index := make(map[int32]int32)
	var (
		rows1  *sql.Rows
		rows2  *sql.Rows
		temp   string
		tempId int32
		mins   int32
		ids    []int32
	)

	if rows1, err = config.DB.Query(query, userId); err != nil {
		return
	}
	defer rows1.Close()

	for rows1.Next() {
		userService := modules.UserService{}
		service := modules.Service{}

		if err = rows1.Scan(
			&userService.Id, &userService.ReservationId, &userService.UserId,
			&userService.CarId, &userService.CarPlate, &userService.PaymentId,
			&userService.Time, &userService.Price, &userService.ReserveStartTime,
			&mins, &userService.StartTime, &userService.EndTime, &userService.Note,
			&userService.Status, &userService.ReserveTime, &service.Id, &service.Name,
			&service.Type, &temp, &service.Description, &service.Time, &service.Price,
		); err != nil {
			return
		}

		if err = json.Unmarshal([]byte(temp), &service.Items); err != nil {
			return
		}

		if len(userServices) != 0 && userServices[len(userServices)-1].Id == userService.Id {
			userServices[len(userServices)-1].ServiceList = append(
				userServices[len(userServices)-1].ServiceList, service,
			)
		} else {
			if mins >= 1440 {
				userService.HowLong = int32(mins / 1440)
				userService.Unit = "DAY"
			} else if mins >= 60 {
				userService.HowLong = int32(mins / 60)
				userService.Unit = "HOUR"
			} else {
				userService.HowLong = mins
				userService.Unit = "MINUTE"
			}

			userService.ServiceList = append(userService.ServiceList, service)
			userServices = append(userServices, userService)

			ids = append(ids, userService.Id)
			index[userService.Id] = int32(len(userServices) - 1)
		}
	}

	query = `
		select sa.id, sa.service_id, sa.name, sa.note,
				sa.price, sa.time, sa.max, sa.unit,
				usal.amount, usal.user_service_id
		from service_addon sa
		inner join user_service_addon_list usal on usal.service_addon_id = sa.id
		where usal.user_service_id in (
	`

	for i, id := range ids {
		if i == 0 {
			query += strconv.Itoa(int(id))
		} else {
			query += (", " + strconv.Itoa(int(id)))
		}
	}

	query += ")"

	if rows2, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		addon := modules.AddOn{}

		if err = rows2.Scan(
			&addon.Id, &addon.ServiceId, &addon.Name, &addon.Note,
			&addon.Price, &addon.Time, &addon.Max, &addon.Unit,
			&addon.Amount, &tempId,
		); err != nil {
			return
		}

		userServices[index[tempId]].AddonList = append(
			userServices[index[tempId]].AddonList, addon,
		)
	}

	return userServices, nil
}

func GetUserServiceReserved(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		err  error
		body []byte
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if userServices, err := getUserService(
		request.UserId, "status = 'RESERVED'",
	); err == nil {
		c.IndentedJSON(http.StatusOK, userServices)
	}
}

func GetUserServiceDone(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		err  error
		body []byte
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if userServices, err := getUserService(
		request.UserId, "status = 'DONE'",
	); err == nil {
		c.IndentedJSON(http.StatusOK, userServices)
	}
}

func OnDemand(c *gin.Context) {
	query := `
		select COUNT(*), SUM(estimated_time)
		from user_service
		where DATE_FORMAT(create_timestamp, '%Y-%m-%d') = DATE_FORMAT(CURDATE(), '%Y-%m-%d')
		and status != 'DONE'
	`
	temp := struct {
		count int32
		time  int32
	}{}

	var err error

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if err := config.DB.QueryRow(query).Scan(
		&temp.count, &temp.time,
	); err == nil {
		c.IndentedJSON(http.StatusOK, err.Error())
	}
}

func GetOpening(c *gin.Context) {
	query := `
		select id, day, period
		from opening
		where count > 0 and day > DATE_FORMAT(CURDATE(), '%Y-%m-%d')
		order by day, period
	`
	request := modules.OpeningRequest{}
	var (
		rows     *sql.Rows
		body     []byte
		err      error
		preDay   string
		openings []modules.Opening
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if len(request.Services) == 0 {
		err = errors.New("Please provide services")
		return
	}

	if rows, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		temp := struct {
			Id     int32
			Day    string
			Period int32
		}{}

		if err = rows.Scan(&temp.Id, &temp.Day, &temp.Period); err != nil {
			return
		} else {
			if preDay != temp.Day {
				preDay = temp.Day
				openings = append(openings, modules.Opening{})
				openings[len(openings)-1].Day = temp.Day
			}

			openings[len(openings)-1].Range = append(
				openings[len(openings)-1].Range,
				modules.Period{
					temp.Id,
					OPENING_BASE + float64(temp.Period-1)*OPENING_GAP,
					OPENING_BASE + float64(temp.Period)*OPENING_GAP,
				},
			)
		}
	}

	if openings, err = filterOpening(
		request.Services, request.Addons, openings,
	); err == nil {
		c.IndentedJSON(http.StatusOK, openings)
	}
}

func filterOpening(services, addons []int32, openings []modules.Opening) (result []modules.Opening, err error) {
	var (
		time int32
		p1   int32
		p2   int32
		pre  int32
	)

	if time, err = getTotalTime(services, addons); err != nil {
		return
	}

	if time%30 != 0 {
		time = (time / 30) + 2
	} else {
		time = (time / 30) + 1
	}

	for _, opening := range openings {
		o := modules.Opening{}
		o.Day = opening.Day

		p1 = 0
		p2 = 0
		pre = opening.Range[p1].Id
		size := int32(len(opening.Range))

		if p1 < (size - time + 1) {
			for p2 < size {
				if opening.Range[p2].Id-pre > 1 {
					p1 = p2
					pre = opening.Range[p1].Id
				} else {
					if opening.Range[p2].Id-opening.Range[p1].Id+1 == time {
						o.Range = append(o.Range, opening.Range[p1])
						p1 += 1
					}

					pre = opening.Range[p2].Id
				}

				p2 += 1
			}
		}

		if len(o.Range) != 0 {
			result = append(result, o)
		}
	}

	return result, nil
}

func getTotalTime(services, addons []int32) (time int32, err error) {
	query1 := `
		select sum(estimated_time) from service where id in (
	`
	query2 := `
		select sum(time) from service_addon where id in (
	`
	var (
		time1 int32
		time2 int32
	)

	for i, id := range services {
		if i == 0 {
			query1 += strconv.Itoa(int(id))
		} else {
			query1 += "," + strconv.Itoa(int(id))
		}
	}

	query1 += ")"

	if err = config.DB.QueryRow(query1).Scan(&time1); err != nil {
		return
	}

	if len(addons) > 0 {
		for i, id := range addons {
			if i == 0 {
				query2 += strconv.Itoa(int(id))
			} else {
				query2 += "," + strconv.Itoa(int(id))
			}
		}

		query2 += ")"

		if err = config.DB.QueryRow(query2).Scan(&time2); err != nil {
			return
		}
	}

	time = time1 + time2

	return time, nil
}

func PlaceOrder(c *gin.Context) {
	request := modules.OrderRequest{}
	car := modules.Car{}
	payment := modules.Payment{}
	user := modules.User{}
	info := modules.ServiceInfo{}

	var (
		result      sql.Result
		tx          *sql.Tx
		rows        *sql.Rows
		body        []byte
		err         error
		price       float32
		time        int32
		count       int32
		gap         int32
		insertedId  int64
		affectedRow int64
		reserved    string
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	updateOpeningDemand(request.Opening)

	if rows, err = config.DB.Query(
		buildServicesQuery(request.Services),
	); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&info.Type, &info.Count, &info.Price, &info.Time,
		); err != nil {
			return
		}

		if info.Count > 1 {
			err = errors.New("You can only select one service for each type")
			return
		}

		count++
		time += info.Time
		price += info.Price
	}

	if user, err = getUserById(request.UserId); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("User not found")
		}

		return
	}

	gap = time/30 + 1

	if time%30 != 0 {
		gap += 1
	}

	if payment, err = getPaymentByIdAndUserId(
		request.PaymentId, request.UserId,
	); err != nil {
		switch {
		case err == sql.ErrNoRows:
			c.IndentedJSON(http.StatusBadRequest, "Payment not found")
		default:
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		c.Abort()
		return
	}

	if car, err = getCarByIdAndUserId(
		request.CarId, request.UserId,
	); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("Car not found")
		}

		return
	}

	if tx, err = config.DB.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				fmt.Println("Error -Rollback - ", err1.Error())
			}
		} else {
			if err1 := tx.Commit(); err1 != nil {
				fmt.Println("Error - Commit - ", err1.Error())
			}
		}
	}()

	updateOpening := `
		update opening set count = count - 1 where id >= ? and id < ? and count > 0
	`

	if result, err = tx.Exec(
		updateOpening, request.Opening, request.Opening+gap,
	); err != nil {
		return
	}

	if affectedRow, err = result.RowsAffected(); err != nil {
		return
	} else if affectedRow != int64(gap) {
		err = errors.New("Opening is not available")
		return
	}

	temp := struct {
		day    string
		period int32
	}{}

	if err = tx.QueryRow(
		"select day, period from opening where id = ?", request.Opening,
	).Scan(&temp.day, &temp.period); err != nil {
		return
	} else {
		total := (temp.period - 1) * 30
		hour := strconv.Itoa(int(OPENING_BASE) + int(total/60))
		minute := total % 60

		if minute == 0 {
			reserved = temp.day + " " + hour + ":00:00"
		} else {
			reserved = temp.day + " " + hour + ":30:00"
		}
	}

	insertUserService := `
		insert into user_service (
			user_id, user_car_id, user_payment_id, opening_id,
			reserved_start_timestamp, gap,
			estimated_time, estimated_price, status
		) values (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	if result, err = tx.Exec(insertUserService,
		user.Id, car.Id, payment.Id, request.Opening,
		reserved, gap, time, price, "RESERVED",
	); err != nil {
		return
	}

	if insertedId, err = result.LastInsertId(); err != nil {
		return
	}

	user_service_id := int32(insertedId)

	queryUserServiceList := `
		insert into user_service_list (
			service_id, user_service_id
		) values (?, ?)
	`

	for _, id := range request.Services {
		if _, err = config.DB.Exec(
			queryUserServiceList, id, user_service_id,
		); err != nil {
			return
		}
	}

	queryUserServiceAddonList := `
		insert into user_service_addon_list (
			service_addon_id, user_service_id, amount
		) values (?, ?, ?)
	`

	for _, addon := range request.Addons {
		if _, err = config.DB.Exec(
			queryUserServiceAddonList, addon.Id, user_service_id, addon.Amount,
		); err != nil {
			return
		}
	}

	if err = lockCar(tx, request.CarId, request.UserId); err != nil {
		return
	}

	if err = lockPayment(tx, request.PaymentId, request.UserId); err != nil {
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func buildServicesQuery(ids []int32) string {
	queryServices := `
		select type, count(*), sum(estimated_price), sum(estimated_time)
		from service where id in (
	`

	for index, id := range ids {
		if index == 0 {
			queryServices += strconv.Itoa(int(id))
		} else {
			queryServices += ("," + strconv.Itoa(int(id)))
		}
	}

	return queryServices + ") group by type"
}

func CancelOrder(c *gin.Context) {
	checkQuery := `
		select user_car_id, user_payment_id, opening_id, gap
		from user_service
		where DATE_ADD(CURRENT_TIMESTAMP(), INTERVAL 1 DAY) < reserved_start_timestamp
		and id = ? and user_id = ?
	`
	query := `
		update user_service set status = 'CANCEL' where id = ? and user_id = ?
	`
	request := modules.CancelRequest{}
	temp := struct {
		CarId     int32
		PaymentId int32
		Opening   int32
		Gap       int32
	}{}

	var (
		tx   *sql.Tx
		body []byte
		err  error
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
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
			if err1 := tx.Commit(); err1 != nil {
				fmt.Println("Error - Commit - ", err1.Error())
			}
		}
	}()

	if err = tx.QueryRow(
		checkQuery, request.Id, request.UserId,
	).Scan(
		&temp.CarId, &temp.PaymentId, &temp.Opening, &temp.Gap,
	); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("Cannot cancel this order")
		}

		return
	}

	if _, err = tx.Exec(
		query, request.Id, request.UserId,
	); err != nil {
		return
	}

	if err = unlockCar(tx, temp.CarId, request.UserId); err != nil {
		return
	}

	if err = unlockPayment(tx, temp.PaymentId, request.UserId); err != nil {
		return
	}

	if err = releaseOpening(tx, temp.Opening, temp.Gap); err != nil {
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func OpeningDemand(c *gin.Context) {
	if id, err := strconv.ParseInt(c.Param("id"), 10, 32); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		updateOpeningDemand(int32(id))
		c.IndentedJSON(http.StatusOK, "OK")
	}
}

func updateOpeningDemand(id int32) {
	query := `update opening set demand = demand + 1 where id = ?`

	if _, err := config.DB.Exec(query, id); err != nil {
		fmt.Println("Error - ", err.Error())
	}
}

func GetService(c *gin.Context) {
	query := `
		select id, name, type, items, description, note,
			estimated_price, estimated_time
		from service
		order by id
	`
	index := make(map[int32]int32)
	var (
		rows1    *sql.Rows
		rows2    *sql.Rows
		services []modules.Service
		err      error
		temp     string
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if rows1, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows1.Close()

	for rows1.Next() {
		service := modules.Service{}
		if err = rows1.Scan(
			&service.Id, &service.Name, &service.Type, &temp,
			&service.Description, &service.Note, &service.Price, &service.Time,
		); err != nil {
			return
		}

		if err = json.Unmarshal([]byte(temp), &service.Items); err != nil {
			return
		}

		index[service.Id] = int32(len(services))
		services = append(services, service)
	}

	if rows2, err = config.DB.Query(`
		select id, service_id, name, note,
			price, time, max, unit
		from service_addon
	`); err != nil {
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		addOn := modules.AddOn{}
		if err = rows2.Scan(
			&addOn.Id, &addOn.ServiceId, &addOn.Name, &addOn.Note,
			&addOn.Price, &addOn.Time, &addOn.Max, &addOn.Unit,
		); err != nil {
			return
		}

		addOn.Amount = 1

		if addOn.Price == 0 {
			services[index[addOn.ServiceId]].Free = append(
				services[index[addOn.ServiceId]].Free, addOn,
			)
		} else if addOn.Time == 0 {
			services[index[addOn.ServiceId]].Charge = append(
				services[index[addOn.ServiceId]].Charge, addOn,
			)
		} else {
			services[index[addOn.ServiceId]].Addons = append(
				services[index[addOn.ServiceId]].Addons, addOn,
			)
		}
	}

	c.IndentedJSON(http.StatusOK, services)
}

func AddonDemand(c *gin.Context) {
	request := modules.AddonDemandRequest{}
	var (
		body []byte
		err  error
	)

	defer func() {
		if err != nil {
			fmt.Println(err.Error())
		}

		c.IndentedJSON(http.StatusOK, "OK")
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	updateAddonDemand(request.Addons)
}

func updateAddonDemand(ids []int32) {
	query := `update service_addon set demand = demand + 1 where id in (`
	last := len(ids) - 1

	for i, id := range ids {
		query += strconv.Itoa(int(id))
		if i != last {
			query += ","
		}
	}

	query += ")"

	if _, err := config.DB.Exec(query); err != nil {
		fmt.Println("Error - Addon Demand - ", err.Error())
	}
}

func ServiceReading(c *gin.Context) {
	query := `update service set reading = reading + 1 where id = ?`
	var (
		err error
		id  int64
	)

	defer func() {
		if err != nil {
			fmt.Println(err.Error())
		}

		c.IndentedJSON(http.StatusOK, "OK")
	}()

	if id, err = strconv.ParseInt(c.Param("id"), 10, 32); err != nil {
		return
	}

	if _, err = config.DB.Exec(query, int32(id)); err != nil {
		return
	}
}

func ServiceDemand(c *gin.Context) {
	request := modules.ServiceDemandRequest{}
	var (
		body []byte
		err  error
	)

	defer func() {
		if err != nil {
			fmt.Println(err.Error())
		}

		c.IndentedJSON(http.StatusOK, "OK")
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	updateServiceDemand(request.Services)
}

func updateServiceDemand(ids []int32) {
	query := `update service set demand = demand + 1 where id in (`
	last := len(ids) - 1

	for i, id := range ids {
		query += strconv.Itoa(int(id))
		if i != last {
			query += ","
		}
	}

	query += ")"

	if _, err := config.DB.Exec(query); err != nil {
		fmt.Println("Error - Service Demand - ", err.Error())
	}
}

func makeServicePaid(userId, serviceId, paymentId int32) (err error) {
	_, err = config.DB.Exec(`
		update user_service set paid = 1
		where id = ? and user_id = ? and user_payment_id = ?
			and status = "DONE" and paid = 0
	`, serviceId, userId, paymentId)

	if err != nil {
		fmt.Println("Error - Make Service Paid - ", err.Error())
	}

	return
}

func releaseOpening(tx *sql.Tx, id, gap int32) (err error) {
	_, err = tx.Exec(`
		update opening set count = count + 1 where id >= ? and id < ?
	`, id, id+gap)

	return
}
