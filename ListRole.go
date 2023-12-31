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
)

type JListRoleRequest struct {
	Username string
	ParamKey string
	RoleName string
	Page     int
	RowPage  int
	OrderBy  string
	Order    string
}

type JListRoleResponse struct {
	RoleName string
	ListMenu string
}

func ListRole(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	StartTime := time.Now()
	StartTimeStr := StartTime.String()
	PageGo := "LISTROLE"
	Role := ""

	var (
		bodyBytes    []byte
		XRealIp      string
		IP           string
		LogFile      string
		totalPage    float64
		totalRecords float64
	)

	jListRoleRequest := JListRoleRequest{}
	jListRoleResponse := JListRoleResponse{}
	jListRoleResponses := []JListRoleResponse{}

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
	LogFILE := LogFile + "ListRole_" + DateNow + ".log"
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
		returnDataJsonListRole(jListRoleResponses, Role, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jListRoleRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage := "Error, Body - invalid json data"
		returnDataJsonListRole(jListRoleResponses, Role, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
		helper.SendLogError(jListRoleRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jListRoleRequest); err != nil {
			errorMessage := "Error, Bind Json Data"
			returnDataJsonListRole(jListRoleResponses, Role, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
			helper.SendLogError(jListRoleRequest.Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
			return
		} else {

			Page := 0
			RowPage := 0

			UsernameSession := jListRoleRequest.Username
			ParamKeySession := jListRoleRequest.ParamKey
			Username := jListRoleRequest.Username

			RoleName := jListRoleRequest.RoleName

			Page = jListRoleRequest.Page
			RowPage = jListRoleRequest.RowPage
			Order := jListRoleRequest.Order
			OrderBy := jListRoleRequest.OrderBy

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(UsernameSession, ParamKeySession, c)
			if checkAccessVal != "1" {
				checkAccessValErrorMsg := checkAccessVal
				checkAccessValErrorMsgReturn := "Session Expired"
				returnDataJsonListRole(jListRoleResponses, Role, totalPage, "2", "2", checkAccessValErrorMsgReturn, checkAccessValErrorMsgReturn, logData, c)
				helper.SendLogError(Username, PageGo, checkAccessValErrorMsg, "", "", "2", AllHeader, Method, Path, IP, c)
				return
			}
			// ------ end of check session paramkey ------

			// ---------- start cek akses role ----------
			ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, Role := helper.GetRole(Username, c)
			if ErrorCodeGetRole != "" {
				returnDataJsonListRole(jListRoleResponses, Role, totalPage, ErrorCodeGetRole, ErrorCodeGetRole, ErrorMessageGetRole, ErrorMessageReturnGetRole, logData, c)
				helper.SendLogError(Username, PageGo, ErrorMessageGetRole, "", "", ErrorCodeGetRole, AllHeader, Method, Path, IP, c)
				return
			}
			// ---------- end of cek akses role ----------

			PageNow := (Page - 1) * RowPage

			// ---------- start query where ----------
			queryWhere := ""
			if RoleName != "" {
				if queryWhere != "" {
					queryWhere += " AND "
				}

				queryWhere += " nama_role LIKE '%" + RoleName + "%' "
			}

			if queryWhere != "" {
				queryWhere = " WHERE " + queryWhere
			}
			// ---------- end of query where ----------

			queryOrder := ""

			if OrderBy != "" {
				queryOrder = " ORDER BY " + OrderBy + " " + Order
			}

			totalRecords = 0
			totalPage = 0
			query := fmt.Sprintf(`SELECT count(1) as cnt FROM (SELECT nama_role, group_concat(menu separator ', ') AS list_menu FROM siam_role sr LEFT JOIN siam_menu_akses sma ON sr.nama_role = sma.role LEFT JOIN siam_menu sm ON sm.id = sma.id_menu GROUP BY nama_role) tab_List_menu %s`, queryWhere)
			if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
				errorMessage := "Error query, " + err.Error()
				returnDataJsonListRole(jListRoleResponses, Role, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
			totalPage = math.Ceil(float64(totalRecords) / float64(RowPage))

			query1 := fmt.Sprintf(`SELECT nama_role, ifnull(list_menu,'') list_menu FROM (SELECT nama_role, group_concat(menu separator ', ') AS list_menu FROM siam_role sr LEFT JOIN siam_menu_akses sma ON sr.nama_role = sma.role LEFT JOIN siam_menu sm ON sm.id = sma.id_menu GROUP BY nama_role) tab_List_menu %s %s LIMIT %d,%d;`, queryWhere, queryOrder, PageNow, RowPage)
			rows, err := db.Query(query1)
			defer rows.Close()
			if err != nil {
				errorMessage := "Error query, " + err.Error()
				returnDataJsonListRole(jListRoleResponses, Role, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
				helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
				return
			}
			for rows.Next() {
				err = rows.Scan(
					&jListRoleResponse.RoleName,
					&jListRoleResponse.ListMenu,
				)

				jListRoleResponses = append(jListRoleResponses, jListRoleResponse)

				if err != nil {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err)
					returnDataJsonListRole(jListRoleResponses, Role, totalPage, "1", "1", errorMessage, errorMessage, logData, c)
					helper.SendLogError(Username, PageGo, errorMessage, "", "", "1", AllHeader, Method, Path, IP, c)
					return
				}
			}

			returnDataJsonListRole(jListRoleResponses, Role, totalPage, "0", "0", "", "", logData, c)
			return
		}
	}
}

func returnDataJsonListRole(jListRoleResponse []JListRoleResponse, Role string, TotalPage float64, ErrorCode string, ErrorCodeReturn string, ErrorMessage string, ErrorMessageReturn string, logData string, c *gin.Context) {
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
			"Result":     jListRoleResponse,
			"TotalPage":  TotalPage,
			"Role": Role,
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
