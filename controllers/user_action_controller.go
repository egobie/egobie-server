package controllers

import (
	"encoding/json"
	"io/ioutil"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

func createUserAction(c *gin.Context, action string) {
	request := modules.ActionRequest{}
	var (
		data []byte
		err  error
	)

	if data, err = ioutil.ReadAll(c.Request.Body); err == nil {
		if err = json.Unmarshal(data, &request); err == nil {
			if request.Data != "" {
				createWithData(request.UserId, action, request.Data)
			} else {
				create(request.UserId, action)
			}
		}
	}
}

func create(userId int32, action string) {
	config.DB.Exec(`
		insert into user_action (user_id, action) values (?, ?)`,
		userId,
		action,
	)
}

func createWithData(userId int32, action, data string) {
	config.DB.Exec(`
		insert into user_action (user_id, action, data) values (?, ?, ?)`,
		userId,
		action,
		data,
	)
}

func Login(c *gin.Context) {
	go createUserAction(c, "Sign In")
}

func Logout(c *gin.Context) {
	go createUserAction(c, "Sign Out")
}

func GoHome(c *gin.Context) {
	go createUserAction(c, "Go to <Home>")
}

func GoNotifications(c *gin.Context) {
	go createUserAction(c, "Go to <Notifications>")
}

func GoServices(c *gin.Context) {
	go createUserAction(c, "Go to <Services>")
}

func GoHistory(c *gin.Context) {
	go createUserAction(c, "Go to <History>")
}

func GoCoupon(c *gin.Context) {
	go createUserAction(c, "Go to <Coupon>")
}

func GoCars(c *gin.Context) {
	go createUserAction(c, "Go to <Vehicles>")
}

func GoPayments(c *gin.Context) {
	go createUserAction(c, "Go to <Payments>")
}

func GoSettings(c *gin.Context) {
	go createUserAction(c, "Go to <Settings>")
}

func GoAbout(c *gin.Context) {
	go createUserAction(c, "Go to <About>")
}

func GoReservation(c *gin.Context) {
	go createUserAction(c, "Go to <Reservation>")
}

func GoOnDemand(c *gin.Context) {
	go createUserAction(c, "Go to <On Demand>")
}

func OpenRating(c *gin.Context) {
	go createUserAction(c, "Open <Rating>")
}

func OpenFeedback(c *gin.Context) {
	go createUserAction(c, "Open <Feedback>")
}

func OpenChooseServices(c *gin.Context) {
	go createUserAction(c, "Open <Choose Services>")
}

func OpenChooseExtraServices(c *gin.Context) {
	go createUserAction(c, "Open <Choose Extra Services>")
}

func OpenChooseOpening(c *gin.Context) {
	go createUserAction(c, "Open <Choose Opening>")
}

func OpenChooseCar(c *gin.Context) {
	go createUserAction(c, "Open <Choose Vehicle>")
}

func OpenChoosePayment(c *gin.Context) {
	go createUserAction(c, "Open <Choose Payment>")
}

func ClickOpening(c *gin.Context) {
	go createUserAction(c, "Click <Opening>")
}

func ReadHistory(c *gin.Context) {
	go createUserAction(c, "Read <History>")
}

func UnselectService(c *gin.Context) {
	go createUserAction(c, "Unselect <Service>")
}

func UnselectExtraService(c *gin.Context) {
	go createUserAction(c, "Unselect <Extra Service>")
}

func ReloadOpening(c *gin.Context) {
	go createUserAction(c, "Reload <Opening>")
}

func readService(userId int32, data string) {
	createWithData(userId, "Read Service", data)
}

func chooseOpening(userId int32, data string) {
	createWithData(userId, "Choose Opening", data)
}

func changeAddress(userId int32) {
	create(userId, "Change Address")
}

func changePassword(userId int32) {
	create(userId, "Change Password")
}

func changeUser(userId int32) {
	create(userId, "Change User")
}

func makeReservation(userId int32) {
	create(userId, "Make Reservation")
}

func cancelReservation(userId int32) {
	create(userId, "Cancel Reservation")
}

func rateService(userId int32) {
	create(userId, "Rate Service")
}

func checkAvailability(userId int32) {
	create(userId, "Check Availability")
}

func applyExtraService(userId int32) {
	create(userId, "Apply Extra Service")
}

func addPayment(userId int32) {
	create(userId, "Add Payment")
}

func editPayment(userId int32) {
	create(userId, "Edit Payment")
}

func deletePayment(userId int32) {
	create(userId, "Delete Payment")
}

func addCar(userId int32) {
	create(userId, "Add Car")
}

func editCar(userId int32) {
	create(userId, "Edit Car")
}

func deleteCar(userId int32) {
	create(userId, "Delete Car")
}
