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

type JSiswaRequest struct {
	Username string
	ParamKey string
	Method string
	NISN string
	Nama string
	JenisKelamin string
	NomorHP string
	Kelas string
	Status string
	Page int
	RowPage int
	OrderBy string
	Order string
}

type JSiswaResponse struct {
	NISN string
	Nama string
	JenisKelamin string
	TanggalLahir string
	Alamat string
	NomorHP string
	Kelas string
	Status string
	TanggalInput string
}

func Siswa(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "SISWA"

	var (
		bodyBytes   []byte
		XRealIp string
		IP      string
		LogFile string
		totalPage    float64
		totalRecords float64
	)

	jSiswaRequest := JSiswaRequest{}
	jSiswaResponse := JSiswaResponse{}
	jSiswaResponses := []JSiswaResponse{}

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
		returnDataJsonSiswa(jSiswaResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jSiswaRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonSiswa(jSiswaResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jSiswaRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jSiswaRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonSiswa(jSiswaResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jSiswaRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jSiswaRequest.Username
			ParamKeySession := jSiswaRequest.ParamKey
			Username := jSiswaRequest.Username
			Method := jSiswaRequest.Method
			Nama := jSiswaRequest.Nama
			JenisKelamin := jSiswaRequest.JenisKelamin
			Status := jSiswaRequest.Status
			Page = jSiswaRequest.Page
			RowPage = jSiswaRequest.RowPage
			Order := jSiswaRequest.Order
			OrderBy := jSiswaRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(UsernameSession, ParamKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonSiswa(jSiswaResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(jSiswaRequest.Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			if Method == "INSERT" {

			} else if Method == "UPDATE" {

			} else if Method == "DELETE" {

			} else if Method == "SELECT" {

				PageNow := (Page - 1) * RowPage

				// ---------- start query where ----------
				queryWhere := ""
				if Nama != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " nama_siswa LIKE '%" + Nama + "%' "
				}

				if JenisKelamin != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " jenis_kelamin = '" + JenisKelamin + "' "
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
					queryOrder = " ORDER BY " + OrderBy + " " + Order
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_siswa %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonSiswa(jSiswaResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

				query1 := fmt.Sprintf(`SELECT nisn, nama_siswa, jenis_kelamin, tanggal_lahir, alamat, nomor_hp, kelas, status_siswa, tgl_input FROM siam_siswa %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, RowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonSiswa(jSiswaResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				for rows.Next() {
					err = rows.Scan(
						&jSiswaResponse.NISN,
						&jSiswaResponse.Nama,
						&jSiswaResponse.JenisKelamin,
						&jSiswaResponse.TanggalLahir,
						&jSiswaResponse.Alamat,
						&jSiswaResponse.NomorHP,
						&jSiswaResponse.Kelas,
						&jSiswaResponse.Status,
						&jSiswaResponse.TanggalInput,
					)

					jSiswaResponses = append(jSiswaResponses, jSiswaResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonSiswa(jSiswaResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				returnDataJsonSiswa(jSiswaResponses, totalPage, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonSiswa(jSiswaResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
		}
	}
}

func returnDataJsonSiswa(jSiswaResponse []JSiswaResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
	if strings.Contains(ErrorMessage, "Error running") {
		ErrorMessage = "Error Execute data"
	}

	if ErrorCode == "504" {
		c.String(http.StatusUnauthorized, "")
	} else {
		currentTime := time.Now()
		currentTime1 := currentTime.Format("01/02/2006 15:04:05")
		
		c.PureJSON(http.StatusOK, gin.H{
			"ErrCode":      ErrorCode,
			"ErrMessage":   ErrorMessage,
			"DateTime": currentTime1,
			"Result": jSiswaResponse,
			"TotalPage": TotalPage,
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
