package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func CheckMenuAccess(Role string, RoleId string, Menu string, c *gin.Context) (string, string, string, string, string) {

	db := Connect(c)
	defer db.Close()

	IdMenu := 0
	query := fmt.Sprintf("SELECT id FROM siam_menu WHERE menu = '%s'", Menu)
	if err := db.QueryRow(query).Scan(&IdMenu); err != nil {
		errorMessageReturn := "Data tidak ditemukan"
		errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
		return "1", errorMessage, errorMessageReturn, "", ""
	}

	Read := ""
	Write := ""
	query2 := fmt.Sprintf("SELECT `read`, `write` FROM siam_role_menu_akses WHERE role_id = '%s' AND menu_id = %d", RoleId, IdMenu)
	if err := db.QueryRow(query2).Scan(&Read, &Write); err != nil {
		errorMessageReturn := "Data tidak ditemukan"
		errorMessage := fmt.Sprintf("Error running %q: %+v", query2, err)
		return "1", errorMessage, errorMessageReturn, "", ""
	}

	CntAccess := 0
	query1 := fmt.Sprintf("SELECT COUNT(*) FROM siam_menu_akses WHERE role = '%s' AND id_menu = %d", Role, IdMenu)
	if err := db.QueryRow(query1).Scan(&CntAccess); err != nil {
		errorMessageReturn := "Data tidak ditemukan"
		errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
		return "1", errorMessage, errorMessageReturn, "", ""
	}

	if CntAccess == 0 {
		errorMessageReturn := "Anda tidak diijinkan mengakses halaman ini"
		return "3", errorMessageReturn, errorMessageReturn, "", ""
	}

	return "", "", "", Read, Write
}
