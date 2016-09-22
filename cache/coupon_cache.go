package cache

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/egobie/egobie-server/config"
)

type Coupon struct {
	Id       int32
	Discount float32
	Percent  int32
}

var COUPON_CACHE map[string]Coupon

func init() {
	COUPON_CACHE = make(map[string]Coupon)

	cacheCoupon()
}

func cacheCoupon() {
	query := `
		select id, coupon, discount, percent from coupon
		where expired = 0 and applied = 0
	`
	var (
		code string
		rows *sql.Rows
		err  error
	)

	if rows, err = config.DB.Query(query); err != nil {
		fmt.Println("Error loading coupon - ", err.Error())
		os.Exit(0)
	}

	for rows.Next() {
		coupon := Coupon{}
		if err = rows.Scan(
			&coupon.Id, &code, &coupon.Discount, &coupon.Percent,
		); err != nil {
			return
		}
		COUPON_CACHE[code] = coupon
	}
}
