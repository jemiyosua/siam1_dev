package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func LogActivity(Username string, Page string, IP string, JsonRequest string, Method, Log string, LogStatus string, Role string, c *gin.Context) {
	db := Connect(c)
	defer db.Close()

	query := fmt.Sprintf("INSERT into siam_log_activity (username, page, ip, json_request, method, log, log_status, role, tgl_input) values ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', NOW())", Username, Page, IP, JsonRequest, Method, Log, LogStatus, Role)
	_, err := db.Exec(query)
	if err != nil {
		return
	}
}
