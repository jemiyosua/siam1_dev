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

type JSubjectsRequest struct {
	Username     string
	ParamKey     string
	Method       string
	SubjectID    string
	SubjectName  string
	SubjectClass string
	Status       string
	Page         int
	RowPage      int
	OrderBy      string
	Order        string
}

type JSubjectsResponse struct {
	SubjectID    int
	SubjectName  string
	SubjectClass string
	Status       string
	TanggalInput string
}

func Subjects(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "SUBJECTS"

	var (
		bodyBytes           []byte
		XRealIp             string
		IP                  string
		LogFile             string
		totalPage           float64
		totalRecords        float64
		totalRecordsID      float64
		totalRecordsSubject float64
	)

	jSubjectsRequest := JSubjectsRequest{}
	jSubjectsResponse := JSubjectsResponse{}
	jSubjectsResponses := []JSubjectsResponse{}

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
		returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jSubjectsRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jSubjectsRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jSubjectsRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jSubjectsRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jSubjectsRequest.Username
			ParamKeySession := jSubjectsRequest.ParamKey
			Username := jSubjectsRequest.Username
			SubjectID := jSubjectsRequest.SubjectID
			SubjectName := strings.TrimSpace(jSubjectsRequest.SubjectName)
			SubjectClass := jSubjectsRequest.SubjectClass
			Method := jSubjectsRequest.Method
			Status := jSubjectsRequest.Status
			Page = jSubjectsRequest.Page
			RowPage = jSubjectsRequest.RowPage
			Order := jSubjectsRequest.Order
			OrderBy := jSubjectsRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(UsernameSession, ParamKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonSubjects(jSubjectsResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(jSubjectsRequest.Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			if Method == "INSERT" {
				queryWhere := ""
				if SubjectName != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " upper(nama_mapel)  = upper('" + SubjectName + "') "
				} else {
					errorMessage := "Subject Name can not be empty!"
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}

				if Status == "" {
					errorMessage := "Status can not be empty!"
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}
				// ---------- end of query where ----------

				totalRecords = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_mata_pelajaran %s", queryWhere)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecords > 0 {
						errorMessage := "Subject " + SubjectName + " is already existed"
						returnDataJsonSubjects(jSubjectsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					} else {
						queryInsert := fmt.Sprintf("INSERT INTO siam_mata_pelajaran ( nama_mapel, kelas, status_mapel, tgl_input) values (upper('%s'), %s,'1',sysdate())", SubjectName, SubjectClass)
						if _, err := db.Exec(queryInsert); err != nil {
							errorMessage := "Error insert, " + err.Error()
							returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {
							successMessage := "Success to insert data Subject!"
							returnDataJsonSubjects(jSubjectsResponses, totalPage, "0", "0", successMessage, successMessage, logData, c)
							return
						}
					}
				}

			} else if Method == "UPDATE" {
				querySet := ""
				// --------------- Query Where --------------
				if SubjectID == "" {
					errorMessage := "Subject ID can not be empty!"
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}

				// --------------- End of query where -------

				// --------------- Query Set ----------------
				if SubjectName != "" {
					if querySet != "" {
						querySet += " , "
					}

					querySet += " nama_mapel  = upper('" + SubjectName + "') "
				}

				if Status != "" {
					if querySet != "" {
						querySet += " , "
					}

					querySet += " status_mapel  = '" + SubjectName + "' "
				}

				if SubjectClass != "" {
					if querySet != "" {
						querySet += " , "
					}

					querySet += " kelas  = '" + SubjectClass + "' "
				}

				if querySet != "" {
					querySet = " SET " + querySet
				}

				// -------- end of query for update ---------
				totalRecords = 0
				totalRecordsSubject = 0
				queryCekID := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_mata_pelajaran WHERE id_mapel = '%s'", SubjectID)

				if err := db.QueryRow(queryCekID).Scan(&totalRecordsID); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecordsID == 0 {
						errorMessage := "Subject ID Not Found!"
						returnDataJsonSubjects(jSubjectsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					} else {
						totalRecordsSubject = 0
						queryCekSubjects := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_mata_pelajaran WHERE upper(nama_mapel) = upper('%s') and id_mapel <> '%s'", SubjectName, SubjectID)

						if err := db.QueryRow(queryCekSubjects).Scan(&totalRecordsSubject); err != nil {
							errorMessage := "Error query, " + err.Error()
							returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {

							if totalRecordsSubject > 0 {
								errorMessage := "Subject " + SubjectName + " is already existed"
								returnDataJsonSubjects(jSubjectsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
								helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
								return
							} else {

								queryUpdate := fmt.Sprintf("UPDATE siam_mata_pelajaran  %s WHERE id_mapel = '%s' ", querySet, SubjectID)
								if _, err := db.Exec(queryUpdate); err != nil {
									errorMessage := "Error update, " + err.Error()
									returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
									helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
									return
								} else {
									successMessage := "Success to update data Subject!"
									returnDataJsonSubjects(jSubjectsResponses, totalPage, "0", "0", successMessage, successMessage, logData, c)
									return
								}

							}
						}
					}
				}

			} else if Method == "DELETE" {
				if SubjectID == "" {
					errorMessage := "Subject ID can not be null!"
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
					return
				}
				// ---------- end of query where ----------

				totalRecords = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_mata_pelajaran WHERE id_mapel = '%s'", SubjectID)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if totalRecords == 1 {

						queryInsert := fmt.Sprintf("DELETE FROM siam_mata_pelajaran WHERE id_mapel = '%s'", SubjectID)
						if _, err := db.Exec(queryInsert); err != nil {
							errorMessage := "Error insert, " + err.Error()
							returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {
							successMessage := "Success to delete data Subject!"
							returnDataJsonSubjects(jSubjectsResponses, totalPage, "0", "0", successMessage, successMessage, logData, c)
							return
						}

					} else if totalRecords == 0 {
						errorMessage := "No Data Found To Delete"
						returnDataJsonSubjects(jSubjectsResponses, totalPage, "3", "3", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "3", AllHeader, Method, Path, IP, c)
						return
					}

				}
			} else if Method == "SELECT" {

				PageNow := (Page - 1) * RowPage

				// ---------- start query where ----------
				queryWhere := ""
				if SubjectName != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " upper(nama_mapel) LIKE upper('%" + SubjectName + "%') "
				}

				if Status != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " status_mapel = '" + Status + "' "
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
						returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					} else {
						queryOrder = " ORDER BY " + OrderBy + " " + Order
					}
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_mata_pelajaran %s", queryWhere)

				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

				query1 := fmt.Sprintf(`SELECT id_mapel, nama_mapel, kelas,status_mapel, tgl_input FROM siam_mata_pelajaran %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, RowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				for rows.Next() {
					err = rows.Scan(
						&jSubjectsResponse.SubjectID,
						&jSubjectsResponse.SubjectName,
						&jSubjectsResponse.SubjectClass,
						&jSubjectsResponse.Status,
						&jSubjectsResponse.TanggalInput,
					)

					jSubjectsResponses = append(jSubjectsResponses, jSubjectsResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				returnDataJsonSubjects(jSubjectsResponses, totalPage, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonSubjects(jSubjectsResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
		}
	}
}

func returnDataJsonSubjects(jSubjectsResponse []JSubjectsResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
	if strings.Contains(ErrorMessage, "Error running") {
		ErrorMessage = "Error Execute data"
	}

	if ErrorCode == "504" {
		c.String(http.StatusUnauthorized, "")
	} else {
		currentTime := time.Now()
		currentTime1 := currentTime.Format("01/02/2006 15:04:05")

		c.PureJSON(http.StatusOK, gin.H{
			"ErrCode":    ErrorCode,
			"ErrMessage": ErrorMessage,
			"DateTime":   currentTime1,
			"Result":     jSubjectsResponse,
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
