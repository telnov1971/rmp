package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	//	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
)

type Config struct {
	server_string     string
	port_string       string
	login_string      string
	pass_string       string
	path_string       string
	database_string   string
	user_string       string
	meter_string      string
	indication_string string
	loadonstart       bool
}

type Runner struct {
	db     *sql.DB
	logger *log.Logger
}

type IndicationRow struct {
	Nom_pu  string
	Marka   string
	Mt      int
	Koef    int
	Los_per float64
	Ktp     string
	Data    float64
	Tz      string
	I_date  string
	Vid_en  string
}

type Usr struct {
	Id               int64     `json:"id"`
	Username         string    `json:"username"`
	Password         string    `json:"password"`
	Personal_account string    `json:"personalaccount"`
	Address          string    `json:"addres"`
	Fio              string    `json:"fio"`
	Contact          string    `json:"contact"`
	Lich_id          int64     `json:"lichid"`
	Visit_date       time.Time `json:"visitdate"`
}

type IndicationOfPersonalAccount struct {
	personal_account string
	address          string
	fio              string
	table            []IndicationRow
}

var config Config
var runner Runner
var router *gin.Engine
var ipa IndicationOfPersonalAccount

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	cfg, err := ini.Load("rmp.ini")
	if err != nil {
		panic(err)
	}

	log_file, _ := os.OpenFile("rmp.log", os.O_APPEND|os.O_CREATE, 0755)
	runner.logger = log.New(log_file, "", log.Ldate|log.Ltime|log.Lshortfile)

	// Classic read of values, default section can be represented as empty string
	config.server_string = cfg.Section("").Key("server").String()
	config.port_string = cfg.Section("").Key("port").String()
	config.login_string = cfg.Section("").Key("login").String()
	config.pass_string = cfg.Section("").Key("pass").String()
	config.path_string = cfg.Section("").Key("path").String()
	config.database_string = cfg.Section("").Key("database").String()
	config.user_string = cfg.Section("").Key("user").String()
	config.meter_string = cfg.Section("").Key("meter").String()
	config.indication_string = cfg.Section("").Key("indication").String()
	config.loadonstart, _ = cfg.Section("").Key("loadonstart").Bool()

	// Create the database handle, confirm driver is present
	connect_string := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true",
		config.login_string, config.pass_string, config.server_string, config.database_string)
	runner.db, err = sql.Open("mysql", connect_string)
	if err != nil {
		runner.logger.Fatal(err)
		panic(err)
	}
	defer runner.db.Close()

	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")

	// Connect and check the server version
	var version string
	err = runner.db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		runner.logger.Println("Error in answer of query server version")
	}
	runner.logger.Println("Connected to:", version)

	if config.loadonstart {
		go firstLoad()
	}

	go loadAll()

	// Initialize the routes
	initializeRoutes()

	// Start serving the application
	runner.logger.Println("Web server started", time.Now().Local())
	runner.logger.Fatal(router.Run(fmt.Sprintf(":%s", config.port_string)))
}

// Render one of HTML, JSON or CSV based on the 'Accept' header of the request
// If the header doesn't specify this, HTML is rendered, provided that
// the template name is present
func render(c *gin.Context, data gin.H, templateName string) {
	loggedInInterface, _ := c.Get("is_logged_in")
	data["is_logged_in"] = loggedInInterface.(bool)

	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		// Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}
}
