package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

var (
	router        = gin.New()
	userRouter    = router.Group("/user")
	carRouter     = router.Group("/car")
	paymentRouter = router.Group("/payment")
	serviceRouter = router.Group("/service")
	historyRouter = router.Group("/history")

	fleetRouter = router.Group("/fleet")
	saleRouter  = router.Group("/sale")

	userActionRouter = router.Group("/action")

	egobieRouter = router.Group("/egobie")
)

func init() {
	// Release
	gin.SetMode(gin.ReleaseMode)

	// CORS
	router.Use(cors)

	// CORS, Authorize User
	userRouter.Use(cors, authorizeResidentialUser)
	carRouter.Use(cors, authorizeResidentialUser)
	paymentRouter.Use(cors, authorizeResidentialUser)
	serviceRouter.Use(cors, authorizeResidentialUser)
	historyRouter.Use(cors, authorizeResidentialUser)

	fleetRouter.Use(cors, authorizeFleetUser)
	saleRouter.Use(cors, authorizeSaleUser)

	egobieRouter.Use(cors, authorizeEgobieUser)

	userActionRouter.Use(cors)

	router.GET("/hc", func(c *gin.Context) {
		c.JSON(http.StatusOK, "OK")
	})

	initSignRoutes()
	initUserRoutes()
	initServiceRoutes()
	initCarRoutes()
	initPaymentRoutes()
	initHistoryRoutes()
	initUserActionRoutes()

	initEgobieRoutes()

	initFleetRoutes()
	initSaleRoutes()
}

func cors(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set(
		"Access-Control-Allow-Headers",
		"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"+
			", "+config.EGOBIE_HEADER_TOKEN+
			", "+config.EGOBIE_HEADER_USERID,
	)
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, GET, PUT")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

func authorizeResidentialUser(c *gin.Context) {
	if err := authorizeUser(
		c, modules.USER_RESIDENTIAL_TOKEN, modules.IsResidential,
	); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.Next()
}

func authorizeEgobieUser(c *gin.Context) {
	if err := authorizeUser(
		c, modules.USER_EGOBIE_TOKEN, modules.IsEgobie,
	); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.Next()
}

func authorizeSaleUser(c *gin.Context) {
	if err := authorizeUser(
		c, modules.USER_SALE_TOKEN, modules.IsSale,
	); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.Next()
}

func authorizeFleetUser(c *gin.Context) {
	if err := authorizeUser(
		c, modules.USER_FLEET_TOKEN, modules.IsFleet,
	); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.Next()
}

func authorizeUser(c *gin.Context, expectTokenLen int32, checkFunc modules.CheckUserFunc) (err error) {
	var (
		token    string
		userId   int32
		userType string
	)

	if userId, token, err = parseUser(c); err != nil {
		return
	} else if int32(len(token)) != expectTokenLen {
		err = errors.New("Invalid user")
		return
	}

	if userType, err = readUser(userId, token); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("User not found")
		}
		return
	} else if !checkFunc(userType) {
		err = errors.New("Invalid user")
		return
	}

	return nil
}

func parseUser(c *gin.Context) (int32, string, error) {
	request := modules.BaseRequest{}
	var (
		data []byte
		err  error
	)

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return -1, "", err
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return -1, "", err
	}

	// Put request body back
	c.Request.Body = ioutil.NopCloser(bytes.NewReader(data))

	return request.UserId, request.UserToken, nil
}

func readUser(userId int32, token string) (userType string, err error) {
	if err = config.DB.QueryRow(
		"select type from user where id = ? and password like ?",
		userId, token+"%",
	).Scan(&userType); err != nil {
		return "", err
	}

	return userType, nil
}

func Serve(port string) {
	router.Run(port)
}
