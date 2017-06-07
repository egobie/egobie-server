package cache

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
)

var CAR_MAKES_MAP map[int32]modules.CarMake
var CAR_MAKES_ARRAY []modules.CarMake

var CAR_MODELS_MAP map[int32]modules.CarModel
var CAR_MODELS_ARRAY []modules.CarModel

func init() {
	CAR_MAKES_MAP = make(map[int32]modules.CarMake)
	CAR_MODELS_MAP = make(map[int32]modules.CarModel)

	cacheCarMake()
	cacheCarModel()
}

func cacheCarMake() {
	query := `
		select id, title from car_maker order by title;
	`
	var (
		err  error
		rows *sql.Rows
	)

	defer func() {
		if err != nil {
			fmt.Println("Failed to load car make - ", err.Error())
			os.Exit(0)
		}
	}()

	if rows, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		make := modules.CarMake{}

		if err = rows.Scan(&make.Id, &make.Title); err != nil {
			return
		}

		CAR_MAKES_ARRAY = append(CAR_MAKES_ARRAY, make)
		CAR_MAKES_MAP[make.Id] = make
	}
}

func cacheCarModel() {
	query := `
		select id, car_maker_id, title from car_model
	`
	var (
		err  error
		rows *sql.Rows
	)

	defer func() {
		if err != nil {
			fmt.Println("Failed to load car model - ", err.Error())
			os.Exit(0)
		}
	}()

	if rows, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		model := modules.CarModel{}

		if err = rows.Scan(
			&model.Id, &model.MakeId, &model.Title,
		); err != nil {
			return
		}

		CAR_MODELS_ARRAY = append(CAR_MODELS_ARRAY, model)
		CAR_MODELS_MAP[model.Id] = model
	}
}
