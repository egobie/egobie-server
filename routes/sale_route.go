package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initSaleRoutes() {

	saleRouter.POST("/fleet/new", controllers.NewFleetUser)
}
