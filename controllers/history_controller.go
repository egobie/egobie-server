package controllers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/eGobie/server/config"
	"github.com/eGobie/server/modules"

	"github.com/gin-gonic/gin"
)

func GetHistory(c *gin.Context) {
	size := 6
	query := `
		select uh.id, uh.ratting, uh.note,
				us.estimated_price, us.start_timestamp, us.end_timestamp
		from user_history uh
		inner join user_service us on us.id = uh.user_service_id
		where uh.user_id = ?
		order by uh.create_timestamp DESC
		limit ?, ?
	`
	request := modules.HistoryRequest{}
	var (
		stmt      *sql.Stmt
		rows      *sql.Rows
		err       error
		histories []modules.History
		body      []byte
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer stmt.Close()

	if rows, err = stmt.Query(request.UserId, request.Page*size, size); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		history := modules.History{}

		if err = rows.Scan(
			&history.Id, &history.Ratting, &history.Note,
			&history.Price, &history.StartTime, &history.EndTime,
		); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		histories = append(histories, history)
	}

	c.IndentedJSON(http.StatusOK, histories)
}
