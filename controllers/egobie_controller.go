package controllers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"fmt"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

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
	request := modules.ChangeServiceStatus{}
	var (
		data []byte
		tx   *sql.Tx
	)

	if status == "IN_PROGRESS" {
		query += ", start_timestamp = CURRENT_TIMESTAMP()"
	} else if status == "DONE" {
		query += ", end_timestamp = CURRENT_TIMESTAMP()"
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
				fmt.Println("Error - Rollback - ", err1.Error());
			}
		} else {
			if err1 := tx.Commit(); err1 != nil {
				fmt.Println("Error - Commit - ", err1.Error());
			}
		}
	}()

	if _, err = tx.Exec(
		query, status, request.ServiceId, request.UserId,
	); err != nil {
		return
	}

	if status == "DONE" {
		if err = unlockCar(tx, request.CarId, request.UserId); err != nil {
			return
		}

		if err = unlockPayment(tx, request.PaymentId, request.UserId); err != nil {
			return
		}

		if err = createHistory(tx, request.UserId, request.ServiceId); err != nil {
			return
		}
	}

	return
}
