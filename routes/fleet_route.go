package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initFleetRoutes() {

	fleetRouter.POST("/order", controllers.PlaceFleetOrder)

	//	fleetRouter.POST("/cancel", controllers.CancelOrder)

	//	fleetRouter.POST("/cancel/force", controllers.ForceCancelOrder)

	fleetRouter.POST("/reservation", controllers.GetFleetReservation)

	fleetRouter.POST("/reservation/detail", controllers.GetFleetReservationDetail)

	fleetRouter.POST("/history", controllers.GetFleetHistory)

	fleetRouter.POST("/rating", controllers.RatingFleet)
}
