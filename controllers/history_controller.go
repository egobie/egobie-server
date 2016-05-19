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

func GetHistory(c *gin.Context) {
	size := 6
	query := `
		select uh.id, uh.rating, uh.note,
				uh.car_plate, uh.car_state, uh.car_maker, uh.car_model, uh.car_year, uh.car_color,
				uh.payment_holder, uh.payment_number, uh.payment_type, uh.payment_price,
				us.id, us.reservation_id, us.start_timestamp, us.end_timestamp
		from user_history uh
		inner join user_service us on us.id = uh.user_service_id and us.status = 'DONE'
		where uh.user_id = ?
		order by uh.create_timestamp DESC
		limit ?, ?
	`
	request := modules.HistoryRequest{}
	index := make(map[int32]int32)
	var (
		rows            *sql.Rows
		err             error
		histories       []modules.History
		body            []byte
		userServices    []int32
		hisotryServices []modules.SimpleService
		historyAddons   []modules.SimpleAddon
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if rows, err = config.DB.Query(query,
		request.UserId, request.Page*int32(size), size,
	); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		history := modules.History{}

		if err = rows.Scan(
			&history.Id, &history.Rating, &history.Note, &history.Plate,
			&history.State, &history.Maker, &history.Model, &history.Year,
			&history.Color, &history.AccountName, &history.AccountNumber,
			&history.AccountType, &history.Price, &history.UserServiceId,
			&history.ReservationId, &history.StartTime, &history.EndTime,
		); err != nil {
			if err == sql.ErrNoRows {
				err = nil
			}

			return
		}

		if history.AccountNumber, err = decryptAccountNumber(
			history.AccountType, history.AccountNumber,
		); err != nil {
			return
		}

		history.AccountNumber = getPaymentLastFour(history.AccountNumber)

		index[history.UserServiceId] = int32(len(histories))
		userServices = append(userServices, history.UserServiceId)
		histories = append(histories, history)
	}

	if hisotryServices, err = getSimpleService(userServices); err != nil {
		return
	}

	if historyAddons, err = getSimpleAddon(userServices); err != nil {
		return
	}

	for _, hisotryService := range hisotryServices {
		histories[index[hisotryService.UserServiceId]].Services = append(
			histories[index[hisotryService.UserServiceId]].Services, hisotryService,
		)
	}

	for _, hisotryAddon := range historyAddons {
		histories[index[hisotryAddon.UserServiceId]].Addons = append(
			histories[index[hisotryAddon.UserServiceId]].Addons, hisotryAddon,
		)
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
		err  error
		tx   *sql.Tx
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
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
		historyQuery, request.Rating, request.Note,
		request.Id, request.UserId, request.ServiceId,
	); err != nil {
		return
	}

	if _, err = tx.Exec(
		serviceQuery, request.ServiceId, request.UserId,
	); err != nil {
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func createHistory(tx *sql.Tx, userId, serviceId int32) (err error) {
	query := `
		insert into user_history (user_id, user_service_id, car_plate, car_state, car_year, car_color, car_maker, car_model, payment_holder, payment_number, payment_type, payment_price)
		select us.user_id, us.id, uc.plate, uc.state, uc.year, uc.color, cma.title, cmo.title, up.account_name, up.account_number, up.account_type, us.estimated_price
		from user_service us
		inner join user_car uc on uc.id = us.user_car_id
		inner join car_maker cma on cma.id = uc.car_maker_id
		inner join car_model cmo on cmo.id = uc.car_model_id
		inner join user_payment up on up.id = us.user_payment_id
		where us.id = ? and us.user_id = ?;
	`

	_, err = tx.Exec(query, serviceId, userId)

	return
}
