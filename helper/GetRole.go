package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetRole(Username string, c *gin.Context) (string, string, string, string) {

	db := Connect(c)
	defer db.Close()

	RoleDB := ""
	query := fmt.Sprintf("SELECT role FROM siam_login WHERE username = '%s'", Username)
	if err := db.QueryRow(query).Scan(&RoleDB); err != nil {
		errorMessageReturn := "Data tidak ditemukan"
		errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
		return "1", errorMessage, errorMessageReturn, ""
	}

	return "", "", "", RoleDB
}
