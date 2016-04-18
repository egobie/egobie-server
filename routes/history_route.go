package routes

import (
	"github.com/eGobie/server/controllers"
)

func initHistoryRoutes() {
	historyRouter.POST("/", controllers.GetHistory)
}
