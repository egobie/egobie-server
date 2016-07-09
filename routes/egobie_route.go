package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initEgobieRoutes() {

	egobieRouter.POST("/service/task", controllers.GetTask)

	egobieRouter.POST("/service/task/detail", controllers.GetFleetReservationDetail)

	egobieRouter.POST("/service/user/done", controllers.MakeUserServiceDone)

	egobieRouter.POST("/service/user/progress", controllers.MakeUserServiceInProgress)

	egobieRouter.POST("/service/user/reserved", controllers.MakeUserServiceReserved)

	egobieRouter.POST("/service/user/cancel", controllers.MakeUserServiceCancelled)

	egobieRouter.POST("/service/fleet/done", controllers.MakeFleetServiceDone)

	egobieRouter.POST("/service/fleet/progress", controllers.MakeFleetServiceInProgress)
}
