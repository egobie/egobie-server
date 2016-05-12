package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initServiceRoutes() {
	// Get all services
	serviceRouter.GET("", controllers.GetService)

	serviceRouter.POST("/reservation", controllers.GetUserServiceReserved)

	serviceRouter.POST("/done", controllers.GetUserServiceDone)

	serviceRouter.POST("/opening", controllers.GetOpening)

	serviceRouter.POST("/order", controllers.PlaceOrder)

	serviceRouter.POST("/cancel", controllers.CancelOrder)

	serviceRouter.GET("/demand/opening/:id", controllers.OpeningDemand)

	serviceRouter.GET("/read/:id", controllers.ServiceReading)

	serviceRouter.POST("/demand", controllers.ServiceDemand)
}
