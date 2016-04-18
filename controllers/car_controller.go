package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/eGobie/egobie-server/config"
	"github.com/eGobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

func getCarByIdAndUserId(carId, userId int32) (car modules.Car, err error){
	query := `
		select uc.id, uc.user_id, uc.report_id, uc.plate, uc.state, uc.year, uc.color,
				cma.title, cmo.title, uc.car_maker_id, uc.car_model_id
		from user_car uc
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		where uc.id = ? and uc.user_id = ?
	`
	var (
		stmt *sql.Stmt
	)

	if stmt, err = config.DB.Prepare(query); err != nil {
		return
	}
	defer stmt.Close()

	if err = stmt.QueryRow(carId, userId).Scan(
		&car.Id, &car.UserId, &car.ReportId, &car.Plate, &car.State, &car.Year,
		&car.Color, &car.Maker, &car.Model, &car.MakerId,&car.ModelId,
	); err != nil {
		return
	}

	return car, nil
}

func getCarByUserId(userId int32) (cars []modules.Car, err error){
	query := `
		select uc.id, uc.user_id, uc.report_id, uc.plate, uc.state, uc.year, uc.color,
				cma.title, cmo.title, uc.car_maker_id, uc.car_model_id
		from user_car uc
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		where uc.user_id = ?
	`
	var (
		stmt *sql.Stmt
		rows *sql.Rows
	)

	if stmt, err = config.DB.Prepare(query); err != nil {
		return
	}
	defer stmt.Close()

	if rows, err = stmt.Query(userId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		car := modules.Car{}

		if err = rows.Scan(
			&car.Id, &car.UserId, &car.ReportId, &car.Plate, &car.State, &car.Year,
			&car.Color, &car.Maker, &car.Model, &car.MakerId, &car.ModelId,
		); err != nil {
			return
		}

		cars = append(cars, car)
	}

	return cars, nil
}

func GetCarMaker(c *gin.Context) {
	query := `
		select id, title from car_maker order by title;
	`
	var (
		err    error
		rows   *sql.Rows
		makers []modules.CarMaker
	)

	if rows, err = config.DB.Query(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		maker := modules.CarMaker{}

		if err = rows.Scan(&maker.Id, &maker.Title); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}

		makers = append(makers, maker)
	}

	c.IndentedJSON(http.StatusOK, makers)
}

func GetCarModel(c *gin.Context) {
	query := `
		select id, title from car_model
		where car_maker_id = ?
		order by title
	`
	var (
		err     error
		stmt    *sql.Stmt
		rows    *sql.Rows
		makerId int64
		models  []modules.CarModel
	)

	if makerId, err = strconv.ParseInt(c.Param("makerId"), 10, 32); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer stmt.Close()

	if rows, err = stmt.Query(int32(makerId)); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		model := modules.CarModel{}

		if err = rows.Scan(
			&model.Id, &model.Title,
		); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}

		models = append(models, model)
	}

	c.IndentedJSON(http.StatusOK, models)
}

func GetCarById(c *gin.Context) {
	request := modules.CarRequst{}
	var (
		car  modules.Car
		err  error
		body []byte
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if car, err = getCarByIdAndUserId(request.Id, request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, car)
}

func GetCarForUser(c *gin.Context) {
	request := modules.CarRequstForUser{}
	var (
		cars []modules.Car
		body []byte
		err  error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if cars, err = getCarByUserId(request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, cars)
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
		stmt        *sql.Stmt
		result      sql.Result
		affectedRow int64
		body        []byte
		err         error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer stmt.Close()

	if result, err = stmt.Exec(
		request.Plate, request.State, request.Year,
		request.Color, request.Maker, request.Model,
		request.Id, request.UserId); err != nil {
		return
	} else if affectedRow, err = result.RowsAffected(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else if affectedRow <= 0 {
		c.IndentedJSON(http.StatusBadRequest, errors.New("Car not found"))
		return
	}

	if car, err := getCarByIdAndUserId(request.Id, request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else {
		c.IndentedJSON(http.StatusOK, car)
	}
}

func CreateCar(c *gin.Context) {
	query := `
		insert into user_car (user_id, plate, state, year, color, car_maker_id, car_model_id)
		values (?, ?, ?, ?, ?, ?, ?)
	`
	request := modules.CarNew{}
	var (
		stmt   *sql.Stmt
		result sql.Result
		newId  int64
		body   []byte
		err    error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer stmt.Close()

	if result, err = stmt.Exec(
		request.UserId, request.Plate, request.State,
		request.Year, request.Color, request.Maker, request.Model,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else if newId, err = result.LastInsertId(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if car, err := getCarByIdAndUserId(int32(newId), request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else {
		c.IndentedJSON(http.StatusOK, car)
	}
}

func DeleteCar(c *gin.Context) {
	query := `
		delete from user_car where id = ? and user_id = ?
	`
	request := modules.CarRequst{}
	var (
		stmt        *sql.Stmt
		result      sql.Result
		affectedRow int64
		body        []byte
		err         error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer stmt.Close()

	if result, err = stmt.Exec(request.Id, request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else if affectedRow, err = result.RowsAffected(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else if affectedRow <= 0 {
		c.IndentedJSON(http.StatusBadRequest, "Car not found")
	}

	c.IndentedJSON(http.StatusOK, "OK")
}
