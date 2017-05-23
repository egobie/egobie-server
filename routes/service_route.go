package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initServiceRoutes() {
	// Get all services
	serviceRouter.POST("", controllers.GetService)

	serviceRouter.POST("/reservation", controllers.GetReservation)

	serviceRouter.POST("/reserved", controllers.GetUserServiceReserved)

	serviceRouter.POST("/done", controllers.GetUserServiceDone)

	serviceRouter.POST("/opening", controllers.GetOpening)

	serviceRouter.POST("/place/opening", controllers.GetPlaceOpening)

	serviceRouter.POST("/order", controllers.PlaceOrder)

	serviceRouter.POST("/cancel", controllers.CancelOrder)

	serviceRouter.POST("/cancel/force", controllers.ForceCancelOrder)

	serviceRouter.POST("/cancel/free", controllers.FreeCancelOrder)

	serviceRouter.POST("/add", controllers.AddService)

	serviceRouter.POST("/demand/opening/:id", controllers.OpeningDemand)

	serviceRouter.POST("/read/:id", controllers.ServiceReading)

	serviceRouter.POST("/demand", controllers.ServiceDemand)

	serviceRouter.POST("/demand/addon", controllers.AddonDemand)

	serviceRouter.POST("/now", controllers.OnDemand)
}
