package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	//	"golang.org/x/crypto/bcrypt"

	"golang.org/x/text/encoding/charmap"
	"gopkg.in/ini.v1"
)

func main() {
	cfg, err := ini.Load("rmp.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	log_file, err := os.OpenFile("rmp.log", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log_file, err = os.Create("rmp.log")
		if err != nil {
			panic(err)
		}
	}
	logger := log.New(log_file, "", log.Lshortfile)

	// Classic read of values, default section can be represented as empty string
	server_string := cfg.Section("").Key("server").String()
	//port_string:=cfg.Section("").Key("port").String()
	login_string := cfg.Section("").Key("login").String()
	pass_string := cfg.Section("").Key("pass").String()
	path_string := cfg.Section("").Key("path").String()
	database_string := cfg.Section("").Key("database").String()
	user_string := cfg.Section("").Key("user").String()
	//meter:=cfg.Section("").Key("meter").String()
	//indication:=cfg.Section("").Key("indication").String()

	// Create the database handle, confirm driver is present
	connect_string := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", login_string, pass_string, server_string, database_string)
	db, err := sql.Open("mysql", connect_string)
	if err != nil {
		logger.Println(err)
	}
	defer db.Close()

	timeStartDBload := time.Now()
	logger.Println(timeStartDBload.Local())
	// Connect and check the server version
	var version string
	db.QueryRow("SELECT VERSION()").Scan(&version)
	logger.Println("Connected to:", version)

	file_string := fmt.Sprintf("%s\\%s", path_string, user_string)
	f, err := os.OpenFile(file_string, os.O_RDONLY, 0755)
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()

	decoder := charmap.Windows1251.NewDecoder()
	reader := decoder.Reader(f)
	b, err := io.ReadAll(reader)
	if err != nil {
		logger.Panic(err)
		panic(err)
	}

	r := csv.NewReader(strings.NewReader(string(b)))
	r.Comma = ';'
	var sql_str string
	countInsert := 0
	countUpdate := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Fatal(err)
		}

		var id uint64
		i, err := strconv.ParseInt(record[4], 10, 64)
		if err == nil {
			sql_find_id := fmt.Sprintf("select `id` from `usr` where `lich_id`=%d", i)
			err2 := db.QueryRow(sql_find_id).Scan(&id)
			if err2 != nil {
				// INSERT into table_name
				// [(column1, [, column2] ...)]
				// values (values_list)

				countInsert += 1
				sql_str = fmt.Sprintf("INSERT INTO usr " +
					"(username, password, personal_account, address, fio, contact, lich_id)" +
					" VALUES (?, ?, ?, ?, ?, ?, ?);")
				stmt, err := db.Prepare(sql_str)
				if err == nil {
					_, err = stmt.Exec(record[0], record[0], record[0], record[1], record[2], record[3], i)
					if err != nil {
						logger.Fatal(err)
						panic(err)
					}
					stmt.Close()
				} else {
					logger.Fatal(err)
					panic(err)
				}
			} else {
				// UPDATE [table] table_name
				// SET column1 = value1, column2 = value2, ...
				// [WHERE condition]
				// [ORDER BY expression [ ASC | DESC ]]
				// [LIMIT number_rows];

				countUpdate += 1
				sql_str = fmt.Sprint("UPDATE usr " +
					"SET username=?, password=?, personal_account=?, address=?, fio=?, contact=?, lich_id=? " +
					"WHERE id=?;")
				stmt, err := db.Prepare(sql_str)
				if err == nil {
					_, err = stmt.Exec(record[0], record[0], record[0], record[1], record[2], record[3], i, id)
					if err != nil {
						logger.Fatal(err)
						panic(err)
					}
					stmt.Close()
				} else {
					logger.Fatal(err)
					panic(err)
				}
			}
		}
	}
	logger.Println("Insert: ", countInsert)
	logger.Println("Update: ", countUpdate)

	db.Close()

	logger.Println(time.Until(timeStartDBload))
	logger.Println("Database closed")
}
