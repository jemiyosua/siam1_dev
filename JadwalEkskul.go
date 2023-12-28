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

type JadwalEkskulRequest struct {
	Id           int
	NamaEkskul   string
	TahunAjaran  string
	Semester     int
	Hari         string
	Jam          string
	NamaPengajar string
	Tempat       string
	Status       string
	Username     string
	ParamKey     string
	Method       string
	Page         int
	RowPage      int
	OrderBy      string
	Order        string
}

type JadwalEkskulResponse struct {
	Id           int
	NamaEkskul   string
	TahunAjaran  string
	Semester     int
	Hari         string
	Jam          string
	NamaPengajar string
	Tempat       string
	Status       int
	TanggalInput string
}

func JadwalEkskul(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "JADWAL_EKSKUL"

	var (
		bodyBytes    []byte
		XRealIp      string
		IP           string
		LogFile      string
		totalPage    float64
		totalRecords float64
	)

	jadwalEkskulRequest := JadwalEkskulRequest{}
	jadwalEkskulResponse := JadwalEkskulResponse{}
	jadwalEkskulResponses := []JadwalEkskulResponse{}

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
	LogFILE := LogFile + "JadwalEkskul_" + DateNow + ".log"
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
		returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jadwalEkskulRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jadwalEkskulRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	// if helper.ValidateHeader(bodyString, c) {
	if err := c.ShouldBindJSON(&jadwalEkskulRequest); err != nil {
		errorMessage := "Error, Bind Json Data"
		returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jadwalEkskulRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	} else {
		page := 0
		rowPage := 0

		id := jadwalEkskulRequest.Id
		namaEkskul := jadwalEkskulRequest.NamaEkskul
		tahunAjaran := jadwalEkskulRequest.TahunAjaran
		semester := jadwalEkskulRequest.Semester
		hari := jadwalEkskulRequest.Hari
		jam := jadwalEkskulRequest.Jam
		namaPengajar := jadwalEkskulRequest.NamaPengajar
		tempat := jadwalEkskulRequest.Tempat
		status := jadwalEkskulRequest.Status

		usernameSession := jadwalEkskulRequest.Username
		paramKeySession := jadwalEkskulRequest.ParamKey
		method := jadwalEkskulRequest.Method
		page = jadwalEkskulRequest.Page
		rowPage = jadwalEkskulRequest.RowPage
		order := jadwalEkskulRequest.Order
		orderBy := jadwalEkskulRequest.OrderBy

		// ------ start check session paramkey ------
		checkAccessVal := helper.CheckSession(usernameSession, paramKeySession, c)
		if checkAccessVal != "1" {
			checkAccessValErrorMsg := checkAccessVal
			checkAccessValErrorMsgReturn := "Session Expired"
			returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
			helper.SendLogError(usernameSession, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, method, Path, IP, c)
			return
		}

		if method == "INSERT" {
			if namaEkskul == "" {
				errorMessage := "Nama Ekskul Tidak Boleh Kosong!"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			if tahunAjaran == "" {
				errorMessage := "Tahun Ajaran Tidak Boleh Kosong!"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			if semester == 0 {
				errorMessage := "Semester Tidak Boleh Kosong / harus > 0"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			if hari == "" {
				errorMessage := "Hari Tidak Boleh Kosong!"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			if jam == "" {
				errorMessage := "Jam Tidak Boleh Kosong!"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			if namaPengajar == "" {
				errorMessage := "Nama Pengajar Tidak Boleh Kosong!"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			if tempat == "" {
				errorMessage := "Tempat Tidak Boleh Kosong!"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			iStatus := 1
			if status != "" {
				iStatus, err = strconv.Atoi(status)
				if err != nil {
					errorMessage := "Error convert variable, " + err.Error()
					returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
			}

			msgValidate := validateJadwalEkskul(tahunAjaran, semester, hari, tempat, jam, db)
			if msgValidate != "OK" {
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
				helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			msgValidate = validateNamaPengajar(tahunAjaran, semester, hari, namaPengajar, jam, db)
			if msgValidate != "OK" {
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
				helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			query := fmt.Sprintf("insert into siam_jadwal_ekskul(nama_ekskul, tahun_ajaran, semester, hari, jam, nama_pengajar, tempat, status) values('%s', '%s', %d, '%s', '%s', '%s', '%s', %d)",
				namaEkskul, tahunAjaran, semester, hari, jam, namaPengajar, tempat, iStatus)
			if _, err = db.Exec(query); err != nil {
				errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			jadwalEkskulResponse.NamaEkskul = namaEkskul
			jadwalEkskulResponse.TahunAjaran = tahunAjaran
			jadwalEkskulResponse.Semester = semester
			jadwalEkskulResponse.Hari = hari
			jadwalEkskulResponse.Jam = jam
			jadwalEkskulResponse.NamaPengajar = namaPengajar
			jadwalEkskulResponse.Tempat = tempat
			jadwalEkskulResponse.Status = iStatus
			jadwalEkskulResponse.TanggalInput = StartTimeStr

			jadwalEkskulResponses := append(jadwalEkskulResponses, jadwalEkskulResponse)

			errorMessage := "Sukses insert data!"
			returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)

		} else if method == "UPDATE" {
			if id == 0 {
				errorMessage := "Id kelas tidak boleh kosong!"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			msgValidate := validateIdJadwalEkskul(id, db)
			if msgValidate != "OK" {
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
				helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			querySet := ""
			if namaEkskul != "" {
				querySet += " nama_ekskul = '" + namaEkskul + "'"
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

			if hari != "" {
				if querySet != "" {
					querySet += " , "
				}
				querySet += fmt.Sprintf(" hari = '%s' ", hari)
			}

			if jam != "" {
				if querySet != "" {
					querySet += " , "
				}
				querySet += fmt.Sprintf(" jam = '%s' ", jam)
			}

			if namaPengajar != "" {
				if querySet != "" {
					querySet += " , "
				}
				querySet += fmt.Sprintf(" nama_pengajar = '%s' ", namaPengajar)
			}

			if tempat != "" {
				if querySet != "" {
					querySet += " , "
				}
				querySet += fmt.Sprintf(" tempat = '%s' ", tempat)
			}

			if status != "" {
				if querySet != "" {
					querySet += " , "
				}
				iStatus, err := strconv.Atoi(status)
				if err == nil {
					querySet += fmt.Sprintf(" status = %d ", iStatus)
				} else {
					errorMessage := "Error convert variable, " + err.Error()
					returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

			}

			query1 := fmt.Sprintf(`update siam_jadwal_ekskul set %s where id = %d ;`, querySet, id)
			rows, err := db.Query(query1)
			defer rows.Close()
			if err != nil {
				errorMessage := "Error query, " + err.Error()
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			query1 = fmt.Sprintf(`select id, nama_ekskul, tahun_ajaran, semester, hari, jam, nama_pengajar, tempat, status, tgl_input from siam_jadwal_ekskul where id = %d`, id)
			rows, err = db.Query(query1)
			defer rows.Close()
			if err != nil {
				errorMessage := "Error query, " + err.Error()
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			for rows.Next() {
				err = rows.Scan(
					&jadwalEkskulResponse.Id,
					&jadwalEkskulResponse.NamaEkskul,
					&jadwalEkskulResponse.TahunAjaran,
					&jadwalEkskulResponse.Semester,
					&jadwalEkskulResponse.Hari,
					&jadwalEkskulResponse.Jam,
					&jadwalEkskulResponse.NamaPengajar,
					&jadwalEkskulResponse.Tempat,
					&jadwalEkskulResponse.Status,
					&jadwalEkskulResponse.TanggalInput,
				)
			}

			jadwalEkskulResponses = append(jadwalEkskulResponses, jadwalEkskulResponse)

			errorMessage := "Sukses update data!"
			returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)

		} else if method == "DELETE" {
			if id == 0 {
				errorMessage := "Id kelas tidak boleh kosong!"
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			msgValidate := validateIdJadwalEkskul(id, db)
			if msgValidate != "OK" {
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", msgValidate, msgValidate, logData, c)
				helper.SendLogError(usernameSession, PageGo, msgValidate, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			query1 := fmt.Sprintf(`update siam_jadwal_ekskul set status = 0 where id = %d ;`, id)
			rows, err := db.Query(query1)
			defer rows.Close()
			if err != nil {
				errorMessage := "Error query, " + err.Error()
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			errorMessage := "Berhasil hapus data!"
			returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)

		} else if method == "SELECT" {
			PageNow := (page - 1) * rowPage

			queryWhere := ""
			if namaEkskul != "" {
				if queryWhere != "" {
					queryWhere += " AND "
				}

				queryWhere += " nama_ekskul LIKE '%" + namaEkskul + "%' "
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

			if hari != "" {
				if queryWhere != "" {
					queryWhere += " AND "
				}

				queryWhere += fmt.Sprintf(" hari = '%s' ", hari)
			}

			if jam != "" {
				if queryWhere != "" {
					queryWhere += " AND "
				}

				queryWhere += fmt.Sprintf(" jam = '%s' ", jam)
			}

			if namaPengajar != "" {
				if queryWhere != "" {
					queryWhere += " AND "
				}

				queryWhere += " nama_pengajar LIKE '%" + namaPengajar + "%' "
			}

			if tempat != "" {
				if queryWhere != "" {
					queryWhere += " AND "
				}

				queryWhere += " tempat LIKE '%" + tempat + "%' "
			}

			if status != "" {
				if queryWhere != "" {
					queryWhere += " AND "
				}
				iStatus, err := strconv.Atoi(status)
				if err == nil {
					queryWhere += fmt.Sprintf(" status = %d ", iStatus)
				} else {
					errorMessage := "Error convert variable, " + err.Error()
					returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
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

			query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_jadwal_ekskul %s ;", queryWhere)
			if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
				errorMessage := "Error query, " + err.Error()
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
			totalPage = math.Ceil(float64(totalRecords) / float64(rowPage))

			query1 := fmt.Sprintf(`select id, nama_ekskul, tahun_ajaran, semester, hari, jam, nama_pengajar, tempat, status, tgl_input from siam_jadwal_ekskul %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, rowPage)
			rows, err := db.Query(query1)
			defer rows.Close()
			if err != nil {
				errorMessage := "Error query, " + err.Error()
				returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			for rows.Next() {
				err = rows.Scan(
					&jadwalEkskulResponse.Id,
					&jadwalEkskulResponse.NamaEkskul,
					&jadwalEkskulResponse.TahunAjaran,
					&jadwalEkskulResponse.Semester,
					&jadwalEkskulResponse.Hari,
					&jadwalEkskulResponse.Jam,
					&jadwalEkskulResponse.NamaPengajar,
					&jadwalEkskulResponse.Tempat,
					&jadwalEkskulResponse.Status,
					&jadwalEkskulResponse.TanggalInput,
				)

				jadwalEkskulResponses = append(jadwalEkskulResponses, jadwalEkskulResponse)

				if err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
					returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
			}

			errorMessage := "OK"
			returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "0", "0", errorMessage, errorMessage, logData, c)

		} else {
			errorMessage := "Method not found"
			returnJsonJadwalEkskul(jadwalEkskulResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(usernameSession, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		}

	}
	// }

}

func returnJsonJadwalEkskul(jadwalEkskulResponse []JadwalEkskulResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jadwalEkskulResponse,
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

func validateIdJadwalEkskul(id int, db *sql.DB) string {
	msg := "OK"

	cntJadwalEkskul := 0
	query1 := fmt.Sprintf(`select count(1) as cnt from siam_jadwal_ekskul where id = %d ;`, id)
	if err := db.QueryRow(query1).Scan(&cntJadwalEkskul); err != nil {
		msg = "Error query, " + err.Error()
	}

	if msg == "OK" {
		if cntJadwalEkskul == 0 {
			msg = "Data tidak ditemukan!"
		}
	}

	return msg
}

func validateJadwalEkskul(tahunAjaran string, semester int, hari string, tempat string, jam string, db *sql.DB) string {
	msg := "OK"

	cntJadwalEkskul := 0
	query1 := fmt.Sprintf(`select count(1) as cnt from siam_jadwal_ekskul where tahun_ajaran = '%s' and hari = '%s' and tempat = '%s' and jam = '%s' and semester = %d;`, tahunAjaran, hari, tempat, jam, semester)
	if err := db.QueryRow(query1).Scan(&cntJadwalEkskul); err != nil {
		msg = "Error query, " + err.Error()
	}

	if msg == "OK" {
		if cntJadwalEkskul > 0 {
			msg = "Sudah Ada Jadwal Ekskul Dengan Data Yang Sama!"
		}
	}

	return msg
}

func validateNamaPengajar(tahunAjaran string, semester int, hari string, jam string, namaPengajar string, db *sql.DB) string {
	msg := "OK"

	cntJadwalEkskul := 0
	query1 := fmt.Sprintf(`select count(1) as cnt from siam_jadwal_ekskul where tahun_ajaran = '%s' and hari = '%s' and nama_pengajar = '%s' and jam = '%s' and semester = %d;`, tahunAjaran, hari, namaPengajar, jam, semester)
	if err := db.QueryRow(query1).Scan(&cntJadwalEkskul); err != nil {
		msg = "Error query, " + err.Error()
	}

	if msg == "OK" {
		if cntJadwalEkskul > 0 {
			msg = "Pengajar Sudah Ada di Ekskul Lain Dengan Jadwal Yang Sama!"
		}
	}

	return msg
}
