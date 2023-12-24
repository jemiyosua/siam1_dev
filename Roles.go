package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"siam/helper"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type JRolesRequest struct {
	Username string
	ParamKey string
	Method   string
	RoleID   int
	RoleName string
	Status   string
	Page     int
	RowPage  int
	OrderBy  string
	Order    string
}

type JRolesResponse struct {
	RoleID       int
	RoleName     string
	Status       string
	TanggalInput string
}

func Roles(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "ROLES"

	var (
		bodyBytes         []byte
		XRealIp           string
		IP                string
		LogFile           string
		totalPage         float64
		totalRecords      float64
		totalRecordsID    float64
		totalRecordsRoles float64
	)

	jRolesRequest := JRolesRequest{}
	jRolesResponse := JRolesResponse{}
	jRolesResponses := []JRolesResponse{}

	AllHeader := helper.ReadAllHeader(c)
	LogFile = os.Getenv("LOGFILE")
	Method := c.Request.Method
	Path := c.Request.URL.EscapedPath()

	// ---------- start get ip ----------
	if Values, _ := c.Request.Header["X-Real-Ip"]; len(Values) > 0 {
		XRealIp = Values[0]
	}

	if XRealIp != "" {
		IP = XRealIp
	} else {
		IP = c.ClientIP()
	}
	// ---------- end of get ip ----------

	// ---------- start log file ----------
	DateNow := StartTime.Format("2006-01-02")
	LogFILE := LogFile + "Login_" + DateNow + ".log"
	file, err := os.OpenFile(LogFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	// ---------- end of log file ----------

	// ------ start body json validation ------
	if c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	bodyString := string(bodyBytes)

	bodyJson := helper.TrimReplace(string(bodyString))
	logData := StartTimeStr + "~" + IP + "~" + Method + "~" + Path + "~" + AllHeader + "~"
	rex := regexp.MustCompile(`\r?\n`)
	logData = logData + rex.ReplaceAllString(bodyJson, "") + "~"

	if string(bodyString) == "" {
		errorMessage := "Error, Body is empty"
		returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jRolesRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jRolesRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jRolesRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jRolesRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jRolesRequest.Username
			ParamKeySession := jRolesRequest.ParamKey
			Username := jRolesRequest.Username
			RoleID := jRolesRequest.RoleID
			RoleName := strings.TrimSpace(jRolesRequest.RoleName)
			Method := jRolesRequest.Method
			Status := jRolesRequest.Status
			Page = jRolesRequest.Page
			RowPage = jRolesRequest.RowPage
			Order := jRolesRequest.Order
			OrderBy := jRolesRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(UsernameSession, ParamKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(jRolesRequest.Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			if Method == "INSERT" {
				queryWhere := ""
				if RoleName != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " upper(nama_role)  = upper('" + RoleName + "') "
				} else {
					errorMessage := "Role Name can not be empty!"
					returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}
				// ---------- end of query where ----------

				totalRecords = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_role %s", queryWhere)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecords > 0 {
						errorMessage := "Role " + RoleName + " is already existed"
						returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					} else {
						Status = "1"
						queryInsert := fmt.Sprintf("INSERT INTO siam_role ( nama_role, status, tgl_input) values ('%s','%s',sysdate())", RoleName, Status)
						if _, err := db.Exec(queryInsert); err != nil {
							errorMessage := "Error insert, " + err.Error()
							returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						}
						// else {
						// 	successMessage := "Success to insert data Role!"
						// 	returnDataJsonRoles(jRolesResponses, totalPage, "0", "0", successMessage, successMessage, logData, c)
						// 	return
						// }

						// update 11/12/2023 - J
						successMessage := "Success to insert data Role!"
						returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "0", "0", successMessage, successMessage, logData, c)
						return
					}
				}

			} else if Method == "UPDATE" {
				querySet := ""
				// --------------- Query Where --------------
				if RoleID == 0 {
					errorMessage := "Role ID can not be empty!"
					returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}

				// --------------- End of query where -------

				// --------------- Query Set ----------------
				if RoleName != "" {
					if querySet != "" {
						querySet += " , "
					}

					querySet += " nama_role  = upper('" + RoleName + "') "
				}

				if Status != "" {
					if querySet != "" {
						querySet += " , "
					}

					querySet += " status  = '" + Status + "' "
				}

				if querySet != "" {
					querySet = " SET " + querySet
				}

				// -------- end of query for update ---------
				totalRecords = 0
				totalRecordsRoles = 0
				queryCekID := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_role WHERE id = %d", RoleID)
				if err := db.QueryRow(queryCekID).Scan(&totalRecordsID); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecordsID == 0 {
						errorMessage := "Role ID Not Found!"
						returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					} else {
						totalRecordsRoles = 0
						queryCekRoles := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_role WHERE upper(nama_role) = upper('%s')", RoleName)
						if err := db.QueryRow(queryCekRoles).Scan(&totalRecordsRoles); err != nil {
							errorMessage := "Error query, " + err.Error()
							returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {

							if totalRecordsRoles > 0 {
								errorMessage := "Role " + RoleName + " is already existed"
								returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "3", "3", errorMessage, errorMessage, logData, c)
								helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
								return
							} else {

								queryUpdate := fmt.Sprintf("UPDATE siam_role %s WHERE id = %d ", querySet, RoleID)
								if _, err := db.Exec(queryUpdate); err != nil {
									errorMessage := "Error update, " + err.Error()
									returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
									helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
									return
								}
								// else {
								// 	successMessage := "Success to update data Role!"
								// 	returnDataJsonRoles(jRolesResponses, totalPage, "0", "0", successMessage, successMessage, logData, c)
								// 	return
								// }

								successMessage := "Success to update data Role!"
								returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "0", "0", successMessage, successMessage, logData, c)
								return
							}
						}
					}
				}

			} else if Method == "DELETE" {
				if RoleID == 0 {
					errorMessage := "Role ID can not be null!"
					returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}
				// ---------- end of query where ----------

				totalRecords = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_role WHERE id = %d", RoleID)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecords == 1 {

						queryInsert := fmt.Sprintf("DELETE FROM siam_role WHERE id = %d", RoleID)
						if _, err := db.Exec(queryInsert); err != nil {
							errorMessage := "Error insert, " + err.Error()
							returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {
							successMessage := "Success to delete data Role!"
							returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "0", "0", successMessage, successMessage, logData, c)
							return
						}

					} else if totalRecords == 0 {
						errorMessage := "No Data Found To Delete"
						returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					}

				}
			} else if Method == "SELECT" {

				PageNow := (Page - 1) * RowPage

				// ---------- start query where ----------
				queryWhere := ""
				if RoleID != 0 {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" id = %d ", RoleID)
				}

				if RoleName != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " upper(nama_role) LIKE upper('%" + RoleName + "%') "
				}

				if Status != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " status = '" + Status + "' "
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}
				// ---------- end of query where ----------

				queryOrder := ""

				if OrderBy == "" {
					queryOrder = " ORDER BY tgl_input desc"
				} else {
					if Order == "" {
						errorMessage := "Order tidak boleh kosong"
						returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					} else {
						queryOrder = " ORDER BY " + OrderBy + " " + Order
					}
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_role %s", queryWhere)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

				query1 := fmt.Sprintf(`SELECT id, nama_role, status, tgl_input FROM siam_role %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, RowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				for rows.Next() {
					err = rows.Scan(
						&jRolesResponse.RoleID,
						&jRolesResponse.RoleName,
						&jRolesResponse.Status,
						&jRolesResponse.TanggalInput,
					)

					jRolesResponses = append(jRolesResponses, jRolesResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonRoles(jRolesResponses, totalPage, totalRecords, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
		}
	}
}

func returnDataJsonRoles(jRolesResponse []JRolesResponse, TotalPage float64, TotalData float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
	if strings.Contains(ErrorMessage, "Error running") {
		ErrorMessage = "Error Execute data"
	}

	if ErrorCode == "504" {
		c.String(http.StatusUnauthorized, "")
	} else {
		currentTime := time.Now()
		currentTime1 := currentTime.Format("01/02/2006 15:04:05")

		c.PureJSON(http.StatusOK, gin.H{
			"ErrCode":    ErrorCodeReturn,
			"ErrMessage": ErrorMessageReturn,
			"DateTime":   currentTime1,
			"Result":     jRolesResponse,
			"TotalPage":  TotalPage,
			"TotalData":  TotalData,
		})
	}

	startTime := time.Now()

	rex := regexp.MustCompile(`\r?\n`)
	endTime := time.Now()
	codeError := "200"

	diff := endTime.Sub(startTime)

	logDataNew := rex.ReplaceAllString(logData+codeError+"~"+endTime.String()+"~"+diff.String()+"~"+ErrorMessage, "")
	log.Println(logDataNew)

	runtime.GC()

	return
}
