package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

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
		body  []byte
		err   error
		count int64
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if err = config.DB.QueryRow(query, request.Value).Scan(&count); err != nil {
		return
	} else if count >= 1 {
		c.JSON(http.StatusAccepted, errorMessage+" is already in use")
	} else {
		c.JSON(http.StatusOK, "OK")
	}
}

func CheckEmail(c *gin.Context) {
	check(c, "select count(*) from user where email = ?", "Email address")
}

func CheckUsername(c *gin.Context) {
	check(c, "select count(*) from user where username = ?", "Username")
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

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if enPassword, err = secures.EncryptPassword(request.Password); err != nil {
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
		if isDuplicateEntryError(err) {
			err = errors.New("user already exists!")
		}

		return
	}

	if lastInsertId, err = result.LastInsertId(); err != nil {
		return
	}

	if user, err := getUserById(int32(lastInsertId)); err != nil {
		return
	} else {
		user.Password = getUserToken(modules.USER_RESIDENTIAL, user.Password)
		c.JSON(http.StatusOK, user)
	}

	go updateUserSignIn(int32(lastInsertId))

	if matched {
		go updateUserCoupon(request.Coupon)
	}
}

func SignUpFleet(c *gin.Context) {
	query := `
		select u.id, f.id, f.name, f.setup
		from user u
		inner join fleet f on f.user_id = u.id
		where u.email = ? and f.token = ?
	`
	queryUser := `
		update user set username = ?, password = ? where id = ?
	`
	querySetUp := `
		update fleet set setup = 1 where id = ? and user_id = ?
	`
	request := modules.SignUpFleet{}
	pattern := "^([A-Z0-9]{5})$"
	user := modules.FleetUser{}
	temp := struct{
		UserId     int32
		FleetId    int32
		SetUp      int32
		Name       string
	}{}
	var (
		tx *sql.Tx
		enPassword string
		body       []byte
		err        error
		matched    bool
	)

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		go updateUserSignIn(temp.UserId)
		user.Password = getUserToken(modules.USER_FLEET, user.Password)

		c.JSON(http.StatusOK, user)
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if enPassword, err = secures.EncryptPassword(request.Password); err != nil {
		return
	}

	request.Token = strings.ToUpper(strings.TrimSpace(request.Token))

	if matched, _ = regexp.MatchString(
		pattern, request.Token,
	); !matched {
		err = errors.New("Invalid invitation code")
		return
	}

	if tx, err = config.DB.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				fmt.Println("Error - Rollback - ", err1.Error())
			}
		} else {
			if err = tx.Commit(); err == nil {
				user, err = getFleetUserByUserId(temp.UserId)
			}
		}
	}()

	if err = tx.QueryRow(query, request.Email, request.Token).Scan(
		&temp.UserId, &temp.FleetId, &temp.Name, &temp.SetUp,
	); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("Fleet user not found")
		}
		return
	} else if temp.SetUp == 1 {
		err = errors.New("Fleet user sign-up twice")
		return
	}

	if _, err = tx.Exec(
		queryUser, request.Username, enPassword, temp.UserId,
	); err != nil {
		return
	}

	if _, err = tx.Exec(querySetUp, temp.FleetId, temp.UserId); err != nil {
		return
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

	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if user, err = getUserByUsername(request.Username); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("User not found")
		}

		return
	}

	if dePassword, err = secures.DecryptPassword(user.Password); err != nil {
		return
	}

	if dePassword != request.Password {
		err = errors.New("Password not match")
		return
	}

	updateUserSignIn(user.Id)

	user.Password = getUserToken(user.Type, user.Password)

	if (user.Type == modules.USER_FLEET) {
		var ui modules.FleetUserBasicInfo

		if ui, err = getFleetUserBasicInfo(user.Id); err == nil {
			c.JSON(http.StatusOK, modules.FleetUser{
				User: user,
				FleetUserBasicInfo: ui,
			})
		}
	} else {
		c.JSON(http.StatusOK, user)
	}
}

func Secure(c *gin.Context) {
	if code, err := secures.EncryptPassword(c.Param("code")); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		c.JSON(http.StatusOK, code)
	}
}
