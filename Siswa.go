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
	Username     string
	ParamKey     string
	Method       string
	NISN         string
	Nama         string
	JenisKelamin string
	TanggalLahir string
	Alamat       string
	NomorHP      string
	Kelas        string
	Status       string
	Page         int
	RowPage      int
	OrderBy      string
	Order        string
}

type JSiswaResponse struct {
	NISN         string
	Nama         string
	JenisKelamin string
	TanggalLahir string
	Alamat       string
	NomorHP      string
	Kelas        string
	Status       string
	TanggalInput string
}

func Siswa(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "SISWA"
	PageMenu := "Siswa"

	var (
		bodyBytes    []byte
		XRealIp      string
		IP           string
		LogFile      string
		totalPage    float64
		totalRecords float64
		ErrorCodeAccess string
		ErrorMessageAccess string
		ErrorMessageReturnAccess string
		Read string
		Write string
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
		returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jSiswaRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jSiswaRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jSiswaRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jSiswaRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jSiswaRequest.Username
			ParamKeySession := jSiswaRequest.ParamKey
			Username := jSiswaRequest.Username
			Method := jSiswaRequest.Method

			NISN := jSiswaRequest.NISN
			Nama := jSiswaRequest.Nama
			JenisKelamin := jSiswaRequest.JenisKelamin
			TanggalLahir := jSiswaRequest.TanggalLahir
			Alamat := jSiswaRequest.Alamat
			NomorHP := jSiswaRequest.NomorHP
			Kelas := jSiswaRequest.Kelas
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
				returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			// ---------- start cek akses role ----------
			ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, Role, RoleId := helper.GetRole(Username, c)
			if ErrorCodeGetRole != "" {
				returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, ErrorCodeGetRole, ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, logData, c)
				helper.SendLogError(Username, PageGo, ErrorMessageGetRole, "", "", ErrorCodeGetRole, AllHeader, Method, Path, IP, c)
				return
			}

			ErrorCodeAccess, ErrorMessageAccess, ErrorMessageReturnAccess, Read, Write = helper.CheckMenuAccess(Role, RoleId, PageMenu, c)
			if ErrorCodeAccess != "" {
				returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, ErrorCodeAccess, ErrorCodeAccess, ErrorMessageAccess, ErrorMessageReturnAccess, logData, c)
				helper.SendLogError(Username, PageGo, ErrorMessageAccess, "", "", ErrorCodeAccess, AllHeader, Method, Path, IP, c)
				return
			}
			// ---------- end of cek akses role ----------

			if Method == "INSERT" {

				ErrorMessage := ""
				if Nama == "" {
					ErrorMessage = "Nama tidak boleh kosong"
				} else if JenisKelamin == "" {
					ErrorMessage = "JenisKelamin tidak boleh kosong"
				} else if TanggalLahir == "" {
					ErrorMessage = "TanggalLahir tidak boleh kosong"
				} else if Alamat == "" {
					ErrorMessage = "Alamat tidak boleh kosong"
				} else if NomorHP == "" {
					ErrorMessage = "NomorHP tidak boleh kosong"
				} else if Kelas == "" {
					ErrorMessage = "Kelas tidak boleh kosong"
				}

				if ErrorMessage != "" {
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				currentTime := time.Now()
				currentTime1 := currentTime.Format("01/02/2006 15:04:05")

				TimeSplit := strings.Split(currentTime1, " ")
				TimeSplit1 := TimeSplit[0]
				TimeSplit2 := TimeSplit[1]
				TimeSplit1Replace := strings.Replace(TimeSplit1, "/", "", -1)
				TimeSplit2Replace := strings.Replace(TimeSplit2, ":", "", -1)
				CreateNISN := TimeSplit1Replace + TimeSplit2Replace

				query := fmt.Sprintf("INSERT INTO siam_siswa (nisn, nama_siswa, jenis_kelamin, tanggal_lahir, alamat, nomor_hp, kelas, status_siswa, tgl_input) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', NOW())", CreateNISN, Nama, JenisKelamin, TanggalLahir, Alamat, NomorHP, Kelas, Status)
				_, err1 := db.Exec(query)
				if err1 != nil {
					errorMessageReturn := "Gagal INSERT ke tabel siam_siswa"
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err1)
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessageReturn, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				Log := "Berhasil insert data siswa"
				helper.LogActivity(Username, PageGo, IP, bodyString, Method, Log, "Sukses", Role, c)

				c.JSON(http.StatusOK, gin.H{
					"ErrorCode":    "0",
					"ErrorMessage": "",
					"DateTime":     currentTime1,
					"Method":       Method,
					"UserName":     Username,
					"Result":       "ok",
				})
				return

			} else if Method == "UPDATE" {

				ErrorMessage := ""
				queryUpdate := ""
				if NISN == "" {
					ErrorMessage = "NISN cannot null"
				}

				if ErrorMessage != "" {
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "3", "3", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jSiswaRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				CntNISN := 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_siswa WHERE nisn = '%s'", NISN)
				if err := db.QueryRow(query).Scan(&CntNISN); err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if CntNISN == 0 {
					ErrorMessage := "NISN tidak ditemukan"
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jSiswaRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				CntNama := 0
				query2 := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_siswa WHERE nama_siswa = '%s'", Nama)
				if err := db.QueryRow(query2).Scan(&CntNama); err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query2, err)
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if CntNama > 0 {
					ErrorMessage := fmt.Sprintf("%s already exist", Nama)
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jSiswaRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if Nama != "" {
					queryUpdate += fmt.Sprintf(" , nama_siswa = '%s' ", Nama)
				}

				if JenisKelamin != "" {
					queryUpdate += fmt.Sprintf(" , jenis_kelamin = '%s' ", JenisKelamin)
				}

				if TanggalLahir != "" {
					queryUpdate += fmt.Sprintf(" , tanggal_lahir = '%s' ", TanggalLahir)
				}

				if Alamat != "" {
					queryUpdate += fmt.Sprintf(" , alamat = '%s' ", Alamat)
				}

				if NomorHP != "" {
					queryUpdate += fmt.Sprintf(" , nomor_hp = '%s' ", NomorHP)
				}

				if Kelas != "" {
					queryUpdate += fmt.Sprintf(" , kelas = '%s' ", Kelas)
				}

				query1 := fmt.Sprintf("UPDATE siam_siswa SET tgl_input = NOW() %s WHERE nisn = '%s'", queryUpdate, NISN)
				_, err1 := db.Exec(query1)
				if err1 != nil {
					errorMessageReturn := "Gagal UPDATE ke tabel siam_siswa"
					errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessageReturn, logData, c)
					helper.SendLogError(jSiswaRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				currentTime := time.Now()
				currentTime1 := currentTime.Format("01/02/2006 15:04:05")

				Log := "Berhasil update data siswa"
				helper.LogActivity(Username, PageGo, IP, bodyString, Method, Log, "Sukses", Role, c)

				c.JSON(http.StatusOK, gin.H{
					"ErrorCode":    "0",
					"ErrorMessage": "",
					"DateTime":     currentTime1,
					"Method":       Method,
					"UserName":     Username,
					"Result":       "ok",
				})
				return

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
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

				query1 := fmt.Sprintf(`SELECT nisn, nama_siswa, jenis_kelamin, tanggal_lahir, alamat, nomor_hp, kelas, status_siswa, tgl_input FROM siam_siswa %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, RowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
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
						returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonSiswa(jSiswaResponses, totalPage, Read, Write, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
		}
	}
}

func returnDataJsonSiswa(jSiswaResponse []JSiswaResponse, TotalPage float64, Read string, Write string, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jSiswaResponse,
			"TotalPage":  TotalPage,
			"Read":  Read,
			"Wrtie":  Write,
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
