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

type JMajorsRequest struct {
	Username  string
	ParamKey  string
	Method    string
	MajorID   string
	MajorName string
	Status    string
	Page      int
	RowPage   int
	OrderBy   string
	Order     string
}

type JMajorsResponse struct {
	MajorID      int
	MajorName    string
	Status       string
	TanggalInput string
}

func Majors(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "SUBJECTS"

	var (
		bodyBytes         []byte
		XRealIp           string
		IP                string
		LogFile           string
		totalPage         float64
		totalRecords      float64
		totalRecordsID    float64
		totalRecordsMajor float64
	)

	jMajorsRequest := JMajorsRequest{}
	jMajorsResponse := JMajorsResponse{}
	jMajorsResponses := []JMajorsResponse{}

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
	fmt.Println(err)
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
		returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jMajorsRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jMajorsRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jMajorsRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jMajorsRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jMajorsRequest.Username
			ParamKeySession := jMajorsRequest.ParamKey
			Username := jMajorsRequest.Username
			MajorID := jMajorsRequest.MajorID
			MajorName := strings.TrimSpace(jMajorsRequest.MajorName)
			Method := jMajorsRequest.Method
			Status := jMajorsRequest.Status
			Page = jMajorsRequest.Page
			RowPage = jMajorsRequest.RowPage
			Order := jMajorsRequest.Order
			OrderBy := jMajorsRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(UsernameSession, ParamKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonMajors(jMajorsResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(jMajorsRequest.Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			if Method == "INSERT" {
				queryWhere := ""
				if MajorName != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " upper(nama_jurusan)  = upper('" + MajorName + "') "
				} else {
					errorMessage := "Major Name can not be empty!"
					returnDataJsonMajors(jMajorsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}

				if Status == "" {
					errorMessage := "Status can not be empty!"
					returnDataJsonMajors(jMajorsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}
				// ---------- end of query where ----------

				totalRecords = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_jurusan %s", queryWhere)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecords > 0 {
						errorMessage := "Major " + MajorName + " is already existed"
						returnDataJsonMajors(jMajorsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					} else {
						queryInsert := fmt.Sprintf("INSERT INTO siam_jurusan ( nama_jurusan, status_jurusan, tgl_input) values (upper('%s'),'1',sysdate())", MajorName)
						if _, err := db.Exec(queryInsert); err != nil {
							errorMessage := "Error insert, " + err.Error()
							returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {
							successMessage := "Success to insert data Major!"
							returnDataJsonMajors(jMajorsResponses, totalPage, "0", "0", successMessage, successMessage, logData, c)
							return
						}
					}
				}

			} else if Method == "UPDATE" {
				querySet := ""
				// --------------- Query Where --------------
				if MajorID == "" {
					errorMessage := "Major ID can not be empty!"
					returnDataJsonMajors(jMajorsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}

				// --------------- End of query where -------

				// --------------- Query Set ----------------
				if MajorName != "" {
					if querySet != "" {
						querySet += " , "
					}

					querySet += " nama_jurusan  = upper('" + MajorName + "') "
				}

				if Status != "" {
					if querySet != "" {
						querySet += " , "
					}

					querySet += " status_jurusan  = '" + MajorName + "' "
				}

				if querySet != "" {
					querySet = " SET " + querySet
				}

				// -------- end of query for update ---------
				totalRecords = 0
				totalRecordsMajor = 0
				queryCekID := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_jurusan WHERE id_jurusan = '%s'", MajorID)

				if err := db.QueryRow(queryCekID).Scan(&totalRecordsID); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecordsID == 0 {
						errorMessage := "Major ID Not Found!"
						returnDataJsonMajors(jMajorsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					} else {
						totalRecordsMajor = 0
						queryCekMajors := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_jurusan WHERE upper(nama_jurusan) = upper('%s')", MajorName)

						if err := db.QueryRow(queryCekMajors).Scan(&totalRecordsMajor); err != nil {
							errorMessage := "Error query, " + err.Error()
							returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {

							if totalRecordsMajor > 0 {
								errorMessage := "Major " + MajorName + " is already existed"
								returnDataJsonMajors(jMajorsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
								helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
								return
							} else {

								queryUpdate := fmt.Sprintf("UPDATE siam_jurusan  %s WHERE id_jurusan = '%s' ", querySet, MajorID)
								if _, err := db.Exec(queryUpdate); err != nil {
									errorMessage := "Error update, " + err.Error()
									returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
									helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
									return
								} else {
									successMessage := "Success to update data Major!"
									returnDataJsonMajors(jMajorsResponses, totalPage, "0", "0", successMessage, successMessage, logData, c)
									return
								}

							}
						}
					}
				}

			} else if Method == "DELETE" {
				if MajorID == "" {
					errorMessage := "Major ID can not be null!"
					returnDataJsonMajors(jMajorsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}
				// ---------- end of query where ----------

				totalRecords = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_jurusan WHERE id_jurusan = '%s'", MajorID)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecords == 1 {

						queryInsert := fmt.Sprintf("DELETE FROM siam_jurusan WHERE id_jurusan = '%s'", MajorID)
						if _, err := db.Exec(queryInsert); err != nil {
							errorMessage := "Error insert, " + err.Error()
							returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {
							successMessage := "Success to delete data Major!"
							returnDataJsonMajors(jMajorsResponses, totalPage, "0", "0", successMessage, successMessage, logData, c)
							return
						}

					} else if totalRecords == 0 {
						errorMessage := "No Data Found To Delete"
						returnDataJsonMajors(jMajorsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					}

				}
			} else if Method == "SELECT" {

				PageNow := (Page - 1) * RowPage

				// ---------- start query where ----------
				queryWhere := ""
				if MajorName != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " upper(nama_jurusan) LIKE upper('%" + MajorName + "%') "
				}

				if Status != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " status_jurusan = '" + Status + "' "
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
						returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					} else {
						queryOrder = " ORDER BY " + OrderBy + " " + Order
					}
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_jurusan %s", queryWhere)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

				query1 := fmt.Sprintf(`SELECT id_jurusan, nama_jurusan, status_jurusan, tgl_input FROM siam_jurusan %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, RowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				for rows.Next() {
					err = rows.Scan(
						&jMajorsResponse.MajorID,
						&jMajorsResponse.MajorName,
						&jMajorsResponse.Status,
						&jMajorsResponse.TanggalInput,
					)

					jMajorsResponses = append(jMajorsResponses, jMajorsResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				returnDataJsonMajors(jMajorsResponses, totalPage, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonMajors(jMajorsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
		}
	}
}

func returnDataJsonMajors(jMajorsResponse []JMajorsResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jMajorsResponse,
			"TotalPage":  TotalPage,
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
