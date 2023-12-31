package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"siam/helper"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type JKelasRequest struct {
	IdKelas     int
	NamaKelas   string
	StatusKelas string
	Username    string
	ParamKey    string
	Method      string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JKelasResponse struct {
	IdKelas      int
	NamaKelas    string
	StatusKelas  int
	TanggalInput string
}

func Kelas(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "KELAS"

	var (
		bodyBytes    []byte
		XRealIp      string
		IP           string
		LogFile      string
		totalPage    float64
		totalRecords float64
	)

	jKelasRequest := JKelasRequest{}
	jKelasResponse := JKelasResponse{}
	jKelasResponses := []JKelasResponse{}

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
	LogFILE := LogFile + "Kelas_" + DateNow + ".log"
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
		returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jKelasRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jKelasRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jKelasRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jKelasRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {
			page := 0
			rowPage := 0

			usernameSession := jKelasRequest.Username
			paramKeySession := jKelasRequest.ParamKey
			method := jKelasRequest.Method
			idKelas := jKelasRequest.IdKelas
			namaKelas := jKelasRequest.NamaKelas
			statusKelas := jKelasRequest.StatusKelas
			page = jKelasRequest.Page
			rowPage = jKelasRequest.RowPage
			order := jKelasRequest.Order
			orderBy := jKelasRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(usernameSession, paramKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonKelas(jKelasResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(usernameSession, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, method, Path, IP, c)
				return
			}

			if method == "INSERT" {
				if namaKelas == "" {
					errorMessage := "Nama Kelas Tidak Boleh Kosong!"
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
					return
				} else {
					msgValidate := validateNamaKelas(namaKelas, db)
					if msgValidate != "OK" {
						returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
						helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, method, Path, IP, c)
						return
					}

					query := fmt.Sprintf("INSERT into siam_kelas(nama_kelas,status_kelas)values('%s',1)", namaKelas)
					if _, err = db.Exec(query); err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
						return
					}

					jKelasResponse.NamaKelas = namaKelas
					jKelasResponse.StatusKelas = 1
					jKelasResponse.TanggalInput = StartTimeStr

					errorMessage := "Sukses insert data!"
					returnDataJsonKelas(jKelasResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)

				}

			} else if method == "UPDATE" {
				querySet := ""
				if namaKelas != "" {
					querySet += " nama_kelas = '" + namaKelas + "'"

					msgValidate := validateNamaKelas(namaKelas, db)
					if msgValidate != "OK" {
						returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
						helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, method, Path, IP, c)
						return
					}
				}

				if idKelas == 0 {
					errorMessage := "Id kelas tidak boleh kosong!"
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
					return
				}

				msgValidate := validateIdKelas(idKelas, db)
				if msgValidate != "OK" {
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
					helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, method, Path, IP, c)
					return
				}

				if statusKelas != "" {
					if querySet != "" {
						querySet += " , "
					}
					iStatus, err := strconv.Atoi(statusKelas)
					if err == nil {
						querySet += fmt.Sprintf(" status_kelas = %d ", iStatus)
					} else {
						errorMessage := "Error convert variable, " + err.Error()
						returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
						return
					}

				}

				query1 := fmt.Sprintf(`update siam_kelas set %s where id_kelas = %d ;`, querySet, idKelas)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
					return
				}

				jKelasResponse.IdKelas = idKelas
				jKelasResponse.NamaKelas = namaKelas
				jKelasResponse.StatusKelas = 1
				jKelasResponse.TanggalInput = StartTimeStr

				jKelasResponses = append(jKelasResponses, jKelasResponse)
				errorMessage := "Sukses update data!"
				returnDataJsonKelas(jKelasResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)

			} else if method == "DELETE" {
				if idKelas == 0 {
					errorMessage := "Id kelas tidak boleh kosong!"
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
					return
				}

				msgValidate := validateIdKelas(idKelas, db)
				if msgValidate != "OK" {
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
					helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, method, Path, IP, c)
					return
				}

				query1 := fmt.Sprintf(`update siam_kelas set status_kelas = 0 where id_kelas = %d ;`, idKelas)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
					return
				}

				errorMessage := "Berhasil hapus data!"
				returnDataJsonKelas(jKelasResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)

			} else if method == "SELECT" {
				PageNow := (page - 1) * rowPage

				queryWhere := ""
				if namaKelas != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " nama_kelas LIKE '%" + namaKelas + "%' "
				}

				if statusKelas != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}
					iStatus, err := strconv.Atoi(statusKelas)
					if err == nil {
						queryWhere += fmt.Sprintf(" status_kelas = %d ", iStatus)
					} else {
						errorMessage := "Error convert variable, " + err.Error()
						returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
						return
					}
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}

				queryOrder := ""

				if orderBy != "" {
					queryOrder = " ORDER BY " + orderBy + " " + order
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_kelas %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(rowPage))

				query1 := fmt.Sprintf(`SELECT id_kelas, nama_kelas, status_kelas, tgl_input FROM siam_kelas %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, rowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jKelasResponse.IdKelas,
						&jKelasResponse.NamaKelas,
						&jKelasResponse.StatusKelas,
						&jKelasResponse.TanggalInput,
					)

					jKelasResponses = append(jKelasResponses, jKelasResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
						return
					}
				}

				returnDataJsonKelas(jKelasResponses, totalPage, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonKelas(jKelasResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, method, Path, IP, c)
				return
			}

		}
	}

}

func returnDataJsonKelas(jKelasResponse []JKelasResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jKelasResponse,
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

func validateIdKelas(id int, db *sql.DB) string {
	msg := "OK"

	var cntKelas int
	query1 := fmt.Sprintf(`select count(1) as cnt from siam_kelas where id_kelas = %d ;`, id)
	if err := db.QueryRow(query1).Scan(&cntKelas); err != nil {
		msg = "Error query, " + err.Error()
	}

	if msg == "OK" {
		if cntKelas == 0 {
			msg = "Data tidak ditemukan!"
		}
	}

	return msg
}

func validateNamaKelas(namaKelas string, db *sql.DB) string {
	msg := "OK"
	cntKelas := 0
	query := fmt.Sprintf("SELECT ifnull(count(1),0)cnt FROM siam_kelas WHERE status_kelas = 1 and nama_kelas = '%s'", namaKelas)
	if err := db.QueryRow(query).Scan(&cntKelas); err != nil {
		msg = fmt.Sprintf("Error running %q: %+v", query, err)
	}

	if msg == "OK" {
		if cntKelas > 0 {
			msg = "Sudah ada nama kelas yang sama!"
		}
	}

	return msg
}
