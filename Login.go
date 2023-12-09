package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
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

type JLoginRequest struct {
	Username string
	Password string
}

func Login(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "LOGIN"
	ParamKey := ""

	var (
		bodyBytes []byte
		XRealIp   string
		IP        string
		LogFile   string
	)

	reqBody := JLoginRequest{}

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
		returnDataJsonlogin(reqBody.Username, ParamKey, "1", errorMessage, logData, c)
		helper.SendLogError(reqBody.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonlogin(reqBody.Username, ParamKey, "1", errorMessage, logData, c)
		helper.SendLogError(reqBody.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonlogin(reqBody.Username, ParamKey, "1", errorMessage, logData, c)
			helper.SendLogError(reqBody.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {
			Username := reqBody.Username
			Password := reqBody.Password
			errorMessage := ""

			// ------ Param Validation ------
			if Username == "" {
				errorMessage += "Username tidak boleh kosong"
			}

			if Password == "" {
				errorMessage += "Password tidak boleh kosong"
			}

			if errorMessage != "" {
				returnDataJsonlogin(Username, ParamKey, "1", errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of Param Validation ------

			CountLogin := 0
			PasswordDB := ""
			StatusAdmin := 0
			Role := ""
			CountBlock := 0
			query := fmt.Sprintf("SELECT COUNT(1) AS cnt, password, status, role, count_block FROM siam_login WHERE username = '%s';", Username)
			if err := db.QueryRow(query).Scan(&CountLogin, &PasswordDB, &StatusAdmin, &Role, &CountBlock); err != nil {
				errorMessage := "Error query, " + err.Error()
				returnDataJsonlogin(Username, ParamKey, "1", errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}

			if CountLogin == 0 {
				errorMessage := "Akun Anda tidak terdaftar"
				returnDataJsonlogin(Username, ParamKey, "1", errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			} else {
				if CountBlock >= 3 {
					errorMessage := "Akun Anda terblokir, harap hubungi Admin SIAM"
					returnDataJsonlogin(Username, ParamKey, "1", errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				} else {
					if StatusAdmin == 0 {
						errorMessage := "Akun Anda sudah tidak aktif"
						returnDataJsonlogin(Username, ParamKey, "1", errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					} else {
						if Password != PasswordDB {
							CountBlock += 1
							query := fmt.Sprintf("UPDATE siam_login SET count_block = '%d'", CountBlock)
							_, err := db.Exec(query)
							if err != nil {
								ParamKey = ""
								errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
								returnDataJsonlogin(Username, ParamKey, "1", errorMessage, logData, c)
								helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
								return
							}

							errorMessage := "Password Anda tidak sesuai"
							returnDataJsonlogin(Username, ParamKey, "1", errorMessage, logData, c)
							helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
							return
						} else {
							ParamKey = helper.Token()

							query := fmt.Sprintf("INSERT INTO siam_login_session (username,paramKey, tgl_input) values ('%s','%s', ADDTIME(NOW(), '0:20:0'))", Username, ParamKey)
							_, err := db.Exec(query)
							if err != nil {
								ParamKey = ""
								errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
								helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
								returnDataJsonlogin(Username, ParamKey, "1", errorMessage, logData, c)
								return
							}

							currentTime := time.Now()
							TimeNow := currentTime.Format("15:04:05")
							TimeNowSplit := strings.Split(TimeNow, ":")
							Hour := TimeNowSplit[0]
							Minute := TimeNowSplit[1]
							State := ""
							if Hour < "12" {
								State = "AM"
							} else {
								State = "PM"
							}

							Log := "Login Pukul " + Hour + ":" + Minute + " " + State
							helper.LogActivity(Username, PageGo, IP, bodyString, Method, Log, "Sukses", Role, c)
							returnDataJsonlogin(Username, ParamKey, "0", errorMessage, logData, c)
							return
						}
					}
				}
			}

		}
	}
}

func returnDataJsonlogin(UserName string, ParamKey string, ErrorCode string, ErrorMessage string, logData string, c *gin.Context) {
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
			"UserName":   UserName,
			"ParamKey":   ParamKey,
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
