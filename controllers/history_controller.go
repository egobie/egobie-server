package controllers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/eGobie/egobie-server/config"
	"github.com/eGobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

func GetHistory(c *gin.Context) {
	size := 6
	query := `
		select uh.id, uh.rating, us.id, us.user_payment_id,
				us.estimated_price, us.end_timestamp,
				uc.id, uc.plate, cma.title, cmo.title
		from user_history uh
		inner join user_service us on us.id = uh.user_service_id and us.status = 'DONE'
		inner join user_car uc on uc.id = us.user_car_id
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		where uh.user_id = ?
		order by uh.create_timestamp DESC
		limit ?, ?
	`
	request := modules.HistoryRequest{}
	var (
		rows      *sql.Rows
		err       error
		histories []modules.History
		body      []byte
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
		request.UserId, request.Page*size, size,
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
			&history.UserPaymentId, &history.Price, &history.EndTime,
			&history.UserCarId, &history.Plate, &history.Maker, &history.Model,
		); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		histories = append(histories, history)
	}

	c.IndentedJSON(http.StatusOK, histories)
}
