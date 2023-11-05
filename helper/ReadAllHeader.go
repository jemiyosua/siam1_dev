package helper

import (
	"regexp"

	"github.com/gin-gonic/gin"
)

func ReadAllHeader(c *gin.Context) string {
	var allHeader string
	var cnt, cnt1 int

	cnt = 0
	allHeader = "\n\"Header=>"
	for k, vals := range c.Request.Header {
		if cnt == 0 {
			allHeader = allHeader + k + ":"
		} else {
			allHeader = allHeader + " | " + k + ":"
		}
		cnt1 = 0
		for _, v := range vals {
			if cnt1 == 0 {
				allHeader = allHeader + v
			} else {
				allHeader = allHeader + ";" + v
			}
			cnt1 = cnt + 1
		}
		cnt = cnt + 1
	}
	allHeader = allHeader + "\""

	rex := regexp.MustCompile(`\r?\n`)
	allHeader = rex.ReplaceAllString(allHeader, " ")
	return allHeader
}
