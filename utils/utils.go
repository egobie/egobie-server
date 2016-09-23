package utils

import (
	"bytes"
	"strconv"
)

func ToStringList(ints []int32) string {
	var buffer bytes.Buffer

	for i, val := range ints {
		if i != 0 {
			buffer.WriteString(",")
		}

		buffer.WriteString(strconv.Itoa(int(val)))
	}

	return buffer.String()
}

func Contains(arr []int32, target int32) bool {
	if len(arr) == 0 {
		return false
	}

	for _, v := range arr {
		if v == target {
			return true
		}
	}

	return false
}

func FormatPhone(phone string) string {
	return formatNumber(phone)
}

func FormatCardNumber(card string) string {
	return formatNumber(card)
}

func FormatZipcode(zip string) string {
	return formatNumber(zip)
}

func formatNumber(num string) string {
	var buf bytes.Buffer
	var char string

	for _, c := range num {
		char = string(c)
		if "0" <= char && char <= "9" {
			buf.WriteString(char)
		}
	}

	return buf.String()
}
