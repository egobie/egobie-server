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
