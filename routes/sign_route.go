package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initSignRoutes() {
	router.POST("/check/email", controllers.CheckEmail)

	router.POST("/check/name", controllers.CheckUsername)

	router.POST("/signup", controllers.SignUp)

	router.POST("/signin", controllers.SignIn)

	router.POST("/signup/fleet", controllers.SignUpFleet)

	router.POST("/reset/step1", controllers.ResetPasswordStep1)

	router.POST("/reset/step2", controllers.ResetPasswordStep2)

	router.POST("/reset/step3", controllers.ResetPasswordStep3)

	router.POST("/reset/resend", controllers.ResetPasswordResend)

	router.GET("/secure/:code", controllers.Secure)

//	router.POST("/test", controllers.Test)
}
