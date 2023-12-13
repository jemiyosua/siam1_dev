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

type JGuruRequest struct {
	IdGuru        string
	NamaGuru      string
	JenisKelamin  string
	TanggalLahir  string
	Alamat        string
	NomorHP       string
	StatusGuru    string
	IdWaliKelas   string
	WaliKelas     string
	MapelPengampu string
	TglInput      string
	Username      string
	ParamKey      string
	Method        string
	Page          int
	RowPage       int
	OrderBy       string
	Order         string
}

type JGuruResponse struct {
	IdGuru        string
	NamaGuru      string
	JenisKelamin  string
	TanggalLahir  string
	Alamat        string
	NomorHP       string
	StatusGuru    string
	IdWaliKelas   string
	WaliKelas     string
	MapelPengampu string
	TglInput      string
}

func Guru(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "GURU"
	PageMenu := "Guru"

	var (
		bodyBytes    []byte
		XRealIp      string
		IP           string
		LogFile      string
		totalPage    float64
		totalRecords float64
	)

	jGuruRequest := JGuruRequest{}
	jGuruResponse := JGuruResponse{}
	jGuruResponses := []JGuruResponse{}

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
	LogFILE := LogFile + "Guru_" + DateNow + ".log"
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
		errorMessage := "Error, Body is empty!"
		returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jGuruRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data!"
		returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jGuruRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jGuruRequest); err != nil {
			errorMessage := "Error, Bind Json Data!"
			returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jGuruRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jGuruRequest.Username
			ParamKeySession := jGuruRequest.ParamKey
			Username := jGuruRequest.Username
			Method := jGuruRequest.Method

			IdGuru := jGuruRequest.IdGuru
			NamaGuru := jGuruRequest.NamaGuru
			JenisKelamin := jGuruRequest.JenisKelamin
			TanggalLahir := jGuruRequest.TanggalLahir
			Alamat := jGuruRequest.Alamat
			NomorHP := jGuruRequest.NomorHP
			StatusGuru := jGuruRequest.StatusGuru
			IdWaliKelas := jGuruRequest.IdWaliKelas
			WaliKelas := jGuruRequest.WaliKelas
			MapelPengampu := jGuruRequest.MapelPengampu

			Page = jGuruRequest.Page
			RowPage = jGuruRequest.RowPage
			Order := jGuruRequest.Order
			OrderBy := jGuruRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(UsernameSession, ParamKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired!"
				returnDataJsonGuru(jGuruResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			// ---------- start cek akses role ----------
			ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, Role := helper.GetRole(Username, c)
			if ErrorCodeGetRole != "" {
				returnDataJsonGuru(jGuruResponses, totalPage, ErrorCodeGetRole, ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, logData, c)
				helper.SendLogError(Username, PageGo, ErrorMessageGetRole, "", "", ErrorCodeGetRole, AllHeader, Method, Path, IP, c)
				return
			}

			ErrorCodeAccess, ErrorMessageAccess, ErrorMessageReturnAccess := helper.CheckMenuAccess(Role, PageMenu, c)
			if ErrorCodeAccess != "" {
				returnDataJsonGuru(jGuruResponses, totalPage, ErrorCodeAccess, ErrorCodeAccess, ErrorMessageAccess, ErrorMessageReturnAccess, logData, c)
				helper.SendLogError(Username, PageGo, ErrorMessageAccess, "", "", ErrorCodeAccess, AllHeader, Method, Path, IP, c)
				return
			}
			// ---------- end of cek akses role ----------

			if Method == "INSERT" {

				ErrorMessage := ""
				if NamaGuru == "" {
					ErrorMessage = "Nama Guru tidak boleh kosong!"
				} else if JenisKelamin == "" {
					ErrorMessage = "JenisKelamin tidak boleh kosong!"
				} else if TanggalLahir == "" {
					ErrorMessage = "TanggalLahir tidak boleh kosong!"
				} else if Alamat == "" {
					ErrorMessage = "Alamat tidak boleh kosong!"
				} else if NomorHP == "" {
					ErrorMessage = "NomorHP tidak boleh kosong!"
				}

				if ErrorMessage != "" {
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				currentTime := time.Now()
				currentTime1 := currentTime.Format("01/02/2006 15:04:05")

				query := fmt.Sprintf("insert into siam_guru(nama_guru, jenis_kelamin, tanggal_lahir, alamat, nomor_hp, status_guru, id_wali_kelas, wali_kelas, mapel_pengampu, tgl_input) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s','%s', NOW())", NamaGuru, JenisKelamin, TanggalLahir, Alamat, NomorHP, StatusGuru, IdWaliKelas, WaliKelas, MapelPengampu)
				_, err1 := db.Exec(query)
				if err1 != nil {
					errorMessageReturn := "Gagal INSERT ke tabel siam_guru!"
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err1)
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessageReturn, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				Log := "Berhasil insert data guru!"
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
				if IdGuru == "" {
					ErrorMessage = "Id Guru cannot null"
					returnDataJsonGuru(jGuruResponses, totalPage, "3", "3", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jGuruRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				CntIdGuru := 0
				query := fmt.Sprintf("SELECT COUNT(1) FROM siam_guru WHERE id_guru = '%s'", IdGuru)

				if err := db.QueryRow(query).Scan(&CntIdGuru); err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if CntIdGuru == 0 {
					ErrorMessage := "Guru dengan ID : " + IdGuru + " tidak di temukan!"
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jGuruRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if NamaGuru != "" {
					queryUpdate += fmt.Sprintf(" , nama_guru = '%s' ", NamaGuru)
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

				if StatusGuru != "" {
					queryUpdate += fmt.Sprintf(" , status_guru = '%s' ", StatusGuru)
				}

				if IdWaliKelas != "" {
					queryUpdate += fmt.Sprintf(" , id_wali_kelas = '%s' ", IdWaliKelas)
				}

				if WaliKelas != "" {
					queryUpdate += fmt.Sprintf(" , wali_kelas = '%s' ", WaliKelas)
				}

				if MapelPengampu != "" {
					queryUpdate += fmt.Sprintf(" , mapel_pengampu = '%s' ", MapelPengampu)
				}

				query1 := fmt.Sprintf("UPDATE siam_guru SET tgl_input = NOW() %s WHERE id_guru = '%s'", queryUpdate, IdGuru)
				_, err1 := db.Exec(query1)
				if err1 != nil {
					errorMessageReturn := "Gagal UPDATE ke tabel siam_guru!"
					errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessageReturn, logData, c)
					helper.SendLogError(jGuruRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				currentTime := time.Now()
				currentTime1 := currentTime.Format("01/02/2006 15:04:05")

				Log := "Berhasil update data Guru!"
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
				if IdGuru == "" {
					ErrorMessage := "Id Guru cannot null"
					returnDataJsonGuru(jGuruResponses, totalPage, "3", "3", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jGuruRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				CntIdGuru := 0
				query := fmt.Sprintf("SELECT COUNT(1) AS CNT FROM siam_guru WHERE id_guru = '%s'", IdGuru)

				if err := db.QueryRow(query).Scan(&CntIdGuru); err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if CntIdGuru == 0 {
					ErrorMessage := "Guru dengan ID : " + IdGuru + " tidak di temukan!"
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jGuruRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				query1 := fmt.Sprintf("DELETE FROM siam_guru WHERE id_guru = '%s'", IdGuru)
				_, err1 := db.Exec(query1)
				if err1 != nil {
					errorMessageReturn := "Gagal DELETE ke tabel siam_guru"
					errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessageReturn, logData, c)
					helper.SendLogError(jGuruRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				currentTime := time.Now()
				currentTime1 := currentTime.Format("01/02/2006 15:04:05")

				Log := "Berhasil menghapus data guru!"
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

			} else if Method == "SELECT" {

				PageNow := (Page - 1) * RowPage

				// ---------- start query where ----------
				queryWhere := ""
				if IdGuru == "" {
					if NamaGuru != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}

						queryWhere += " nama_guru LIKE '%" + NamaGuru + "%' "
					}

					if JenisKelamin != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}

						queryWhere += " jenis_kelamin = '" + JenisKelamin + "' "
					}

					if StatusGuru != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}

						queryWhere += " status = '" + StatusGuru + "' "
					}

					if queryWhere != "" {
						queryWhere = " WHERE " + queryWhere
					}
				}
				if IdGuru != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += " id_guru = '" + IdGuru + "' "

					if queryWhere != "" {
						queryWhere = " WHERE " + queryWhere
					}
				}
				// ---------- end of query where ----------

				queryOrder := ""

				if OrderBy != "" {
					queryOrder = " ORDER BY " + OrderBy + " " + Order
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_guru %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

				query1 := fmt.Sprintf(`SELECT id_guru,nama_guru,jenis_kelamin,tanggal_lahir,alamat,nomor_hp,status_guru,id_wali_kelas,wali_kelas,mapel_pengampu,tgl_input FROM siam_guru %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, RowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				for rows.Next() {
					err = rows.Scan(
						&jGuruResponse.IdGuru,
						&jGuruResponse.NamaGuru,
						&jGuruResponse.JenisKelamin,
						&jGuruResponse.TanggalLahir,
						&jGuruResponse.Alamat,
						&jGuruResponse.NomorHP,
						&jGuruResponse.StatusGuru,
						&jGuruResponse.IdWaliKelas,
						&jGuruResponse.WaliKelas,
						&jGuruResponse.MapelPengampu,
						&jGuruResponse.TglInput,
					)

					jGuruResponses = append(jGuruResponses, jGuruResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				returnDataJsonGuru(jGuruResponses, totalPage, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonGuru(jGuruResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
		}
	}
}

func returnDataJsonGuru(jGuruResponse []JGuruResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jGuruResponse,
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
