package helper

import (
	"strings"
)

func TrimReplace(txt_value string) string {

	text1 := strings.Trim(txt_value, " ")

	text2 := strings.ReplaceAll(text1, "'", "''")

	text3 := strings.ReplaceAll(text2, "\"", "\"")
	text4 := strings.ReplaceAll(text3, "\t", "")
	text5 := strings.ReplaceAll(text4, "\n", "")

	return text5
}