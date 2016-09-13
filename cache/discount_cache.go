package cache

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/egobie/egobie-server/config"
)

var DISCOUNT_MAP map[string]int32

func init() {
	DISCOUNT_MAP = make(map[string]int32)

	cacheDiscount()
}

func cacheDiscount() {
	query := `
		select type, discount from discount
	`
	var (
		rows *sql.Rows
		err  error
		t    string
		d    int32
	)

	defer func() {
		if err != nil {
			fmt.Println("Error loading discount - ", err.Error())
			os.Exit(0)
		}
	}()

	if rows, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&t, &d); err != nil {
			return
		}

		DISCOUNT_MAP[t] = d
	}
}
