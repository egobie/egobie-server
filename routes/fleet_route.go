package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initFleetRoutes() {

	fleetRouter.POST("/order", controllers.PlaceFleetOrder)

//	fleetRouter.POST("/cancel", controllers.CancelOrder)

//	fleetRouter.POST("/cancel/force", controllers.ForceCancelOrder)
}
