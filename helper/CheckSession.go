package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func CheckSession(Username string, ParamKey string, c *gin.Context) string {

	db := Connect(c)
	defer db.Close()

	ParamKeyDB := ""
	query := fmt.Sprintf("SELECT paramkey FROM siam_login_session WHERE username = '%s' AND paramkey = '%s' AND tgl_input >= NOW() LIMIT 1", Username, ParamKey)
	if err := db.QueryRow(query).Scan(&ParamKeyDB); err != nil {
		errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
		return errorMessage
	}

	if ParamKeyDB != "" {
		query1 := fmt.Sprintf("UPDATE siam_login_session SET tgl_input = (ADDTIME(NOW(), '0:20:0')) WHERE username = '%s' AND paramkey = '%s' AND tgl_input >= NOW() LIMIT 1", Username, ParamKey)
		_, err1 := db.Exec(query1)
		if err1 != nil {
			errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
			return errorMessage
		}
		return "1"
	}
	return "2"
}
