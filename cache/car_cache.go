package cache

import (
	"os"
	"database/sql"
	"fmt"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
)

var CAR_MAKES_MAP map[int32]modules.CarMaker
var CAR_MAKES_ARRAY []modules.CarMaker

var CAR_MODELS_MAP map[int32]modules.CarModel
var CAR_MODELS_ARRAY []modules.CarModel

func init() {
	CAR_MAKES_MAP = make(map[int32]modules.CarMaker)
	CAR_MODELS_MAP = make(map[int32]modules.CarModel)

	cacheCarMake()
	cacheCarModel()
}

func cacheCarMake() {
	query := `
		select id, title from car_maker order by title;
	`
	var (
		err    error
		rows   *sql.Rows
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
		maker := modules.CarMaker{}

		if err = rows.Scan(&maker.Id, &maker.Title); err != nil {
			return
		}

		CAR_MAKES_ARRAY = append(CAR_MAKES_ARRAY, maker)
		CAR_MAKES_MAP[maker.Id] = maker
	}
}

func cacheCarModel() {
	query := `
		select id, car_maker_id, title from car_model
	`
	var (
		err    error
		rows   *sql.Rows
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
			&model.Id, &model.MakerId, &model.Title,
		); err != nil {
			return
		}

		CAR_MODELS_ARRAY = append(CAR_MODELS_ARRAY, model)
		CAR_MODELS_MAP[model.Id] = model
	}
}
