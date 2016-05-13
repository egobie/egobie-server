package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
	"github.com/egobie/egobie-server/secures"

	"github.com/gin-gonic/gin"
	"github.com/lionelbarrow/braintree-go"
)

func EncryptAccount(accountType, accountNumber, accountCode string) (enNumber, enCode string, err error) {
	switch {
	case accountType == config.CREDIT_CARD:
		if enNumber, err = secures.EncryptCredit(accountNumber); err != nil {
			return
		}

		if enCode, err = secures.EncryptCreditCVV(accountCode); err != nil {
			return
		}
	case accountType == config.DEBIT_CARD:
		if enNumber, err = secures.EncryptDebit(accountNumber); err != nil {
			return
		}

		if enCode, err = secures.EncryptDebitPin(accountCode); err != nil {
			return
		}
	default:
		return "", "", errors.New("Unknown payment type - " + accountType)
	}

	return enNumber, enCode, nil
}

func DecryptAccountNumber(accountType, accountNumber string) (deNumber string, err error) {
	switch {
	case accountType == config.CREDIT_CARD:
		if deNumber, err = secures.DecryptCredit(accountNumber); err != nil {
			return
		}
	case accountType == config.DEBIT_CARD:
		if deNumber, err = secures.DecryptDebit(accountNumber); err != nil {
			return
		}
	default:
		return "", errors.New("Unknown payment type - " + accountType)
	}

	return deNumber, nil
}

func DecryptAccountCode(accountType, code string) (deNumber string, err error) {
	switch {
	case accountType == config.CREDIT_CARD:
		if deNumber, err = secures.DecryptCreditCVV(code); err != nil {
			return
		}
	case accountType == config.DEBIT_CARD:
		if deNumber, err = secures.DecryptDebitPin(code); err != nil {
			return
		}
	default:
		return "", errors.New("Unknown payment type - " + accountType)
	}

	return deNumber, nil
}

func getPaymentLastFour(accountNumber string) string {
	return accountNumber[len(accountNumber)-4:]
}

func getPaymentByIdAndUserId(id, userId int32) (payment modules.Payment, err error) {
	query := `
		select id, user_id, account_name, account_number,
				account_type, account_zip, code,
				expire_month, expire_year, reserved
		from user_payment
		where id = ? and user_id = ?
	`

	if err = config.DB.QueryRow(query, id, userId).Scan(
		&payment.Id, &payment.UserId, &payment.AccountName,
		&payment.AccountNumber, &payment.AccountType, &payment.AccountZip,
		&payment.Code, &payment.ExpireMonth, &payment.ExpireYear, &payment.Reserved,
	); err != nil {
		return
	}

	if payment.AccountNumber, err = DecryptAccountNumber(
		payment.AccountType, payment.AccountNumber,
	); err != nil {
		return
	}

	if payment.Code, err = DecryptAccountCode(
		payment.AccountType, payment.Code,
	); err != nil {
		return
	}

	return payment, nil
}

func getPaymentForUser(userId int32) (payments []modules.Payment, err error) {
	query := `
		select id, user_id, account_name, account_number,
				account_type, account_zip, expire_month, expire_year, reserved
		from user_payment
		where user_id = ?
	`
	var (
		rows     *sql.Rows
		deNumber string
	)

	if rows, err = config.DB.Query(query, userId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		payment := modules.Payment{}

		if err = rows.Scan(
			&payment.Id, &payment.UserId, &payment.AccountName,
			&payment.AccountNumber, &payment.AccountType, &payment.AccountZip,
			&payment.ExpireMonth, &payment.ExpireYear, &payment.Reserved,
		); err != nil {
			return
		}

		if deNumber, err = DecryptAccountNumber(
			payment.AccountType, payment.AccountNumber,
		); err != nil {
			return
		}

		payment.AccountNumber = getPaymentLastFour(deNumber)

		payments = append(payments, payment)
	}

	return payments, nil
}

func GetPaymentById(c *gin.Context) {
	request := modules.PaymentRequest{}
	var (
		payment modules.Payment
		body    []byte
		err     error
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

	if payment, err = getPaymentByIdAndUserId(
		request.Id, request.UserId,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		payment.AccountNumber = getPaymentLastFour(payment.AccountNumber)
		payment.Code = ""

		c.IndentedJSON(http.StatusOK, payment)
	}
}

func GetPaymentByUserId(c *gin.Context) {
	request := modules.PaymentRequestForUser{}
	var (
		body     []byte
		payments []modules.Payment
		err      error
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

	if payments, err = getPaymentForUser(request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		c.IndentedJSON(http.StatusOK, payments)
	}
}

func CreatePayment(c *gin.Context) {
	query := `
		insert into user_payment (user_id, account_name, account_number,
			account_type, account_zip, code, expire_month, expire_year)
		values (?, ?, ?, ?, ?, ?, ?, ?)
	`
	request := modules.PaymentNew{}
	var (
		result   sql.Result
		insertId int64
		enNumber string
		enCode   string
		body     []byte
		err      error
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

	if enNumber, enCode, err = EncryptAccount(
		request.AccountType, request.AccountNumber, request.Code,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if result, err = config.DB.Exec(query,
		request.UserId, request.AccountName, enNumber, request.AccountType,
		request.AccountZip, enCode, request.ExpireMonth, request.ExpireYear,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if insertId, err = result.LastInsertId(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if payment, err := getPaymentByIdAndUserId(
		int32(insertId), request.UserId,
	); err != nil {
		c.IndentedJSON(http.StatusOK, int32(insertId))
	} else {
		c.IndentedJSON(http.StatusOK, payment)
	}
}

func UpdatePayment(c *gin.Context) {
	query := `
		update user_payment set account_name = ?, account_number = ?,
		account_type = ?, account_zip = ?, code = ?, expire_month = ?, expire_year = ?
		where id = ? and user_id = ?
	`
	request := modules.UpdatePayment{}
	var (
		result      sql.Result
		affectedRow int64
		enNumber    string
		enCode      string
		body        []byte
		err         error
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

	if enNumber, enCode, err = EncryptAccount(
		request.AccountType, request.AccountNumber, request.Code,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if result, err = config.DB.Exec(query,
		request.AccountName, enNumber, request.AccountType, request.AccountZip,
		enCode, request.ExpireMonth, request.ExpireYear, request.Id, request.UserId,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else if affectedRow, err = result.RowsAffected(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else if affectedRow <= 0 {
		c.IndentedJSON(http.StatusBadRequest, "Payment not found")
		c.Abort()
		return
	}

	if payment, err := getPaymentByIdAndUserId(
		request.Id, request.UserId,
	); err == nil {
		c.IndentedJSON(http.StatusOK, payment)
	} else {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}
}

func DeletePayment(c *gin.Context) {
	query := `
		delete from user_payment where id = ? and user_id = ?
	`
	request := modules.PaymentRequest{}
	var (
		result      sql.Result
		affectedRow int64
		body        []byte
		err         error
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

	if status, msg := checkPaymentStatus(
		request.Id, request.UserId,
	); status {
		c.IndentedJSON(http.StatusBadRequest, msg)
		c.Abort()
		return
	}

	if result, err = config.DB.Exec(
		query, request.Id, request.UserId,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else if affectedRow, err = result.RowsAffected(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else if affectedRow <= 0 {
		c.IndentedJSON(http.StatusBadRequest, "Payment not found")
		c.Abort()
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func checkPaymentStatus(id, userId int32) (bool, string) {
	var temp int32

	query := `
		select reserved from user_payment
		where id = ? and user_id = ?
	`

	if err := config.DB.QueryRow(
		query, id, userId,
	).Scan(&temp); err != nil {
		fmt.Println("Check Payment Status - Error - ", err.Error())
		return true, err.Error()
	} else if temp > 0 {
		return true, `
			This payment method cannot be deleted since you have one reservation on it.
		`
	}

	query = `
		select count(*)
		from user_service us
		inner join user_payment up on up.id = us.user_payment_id and up.user_id = us.user_id
		where up.id = ? and up.user_id = ? and us.pay = 0 and us.status = 'DONE'
	`

	if err := config.DB.QueryRow(
		query, id, userId,
	).Scan(&temp); err != nil {
		fmt.Println("Check Payment Status - Error - ", err.Error())
		return true, err.Error()
	} else if temp > 0 {
		return true, `
			This payment method cannot be deleted since you need to process your payment.
		`
	}

	return false, ""
}

func lockPayment(id, userId int32) {
	query := `
		update user_payment set reserved = reserved + 1 where id = ? and user_id = ?
	`

	if _, err := config.DB.Exec(
		query, id, userId,
	); err != nil {
		fmt.Println("Lock Pyment - Error - ", err)
	}
}

func unlockPayment(id, userId int32) {
	query := `
		update user_payment set reserved = reserved - 1 where id = ? and user_id = ?
	`

	if _, err := config.DB.Exec(
		query, id, userId,
	); err != nil {
		fmt.Println("Unlock Pyment - Error - ", err)
	}
}

func ProcessPayment(c *gin.Context) {
	query := `
		select up.id, us.estimated_price, up.account_number, up.account_zip,
				up.code, up.expire_month, up.expire_year, up.account_type
		from user_service us
		inner join user_payment up on up.id = us.user_payment_id
		where us.id = ? and us.user_id = ? and us.status = 'DONE' and us.pay = 0
	`
	request := modules.ProcessRequest{}
	process := struct {
		PaymentId     int32
		Price         float32
		Code          string
		Zip           string
		Year          string
		Month         string
		AccountNumber string
		AccountType   string
	}{}
	var (
		data []byte
		err  error
		tx   *braintree.Transaction
	)

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = config.DB.QueryRow(
		query, request.ServiceId, request.UserId,
	).Scan(
		&process.PaymentId, &process.Price, &process.AccountNumber,
		&process.Zip, &process.Code, &process.Month, &process.Year,
		&process.AccountType,
	); err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusOK, "UserService (payment) not found")
			return
		} else {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}
	}

	if process.AccountType == "CREDIT" {
		if process.AccountNumber, err = secures.DecryptCredit(
			process.AccountNumber,
		); err != nil {
			fmt.Println("Cannot process credit payment - number - ", err.Error())
			c.IndentedJSON(http.StatusBadRequest, "Cannot process payment now")
			c.Abort()
			return
		}

		if process.Code, err = secures.DecryptCreditCVV(
			process.Code,
		); err != nil {
			fmt.Println("Cannot process credit payment - code - ", err.Error())
			c.IndentedJSON(http.StatusBadRequest, "Cannot process payment now")
			c.Abort()
			return
		}
	} else {
		if process.AccountNumber, err = secures.DecryptDebit(
			process.AccountNumber,
		); err != nil {
			fmt.Println("Cannot process debit payment - number - ", err.Error())
			c.IndentedJSON(http.StatusBadRequest, "Cannot process payment now")
			c.Abort()
			return
		}

		if process.Code, err = secures.DecryptDebitPin(
			process.Code,
		); err != nil {
			fmt.Println("Cannot process debit payment - code - ", err.Error())
			c.IndentedJSON(http.StatusBadRequest, "Cannot process payment now")
			c.Abort()
			return
		}
	}

	fmt.Println("Process Payment ------ Start")
	fmt.Println("Decimal - ", int64(process.Price*100))
	fmt.Println("Number - ", process.AccountNumber)
	fmt.Println("CVV - ", process.Code)
	fmt.Println("ExpirationMonth - ", process.Month)
	fmt.Println("ExpirationYear - ", process.Year[2:])
	fmt.Println("Process Payment ------ End\n")

	if tx, err = config.BT.Transaction().Create(
		&braintree.Transaction{
			Type:   "sale",
			Amount: braintree.NewDecimal(int64(process.Price*100), 2),
			CreditCard: &braintree.CreditCard{
				Number:          process.AccountNumber,
				CVV:             process.Code,
				ExpirationMonth: process.Month,
				ExpirationYear:  process.Year[2:],
			},
		},
	); err != nil {
		fmt.Println("Error when processing payment - ", err.Error())
		c.IndentedJSON(http.StatusBadRequest, "Cannot process payment now")
		c.Abort()
		return
	}

	makeServicePay(request.UserId, request.ServiceId)

	fmt.Println("Transaction Info - ", tx)

	c.IndentedJSON(http.StatusOK, "OK")
}
