package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/cache"
	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
	"github.com/egobie/egobie-server/secures"
	"github.com/egobie/egobie-server/utils"

	"github.com/gin-gonic/gin"
)

func getUser(condition string, args ...interface{}) (user modules.User, err error) {
	query := `
		select id, type, username, password, first_time,
			email, phone_number, coupon, discount,
			first_name, last_name, middle_name,
			home_address_state, home_address_zip,
			home_address_city, home_address_street,
			work_address_state, work_address_zip,
			work_address_city, work_address_street
		from user where
	`

	if err = config.DB.QueryRow(query+" "+condition, args...).Scan(
		&user.Id, &user.Type, &user.Username, &user.Password, &user.FirstTime,
		&user.Email, &user.PhoneNumber, &user.Coupon, &user.Discount,
		&user.FirstName, &user.LastName, &user.MiddleName,
		&user.HomeAddressState, &user.HomeAddressZip,
		&user.HomeAddressCity, &user.HomeAddressStreet,
		&user.WorkAddressState, &user.WorkAddressZip,
		&user.WorkAddressCity, &user.WorkAddressStreet,
	); err != nil {
		return
	}

	return user, nil
}

func getUserById(id int32) (user modules.User, err error) {
	return getUser("id = ?", id)
}

func getUserByUsername(username string) (user modules.User, err error) {
	return getUser("username = ?", username)
}

func getUserToken(userType, password string) string {
	if modules.IsResidential(userType) {
		return password[:modules.USER_RESIDENTIAL_TOKEN]

	} else if modules.IsEgobie(userType) {
		return password[:modules.USER_EGOBIE_TOKEN]

	} else if modules.IsFleet(userType) {
		return password[:modules.USER_FLEET_TOKEN]

	} else if modules.IsBusiness(userType) {
		return password[:modules.USER_BUSINESS_TOKEN]
	} else if modules.IsSale(userType) {
		return password[:modules.USER_SALE_TOKEN]
	}

	return ""
}

func updateAddress(body []byte, setClause string) (err error) {
	query := "update user set " + setClause + " where id = ? and password like ?"
	request := modules.UpdateAddress{}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if _, err = config.DB.Exec(query,
		request.State, utils.FormatZipcode(request.Zip), request.City,
		request.Street, request.UserId, request.UserToken+"%",
	); err != nil {
		return
	}

	go changeAddress(request.UserId)

	return nil
}

func GetUser(c *gin.Context) {
	request := modules.UserRequest{}
	var (
		body []byte
		user modules.User
		err  error
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

	if user, err = getUser(
		"id = ? and password like ?", request.UserId, request.UserToken+"%",
	); err != nil {
		return
	}

	user.Password = ""

	c.JSON(http.StatusOK, user)
}

func GetDiscount(c *gin.Context) {
	c.JSON(http.StatusOK, cache.DISCOUNT_MAP)
}

func UpdateUser(c *gin.Context) {
	query := `
		update user
		set first_name = ?, last_name = ?, middle_name = ?,
			email = ?, phone_number = ?
		where id = ? and password like ?
	`
	request := modules.UpdateUser{}
	var (
		body []byte
		err  error
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

	if _, err = config.DB.Exec(query,
		request.FirstName, request.LastName, request.MiddleName, request.Email,
		utils.FormatPhone(request.PhoneNumber), request.UserId, request.UserToken+"%",
	); err != nil {
		return
	}

	go changeUser(request.UserId)

	if user, err := getUserById(request.UserId); err == nil {
		user.Password = getUserToken(user.Type, user.Password)
		c.JSON(http.StatusOK, user)
	}
}

func UpdatePassword(c *gin.Context) {
	query := "update user set password = ? where id = ?"
	request := modules.UpdatePassword{}
	var (
		user       modules.User
		dePassword string
		enPassword string
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

	if user, err = getUserById(request.UserId); err != nil {
		err = errors.New("User not found")
		return
	}

	if dePassword, err = secures.DecryptPassword(user.Password); err != nil {
		return
	}

	if dePassword != request.Password {
		err = errors.New("Password is not valid")
		return
	}

	if enPassword, err = secures.EncryptPassword(request.NewPassword); err != nil {
		return
	}

	if _, err = config.DB.Exec(query,
		enPassword, request.UserId,
	); err != nil {
		return
	}

	go changePassword(request.UserId)

	c.JSON(http.StatusOK, modules.UserInfo{
		request.UserId, getUserToken(user.Type, enPassword),
	})
}

func UpdateHome(c *gin.Context) {
	var (
		body []byte
		err  error
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

	if err = updateAddress(
		body, `home_address_state = ?, home_address_zip = ?, home_address_city = ?, home_address_street = ?`,
	); err != nil {
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func UpdateWork(c *gin.Context) {
	var (
		body []byte
		err  error
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

	if err = updateAddress(
		body, `work_address_state = ?, work_address_zip = ?, work_address_city = ?, work_address_street = ?`,
	); err != nil {
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func GetCoupon(c *gin.Context) {
	query := `
		select discount from coupon c
		inner join user_coupon uc on uc.coupon_id = c.id
		where c.expired = 0 and uc.user_id = ? and uc.used = 0
		order by create_timestamp
	`
	request := modules.BaseRequest{}
	var (
		discount int32
		err      error
		body     []byte
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

	if err = config.DB.QueryRow(query, request.UserId).Scan(&discount); err != nil {
		return
	}

	c.JSON(http.StatusOK, discount)
}

func ApplyCoupon(c *gin.Context) {
	query := `
		select coupon_id from user_coupon
		where user_id = ? and (used = 0 or coupon_id = ?)
		order by create_timestamp
	`
	request := modules.ApplyCouponRequest{}
	var (
		temp   int32
		err    error
		body   []byte
		coupon cache.Coupon
		ok     bool
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

	if coupon, ok = cache.COUPON_CACHE[request.Coupon]; !ok {
		err = errors.New("Invalid Coupon Code")
		return
	}

	err = config.DB.QueryRow(query, request.UserId, coupon.Id).Scan(&temp)

	if err != sql.ErrNoRows {
		return
	} else if err == nil {
		err = errors.New("You already have a coupon activated")
		return
	}

	query = `
		insert into user_coupon (user_id, coupon_id) values (?, ?)
	`

	if _, err = config.DB.Exec(query, request.UserId, coupon.Id); err != nil {
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func Feedback(c *gin.Context) {
	query := `
		insert into user_feedback (user_id, title, feedback) values (?, ?, ?)
	`
	request := modules.Feedback{}
	var (
		err  error
		body []byte
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

	if _, err = config.DB.Exec(
		query, request.UserId, request.Title, request.Feedback,
	); err != nil {
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func useDiscount(tx *sql.Tx, userId int32) (err error) {
	query := `
		update user set discount = CASE
				WHEN discount > 0 and first_time <= 0 THEN discount - 1
				ELSE discount
			END, first_time = CASE
				WHEN first_time > 0 THEN first_time - 1
				ELSE 0
			END
		where id = ?
	`

	_, err = tx.Exec(query, userId)

	return
}

func useCoupon(tx *sql.Tx, userId, couponId int32) (err error) {
	query := `
		update user_coupon set used = 1
		where user_id = ? and coupon_id = ?
	`

	_, err = tx.Exec(query, userId, couponId)

	return
}
