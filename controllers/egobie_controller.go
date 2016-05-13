package controllers

import (
	"io/ioutil"
	"net/http"
	"encoding/json"

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

func changeServiceStatus(c *gin.Context, status string) (error) {
	query := `
		update user_service set status = ?
	`
	request := modules.ChangeServiceStatus{}
	var (
		data []byte
		err error
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
		return err
	}

	if err = json.Unmarshal(data, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return err
	}

	if _, err = config.DB.Exec(
		query, status, request.ServiceId, request.UserId,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return err
	}

	if (status == "DONE") {
		unlockCar(request.CarId, request.UserId)
		unlockPayment(request.PaymentId, request.UserId)
	}

	return nil
}
