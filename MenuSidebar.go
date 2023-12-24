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

type JMenuSidebarRequest struct {
	Username string
	ParamKey string
	Method   string
	Page     int
	RowPage  int
	OrderBy  string
	Order    string
}

type JMenuSidebarResponse struct {
	Menu       string
	IdMenu     string
	PageActive string
	Href       string
	Status     string
}

func MenuSidebar(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "MENU_SIDEBAR"
	// PageMenu := "Menu Sidebar"

	var (
		bodyBytes    []byte
		XRealIp      string
		IP           string
		LogFile      string
		totalPage    float64
		totalRecords float64
	)

	jMenuSidebarRequest := JMenuSidebarRequest{}
	jMenuSidebarResponse := JMenuSidebarResponse{}
	jMenuSidebarResponses := []JMenuSidebarResponse{}

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
	LogFILE := LogFile + "MenuSidebar_" + DateNow + ".log"
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
		returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jMenuSidebarRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jMenuSidebarRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jMenuSidebarRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jMenuSidebarRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jMenuSidebarRequest.Username
			ParamKeySession := jMenuSidebarRequest.ParamKey
			Username := jMenuSidebarRequest.Username
			Method := jMenuSidebarRequest.Method

			Page = jMenuSidebarRequest.Page
			RowPage = jMenuSidebarRequest.RowPage
			Order := jMenuSidebarRequest.Order
			OrderBy := jMenuSidebarRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(UsernameSession, ParamKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			// ---------- start cek akses role ----------
			ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, Role := helper.GetRole(Username, c)
			if ErrorCodeGetRole != "" {
				returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, ErrorCodeGetRole, ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, logData, c)
				helper.SendLogError(Username, PageGo, ErrorMessageGetRole, "", "", ErrorCodeGetRole, AllHeader, Method, Path, IP, c)
				return
			}
			// ---------- end of cek akses role ----------

			if Method == "INSERT" {

			} else if Method == "UPDATE" {

			} else if Method == "DELETE" {

			} else if Method == "SELECT" {

				PageNow := (Page - 1) * RowPage

				// ---------- start query where ----------
				queryWhere := fmt.Sprintf(" a.id_menu = b.id AND status = 1 AND role = '%s' ", Role)

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
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM siam_menu_akses a, siam_menu b %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

				query1 := fmt.Sprintf("SELECT b.id, menu, page_active, href, status FROM siam_menu_akses a, siam_menu b %s %s LIMIT %d,%d;", queryWhere, queryOrder, PageNow, RowPage)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage := "Error query, " + err.Error()
					returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
				for rows.Next() {
					err = rows.Scan(
						&jMenuSidebarResponse.IdMenu,
						&jMenuSidebarResponse.Menu,
						&jMenuSidebarResponse.PageActive,
						&jMenuSidebarResponse.Href,
						&jMenuSidebarResponse.Status,
					)

					jMenuSidebarResponses = append(jMenuSidebarResponses, jMenuSidebarResponse)

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
						returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
						helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
						return
					}
				}

				returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "0", "0", "", "", logData, c)
				return

			} else {
				errorMessage := "Method not found"
				returnDataJsonMenuSidebar(jMenuSidebarResponses, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
		}
	}
}

func returnDataJsonMenuSidebar(jMenuSidebarResponse []JMenuSidebarResponse, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jMenuSidebarResponse,
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
