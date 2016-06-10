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

func getCarByIdAndUserId(carId, userId int32) (car modules.Car, err error) {
	query := `
		select uc.id, uc.user_id, uc.report_id, uc.plate, uc.state, uc.year, uc.color,
				cma.title, cmo.title, uc.car_maker_id, uc.car_model_id, uc.reserved
		from user_car uc
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		where uc.id = ? and uc.user_id = ?
	`

	if err = config.DB.QueryRow(query, carId, userId).Scan(
		&car.Id, &car.UserId, &car.ReportId, &car.Plate,
		&car.State, &car.Year, &car.Color, &car.Maker,
		&car.Model, &car.MakerId, &car.ModelId, &car.Reserved,
	); err != nil {
		return
	}

	return car, nil
}

func getCarByUserId(userId int32) (cars []modules.Car, err error) {
	query := `
		select uc.id, uc.user_id, uc.report_id, uc.plate, uc.state, uc.year, uc.color,
				cma.title, cmo.title, uc.car_maker_id, uc.car_model_id, uc.reserved
		from user_car uc
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		where uc.user_id = ?
	`
	var (
		rows *sql.Rows
	)

	if rows, err = config.DB.Query(query, userId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		car := modules.Car{}

		if err = rows.Scan(
			&car.Id, &car.UserId, &car.ReportId, &car.Plate,
			&car.State, &car.Year, &car.Color, &car.Maker,
			&car.Model, &car.MakerId, &car.ModelId, &car.Reserved,
		); err != nil {
			return
		}

		cars = append(cars, car)
	}

	return cars, nil
}

func GetCarMake(c *gin.Context) {
	c.JSON(http.StatusOK, cache.CAR_MAKES_ARRAY)
}

func GetCarModel(c *gin.Context) {
	c.JSON(http.StatusOK, cache.CAR_MODELS_ARRAY)
}

func GetCarById(c *gin.Context) {
	request := modules.CarRequst{}
	var (
		car  modules.Car
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

	if car, err = getCarByIdAndUserId(request.Id, request.UserId); err != nil {
		return
	}

	c.JSON(http.StatusOK, car)
}

func GetCarForUser(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		cars []modules.Car
		body []byte
		err  error
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

	if cars, err = getCarByUserId(request.UserId); err != nil {
		return
	}

	c.JSON(http.StatusOK, cars)
}

func UpdateCar(c *gin.Context) {
	query := `
		update user_car
		set plate = ?, state = ?, year = ?,
			color = ?, car_maker_id = ?, car_model_id = ?
		where id = ? and user_id = ?
	`
	request := modules.UpdateCar{}
	var (
		result      sql.Result
		affectedRow int64
		body        []byte
		err         error
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

	if result, err = config.DB.Exec(query,
		request.Plate, request.State, request.Year,
		request.Color, request.Maker, request.Model,
		request.Id, request.UserId); err != nil {
		return
	} else if affectedRow, err = result.RowsAffected(); err != nil {
		return
	} else if affectedRow <= 0 {
		err = errors.New("Car not found")
		return
	}

	go editCar(request.UserId)

	if car, err := getCarByIdAndUserId(
		request.Id, request.UserId,
	); err == nil {
		c.JSON(http.StatusOK, car)
	}
}

func CreateCar(c *gin.Context) {
	query := `
		insert into user_car (user_id, plate, state, year, color, car_maker_id, car_model_id)
		values (?, ?, ?, ?, ?, ?, ?)
	`
	request := modules.CarNew{}
	var (
		result sql.Result
		newId  int64
		body   []byte
		err    error
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

	if result, err = config.DB.Exec(query,
		request.UserId, request.Plate, request.State,
		request.Year, request.Color, request.Maker, request.Model,
	); err != nil {
		return
	} else if newId, err = result.LastInsertId(); err != nil {
		return
	}

	go addCar(request.UserId)

	if car, err := getCarByIdAndUserId(
		int32(newId), request.UserId,
	); err == nil {
		c.JSON(http.StatusOK, car)
	}
}

func DeleteCar(c *gin.Context) {
	query := `
		delete from user_car where id = ? and user_id = ?
	`
	request := modules.CarRequst{}
	var (
		result      sql.Result
		affectedRow int64
		body        []byte
		err         error
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

	if checkCarStatus(request.Id, request.UserId) {
		err = errors.New("This vehicle cannot be deleted since you have one reservation on it.")
		return
	}

	if result, err = config.DB.Exec(
		query, request.Id, request.UserId,
	); err != nil {
		return
	} else if affectedRow, err = result.RowsAffected(); err != nil {
		return
	} else if affectedRow <= 0 {
		err = errors.New("Car not found")
		return
	}

	go deleteCar(request.UserId)

	c.JSON(http.StatusOK, "OK")
}

func checkCarStatus(id, userId int32) bool {
	query := `
		select reserved from user_car where id = ? and user_id = ?
	`
	var temp int32

	if err := config.DB.QueryRow(
		query, id, userId,
	).Scan(&temp); err != nil {
		fmt.Println("Check Car Status - Error - ", err)
		return true
	}

	fmt.Println("temp = ", temp)

	return temp > 0
}

func lockCar(tx *sql.Tx, id, userId int32) (err error) {
	query := `
		update user_car set reserved = reserved + 1 where id = ? and user_id = ?
	`

	if _, err = config.DB.Exec(
		query, id, userId,
	); err != nil {
		fmt.Println("Lock Car - Error - ", err.Error())
	}

	return
}

func unlockCar(tx *sql.Tx, id, userId int32) (err error) {
	query := `
		update user_car set reserved = reserved - 1 where id = ? and user_id = ?
	`

	if _, err = tx.Exec(
		query, id, userId,
	); err != nil {
		fmt.Println("Unlock Car - Error - ", err.Error())
	}

	return
}
