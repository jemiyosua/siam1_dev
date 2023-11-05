package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func SendLogError(UserName string, Page string, ErrorLog string, JsonRequest string, JsonResponse string, ErrorCode string, HeaderAuth string, MethodRequest string, Path string, IP string, c *gin.Context) {
	db := Connect(c)
	defer db.Close()

	query := fmt.Sprintf("INSERT INTO siam_log_error (username, page, error_log, json_request, json_response, error_code, header_auth, method_request, path, ip, tgl_input) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', NOW());", TrimReplace(UserName), TrimReplace(Page), TrimReplace(ErrorLog), TrimReplace(JsonRequest), TrimReplace(JsonResponse), TrimReplace(ErrorCode), TrimReplace(HeaderAuth), TrimReplace(MethodRequest), TrimReplace(Path), TrimReplace(IP))
	fmt.Println(query)
	_, err := db.Exec(query)
	if err != nil {
		return
	}
}