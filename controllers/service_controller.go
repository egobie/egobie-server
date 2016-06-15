package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/egobie/egobie-server/cache"
	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
	"github.com/egobie/egobie-server/utils"

	"github.com/gin-gonic/gin"
)

var OPENING_GAP = 0.5
var OPENING_BASE = 8.0
var SERVICE_TYPES = map[string]string{
	"CAR_WASH":   "Car Wash",
	"OIL_CHANGE": "Oil & Filter",
	"DETAILING":  "Detailing",
}

func getUserService(userId int32, condition string) (
	userServices []modules.UserService, err error,
) {
	query := `
		select us.id, us.reservation_id, us.user_car_id, uc.plate,
				us.user_payment_id, us.estimated_time, us.estimated_price,
				us.reserved_start_timestamp,
				TIMESTAMPDIFF(MINUTE, CURRENT_TIMESTAMP(), us.reserved_start_timestamp) as mins,
				us.start_timestamp, us.end_timestamp,
				us.note, us.status, us.create_timestamp,
				s.id, s.name, s.type, s.note, s.description,
				s.estimated_time, s.estimated_price
		from user_service us
		inner join user_car uc on uc.id = us.user_car_id
		inner join user_service_list usl on usl.user_service_id = us.id
		inner join service s on s.id = usl.service_id
		where us.user_id = ? and (
	` + condition + ") order by us.id"

	index := make(map[int32]int32)
	var (
		rows1  *sql.Rows
		rows2  *sql.Rows
		tempId int32
		mins   int32
		ids    []int32
	)

	if rows1, err = config.DB.Query(query, userId); err != nil {
		return
	}
	defer rows1.Close()

	for rows1.Next() {
		userService := modules.UserService{}
		service := modules.Service{}

		if err = rows1.Scan(
			&userService.Id, &userService.ReservationId, &userService.CarId,
			&userService.CarPlate, &userService.PaymentId, &userService.Time,
			&userService.Price, &userService.ReserveStartTime, &mins,
			&userService.StartTime, &userService.EndTime, &userService.Note,
			&userService.Status, &userService.ReserveTime, &service.Id,
			&service.Name, &service.Type, &service.Note, &service.Description,
			&service.Time, &service.Price,
		); err != nil {
			return
		}

		if len(userServices) != 0 && userServices[len(userServices)-1].Id == userService.Id {
			userServices[len(userServices)-1].ServiceList = append(
				userServices[len(userServices)-1].ServiceList, service,
			)
		} else {
			userService.HowLong, userService.Unit = calculateHowLong(mins)
			userService.ServiceList = append(userService.ServiceList, service)
			userServices = append(userServices, userService)

			ids = append(ids, userService.Id)
			index[userService.Id] = int32(len(userServices) - 1)
		}
	}

	query = `
		select sa.id, sa.service_id, sa.name, sa.note,
				sa.price, sa.time, sa.max, sa.unit,
				usal.amount, usal.user_service_id
		from service_addon sa
		inner join user_service_addon_list usal on usal.service_addon_id = sa.id
		where usal.user_service_id in (
	` + utils.ToStringList(ids) + ")"

	if rows2, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		addon := modules.AddOn{}

		if err = rows2.Scan(
			&addon.Id, &addon.ServiceId, &addon.Name, &addon.Note,
			&addon.Price, &addon.Time, &addon.Max, &addon.Unit,
			&addon.Amount, &tempId,
		); err != nil {
			return
		}

		userServices[index[tempId]].AddonList = append(
			userServices[index[tempId]].AddonList, addon,
		)
	}

	return userServices, nil
}

func GetReservation(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		err  error
		body []byte
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if userServices, err := getUserService(
		request.UserId, "status = 'RESERVED' or status = 'IN_PROGRESS'",
	); err == nil {
		c.JSON(http.StatusOK, userServices)
	}
}

func GetUserServiceReserved(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		err  error
		body []byte
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if userServices, err := getUserService(
		request.UserId, "status = 'RESERVED'",
	); err == nil {
		c.JSON(http.StatusOK, userServices)
	}
}

func GetUserServiceDone(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		err  error
		body []byte
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if userServices, err := getUserService(
		request.UserId, "status = 'DONE'",
	); err == nil {
		c.JSON(http.StatusOK, userServices)
	}
}

func OnDemand(c *gin.Context) {
	query := `
		select id, day, period
		from opening
		where day = DATE_FORMAT(CURDATE(), '%Y-%m-%d') and period >= ?
	`
	request := modules.OnDemandRequest{}
	curr := getCurrentPeriod()

	if curr < 0 {
		c.JSON(http.StatusBadRequest, "NO")
		c.Abort()
		return
	}

	var (
		body      []byte
		err       error
		addons    []int32
		totalTime int32
		types     string
		openings  []modules.Opening
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if len(request.Services) == 0 {
		err = errors.New("Please provide services")
		return
	}

	go checkAvailability(request.UserId)

	totalTime, _, types = getTotalTimeAndPriceAndTypes(
		request.Services, addons,
	)

	if openings, err = loadOpening(
		query, types, curr,
	); err != nil {
		return
	}

	if openings, err = filterOpening(
		calculateGap(totalTime), openings,
	); err == nil {
		if len(openings) > 0 {
			str := ""
			temp := struct {
				Id    int32   `json:"id"`
				Day   string  `json:"day"`
				Start float64 `json:"start"`
				End   float64 `json:"end"`
				Diff  int32   `json:"diff"`
			}{}

			temp.Id = openings[0].Range[0].Id
			temp.Day = openings[0].Day
			temp.Start = openings[0].Range[0].Start
			temp.End = openings[0].Range[0].End

			if int32(temp.Start*30)%30 == 0 {
				str = temp.Day + "T" + strconv.Itoa(int(temp.Start)) + ":00:00.000Z"
			} else {
				str = temp.Day + "T" + strconv.Itoa(int(temp.Start-0.5)) + ":30:00.000Z"
			}

			temp.Diff = timeDiffInMins(str)

			c.JSON(http.StatusOK, temp)
		} else {
			c.JSON(http.StatusBadRequest, "NO")
			c.Abort()
		}
	}
}

func getCurrentPeriod() int32 {
	newYork, _ := time.LoadLocation("America/New_York")

	//	t, _ := time.Parse("2006-01-02T15:04:05.000Z", "2016-05-16T10:21:26.371Z")
	t := time.Now().In(newYork)

	now := t.Add(30 * time.Minute)
	hour := now.Hour()

	if hour < int(OPENING_BASE) || hour > 19 {
		return -1
	}

	period := int32((float64(hour)-OPENING_BASE)/OPENING_GAP) + 2

	if now.Minute() > 30 {
		period += 1
	}

	return period
}

func timeDiffInMins(str string) int32 {
	newYork, _ := time.LoadLocation("America/New_York")

	//	now, _ := time.Parse("2006-01-02T15:04:05.000Z", "2016-05-16T10:21:26.371Z")
	now := time.Now().In(newYork)

	if t, err := time.ParseInLocation(
		"2006-01-02T15:04:05.000Z", str, newYork,
	); err != nil {
		return -1
	} else {
		return int32(t.Sub(now).Minutes())
	}
}

func GetOpening(c *gin.Context) {
	query := `
		select id, day, period
		from opening
		where day > DATE_FORMAT(CURDATE(), '%Y-%m-%d')
	`
	request := modules.OpeningRequest{}
	var (
		totalTime int32
		body      []byte
		err       error
		types     string
		openings  []modules.Opening
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, openings)
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if len(request.Services) == 0 {
		err = errors.New("Please provide services")
		return
	}

	totalTime, _, types = getTotalTimeAndPriceAndTypes(
		request.Services, request.Addons,
	)

	if openings, err = loadOpening(query, types); err != nil {
		return
	}

	openings, err = filterOpening(
		calculateGap(totalTime), openings,
	)
}

func loadOpening(query, types string, args ...interface{}) (
	openings []modules.Opening, err error,
) {
	var (
		rows   *sql.Rows
		preDay string
	)

	if types == "CAR_WASH" {
		query += " and count_wash > 0 order by day, period"
	} else {
		query += " and count_oil > 0 order by day, period"
	}

	if rows, err = config.DB.Query(query, args...); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		temp := struct {
			Id     int32
			Day    string
			Period int32
		}{}

		if err = rows.Scan(&temp.Id, &temp.Day, &temp.Period); err != nil {
			return
		} else {
			if preDay != temp.Day {
				preDay = temp.Day
				openings = append(openings, modules.Opening{})
				openings[len(openings)-1].Day = temp.Day
			}

			openings[len(openings)-1].Range = append(
				openings[len(openings)-1].Range,
				modules.Period{
					temp.Id,
					OPENING_BASE + float64(temp.Period-1)*OPENING_GAP,
					OPENING_BASE + float64(temp.Period)*OPENING_GAP,
				},
			)
		}
	}

	return openings, nil
}

func filterOpening(gap int32, openings []modules.Opening) (
	result []modules.Opening, err error,
) {
	var (
		p1  int32
		p2  int32
		pre int32
	)

	fmt.Println("Filter by gap - ", gap)

	for _, opening := range openings {
		o := modules.Opening{}
		o.Day = opening.Day

		p1 = 0
		p2 = 0
		pre = opening.Range[p1].Id
		size := int32(len(opening.Range))

		if p1 < (size - gap + 1) {
			for p2 < size {
				if opening.Range[p2].Id-pre > 1 {
					p1 = p2
					pre = opening.Range[p1].Id
				} else {
					if opening.Range[p2].Id-opening.Range[p1].Id+1 == gap {
						o.Range = append(o.Range, opening.Range[p1])
						p1 += 1
					}

					pre = opening.Range[p2].Id
				}

				p2 += 1
			}
		}

		if len(o.Range) != 0 {
			result = append(result, o)
		}
	}

	return result, nil
}

func PlaceOrder(c *gin.Context) {
	request := modules.OrderRequest{}
	car := modules.Car{}
	payment := modules.Payment{}
	user := modules.User{}

	var (
		result            sql.Result
		tx                *sql.Tx
		body              []byte
		err               error
		price             float32
		time              int32
		gap               int32
		insertedId        int64
		reserved          string
		assignee          int32
		reservationNumber string
		types             string
		services          []string
		addons            []string
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	updateOpeningDemand(request.Opening)

	if user, err = getUserById(request.UserId); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("User not found")
		}
		return
	}

	if payment, err = getPaymentByIdAndUserId(
		request.PaymentId, request.UserId,
	); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("User not found")
		}
		return
	}

	if car, err = getCarByIdAndUserId(
		request.CarId, request.UserId,
	); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("Car not found")
		}
		return
	}

	if t, p, ty, err := getServicesTimeAndPriceAndTypes(
		request.Services,
	); err != nil {
		return
	} else {
		time += t
		price += p
		types = ty
	}

	t, p := getAddonsTimeAndPrice(request.Addons)

	time += t
	price += p

	services, addons = getServicesAndAddons(
		request.Services, request.Addons,
	)

	if user.Discount > 0 {
		price = price * 1.07 * 0.9
	} else {
		price = price * 1.07
	}

	price = float32(int(price*100)) / 100
	gap = calculateGap(time)

	if tx, err = config.DB.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				fmt.Println("Error -Rollback - ", err1.Error())
			}

			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			err = nil
		} else {
			if err1 := tx.Commit(); err1 != nil {
				fmt.Println("Error - Commit - ", err1.Error())
			} else {
				// Send Email To User
				go sendPlaceOrderEmail(
					user.Email.String,
					user.FirstName.String,
					reservationNumber,
					reserved,
					services,
					addons,
					price,
				)

				c.JSON(http.StatusOK, "OK")
			}
		}
	}()

	if err = useDiscount(tx, user.Id); err != nil {
		return
	}

	if err = holdOpening(
		tx, request.Opening, request.Opening+gap, types,
	); err != nil {
		return
	}

	if assignee, err = assignService(
		tx, request.Opening, gap, types,
	); err != nil {
		return
	}

	if reserved, err = calculateReservedTime(tx, request.Opening); err != nil {
		return
	}

	insertUserService := `
		insert into user_service (
			user_id, user_car_id, user_payment_id, opening_id,
			reserved_start_timestamp, gap, assignee,
			estimated_time, estimated_price, status, types
		) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	if result, err = tx.Exec(insertUserService,
		user.Id, car.Id, payment.Id, request.Opening, reserved,
		gap, assignee, time, price, "RESERVED", types,
	); err != nil {
		return
	}

	if insertedId, err = result.LastInsertId(); err != nil {
		return
	}

	user_service_id := int32(insertedId)

	if err = tx.QueryRow(`
		select reservation_id from user_service where id = ?`, user_service_id,
	).Scan(&reservationNumber); err != nil {
		return
	}

	queryUserServiceList := `
		insert into user_service_list (
			service_id, user_service_id
		) values (?, ?)
	`

	for _, id := range request.Services {
		if _, err = tx.Exec(
			queryUserServiceList, id, user_service_id,
		); err != nil {
			return
		}
	}

	queryUserServiceAddonList := `
		insert into user_service_addon_list (
			service_addon_id, user_service_id, amount
		) values (?, ?, ?)
	`

	for _, addon := range request.Addons {
		if _, err = tx.Exec(
			queryUserServiceAddonList, addon.Id, user_service_id, addon.Amount,
		); err != nil {
			return
		}
	}

	go makeReservation(request.UserId)

	if err = lockCar(tx, request.CarId, request.UserId); err != nil {
		return
	}

	if err = lockPayment(tx, request.PaymentId, request.UserId); err != nil {
		return
	}
}

func getServicesAndAddons(serviceIds []int32, addonRequests []modules.AddonRequest) (
	services, addons []string,
) {
	if len(serviceIds) > 0 {
		for _, id := range serviceIds {
			if s, ok := cache.SERVICES_MAP[id]; ok {
				services = append(
					services, s.Name+" ("+SERVICE_TYPES[s.Type]+")",
				)
			}
		}
	}

	if len(addonRequests) > 0 {
		for _, addon := range addonRequests {
			if a, ok := cache.ADDONS_MAP[addon.Id]; ok {
				addons = append(addons, a.Name)
			}
		}
	}

	return
}

func getServicesTimeAndPriceAndTypes(ids []int32) (
	time int32, price float32, types string, err error,
) {
	temp := make(map[string]int32)

	var (
		wash bool
		oil  bool
	)

	for _, id := range ids {
		if s, ok := cache.SERVICES_MAP[id]; ok {
			time += s.Time
			price += s.Price

			if s.Type == "CAR_WASH" {
				wash = true
			} else if s.Type == "OIL_CHANGE" {
				oil = true
			}

			if _, ok = temp[s.Type]; ok {
				err = errors.New("Can only choose one service for each type")
				return
			} else {
				temp[s.Type] = 0
			}
		}
	}

	types = calculateOrderTypes(wash, oil)

	return
}

func getAddonsTimeAndPrice(addons []modules.AddonRequest) (
	time int32, price float32,
) {
	if len(addons) > 0 {
		for _, addon := range addons {
			if a, ok := cache.ADDONS_MAP[addon.Id]; ok {
				time += a.Time
				price += (a.Price * float32(addon.Amount))
			}
		}
	}

	return
}

func getTotalTimeAndPriceAndTypes(services, addons []int32) (time int32, price float32, types string) {
	var (
		time1  int32
		time2  int32
		price1 float32
		price2 float32
		wash   bool
		oil    bool
	)

	for _, id := range services {
		if s, ok := cache.SERVICES_MAP[id]; ok {
			time1 += s.Time
			price1 += s.Price

			if s.Type == "CAR_WASH" {
				wash = true
			} else if s.Type == "OIL_CHANGE" {
				oil = true
			}
		}
	}

	if len(addons) > 0 {
		for _, id := range addons {
			if a, ok := cache.ADDONS_MAP[id]; ok {
				time2 += a.Time
				price2 += a.Price
			}
		}
	}

	time = time1 + time2
	price = price1 + price2
	types = calculateOrderTypes(wash, oil)

	return
}

func CancelOrder(c *gin.Context) {
	cancel(c, false)
}

func ForceCancelOrder(c *gin.Context) {
	cancel(c, true)
}

func cancel(c *gin.Context, force bool) {
	checkQuery := `
		select user_car_id, user_payment_id, opening_id, gap, assignee, types
		from user_service
		where id = ? and user_id = ? and status = 'RESERVED'
	`
	query := `
		update user_service set status = 'CANCEL' where id = ? and user_id = ?
	`
	request := modules.CancelRequest{}
	temp := struct {
		CarId     int32
		PaymentId int32
		Opening   int32
		Gap       int32
		Assignee  int32
		Types     string
	}{}

	var (
		tx   *sql.Tx
		body []byte
		err  error
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
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

	if !force {
		checkQuery += `
			and DATE_ADD(CURRENT_TIMESTAMP(), INTERVAL 1 DAY) < reserved_start_timestamp
		`
	}

	if err = tx.QueryRow(
		checkQuery, request.Id, request.UserId,
	).Scan(
		&temp.CarId, &temp.PaymentId, &temp.Opening,
		&temp.Gap, &temp.Assignee, &temp.Types,
	); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			c.JSON(http.StatusAccepted, "Cannot cancel this reservation")
		}

		return
	}

	if _, err = tx.Exec(
		query, request.Id, request.UserId,
	); err != nil {
		return
	}

	go cancelReservation(request.UserId)

	if err = unlockCar(tx, temp.CarId, request.UserId); err != nil {
		return
	}

	if err = unlockPayment(tx, temp.PaymentId, request.UserId); err != nil {
		return
	}

	if err = releaseOpening(tx, temp.Opening, temp.Gap, temp.Types); err != nil {
		return
	}

	if err = revokeService(
		tx, request.Id, temp.Opening, temp.Gap, temp.Assignee,
	); err != nil {
		return
	}

	if force {
		if err = processPayment(
			tx, request.Id, temp.PaymentId, request.UserId, 0.5,
		); err != nil {
			return
		}
	}

	c.JSON(http.StatusOK, "OK")
}

func AddService(c *gin.Context) {
	query := `
		select estimated_price from user_service
		where id = ? and user_id = ?
	`
	request := modules.AddServiceRequest{}
	var (
		err   error
		data  []byte
		price float32
		tx    *sql.Tx
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
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
			if err = tx.Commit(); err != nil {
				return
			}
		}
	}()

	if err = tx.QueryRow(
		query, request.ServiceId, request.UserId,
	).Scan(&price); err != nil {
		return
	}

	queryAdd := `
		insert into user_service_addon_list (
			service_addon_id, user_service_id, amount
		) values (?, ?, ?)
	`

	for _, addon := range request.Addons {
		if _, err = tx.Exec(
			queryAdd, addon, request.ServiceId, 1,
		); err != nil {
			return
		}
	}

	queryUpdate := `
		update user_service set estimated_price = ?
		where id = ? and user_id = ?
	`

	if _, err = tx.Exec(
		queryUpdate, request.RealPrice, request.ServiceId, request.UserId,
	); err != nil {
		return
	}

	go applyExtraService(request.UserId)

	c.JSON(http.StatusOK, "OK")
}

func OpeningDemand(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		id   int64
		err  error
		data []byte
	)

	if id, err = strconv.ParseInt(c.Param("id"), 10, 32); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		if data, err = ioutil.ReadAll(c.Request.Body); err == nil {
			if err = json.Unmarshal(data, &request); err == nil {
				go chooseOpening(request.UserId, strconv.Itoa(int(id)))
			}
		}

		updateOpeningDemand(int32(id))
		c.JSON(http.StatusOK, "OK")
	}
}

func updateOpeningDemand(id int32) {
	query := `update opening set demand = demand + 1 where id = ?`

	if _, err := config.DB.Exec(query, id); err != nil {
		fmt.Println("Error - ", err.Error())
	}
}

func GetService(c *gin.Context) {
	c.JSON(http.StatusOK, cache.SERVICES_ARRAY)
}

func AddonDemand(c *gin.Context) {
	request := modules.AddonDemandRequest{}
	var (
		body []byte
		err  error
	)

	defer func() {
		if err != nil {
			fmt.Println(err.Error())
		}

		c.JSON(http.StatusOK, "OK")
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	updateAddonDemand(request.Addons)
}

func updateAddonDemand(ids []int32) {
	query := `update service_addon set demand = demand + 1 where id in (`
	last := len(ids) - 1

	if last < 0 {
		return
	}

	for i, id := range ids {
		query += strconv.Itoa(int(id))
		if i != last {
			query += ","
		}
	}

	query += ")"

	if _, err := config.DB.Exec(query); err != nil {
		fmt.Println("Error - Addon Demand - ", err.Error())
	}
}

func ServiceReading(c *gin.Context) {
	query := `update service set reading = reading + 1 where id = ?`
	request := modules.BaseRequest{}
	var (
		data []byte
		err  error
		id   int64
	)

	defer func() {
		if err != nil {
			fmt.Println(err.Error())
		}

		c.JSON(http.StatusOK, "OK")
	}()

	if id, err = strconv.ParseInt(c.Param("id"), 10, 32); err != nil {
		return
	}

	if _, err = config.DB.Exec(query, int32(id)); err != nil {
		return
	}

	if data, err = ioutil.ReadAll(c.Request.Body); err == nil {
		if err = json.Unmarshal(data, &request); err == nil {
			go readService(request.UserId, strconv.Itoa(int(id)))
		}
	}
}

func ServiceDemand(c *gin.Context) {
	request := modules.ServiceDemandRequest{}
	var (
		body []byte
		err  error
	)

	defer func() {
		if err != nil {
			fmt.Println(err.Error())
		}

		c.JSON(http.StatusOK, "OK")
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	updateServiceDemand(request.Services)
}

func updateServiceDemand(ids []int32) {
	query := `update service set demand = demand + 1 where id in (`
	last := len(ids) - 1

	if last < 0 {
		return
	}

	for i, id := range ids {
		query += strconv.Itoa(int(id))
		if i != last {
			query += ","
		}
	}

	query += ")"

	if _, err := config.DB.Exec(query); err != nil {
		fmt.Println("Error - Service Demand - ", err.Error())
	}
}

func assignService(tx *sql.Tx, openingId, gap int32, types string) (assignee int32, err error) {
	query := "select day, period from opening where id = ?"
	var (
		day    string
		period int32
	)

	if err = tx.QueryRow(
		query, openingId,
	).Scan(&day, &period); err != nil {
		return
	}

	mask := ((int32(math.Pow(float64(2), float64(gap))) - 1) << uint32(period-1))

	query = `
		select user_id from user_opening
		where day = ? and user_schedule & ? = ?
	`

	if types == "CAR_WASH" {
		query += " and mixed = 0"
	} else {
		query += " and mixed = 1"
	}

	fmt.Println("mask - ", mask)
	fmt.Println("day - ", day)

	if err = tx.QueryRow(
		query, day, mask, mask,
	).Scan(&assignee); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("No service provider available")
		}

		return
	}

	fmt.Println("assignee - ", assignee)

	_, err = tx.Exec(`
		update user_opening set user_schedule = user_schedule ^ ?
		where user_id = ? and day = ?
	`, mask, assignee, day)

	return
}

func revokeService(tx *sql.Tx, userServiceId, openingId, gap, assignee int32) (err error) {
	var (
		day    string
		period int32
	)

	if err = tx.QueryRow(`
		select day, period from opening where id = ?`,
		openingId,
	).Scan(&day, &period); err != nil {
		return
	}

	mask := ((int32(math.Pow(float64(2), float64(gap))) - 1) << uint32(period-1))

	if _, err = tx.Exec(`
		update user_opening set user_schedule = user_schedule ^ ?
		where user_id = ? and day = ?`,
		mask, assignee, day,
	); err != nil {
		return
	}

	_, err = tx.Exec("update user_service set assignee = -1 where id = ?", userServiceId)

	return
}

func holdOpening(tx *sql.Tx, start, end int32, types string) (err error) {
	gap := end - start
	var (
		result        sql.Result
		affectedRow   int64
		updateOpening string
	)

	if types == "CAR_WASH" {
		updateOpening = `
			update opening set count_wash = count_wash - 1
			where id >= ? and id < ? and count_wash > 0
		`
	} else {
		updateOpening = `
			update opening set count_oil = count_oil - 1
			where id >= ? and id < ? and count_oil > 0
		`
	}

	if result, err = tx.Exec(
		updateOpening, start, end,
	); err != nil {
		return
	}

	if affectedRow, err = result.RowsAffected(); err != nil {
		return
	} else if affectedRow != int64(gap) {
		err = errors.New("Opening is not available")
		return
	}

	return
}

func releaseOpening(tx *sql.Tx, id, gap int32, types string) (err error) {
	var query string

	if types == "CAR_WASH" {
		query = `
			update opening set count_wash = count_wash + 1
			where id >= ? and id < ?
		`
	} else {
		query = `
			update opening set count_oil = count_oil + 1
			where id >= ? and id < ?
		`
	}

	_, err = tx.Exec(query, id, id+gap)

	return
}

func getSimpleService(userServices []int32) (services []modules.SimpleService, err error) {
	query := `
		select s.id, s.name, s.note, s.type, usl.user_service_id
		from service s
		inner join user_service_list usl on usl.service_id = s.id
		where usl.user_service_id in (
	`
	last := len(userServices) - 1

	if last < 0 {
		return
	}

	var (
		rows *sql.Rows
	)

	for i, id := range userServices {
		query += strconv.Itoa(int(id))

		if i != last {
			query += ", "
		}
	}

	query += ") order by usl.user_service_id"

	if rows, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		service := modules.SimpleService{}

		if err = rows.Scan(
			&service.Id, &service.Name, &service.Note,
			&service.Type, &service.UserServiceId,
		); err != nil {
			return
		}

		services = append(services, service)
	}

	return
}

func getSimpleAddon(userServices []int32) (addons []modules.SimpleAddon, err error) {
	query := `
		select sa.id, sa.name, sa.note, usal.amount, sa.unit, usal.user_service_id
		from user_service_addon_list usal
		inner join service_addon sa on sa.id = usal.service_addon_id
		where usal.user_service_id in (
	`
	last := len(userServices) - 1

	if last < 0 {
		return
	}

	var (
		rows *sql.Rows
	)

	for i, id := range userServices {
		query += strconv.Itoa(int(id))

		if i != last {
			query += ", "
		}
	}

	query += ") order by usal.user_service_id"

	if rows, err = config.DB.Query(query); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		addon := modules.SimpleAddon{}

		if err = rows.Scan(
			&addon.Id, &addon.Name, &addon.Note, &addon.Amount,
			&addon.Unit, &addon.UserServiceId,
		); err != nil {
			return
		}

		addons = append(addons, addon)
	}

	return
}

func calculateGap(time int32) (gap int32) {
	gap = time/30 + 1

	if time%30 != 0 {
		gap += 1
	}

	return gap
}

func calculateReservedTime(tx *sql.Tx, opening int32) (reserved string, err error) {
	var (
		day    string
		period int32
	)

	if err = tx.QueryRow(
		"select day, period from opening where id = ?", opening,
	).Scan(&day, &period); err != nil {
		return
	} else {
		total := (period - 1) * 30
		hour := strconv.Itoa(int(OPENING_BASE) + int(total/60))
		minute := total % 60

		if minute == 0 {
			reserved = day + " " + hour + ":00:00"
		} else {
			reserved = day + " " + hour + ":30:00"
		}
	}

	return reserved, nil
}

func calculateHowLong(mins int32) (int32, string) {
	if mins >= 1440 {
		return int32(mins / 1440), "DAY"
	} else if mins >= 60 {
		return int32(mins / 60), "HOUR"
	} else {
		return mins, "MINUTE"
	}
}

func calculateOrderTypes(wash, oil bool) (types string) {
	switch {
	case wash && oil:
		types = "BOTH"
	case wash:
		types = "CAR_WASH"
	case oil:
		types = "OIL_CHANGE"
	default:
		types = "BOTH"
	}

	return
}
