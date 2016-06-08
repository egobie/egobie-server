package controllers

import (
	"database/sql"
	"encoding/json"
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
			middle_name, email, phone_number)
		values ('FLEET', ?, ?, ?, ?, ?, ?, ?)
	`
	queryFLeet := `
		insert into fleet (name, user_id)
		values (?, ?)
	`
	selectFleet := `
		select f.id, f.name, f.token, u.id, u.first_name,
				u.last_name, u.middle_name, u.email, u.phone_number
		from fleet f
		inner join user u on u.id = f.user_id
		where f.id = ?
	`
	request := modules.NewFLeetUserRequest{}
	var (
		tx         *sql.Tx
		data       []byte
		err        error
		result     sql.Result
		id         int64
		username   string
		enPassword string
		user       modules.FleetUser
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, user)
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
			err = tx.Commit()
		}
	}()

	if result, err = tx.Exec(
		queryUser, username, enPassword, request.FirstName,
		request.LastName, request.MiddleName, request.Email,
		request.Phone,
	); err != nil {
		return
	} else if id, err = result.LastInsertId(); err != nil {
		return
	}

	if result, err = tx.Exec(
		queryFLeet, request.FleetName, id,
	); err != nil {
		return
	} else if id, err = result.LastInsertId(); err != nil {
		return
	}

	if err = tx.QueryRow(selectFleet, id).Scan(
		&user.FleetId, &user.FleetName, &user.Token,
		&user.UserId, &user.FirstName, &user.LastName,
		&user.MiddleName, &user.Email, &user.Phone,
	); err != nil {
		return
	}
}
