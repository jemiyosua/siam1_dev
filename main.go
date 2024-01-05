package main

import (
	"database/sql"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	db               *sql.DB
	connectionString string
)

func main() {
	router := gin.Default()
	router.Use(CORS())
	router.ForwardedByClientIP = true

	connectionString = os.Getenv("STRINGCONNECTION1")

	apiVersion := "/api/v1/"

	router.POST(apiVersion+"Login", Login)
	router.POST(apiVersion+"Siswa", Siswa)
	router.POST(apiVersion+"Roles", Roles)
	router.POST(apiVersion+"MenuSidebar", MenuSidebar)
	router.POST(apiVersion+"SubMenu", SubMenu)
	router.POST(apiVersion+"UserLogin", UserLogin)
	router.POST(apiVersion+"KelasActive", KelasActive)
	router.POST(apiVersion+"JadwalEkskul", JadwalEkskul)
	router.POST(apiVersion+"ListRole", ListRole)
	router.POST(apiVersion+"Roles", Roles)
	router.POST(apiVersion+"Subjects", Subjects)
	router.POST(apiVersion+"Majors", Majors)

	PORT := os.Getenv("PORT")

	router.Run(":" + PORT)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Signature, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
