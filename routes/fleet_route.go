package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initFleetRoutes() {

	fleetRouter.POST("/service", controllers.GetFleetService)

	fleetRouter.POST("/addon", controllers.GetFleetAddon)

	fleetRouter.POST("/order", controllers.PlaceFleetOrder)

	//	fleetRouter.POST("/cancel", controllers.CancelOrder)

	//	fleetRouter.POST("/cancel/force", controllers.ForceCancelOrder)

	fleetRouter.POST("/reservation", controllers.GetFleetReservation)

	fleetRouter.POST("/reservation/detail", controllers.GetFleetReservationDetail)

	fleetRouter.POST("/history", controllers.GetFleetHistory)

	fleetRouter.POST("/history/rating", controllers.RatingFleet)

	fleetRouter.POST("/demand/opening/:id", controllers.OpeningDemand)

	fleetRouter.POST("/read/:id", controllers.ServiceReading)

	fleetRouter.POST("/demand", controllers.ServiceDemand)

	fleetRouter.POST("/demand/addon", controllers.AddonDemand)
}
