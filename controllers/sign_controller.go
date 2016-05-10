package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
	"github.com/egobie/egobie-server/secures"

	"github.com/gin-gonic/gin"
)

func updateUserSignIn(userId int32) {
	query := `
		update user set sign = sign + 1, sign_in = CURRENT_TIMESTAMP where id = ?
	`

	if _, err := config.DB.Exec(query, userId); err != nil {
		fmt.Println("fail to update sign-in - ", err.Error())
	}
}

func updateUserCoupon(coupon string) {
	query := `
		update user
		set invitation = invitation + 1, discount = discount + 1
		where coupon = ?
	`

	if _, err := config.DB.Exec(query, coupon); err != nil {
		fmt.Println("fail to update coupon - ", coupon, " error - ", err.Error())
	}
}

func check(c *gin.Context, query, errorMessage string) {
	request := modules.Check{}
	var (
		body []byte
		err error
		count int64
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = config.DB.QueryRow(query, request.Value).Scan(&count); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else if count >= 1{
		c.IndentedJSON(http.StatusAccepted, errorMessage + " is already in use")
	} else {
		c.IndentedJSON(http.StatusOK, "OK")
	}
}

func CheckEmail(c *gin.Context) {
	check(c, "select count(*) from user where email = ?", "Email address")
}

func CheckUsername(c *gin.Context) {
	check(c, "select count(*) from user where username = ?", "Username");
}

func SignUp(c *gin.Context) {
	query := `
		insert into user (type, username, password, email, phone_number, referred)
		values ('RESIDENTIAL', ?, ?, ?, ?, ?)
	`
	request := modules.SignUp{}
	pattern := "^([A-Z0-9]{5})$"
	var (
		result       sql.Result
		enPassword   string
		lastInsertId int64
		body         []byte
		err          error
		referred     string
		matched      bool
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if enPassword, err = secures.EncryptPassword(request.Password); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	request.Coupon = strings.ToUpper(strings.TrimSpace(request.Coupon))

	if matched, _ = regexp.MatchString(
		pattern, request.Coupon,
	); matched {
		referred = request.Coupon
	} else {
		referred = ""
	}

	if result, err = config.DB.Exec(
		query, request.Username, enPassword, request.Email,
		request.PhoneNumber, referred,
	); err != nil {
		message := ""

		if isDuplicateEntryError(err) {
			message = "user already exists!"
		} else {
			message = err.Error()
		}

		c.IndentedJSON(http.StatusBadRequest, message)
		c.Abort()
		return
	}

	if lastInsertId, err = result.LastInsertId(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if user, err := getUserById(int32(lastInsertId)); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		user.Password = getUserToken(user.Password)
		c.IndentedJSON(http.StatusOK, user)
	}

	updateUserSignIn(int32(lastInsertId))

	if matched {
		updateUserCoupon(request.Coupon)
	}
}

func SignIn(c *gin.Context) {
	request := modules.SignIn{}
	var (
		dePassword string
		user       modules.User
		body       []byte
		err        error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if user, err = getUser(
		"username = ?", request.Username,
	); err != nil {
		switch {
		case err == sql.ErrNoRows:
			c.IndentedJSON(http.StatusBadRequest, "User not found")
		default:
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}
		c.Abort()
		return
	}

	if dePassword, err = secures.DecryptPassword(user.Password); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if dePassword != request.Password {
		c.IndentedJSON(http.StatusBadRequest, "Password not match")
		c.Abort()
		return
	}

	updateUserSignIn(user.Id)

	user.Password = getUserToken(user.Password)

	c.IndentedJSON(http.StatusOK, user)
}

func Secure(c *gin.Context) {
	if code, err := secures.EncryptPassword(c.Param("code")); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		c.IndentedJSON(http.StatusOK, code)
	}
}
