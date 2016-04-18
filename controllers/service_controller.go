package controllers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/eGobie/server/config"
	"github.com/eGobie/server/modules"

	"github.com/gin-gonic/gin"
)

func GetUserService(c *gin.Context) {
	query := `
		select us.id, us.user_id, us.user_car_id, uc.plate,
				us.user_payment_id, us.estimated_time, us.estimated_price,
				us.start_timestamp, us.end_timestamp,
				us.note, us.status, us.create_timestamp,
				s.id, s.name, s.description, s.estimated_time, s.estimated_price
		from user_service us
		inner join user_car uc on uc.id = us.user_car_id
		inner join user_service_list usl on usl.user_service_id = us.id
		inner join service s on s.id = usl.service_id
		where us.user_id = ? order by us.id
	`
	request := modules.BaseRequest{}
	var (
		stmt         *sql.Stmt
		rows         *sql.Rows
		userServices []modules.UserService
		err          error
		body         []byte
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

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if rows, err = stmt.Query(request.UserId); err != nil {
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
			&service.Id, &service.Name, &service.Description,
			&service.Time, &service.Price,
		); err != nil {
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

func PrepareOrder(c *gin.Context) {
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

//func PlaceOrder(c *gin.Context) {
//	queryUserService := `
//		insert into user_service (
//			user_id, user_car_id, user_payment_id, estimated_time,
//			estimated_price, start_timestamp, end_timestamp, status
//		) values (?, ?, ?, ?, ?, ?, ?, ?)
//	`
//	queryUserServiceList := `
//		insert into user_service_list (service_id, user_service_id) values (?, ?)
//	`
//	request := modules.OrderRequest{}
//	var (
//		stmt   *sql.Stmt
//		result sql.Result
//		tx     *sql.Tx
//		err    error
//		body   []byte
//	)

//	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
//		c.IndentedJSON(http.StatusBadRequest, err.Error())
//		c.Abort()
//		return
//	}

//	if err = json.Unmarshal(body, &request); err != nil {
//		c.IndentedJSON(http.StatusBadRequest, err.Error())
//		c.Abort()
//		return
//	}
//}

func GetService(c *gin.Context) {
	query := `
		select id, name, description, estimated_price, estimated_time from service
	`
	var (
		rows     *sql.Rows
		services []modules.Service
		err      error
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
			&service.Id, &service.Name, &service.Description,
			&service.Price, &service.Time,
		); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		services = append(services, service)
	}

	c.IndentedJSON(http.StatusOK, services)
}
