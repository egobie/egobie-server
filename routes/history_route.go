package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initHistoryRoutes() {
	historyRouter.POST("", controllers.GetHistory)
}
