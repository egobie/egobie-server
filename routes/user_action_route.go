package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initUserActionRoutes() {

	userActionRouter.POST("/login", controllers.Login)

	userActionRouter.POST("/logout", controllers.Logout)

	userActionRouter.POST("/go/home", controllers.GoHome)

	userActionRouter.POST("/go/notification", controllers.GoNotifications)

	userActionRouter.POST("/go/service", controllers.GoServices)

	userActionRouter.POST("/go/history", controllers.GoHistory)

	userActionRouter.POST("/go/coupon", controllers.GoCoupon)

	userActionRouter.POST("/go/car", controllers.GoCars)

	userActionRouter.POST("/go/payment", controllers.GoPayments)

	userActionRouter.POST("/go/setting", controllers.GoSettings)

	userActionRouter.POST("/go/about", controllers.GoAbout)

	userActionRouter.POST("/go/reservation", controllers.GoReservation)

	userActionRouter.POST("/go/ondemand", controllers.GoOnDemand)

	userActionRouter.POST("/open/rating", controllers.OpenRating)

	userActionRouter.POST("/open/feedback", controllers.OpenFeedback)

	userActionRouter.POST("/open/service", controllers.OpenChooseServices)

	userActionRouter.POST("/open/extra", controllers.OpenChooseExtraServices)

	userActionRouter.POST("/open/date", controllers.OpenChooseOpening)

	userActionRouter.POST("/open/car", controllers.OpenChooseCar)

	userActionRouter.POST("/open/payment", controllers.OpenChoosePayment)

	userActionRouter.POST("/read/history", controllers.ReadHistory)

	userActionRouter.POST("/click/opening", controllers.ClickOpening)

	userActionRouter.POST("/unselect/service", controllers.UnselectService)

	userActionRouter.POST("/unselect/extra", controllers.UnselectExtraService)

	userActionRouter.POST("/reload/opening", controllers.ReloadOpening)
}
