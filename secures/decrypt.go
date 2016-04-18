package secures

func DecryptPassword(text string) (string, error) {
	return decrypt(text, PASSWORD_KEY)
}

func DecryptCredit(text string) (string, error) {
	return decrypt(text, CREDIT_KEY)
}

func DecryptCreditCVV(text string) (string, error) {
	return decrypt(text, CREDIT_CVV_KEY)
}

func DecryptDebit(text string) (string, error) {
	return decrypt(text, DEBIT_KEY)
}

func DecryptDebitPin(text string) (string, error) {
	return decrypt(text, DEBIT_PIN_KEY)
}
