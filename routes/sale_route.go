package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initSaleRoutes() {

	saleRouter.POST("/fleet/new", controllers.NewFleetUser)

	saleRouter.POST("/fleet/resend", controllers.ResendEmail);

	saleRouter.POST("/fleet/all", controllers.AllFleetUser)

	saleRouter.POST("/fleet/order", controllers.AllFleetOrder)

	saleRouter.POST("/fleet/order/detail", controllers.GetFleetReservationDetail)

	saleRouter.POST("/fleet/price", controllers.PromotePrice)
}
