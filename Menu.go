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

type JMenuRequest struct {
	Username string
	ParamKey string
	Method   string
	MenuId   string
	Menu     string
	Status   string
	Page     int
	RowPage  int
	OrderBy  string
	Order    string
}

type JMenuResponse struct {
	MenuId       string
	Menu         string
	Status       string
	TanggalInput string
}

func Menu(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "MENU"
	// PageMenu := "Menu"

	var (
		bodyBytes    []byte
		XRealIp      string
		IP           string
		LogFile      string
		totalPage    float64
		totalRecords float64
	)

	jMenuRequest := JMenuRequest{}
	jMenuResponse := JMenuResponse{}
	jMenuResponses := []JMenuResponse{}

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
	LogFILE := LogFile + "Menu_" + DateNow + ".log"
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
		returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jMenuRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jMenuRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jMenuRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jMenuRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jMenuRequest.Username
			ParamKeySession := jMenuRequest.ParamKey
			Username := jMenuRequest.Username
			Method := jMenuRequest.Method

			MenuId := jMenuRequest.MenuId
			Menu := jMenuRequest.Menu
			Status := jMenuRequest.Status

			Page = jMenuRequest.Page
			RowPage = jMenuRequest.RowPage
			Order := jMenuRequest.Order
			OrderBy := jMenuRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(UsernameSession, ParamKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonMenu(jMenuResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			// ---------- start cek akses role ----------
			ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, Role := helper.GetRole(Username, c)
			if ErrorCodeGetRole != "" {
				returnDataJsonMenu(jMenuResponses, totalPage, ErrorCodeGetRole, ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, logData, c)
				helper.SendLogError(Username, PageGo, ErrorMessageGetRole, "", "", ErrorCodeGetRole, AllHeader, Method, Path, IP, c)
				return
			}

			// ErrorCodeAccess, ErrorMessageAccess, ErrorMessageReturnAccess := helper.CheckMenuAccess(Role, PageMenu, c)
			// if ErrorCodeAccess != "" {
			// 	returnDataJsonMenu(jMenuResponses, totalPage, ErrorCodeAccess, ErrorCodeAccess, ErrorMessageAccess, ErrorMessageReturnAccess, logData, c)
			// 	helper.SendLogError(Username, PageGo, ErrorMessageAccess, "", "", ErrorCodeAccess, AllHeader, Method, Path, IP, c)
			// 	return
			// }
			// ---------- end of cek akses role ----------

			if Method == "INSERT" {

				ErrorMessage := ""
				if Menu == "" {
					ErrorMessage = "Menu cannot null"
				}

				if ErrorMessage != "" {
					returnDataJsonMenu(jMenuResponses, totalPage, "3", "3", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				CntMenu := 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_menu WHERE menu = '%s'", Menu)
				if err := db.QueryRow(query).Scan(&CntMenu); err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if CntMenu > 0 {
					ErrorMessageReturn := fmt.Sprintf("%s already exist", Menu)
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", ErrorMessageReturn, ErrorMessageReturn, logData, c)
					helper.SendLogError(Username, PageGo, ErrorMessageReturn, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				Status := "1"
				query2 := fmt.Sprintf("INSERT INTO siam_menu (menu, status, tgl_input) VALUES ('%s', '%s', NOW())", Menu, Status)
				_, err1 := db.Exec(query2)
				if err1 != nil {
					errorMessageReturn := "Gagal INSERT ke tabel siam_menu"
					errorMessage := fmt.Sprintf("Error running %q: %+v", query2, err1)
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessageReturn, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				currentTime := time.Now()
				currentTime1 := currentTime.Format("01/02/2006 15:04:05")

				Log := "Berhasil INSERT data menu"
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
				if MenuId == "" {
					ErrorMessage = "MenuId cannot null"
				}

				if ErrorMessage != "" {
					returnDataJsonMenu(jMenuResponses, totalPage, "3", "3", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jMenuRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				CntMenuId := 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_menu WHERE id = '%s'", MenuId)
				if err := db.QueryRow(query).Scan(&CntMenuId); err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if CntMenuId == 0 {
					ErrorMessage := "MenuId not exist"
					returnDataJsonMenu(jMenuResponses, totalPage, "3", "3", ErrorMessage, ErrorMessage, logData, c)
					helper.SendLogError(jMenuRequest.Username, PageGo, ErrorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				CntMenu := 0
				query2 := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_menu WHERE menu = '%s'", Menu)
				if err := db.QueryRow(query2).Scan(&CntMenu); err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query2, err)
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if CntMenu > 0 {
					ErrorMessageReturn := fmt.Sprintf("%s already exist", Menu)
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", ErrorMessageReturn, ErrorMessageReturn, logData, c)
					helper.SendLogError(jMenuRequest.Username, PageGo, ErrorMessageReturn, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				if Menu != "" {
					queryUpdate += fmt.Sprintf(" , menu = '%s' ", Menu)
				}

				if Status != "" {
					queryUpdate += fmt.Sprintf(" , status = '%s' ", Status)
				}

				query1 := fmt.Sprintf("UPDATE siam_menu SET tgl_input = NOW() %s WHERE id = '%s'", queryUpdate, MenuId)
				_, err1 := db.Exec(query1)
				if err1 != nil {
					errorMessageReturn := "Gagal UPDATE ke tabel siam_menu"
					errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessageReturn, logData, c)
					helper.SendLogError(jMenuRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}

				currentTime := time.Now()
				currentTime1 := currentTime.Format("01/02/2006 15:04:05")

				Log := "Berhasil UPDATE data siswa"
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
				if Menu != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" menu LIKE '%%%s%%' ", Menu)
				}

				if Status != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" status = '%s' ", Status)
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
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_menu %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

				query1 := fmt.Sprintf(`SELECT id, menu, status, tgl_input FROM siam_menu %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, RowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				for rows.Next() {
					err = rows.Scan(
						&jMenuResponse.MenuId,
						&jMenuResponse.Menu,
						&jMenuResponse.Status,
						&jMenuResponse.TanggalInput,
					)

					jMenuResponses = append(jMenuResponses, jMenuResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				returnDataJsonMenu(jMenuResponses, totalPage, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonMenu(jMenuResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
		}
	}
}

func returnDataJsonMenu(jMenuResponse []JMenuResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jMenuResponse,
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
