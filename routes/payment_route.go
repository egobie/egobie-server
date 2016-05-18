package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initPaymentRoutes() {
//	paymentRouter.POST("/", controllers.GetPaymentById)

	paymentRouter.POST("/user", controllers.GetPaymentByUserId)

	paymentRouter.POST("/new", controllers.CreatePayment)

	paymentRouter.POST("/update", controllers.UpdatePayment)

	paymentRouter.POST("/delete", controllers.DeletePayment)

	//paymentRouter.POST("/pay", controllers.MakePayment);
}
