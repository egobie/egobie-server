package cache

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
)

var PLACES_MAP map[int32]modules.Place
var PLACES_ARRAY []modules.Place

func init() {
	PLACES_MAP = make(map[int32]modules.Place)

	cachePlace()
}

func cachePlace() {
	query := `
		select id, name, address, latitude, longitude from place
	`
	var (
		rows *sql.Rows
		err  error
	)

	defer func() {
		if err != nil {
			fmt.Println("Failed to load places - ", err.Error())
			os.Exit(0)
		}
	}()

	if rows, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		place := modules.Place{}
		if err = rows.Scan(
			&place.Id, &place.Name, &place.Address, &place.Latitude, &place.Longitude,
		); err != nil {
			return
		}

		PLACES_MAP[place.Id] = place
		PLACES_ARRAY = append(PLACES_ARRAY, place)
	}
}
