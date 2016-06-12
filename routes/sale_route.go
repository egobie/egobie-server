package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initSaleRoutes() {

	saleRouter.POST("/fleet/new", controllers.NewFleetUser)

	saleRouter.POST("/fleet/all", controllers.AllFleetUser)
}
