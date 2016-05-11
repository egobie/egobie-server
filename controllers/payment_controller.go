package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"fmt"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
	"github.com/egobie/egobie-server/secures"

	"github.com/gin-gonic/gin"
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

func getPaymentLastFour(accountNumber string) (string) {
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

	if checkPaymentStatus(request.Id, request.UserId) {
		c.IndentedJSON(http.StatusBadRequest, `
			This payment method cannot be deleted since you have one reservation on it.
		`)
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

func checkPaymentStatus(id, userId int32) bool {
	query := `
		select reserved from user_payment where id = ? and user_id = ?
	`
	var temp int32

	if err := config.DB.QueryRow(
		query, id, userId,
	).Scan(&temp); err != nil {
		fmt.Println("Check Payment Status - Error - ", err)
		return false
	} else {
		return temp > 0
	}
}

func lockPayment(id int32) {
	query := `
		update user_payment set reserved = reserved + 1 where id = ?
	`

	if _, err := config.DB.Exec(query, id); err != nil {
		fmt.Println("Lock Pyment - Error - ", err);
	}
}

func unlockPayment(id int32) {
	query := `
		update user_payment set reserved = reserved - 1 where id = ?
	`

	if _, err := config.DB.Exec(query, id); err != nil {
		fmt.Println("Unlock Pyment - Error - ", err);
	}
}
