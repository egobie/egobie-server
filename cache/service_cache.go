package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
)

var SERVICES_MAP map[int32]modules.Service
var SERVICES_ARRAY []modules.Service

var ADDONS_MAP map[int32]modules.AddOn

var FLEET_ADDONS_MAP map[int32]modules.AddOn
var FLEET_ADDONS_ARRAY []modules.AddOn

func init() {
	SERVICES_MAP = make(map[int32]modules.Service)
	ADDONS_MAP = make(map[int32]modules.AddOn)
	FLEET_ADDONS_MAP = make(map[int32]modules.AddOn)

	cacheService()
}

func cacheService() {
	query := `
		select id, name, type, items, description, note,
			estimated_price, estimated_time
		from service
		where type != 'DETAILING'
		order by id
	`
	index := make(map[int32]int)
	var (
		rows1 *sql.Rows
		rows2 *sql.Rows
		err   error
		temp  string
	)

	defer func() {
		if err != nil {
			fmt.Println("Failed to load services - ", err.Error())
			os.Exit(1)
		}
	}()

	if rows1, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows1.Close()

	for rows1.Next() {
		service := modules.Service{}
		if err = rows1.Scan(
			&service.Id, &service.Name, &service.Type, &temp,
			&service.Description, &service.Note, &service.Price, &service.Time,
		); err != nil {
			return
		}

		if err = json.Unmarshal([]byte(temp), &service.Items); err != nil {
			return
		}

		index[service.Id] = len(SERVICES_ARRAY)
		SERVICES_ARRAY = append(SERVICES_ARRAY, service)
	}

	if rows2, err = config.DB.Query(`
		select id, service_id, name, note,
			price, time, max, unit
		from service_addon
		where name != 'Engine Cleaning' and name != 'Headlight Reconditioning'
	`); err != nil {
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		addOn := modules.AddOn{}
		if err = rows2.Scan(
			&addOn.Id, &addOn.ServiceId, &addOn.Name, &addOn.Note,
			&addOn.Price, &addOn.Time, &addOn.Max, &addOn.Unit,
		); err != nil {
			return
		}

		if invalidAddon(&addOn) {
			continue
		}

		ADDONS_MAP[addOn.Id] = addOn

		if i, ok := index[addOn.ServiceId]; ok {
			addOn.Amount = 1

			if addOn.Price == 0 {
				SERVICES_ARRAY[i].Free = append(
					SERVICES_ARRAY[i].Free, addOn,
				)
			} else if addOn.Time == 0 {
				SERVICES_ARRAY[i].Charge = append(
					SERVICES_ARRAY[i].Charge, addOn,
				)
			} else {
				SERVICES_ARRAY[i].Addons = append(
					SERVICES_ARRAY[i].Addons, addOn,
				)
			}
		} else if addOn.ServiceId == 0 {
			// Addon for fleet only
			FLEET_ADDONS_MAP[addOn.Id] = addOn
			FLEET_ADDONS_ARRAY = append(FLEET_ADDONS_ARRAY, addOn)
		}
	}

	for _, service := range SERVICES_ARRAY {
		SERVICES_MAP[service.Id] = service
	}
}

func invalidAddon(addOn *modules.AddOn) bool {
	if addOn.Name == "Engine Cleaning" || addOn.Name == "Headlight Reconditioning" {
		return true
	}

	if addOn.Price > 0 {
		if addOn.Name == "Hand Wax" && (addOn.ServiceId == 2 || addOn.ServiceId == 5) {
			return true
		}

		if addOn.Name == "Paint Protectant" && (addOn.ServiceId == 3 || addOn.ServiceId == 6) {
			return true
		}
	}

	return false
}
