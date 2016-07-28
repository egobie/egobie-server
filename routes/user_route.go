package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initUserRoutes() {

	userRouter.POST("", controllers.GetUser)

	userRouter.POST("/discount", controllers.GetDiscount)

	userRouter.POST("/update/password", controllers.UpdatePassword)

	userRouter.POST("/update/user", controllers.UpdateUser)

	userRouter.POST("/update/home", controllers.UpdateHome)

	userRouter.POST("/update/work", controllers.UpdateWork)

	userRouter.POST("/feedback", controllers.Feedback)

	userRouter.POST("/coupon", controllers.GetCoupon)

	userRouter.POST("/coupon/apply", controllers.ApplyCoupon)
}
