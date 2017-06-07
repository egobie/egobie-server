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
	request := modules.TaskRequest{}
	var (
		data []byte
		err  error
		task modules.Task
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, task)
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if task.UserTasks, err = getUserTask(request.UserId); err != nil {
		return
	}

	if task.FleetTasks, err = getFleetTask(request.UserId); err != nil {
		return
	}
}

func getUserTask(userId int32) (tasks []modules.UserTask, err error) {
	//	query := `
	//		select us.id, us.status, us.reserved_start_timestamp, u.first_name, u.middle_name,
	//				u.last_name, u.phone_number, u.home_address_state, u.home_address_zip,
	//				u.home_address_city, u.home_address_street, uc.plate, uc.state,
	//				uc.color, uc.year, cma.title, cmo.title
	//		from user_service us
	//		inner join user u on u.id = us.user_id
	//		inner join user_car uc on uc.id = us.user_car_id
	//		inner join car_maker cma on cma.id = uc.car_maker_id
	//		inner join car_model cmo on cmo.id = uc.car_model_id
	//		inner join user_service_assignee_list usal on usal.user_service_id = us.id
	//		where us.status != "CANCEL" and usal.user_id = ?
	//		order by us.reserved_start_timestamp
	//	`
	query := `
		select us.id, usal.status, us.reserved_start_timestamp, u.first_name, u.middle_name,
				u.last_name, u.phone_number, u.home_address_state, u.home_address_zip,
				u.home_address_city, u.home_address_street, uc.plate, uc.state,
				uc.color, uc.year, cma.title, cmo.title
		from user_service us
		inner join user u on u.id = us.user_id
		inner join user_car uc on uc.id = us.user_car_id
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		inner join user_service_assignee_list usal on usal.user_service_id = us.id
		where us.status != "CANCEL" and usal.user_id = ? and us.opening_id in (
			select id from opening
			where day >= DATE_FORMAT(CURDATE(), '%Y-%m-%d') and (count_wash < 1 or count_oil < 1)
		) order by us.reserved_start_timestamp
	`
	index := make(map[int32]int32)
	var (
		rows         *sql.Rows
		userServices []int32
		taskServices []modules.SimpleService
		taskAddons   []modules.SimpleAddon
	)

	if rows, err = config.DB.Query(query, userId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		task := modules.UserTask{}

		if err = rows.Scan(
			&task.Id, &task.Status, &task.Start, &task.FirstName, &task.MiddleName,
			&task.LastName, &task.Phone, &task.State, &task.Zip, &task.City,
			&task.Street, &task.Plate, &task.CarState, &task.Color, &task.Year,
			&task.Make, &task.Model,
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

	return tasks, nil
}

func getFleetTask(userId int32) (tasks []modules.FleetTask, err error) {
	//	query := `
	//		select fs.id, f.name, fs.note, fsal.status, fs.reserved_start_timestamp,
	//				u.first_name, u.last_name, u.phone_number, u.work_address_state,
	//				u.work_address_city, u.work_address_street, u.work_address_zip
	//		from fleet_service fs
	//		inner join fleet f on f.user_id = fs.user_id
	//		inner join user u on u.id = f.user_id
	//		inner join fleet_service_assignee_list fsal on fsal.fleet_service_id = fs.id
	//		where fs.status in ('RESERVED', 'IN_PROGRESS', 'DONE') and fsal.user_id = ?
	//		order by fs.reserved_start_timestamp
	//	`
	query := `
		select fs.id, f.name, fs.note, fsal.status, fs.reserved_start_timestamp,
				u.first_name, u.last_name, u.phone_number, u.work_address_state,
				u.work_address_city, u.work_address_street, u.work_address_zip
		from fleet_service fs
		inner join fleet f on f.user_id = fs.user_id
		inner join user u on u.id = f.user_id
		inner join fleet_service_assignee_list fsal on fsal.fleet_service_id = fs.id
		where fs.status in ('RESERVED', 'IN_PROGRESS', 'DONE') and fsal.user_id = ? and fs.opening_id in (
			select id from opening
			where day >= DATE_FORMAT(CURDATE(), '%Y-%m-%d') and (count_wash < 1 or count_oil < 1)
		) order by fs.reserved_start_timestamp
	`
	var (
		rows *sql.Rows
	)

	if rows, err = config.DB.Query(query, userId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		task := modules.FleetTask{}
		if err = rows.Scan(
			&task.Id, &task.FleetName, &task.Note, &task.Status,
			&task.Start, &task.FirstName, &task.LastName, &task.Phone,
			&task.State, &task.City, &task.Street, &task.Zip,
		); err != nil {
			return
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func MakeUserServiceDone(c *gin.Context) {
	if err := changeUserServiceStatus(c, "DONE"); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func MakeUserServiceReserved(c *gin.Context) {
	if err := changeUserServiceStatus(c, "RESERVED"); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func MakeUserServiceInProgress(c *gin.Context) {
	if err := changeUserServiceStatus(c, "IN_PROGRESS"); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func MakeUserServiceCancelled(c *gin.Context) {
	if err := changeUserServiceStatus(c, "CANCEL"); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func changeUserServiceStatus(c *gin.Context, status string) (err error) {
	query := `
		update user_service_assignee_list set status = ?
	`
	queryUserService := `
		update user_service set status = ?
	`
	selectQuery := `
		select status, user_id, user_car_id, user_payment_id
		from user_service
		where id = ?
	`

	request := modules.ChangeServiceStatus{}
	taskInfo := modules.TaskInfo{}

	var (
		data []byte
		tx   *sql.Tx
	)

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if tx, err = config.DB.Begin(); err != nil {
		return
	}

	args := []interface{}{
		status, request.ServiceId,
	}

	if status == "IN_PROGRESS" {
		query += ", start_timestamp = CURRENT_TIMESTAMP()"
		queryUserService += `
			, start_timestamp = CURRENT_TIMESTAMP()
			where id = ? and status = 'RESERVED'
		`
	} else if status == "DONE" {
		query += ", end_timestamp = CURRENT_TIMESTAMP()"
		queryUserService += `
			, end_timestamp = CURRENT_TIMESTAMP()
			where id = ? and status = 'IN_PROGRESS' and not exists(
				select user_id from user_service_assignee_list ul
				where ul.user_service_id = ? and ul.status != 'DONE'
			)
		`
		args = append(args, request.ServiceId)
	} else if status == "RESERVED" {
		query += ", start_timestamp = NULL, end_timestamp = NULL"
		queryUserService += `
			, start_timestamp = NULL, end_timestamp = NULL
			where id = ?
		`
	}

	query += " where user_service_id = ? and user_id = ?"

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

	if _, err = tx.Exec(
		query, status, request.ServiceId, request.UserId,
	); err != nil {
		return
	}

	if _, err = tx.Exec(queryUserService, args...); err != nil {
		return
	}

	if err = tx.QueryRow(selectQuery, request.ServiceId).Scan(
		&taskInfo.Status, &taskInfo.UserId, &taskInfo.UserCarId,
		&taskInfo.UserPaymentId,
	); err != nil {
		return
	}

	if taskInfo.Status == "DONE" {
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

		if err = createUserHistory(
			tx, taskInfo.UserId, request.ServiceId,
		); err != nil {
			return
		}

		if err = processPayment(
			tx, request.ServiceId, taskInfo.UserPaymentId, taskInfo.UserId, 1.0,
		); err != nil {
			return
		}
	}

	return
}

func MakeFleetServiceDone(c *gin.Context) {
	if err := changeFleetServiceStatus(c, "DONE"); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func MakeFleetServiceReserved(c *gin.Context) {
	if err := changeFleetServiceStatus(c, "RESERVED"); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func MakeFleetServiceInProgress(c *gin.Context) {
	if err := changeFleetServiceStatus(c, "IN_PROGRESS"); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func changeFleetServiceStatus(c *gin.Context, status string) (err error) {
	query := `
		update fleet_service set status = ?
	`

	request := modules.ChangeServiceStatus{}
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

	query += " where id = ?"

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

	if _, err = tx.Exec(query, status, request.ServiceId); err != nil {
		return
	}

	if status == "DONE" {
		err = createFleetHistory(tx, request.ServiceId)
	}

	return
}
