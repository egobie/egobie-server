package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
	"github.com/egobie/egobie-server/secures"

	"github.com/gin-gonic/gin"
)

func NewFleetUser(c *gin.Context) {
	queryUser := `
		insert into user (type, username, password, first_name, last_name,
			middle_name, email, phone_number, work_address_street,
			work_address_city, work_address_state, work_address_zip)
		values ('FLEET', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	queryFLeet := `
		insert into fleet (name, user_id, sale_user_id)
		values (?, ?, ?)
	`
	request := modules.NewFLeetUser{}
	var (
		tx         *sql.Tx
		data       []byte
		err        error
		result     sql.Result
		userId     int64
		username   string
		enPassword string
		info       modules.FleetUserInfo
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		go sendNewFleetUserEmail(
			request.Email, request.FirstName, info.Token,
		)

		info.Token = ""

		c.JSON(http.StatusOK, info)
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

	username = "fleet-" + secures.RandString(8)

	if enPassword, err = secures.EncryptPassword(username); err != nil {
		return
	}

	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				fmt.Println("Error - Rollback - ", err1.Error())
			}
		} else {
			if err = tx.Commit(); err == nil {
				info, err = getFleetUserInfoByUserId(int32(userId))
			}
		}
	}()

	if result, err = tx.Exec(
		queryUser, username, enPassword, request.FirstName, request.LastName,
		request.MiddleName, request.Email, request.Phone, request.Street,
		request.City, request.State, request.Zip,
	); err != nil {
		return
	} else if userId, err = result.LastInsertId(); err != nil {
		return
	}

	if result, err = tx.Exec(
		queryFLeet, request.FleetName, userId, request.UserId,
	); err != nil {
		return
	}
}

func ResendEmail(c *gin.Context) {
	request := modules.SendEmailRequest{}
	temp := struct {
		Setup int32
		Name  string
		Token string
		Email string
	}{}
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

	query := `
		select f.setup, f.token, u.first_name, u.email
		from fleet f
		inner join user u on u.id = f.user_id
		where u.id = ?
	`

	if err = config.DB.QueryRow(query, request.FleetUserId).Scan(
		&temp.Setup, &temp.Token, &temp.Name, &temp.Email,
	); err != nil {
		return
	}

	if (temp.Setup == 1) {
		err = errors.New("Fleet user had been activated");
		return
	}

	go sendNewFleetUserEmail(
		temp.Email, temp.Name, temp.Token,
	)
}

func AllFleetUser(c *gin.Context) {
	request := modules.GetFleetUserRequest{}
	var (
		data []byte
		err  error
		all  modules.AllFleetUser
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, all)
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	all, err = getFleetUsersBySaleUserId(
		request.UserId, request.Page,
	)
}

func AllFleetOrder(c *gin.Context) {
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

	if fleetServices, err := getFleetServiceBySaleUser(
		request.UserId, "status = 'WAITING' or status = 'NOT_ASSIGNED'",
	); err == nil {
		c.JSON(http.StatusOK, fleetServices)
	}
}

func PromotePrice(c *gin.Context) {
	request := modules.PriceRequest{}
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

	query := `
		update fleet_service set estimated_price = ?, status= CASE
				WHEN status = "WAITING" THEN "RESERVED"
				ELSE "NOT_ASSIGNED"
			END
		where id = ? and (status = "WAITING" or status = "NOT_ASSIGNED")
	`

	_, err = config.DB.Exec(query, request.Price, request.Id)
}
