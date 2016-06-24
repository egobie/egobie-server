package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initFleetRoutes() {

	fleetRouter.POST("/service", controllers.GetFleetService)

	fleetRouter.POST("/addon", controllers.GetFleetAddon)

	fleetRouter.POST("/order", controllers.PlaceFleetOrder)

	fleetRouter.POST("/opening", controllers.GetFleetOpening)

	fleetRouter.POST("/cancel", controllers.CancelFleetOrder)

	fleetRouter.POST("/cancel/force", controllers.ForceCancelFleetOrder)

	fleetRouter.POST("/reservation", controllers.GetFleetReservation)

	fleetRouter.POST("/reservation/detail", controllers.GetFleetReservationDetail)

	fleetRouter.POST("/history", controllers.GetFleetHistory)

	fleetRouter.POST("/history/rating", controllers.RatingFleet)

	fleetRouter.POST("/demand/opening/:id", controllers.OpeningDemand)

	fleetRouter.POST("/read/:id", controllers.ServiceReading)

	fleetRouter.POST("/demand", controllers.ServiceDemand)

	fleetRouter.POST("/demand/addon", controllers.AddonDemand)

	fleetRouter.POST("/update/password", controllers.UpdatePassword)

	fleetRouter.POST("/update/user", controllers.UpdateUser)

	fleetRouter.POST("/update/home", controllers.UpdateHome)

	fleetRouter.POST("/update/work", controllers.UpdateWork)

	fleetRouter.POST("/feedback", controllers.Feedback)

//	fleetRouter.POST("/test", controllers.Test)
}
