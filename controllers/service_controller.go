package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

func GetUserService(c *gin.Context) {
	query := `
		select us.id, us.user_id, us.user_car_id, uc.plate,
				us.user_payment_id, us.estimated_time, us.estimated_price,
				us.start_timestamp, us.end_timestamp,
				us.note, us.status, us.create_timestamp,
				s.id, s.name, s.type, s.items, s.description,
				s.estimated_time, s.estimated_price, s.addons
		from user_service us
		inner join user_car uc on uc.id = us.user_car_id
		inner join user_service_list usl on usl.user_service_id = us.id
		inner join service s on s.id = usl.service_id
		where us.user_id = ? order by us.id
	`
	request := modules.BaseRequest{}
	var (
		rows         *sql.Rows
		userServices []modules.UserService
		err          error
		body         []byte
		temp         string
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if rows, err = config.DB.Query(query, request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	for rows.Next() {
		userService := modules.UserService{}
		service := modules.Service{}

		if err = rows.Scan(
			&userService.Id, &userService.UserId, &userService.CarId,
			&userService.CarPlate, &userService.PaymentId, &userService.Time,
			&userService.Price, &userService.StartTime, &userService.EndTime,
			&userService.Note, &userService.Status, &userService.ReserveTime,
			&service.Id, &service.Name, &service.Type, &temp, &service.Description,
			&service.Time, &service.Price, &service.AddOns,
		); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		if err = json.Unmarshal([]byte(temp), &service.Items); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		if len(userServices) != 0 && userServices[len(userServices)-1].Id == userService.Id {
			userServices[len(userServices)-1].ServiceList = append(
				userServices[len(userServices)-1].ServiceList, service,
			)
		} else {
			userService.ServiceList = append(userService.ServiceList, service)
			userServices = append(userServices, userService)
		}
	}

	c.IndentedJSON(http.StatusOK, userServices)
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

	if err := config.DB.QueryRow(query).Scan(&temp.count, &temp.time); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
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
	gap := 0.5
	var (
		rows     *sql.Rows
		body     []byte
		request  []int32
		err      error
		preDay   string
		openings []modules.Opening
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if len(request) == 0 {
		c.IndentedJSON(http.StatusBadRequest, "Please provide services")
		c.Abort()
		return
	}

	if rows, err = config.DB.Query(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
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
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
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
					8.0 + float32(temp.Period-1)*float32(gap),
					8.0 + float32(temp.Period)*float32(gap),
				},
			)
		}
	}

	if openings, err = filterOpening(request, openings); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.IndentedJSON(http.StatusOK, openings)
}

func filterOpening(services []int32, openings []modules.Opening) (result []modules.Opening, err error) {
	query := `
		select sum(estimated_time) from service where id in (
	`
	var (
		time int32
		p1   int32
		p2   int32
		pre  int32
	)

	for i, id := range services {
		if i == 0 {
			query += strconv.Itoa(int(id))
		} else {
			query += "," + strconv.Itoa(int(id))
		}
	}

	query += ")"

	if err = config.DB.QueryRow(query).Scan(&time); err != nil {
		return
	}

	if time%30 != 0 {
		time = (time / 30) + 2
	} else {
		time = (time / 30) + 1
	}

	for _, opening := range openings {
		o := modules.Opening{}
		o.Day = opening.Day;

		p1 = 0
		p2 = 0
		pre = opening.Range[p1].Id
		size := int32(len(opening.Range))

		if p1 < (size - time + 1) {
			for p2 < size {
				if opening.Range[p2].Id - pre > 1 {
					p1 = p2;
					pre = opening.Range[p1].Id
				} else {
					if opening.Range[p2].Id - opening.Range[p1].Id + 1 == time {
						o.Range = append(o.Range, opening.Range[p1])
						p1 += 1
					}

					pre = opening.Range[p2].Id
				}

				p2 += 1
			}
		}

		if len(o.Range) != 0{
			result = append(result, o)
		}
	}

	return result, nil
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
		insertedId  int64
		affectedRow int64
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	updateOpeningDemand(request.Opening)

	if rows, err = config.DB.Query(
		buildServicesQuery(request.Services),
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&info.Type, &info.AddOns, &info.Count, &info.Price, &info.Time,
		); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		if info.Count > 1 && !info.AddOns {
			c.IndentedJSON(http.StatusBadRequest, "You can only select one service for each type")
			c.Abort()
			return
		}

		count++
		time += info.Time
		price += info.Price
	}

	if user, err = getUserById(request.UserId); err != nil {
		switch {
		case err == sql.ErrNoRows:
			c.IndentedJSON(http.StatusBadRequest, "User not found")
		default:
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		c.Abort()
		return
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
		switch {
		case err == sql.ErrNoRows:
			c.IndentedJSON(http.StatusBadRequest, "Car not found")
		default:
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		c.Abort()
		return
	}

	if tx, err = config.DB.Begin(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	updateOpening := `
		update opening set count = count - 1 where id = ? and count > 0
	`

	if result, err = tx.Exec(updateOpening, request.Opening); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()

		if err = tx.Rollback(); err != nil {
			fmt.Println("Fail to rollback - ", err.Error())
		}

		return
	}

	if affectedRow, err = result.RowsAffected(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()

		if err = tx.Rollback(); err != nil {
			fmt.Println("Fail to rollback - ", err.Error())
		}

		return
	} else if affectedRow != 1 {
		c.IndentedJSON(http.StatusBadRequest, "Opening is not available")
		c.Abort()

		if err = tx.Rollback(); err != nil {
			fmt.Println("Fail to rollback - ", err.Error())
		}

		return
	}

	insertUserService := `
		insert into user_service (
			user_id, user_car_id, user_payment_id, opening_id,
			estimated_time, estimated_price, status
		) values (?, ?, ?, ?, ?, ?, ?)
	`

	if result, err = tx.Exec(insertUserService,
		user.Id, car.Id, payment.Id, request.Opening, time, price, "RESERVED",
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()

		if err = tx.Rollback(); err != nil {
			fmt.Println("Fail to rollback - ", err.Error())
		}

		return
	}

	if insertedId, err = result.LastInsertId(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()

		if err = tx.Rollback(); err != nil {
			fmt.Println("Fail to rollback - ", err.Error())
		}

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
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()

			if err = tx.Rollback(); err != nil {
				fmt.Println("Fail to rollback - ", err.Error())
			}

			return
		}
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("Fail to commit - ", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()

		return
	}

	lockCar(request.CarId)
	lockPayment(request.PaymentId)

	fmt.Println("User - ", user)
	fmt.Println("Payment - ", payment)
	fmt.Println("Car - ", car)

	c.IndentedJSON(http.StatusOK, "OK")
}

func buildServicesQuery(ids []int32) string {
	queryServices := `
		select type, addons, count(*), sum(estimated_price), sum(estimated_time)
		from service where id in (
	`
	fmt.Println("Ids - ", ids)

	for index, id := range ids {
		if index == 0 {
			queryServices += strconv.Itoa(int(id))
		} else {
			queryServices += ("," + strconv.Itoa(int(id)))
		}
	}

	return queryServices + ") group by type, addons"
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
		select id, name, type, items, description,
			estimated_price, estimated_time, addons
		from service
		order by id
	`
	var (
		rows     *sql.Rows
		services []modules.Service
		err      error
		temp     string
	)

	if rows, err = config.DB.Query(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}
	defer rows.Close()

	for rows.Next() {
		service := modules.Service{}
		if err = rows.Scan(
			&service.Id, &service.Name, &service.Type, &temp,
			&service.Description, &service.Price, &service.Time, &service.AddOns,
		); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		if err = json.Unmarshal([]byte(temp), &service.Items); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		services = append(services, service)
	}

	c.IndentedJSON(http.StatusOK, services)
}

func ServiceReading(c *gin.Context) {
	query := `update service set reading = reading + 1 where id = ?`
	var (
		err error
		id  int64
	)

	if id, err = strconv.ParseInt(c.Param("id"), 10, 32); err != nil {
		c.IndentedJSON(http.StatusOK, err.Error())
		c.Abort()
		return
	}

	if _, err = config.DB.Exec(query, int32(id)); err != nil {
		fmt.Println("Error - Service Reading - ", err.Error())
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func ServiceDemand(c *gin.Context) {
	var (
		body    []byte
		err     error
		request []int32
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	updateServiceDemand(request)
	c.IndentedJSON(http.StatusOK, "OK")
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
