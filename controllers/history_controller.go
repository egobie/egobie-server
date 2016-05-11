package controllers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

func GetHistory(c *gin.Context) {
	size := 6
	query := `
		select uh.id, uh.rating, us.id, us.reservation_id, us.user_payment_id,
				us.estimated_price, us.start_timestamp, us.end_timestamp,
				uc.id, uc.plate, cma.title, cmo.title,
				GROUP_CONCAT(usl.service_id) as services
		from user_history uh
		inner join user_service us on us.id = uh.user_service_id and us.status = 'DONE'
		inner join user_car uc on uc.id = us.user_car_id
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		left join user_service_list usl on usl.user_service_id = us.id
		where uh.user_id = ? and us.id is not null and us.cancel = 0
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
			&history.Id, &history.Rating, &history.UserServiceId,
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
