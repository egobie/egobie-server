package controllers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"fmt"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

func GetHistory(c *gin.Context) {
	size := 6
	query := `
		select uh.id, uh.rating, uh.note, us.id, us.reservation_id,
				us.user_payment_id, us.estimated_price, us.start_timestamp,
				us.end_timestamp, uc.id, uc.plate, cma.title, cmo.title,
				GROUP_CONCAT(usl.service_id) as services
		from user_history uh
		inner join user_service us on us.id = uh.user_service_id and us.status = 'DONE'
		inner join user_car uc on uc.id = us.user_car_id
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		left join user_service_list usl on usl.user_service_id = us.id
		where uh.user_id = ? and us.id is not null
		order by uh.create_timestamp DESC
		limit ?, ?
	`
	request := modules.HistoryRequest{}
	var (
		rows      *sql.Rows
		err       error
		histories []modules.History
		body      []byte
		temp string
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

	if rows, err = config.DB.Query(query,
		request.UserId, request.Page*int32(size), size,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}
	defer rows.Close()

	for rows.Next() {
		history := modules.History{}

		if err = rows.Scan(
			&history.Id, &history.Rating, &history.Note, &history.UserServiceId,
			&history.ReservationId, &history.UserPaymentId, &history.Price,
			&history.StartTime, &history.EndTime, &history.UserCarId,
			&history.Plate, &history.Maker, &history.Model, &temp,
		); err != nil {
			if strings.HasPrefix(err.Error(), "sql: Scan error on column index 0") {
				break
			} else {
				c.IndentedJSON(http.StatusBadRequest, err.Error())
				c.Abort()
				return
			}
		}

		if err = json.Unmarshal(
			[]byte("[" + temp + "]"), &history.Services,
		); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		histories = append(histories, history)
	}

	c.IndentedJSON(http.StatusOK, histories)
}

func Rating(c *gin.Context) {
	historyQuery := `
		update user_history set rating = ?, note = ?
		where id = ? and user_id = ? and user_service_id = ?
	`
	serviceQuery := `
		update user_service set status = 'DONE'
		where id = ? and user_id = ?
	`

	request := modules.RatingRequest{}
	var (
		data []byte
		err error
		tx *sql.Tx
	)

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if tx, err = config.DB.Begin(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if _, err = tx.Exec(
		historyQuery, request.Rating, request.Note,
		request.Id, request.UserId, request.ServiceId,
	); err != nil {
		if err = tx.Rollback(); err != nil {
			fmt.Println("Error - rollback - rating history - ", err.Error())
		}

		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if _, err = tx.Exec(
		serviceQuery, request.ServiceId, request.UserId,
	); err != nil {
		if err = tx.Rollback(); err != nil {
			fmt.Println("Error - rollback - done service - ", err.Error())
		}

		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("Error - commit - rating - ", err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}
