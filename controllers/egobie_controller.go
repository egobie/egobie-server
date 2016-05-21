package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

func GetTask(c *gin.Context) {
	query := `
		select us.id, us.status, us.reserved_start_timestamp, u.first_name, u.middle_name,
				u.last_name, u.phone_number, u.home_address_state, u.home_address_zip,
				u.home_address_city, u.home_address_street, uc.plate, uc.state,
				uc.color, cma.title, cmo.title
		from user_service us
		inner join user u on u.id = us.user_id
		inner join user_car uc on uc.id = us.user_car_id
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		where us.status != "CANCEL" and us.assignee = ? and us.opening_id in (
			select id from opening
			where day = DATE_FORMAT(CURDATE(), '%Y-%m-%d') and count < 2
		) order by us.reserved_start_timestamp
	`

	request := modules.TaskRequest{}
	index := make(map[int32]int32)
	var (
		rows1        *sql.Rows
		data         []byte
		err          error
		userServices []int32
		tasks        []modules.Task
		taskServices []modules.SimpleService
		taskAddons   []modules.SimpleAddon
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if rows1, err = config.DB.Query(query, request.UserId); err != nil {
		return
	}

	for rows1.Next() {
		task := modules.Task{}

		if err = rows1.Scan(
			&task.Id, &task.Status, &task.Start, &task.FirstName, &task.MiddleName,
			&task.LastName, &task.Phone, &task.State, &task.Zip, &task.City,
			&task.Street, &task.Plate, &task.CarState, &task.Color, &task.Maker,
			&task.Model,
		); err != nil {
			return
		}

		index[task.Id] = int32(len(tasks))
		tasks = append(tasks, task)
		userServices = append(userServices, task.Id)
	}

	if taskServices, err = getSimpleService(userServices); err != nil {
		return
	}

	if taskAddons, err = getSimpleAddon(userServices); err != nil {
		return
	}

	for _, taskService := range taskServices {
		tasks[index[taskService.UserServiceId]].Services = append(
			tasks[index[taskService.UserServiceId]].Services, taskService,
		)
	}

	for _, taskAddon := range taskAddons {
		tasks[index[taskAddon.UserServiceId]].Addons = append(
			tasks[index[taskAddon.UserServiceId]].Addons, taskAddon,
		)
	}

	c.IndentedJSON(http.StatusOK, tasks)
}

func MakeServiceDone(c *gin.Context) {
	if err := changeServiceStatus(c, "DONE"); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func MakeServiceReserved(c *gin.Context) {
	if err := changeServiceStatus(c, "RESERVED"); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func MakeServiceInProgress(c *gin.Context) {
	if err := changeServiceStatus(c, "IN_PROGRESS"); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func changeServiceStatus(c *gin.Context, status string) (err error) {
	query := `
		update user_service set status = ?
	`
	selectQuery := `
		select user_id, user_car_id, user_payment_id
		from user_service
		where id = ?
	`

	request := modules.ChangeServiceStatus{}
	taskInfo := modules.TaskInfo{}
	var (
		data []byte
		tx   *sql.Tx
	)

	if status == "IN_PROGRESS" {
		query += `
			, start_timestamp = CURRENT_TIMESTAMP()
		`
	} else if status == "DONE" {
		query += `
			, end_timestamp = CURRENT_TIMESTAMP()
		`
	} else if status == "RESERVED" {
		query += ", start_timestamp = NULL, end_timestamp = NULL"
	}

	query += " where id = ? and user_id = ?"

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
			if err1 := tx.Commit(); err1 != nil {
				fmt.Println("Error - Commit - ", err1.Error())
			}
		}
	}()

	if err = tx.QueryRow(selectQuery, request.ServiceId).Scan(
		&taskInfo.UserId, &taskInfo.UserCarId, &taskInfo.UserPaymentId,
	); err != nil {
		return
	}

	if _, err = tx.Exec(
		query, status, request.ServiceId, taskInfo.UserId,
	); err != nil {
		return
	}

	if status == "DONE" {
		if err = unlockCar(
			tx, taskInfo.UserCarId, taskInfo.UserId,
		); err != nil {
			return
		}

		if err = unlockPayment(
			tx, taskInfo.UserPaymentId, taskInfo.UserId,
		); err != nil {
			return
		}

		if err = createHistory(
			tx, taskInfo.UserId, request.ServiceId,
		); err != nil {
			return
		}

		if err = processPayment(
			tx, request.ServiceId, taskInfo.UserPaymentId, taskInfo.UserId,
		); err != nil {
			return
		}
	}

	return
}
