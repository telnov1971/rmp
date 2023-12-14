package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	//	"golang.org/x/crypto/bcrypt"

	"golang.org/x/text/encoding/charmap"
	"gopkg.in/ini.v1"
)

type Config struct {
	server_string   string
	port_string     string
	login_string    string
	pass_string     string
	path_string     string
	database_string string
	user_string     string
	meter_string    string
}

type Runner struct {
	db     *sql.DB
	logger *log.Logger
}

var config Config
var runner Runner

// , , indication string

func main() {
	cfg, err := ini.Load("rmp.ini")
	if err != nil {
		panic(err)
	}

	log_file, _ := os.OpenFile("rmp.log", os.O_APPEND|os.O_CREATE, 0755)
	runner.logger = log.New(log_file, "", log.Lshortfile)

	// Classic read of values, default section can be represented as empty string
	config.server_string = cfg.Section("").Key("server").String()
	config.port_string = cfg.Section("").Key("port").String()
	config.login_string = cfg.Section("").Key("login").String()
	config.pass_string = cfg.Section("").Key("pass").String()
	config.path_string = cfg.Section("").Key("path").String()
	config.database_string = cfg.Section("").Key("database").String()
	config.user_string = cfg.Section("").Key("user").String()
	config.meter_string = cfg.Section("").Key("meter").String()
	//indication = cfg.Section("").Key("indication").String()

	// Create the database handle, confirm driver is present
	connect_string := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
		config.login_string, config.pass_string, config.server_string, config.database_string)
	runner.db, err = sql.Open("mysql", connect_string)
	if err != nil {
		runner.logger.Fatal(err)
		panic(err)
	}
	defer runner.db.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/usr/", usrHandler)
	http.HandleFunc("/meter/", meterHandler)

	// Connect and check the server version
	var version string
	runner.db.QueryRow("SELECT VERSION()").Scan(&version)
	runner.logger.Println("Connected to:", version)

	go loadAll()

	addr := fmt.Sprintf("localhost:%s", config.port_string)
	runner.logger.Fatal(http.ListenAndServe(addr, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the home page RMP application!")
}

func usrHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := runner.db.Query("SELECT " +
		"username, password, personal_account, address, fio, contact, visit_date " +
		"FROM usr")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fmt.Fprintf(w, "Table:")
	count := 0
	for rows.Next() {
		var username, password, personal_account, address, fio, contact string
		var visit_date []uint8
		err := rows.Scan(&username, &password, &personal_account, &address, &fio, &contact, &visit_date)
		if err != nil {
			runner.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		count += 1
		fmt.Fprintf(w, "%d,%s,%s,%s,%s,%s,%s,%s\n",
			count, username, password, personal_account, address, fio, contact, string(visit_date[:]))
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func meterHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := runner.db.Query("SELECT " +
		"nom_pu, marka, mt, koef, los_per, ktp, res, meter_id, usr_id " +
		"FROM meterdevice")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fmt.Fprintf(w, "Table:")
	count := 0
	for rows.Next() {
		var nom_pu, marka, ktp, res string
		var meter_id, usr_id int64
		var mt, koef int
		var los_per float32
		err := rows.Scan(&nom_pu, &marka, &mt, &koef, &los_per, &ktp, &res, &meter_id, &usr_id)
		if err != nil {
			runner.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		count += 1
		fmt.Fprintf(w, "%d,%s,%s,%d,%d,%f,%s,%s,%d,%d\n",
			count, nom_pu, marka, mt, koef, los_per, ktp, res, meter_id, usr_id)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loadAll() {
	//loadUsr()
	loadMeterDevece()
}

func loadUsr() {
	timeStartDBload := time.Now()
	runner.logger.Printf("Start user load: %s", timeStartDBload.Local())

	file_string := fmt.Sprintf("%s\\%s", config.path_string, config.user_string)
	f, err := os.OpenFile(file_string, os.O_RDONLY, 0755)
	if err != nil {
		runner.logger.Fatal(err)
	}
	defer f.Close()

	decoder := charmap.Windows1251.NewDecoder()
	reader := decoder.Reader(f)
	b, err := io.ReadAll(reader)
	if err != nil {
		runner.logger.Panic(err)
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
			runner.logger.Fatal(err)
		}

		var id uint64
		i, err := strconv.ParseInt(record[4], 10, 64)
		if err == nil {
			sql_find_id := fmt.Sprintf("select `id` from `usr` where `lich_id`=%d", i)
			err2 := runner.db.QueryRow(sql_find_id).Scan(&id)
			if err2 != nil {
				// INSERT into table_name
				// [(column1, [, column2] ...)]
				// values (values_list)

				countInsert += 1
				sql_str = fmt.Sprintf("INSERT INTO usr " +
					"(username, password, personal_account, address, fio, contact, lich_id)" +
					" VALUES (?, ?, ?, ?, ?, ?, ?);")
				stmt, err := runner.db.Prepare(sql_str)
				if err == nil {
					_, err = stmt.Exec(record[0], record[0], record[0], record[1], record[2], record[3], i)
					if err != nil {
						runner.logger.Fatal(err)
						panic(err)
					}
					stmt.Close()
				} else {
					runner.logger.Fatal(err)
					panic(err)
				}
			} else {
				continue
				// UPDATE [table] table_name
				// SET column1 = value1, column2 = value2, ...
				// [WHERE condition]
				// [ORDER BY expression [ ASC | DESC ]]
				// [LIMIT number_rows];

				countUpdate += 1
				sql_str = fmt.Sprint("UPDATE usr " +
					"SET username=?, password=?, personal_account=?, address=?, fio=?, contact=?, lich_id=? " +
					"WHERE id=?;")
				stmt, err := runner.db.Prepare(sql_str)
				if err == nil {
					_, err = stmt.Exec(record[0], record[0], record[0], record[1], record[2], record[3], i, id)
					if err != nil {
						runner.logger.Fatal(err)
						panic(err)
					}
					stmt.Close()
				} else {
					runner.logger.Fatal(err)
					panic(err)
				}
			}
		}
	}
	runner.logger.Println("Users insert: ", countInsert)
	runner.logger.Println("Users update: ", countUpdate)
	runner.logger.Println(time.Until(timeStartDBload))
	runner.logger.Println("Users load")
}

func loadMeterDevece() {
	/*
		`meterdevice`
		`id` BIGINT(20), `nom_pu` VARCHAR(21), `marka`,
		`mt` INT(11), `koef` INT(11), `los_per` FLOAT,
		`ktp` VARCHAR(16), `res` VARCHAR(4),
		`meter_id` BIGINT(11), `usr_id` BIGINT(20)

			id, nom_pu, marka, mt, koef, los_per, ktp, res, meter_id, usr_id
	*/
	timeStartDBload := time.Now()
	runner.logger.Printf("Start meter device load: %s", timeStartDBload.Local())

	file_string := fmt.Sprintf("%s\\%s", config.path_string, config.meter_string)
	f, err := os.OpenFile(file_string, os.O_RDONLY, 0755)
	if err != nil {
		runner.logger.Fatal(err)
	}
	defer f.Close()

	decoder := charmap.Windows1251.NewDecoder()
	reader := decoder.Reader(f)
	b, err := io.ReadAll(reader)
	if err != nil {
		runner.logger.Panic(err)
		panic(err)
	}

	r := csv.NewReader(strings.NewReader(string(b)))
	r.Comma = ';'
	var sql_str string
	countInsert := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			runner.logger.Fatal(err)
		}

		var usr_id, meter_id uint64
		lich_id, _ := strconv.ParseInt(record[8], 10, 64)
		mt, _ := strconv.ParseInt(record[2], 10, 8)
		koef, _ := strconv.ParseInt(record[3], 10, 8)
		losPer, _ := strconv.ParseFloat(record[4], 32)
		meter_id, _ = strconv.ParseUint(record[7], 10, 64)
		if err == nil {
			sql_find_id := fmt.Sprintf("select id from usr where lich_id=%d", lich_id)
			err2 := runner.db.QueryRow(sql_find_id).Scan(&usr_id)
			if err2 != nil {
				runner.logger.Printf("For meter device $d not found user $d\n", meter_id, lich_id)
				continue
			}
			sql_find_meter := fmt.Sprintf("select id from meterdevice where meter_id=%d", meter_id)
			err2 = runner.db.QueryRow(sql_find_meter).Scan(&meter_id)
			if err2 == nil {
				continue
			} else {
				countInsert += 1
				sql_str = fmt.Sprintf("INSERT INTO meterdevice " +
					"(nom_pu, marka, mt, koef, los_per, ktp, res, meter_id, usr_id)" +
					" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);")
				stmt, err := runner.db.Prepare(sql_str)
				if err == nil {
					_, err = stmt.Exec(record[0], record[1], mt, koef, losPer, record[5], record[6], meter_id, usr_id)
					if err != nil {
						runner.logger.Fatal(err)
						panic(err)
					}
					stmt.Close()
				} else {
					runner.logger.Fatal(err)
					panic(err)
				}
			}
		}
	}
	runner.logger.Println("Meter devices insert: ", countInsert)
	runner.logger.Println(time.Until(timeStartDBload))
	runner.logger.Println("Meter devices load")
}
