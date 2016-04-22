package routes

import (
	"github.com/eGobie/egobie-server/controllers"
)

func initHistoryRoutes() {
	historyRouter.POST("", controllers.GetHistory)
}
