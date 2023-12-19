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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type JKelasActiveRequest struct {
	IdKelasActive int
	IdKelas       int
	NamaKelas     string
	TahunAjaran   string
	Semester      int
	JumlahSiswa   int
	ParamKey      string
	Method        string
	Username      string
	Page          int
	RowPage       int
	OrderBy       string
	Order         string
}

type JKelasActiveResponse struct {
	IdKelasActive int
	NamaKelas     string
	TahunAjaran   string
	Semester      int
	JumlahSiswa   int
	TanggalInput  string
}

func KelasActive(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "KELAS_ACTIVE"

	var (
		bodyBytes    []byte
		XRealIp      string
		IP           string
		LogFile      string
		totalPage    float64
		totalRecords float64
	)

	jKelasActiveRequest := JKelasActiveRequest{}
	jKelasActiveResponse := JKelasActiveResponse{}
	jKelasActiveResponses := []JKelasActiveResponse{}

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
	LogFILE := LogFile + "KelasActive_" + DateNow + ".log"
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
		returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jKelasActiveRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jKelasActiveRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jKelasActiveRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jKelasActiveRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {
			page := 0
			rowPage := 0

			idKelasActive := jKelasActiveRequest.IdKelasActive
			idKelas := jKelasActiveRequest.IdKelas
			namaKelas := jKelasActiveRequest.NamaKelas
			tahunAjaran := jKelasActiveRequest.TahunAjaran
			semester := jKelasActiveRequest.Semester
			jumlahSiswa := jKelasActiveRequest.JumlahSiswa
			usernameSession := jKelasActiveRequest.Username
			paramKeySession := jKelasActiveRequest.ParamKey
			method := jKelasActiveRequest.Method
			page = jKelasActiveRequest.Page
			rowPage = jKelasActiveRequest.RowPage
			order := jKelasActiveRequest.Order
			orderBy := jKelasActiveRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(usernameSession, paramKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(usernameSession, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}

			if method == "INSERT" {
				errorMessage := "OK"

				if idKelas == 0 {
					errorMessage = "Id Kelas tidak boleh kosong!"
				} else if tahunAjaran == "" {
					errorMessage = "Tahun Ajaran tidak boleh kosong!"
				} else if semester == 0 {
					errorMessage = "Semester tidak boleh kosong!"
				}

				if errorMessage == "OK" {

					namaKelas := ""
					query := fmt.Sprintf("SELECT nama_kelas from siam_kelas WHERE id_kelas = %d", idKelas)
					if err := db.QueryRow(query).Scan(&namaKelas); err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}

					query = fmt.Sprintf("INSERT into siam_kelas_active(id_kelas, nama_kelas, tahun_ajaran, semester, jumlah_siswa)values(%d,'%s','%s',%d,%d)", idKelas, namaKelas, tahunAjaran, semester, jumlahSiswa)
					if _, err = db.Exec(query); err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}

					currentTime := time.Now()
					currentTimeStr := currentTime.Format("01/02/2006 15:04:05")

					jKelasActiveResponse.JumlahSiswa = jumlahSiswa
					jKelasActiveResponse.NamaKelas = namaKelas
					jKelasActiveResponse.Semester = semester
					jKelasActiveResponse.TahunAjaran = tahunAjaran
					jKelasActiveResponse.TanggalInput = currentTimeStr

					jKelasActiveResponses = append(jKelasActiveResponses, jKelasActiveResponse)

					errorMessage = "Sukses insert data!"
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)

				} else {
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

			} else if method == "UPDATE" {
				querySet := ""

				if idKelasActive == 0 {
					errorMessage := "Id Kelas Active tidak boleh kosong!"
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				msgValidate := validateIdKelasAvtive(idKelasActive, db)
				if msgValidate != "OK" {
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
					helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				namaKelas := ""
				if idKelas > 0 {
					if querySet != "" {
						querySet += " , "
					}
					querySet += fmt.Sprintf(" id_kelas = '%d' ", idKelas)

					query := fmt.Sprintf("SELECT nama_kelas from siam_kelas WHERE id_kelas = %d", idKelas)
					if err := db.QueryRow(query).Scan(&namaKelas); err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}

					if namaKelas != "" {
						if querySet != "" {
							querySet += " , "
						}
						querySet += fmt.Sprintf(" nama_kelas = '%s' ", namaKelas)
					}
				}

				if tahunAjaran != "" {
					if querySet != "" {
						querySet += " , "
					}
					querySet += fmt.Sprintf(" tahun_ajaran = '%s' ", tahunAjaran)
				}

				if semester > 0 {
					if querySet != "" {
						querySet += " , "
					}
					querySet += fmt.Sprintf(" semester = %d ", semester)
				}

				if jumlahSiswa > 0 {
					if querySet != "" {
						querySet += " , "
					}
					querySet += fmt.Sprintf(" jumlah_siswa = %d ", jumlahSiswa)
				}

				query1 := fmt.Sprintf(`update siam_kelas_active set %s where id_kelas_active = %d ;`, querySet, idKelasActive)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				query1 = fmt.Sprintf(`select id_kelas_active,nama_kelas,tahun_ajaran,semester,jumlah_siswa, tgl_input from siam_kelas_active ska where id_kelas_active = %d`, idKelasActive)
				rows, err = db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jKelasActiveResponse.IdKelasActive,
						&jKelasActiveResponse.NamaKelas,
						&jKelasActiveResponse.Semester,
						&jKelasActiveResponse.TahunAjaran,
						&jKelasActiveResponse.JumlahSiswa,
						&jKelasActiveResponse.TanggalInput,
					)
				}

				jKelasActiveResponses = append(jKelasActiveResponses, jKelasActiveResponse)

				errorMessage := "Sukses update data!"
				returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)

			} else if method == "DELETE" {
				if idKelasActive > 0 {

					msgValidate := validateIdKelasAvtive(idKelasActive, db)

					if msgValidate == "OK" {
						query1 := fmt.Sprintf(`delete from siam_kelas_active where id_kelas_active = %d ;`, idKelasActive)
						rows, err := db.Query(query1)
						defer rows.Close()
						if err != nil {
							errorMessage := "Error query, " + err.Error()
							returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
							helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						}

						errorMessage := "Sukses delete data!"
						returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)
					} else {
						returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
						helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				} else {
					errorMessage := "Id kelas active tidak boleh kosong!"
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
			} else if method == "SELECT" {
				PageNow := (page - 1) * rowPage

				queryWhere := ""
				if namaKelas != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " nama_kelas LIKE '%" + namaKelas + "%' "
				}

				if tahunAjaran != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" tahun_ajaran = '%s' ", tahunAjaran)
				}

				if semester > 0 {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" semester = %d ", semester)
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

				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_kelas_active %s ;", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(rowPage))

				query1 := fmt.Sprintf(`select id_kelas_active,nama_kelas,tahun_ajaran,semester,jumlah_siswa,tgl_input from siam_kelas_active ska %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, rowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jKelasActiveResponse.IdKelasActive,
						&jKelasActiveResponse.NamaKelas,
						&jKelasActiveResponse.TahunAjaran,
						&jKelasActiveResponse.Semester,
						&jKelasActiveResponse.JumlahSiswa,
						&jKelasActiveResponse.TanggalInput,
					)

					jKelasActiveResponses = append(jKelasActiveResponses, jKelasActiveResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				errorMessage := "OK"
				returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonKelasActive(jKelasActiveResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

		}
	}
}

func returnDataJsonKelasActive(jKelasActiveResponse []JKelasActiveResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jKelasActiveResponse,
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

func validateIdKelasAvtive(id int, db *sql.DB) string {
	msg := "OK"

	var cntKelasActive int
	query1 := fmt.Sprintf(`select count(1) as cnt from siam_kelas_active where id_kelas_active = %d ;`, id)
	if err := db.QueryRow(query1).Scan(&cntKelasActive); err != nil {
		msg = "Error query, " + err.Error()
	}

	if msg == "OK" {
		if cntKelasActive == 0 {
			msg = "Data tidak ditemukan!"
		}
	}

	return msg
}
