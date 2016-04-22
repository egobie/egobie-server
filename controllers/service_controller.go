package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/eGobie/egobie-server/config"
	"github.com/eGobie/egobie-server/modules"

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
		select id, day, period from opening where count > 0 order by day, period
	`
	var (
		rows     *sql.Rows
		err      error
		preDay   string
		openings []modules.Opening
	)

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
					temp.Id, 8 + (temp.Period - 1), 8 + temp.Period,
				},
			)
		}
	}

	c.IndentedJSON(http.StatusOK, openings)
}

func PlaceOrder(c *gin.Context) {
	request := modules.OrderRequest{}
	car := modules.Car{}
	payment := modules.Payment{}
	user := modules.User{}
	info := modules.ServiceInfo{}

	var (
		result      sql.Result
		stmt        *sql.Stmt
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

	updateDemand(request.Opening)

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

	if stmt, err = tx.Prepare(queryUserServiceList); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()

		if err = tx.Rollback(); err != nil {
			fmt.Println("Fail to rollback - ", err.Error())
		}

		return
	}

	for _, id := range request.Services {
		if _, err = stmt.Exec(id, user_service_id); err != nil {
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

func Demand(c *gin.Context) {
	if id, err := strconv.ParseInt(c.Param("id"), 10, 32); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		updateDemand(int32(id))
		c.IndentedJSON(http.StatusOK, "OK")
	}
}

func updateDemand(id int32) {
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
