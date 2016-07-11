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

func FormatPhone(phone string) string {
	var buf bytes.Buffer
	var char string

	for _, c := range phone {
		char = string(c)
		if "0" <= char && char <= "9" {
			buf.WriteString(char)
		}
	}

	return buf.String()
}
