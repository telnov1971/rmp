package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

// Создаем расписание с указанными временами
var schedule = []string{"07:00", "12:00", "18:00"}
var tasks [3]bool

func firstLoad() {
	loadUsr()
	loadMeterDevece()
	loadIndication()
	config.loadonstart = false
}

func loadAll() {
	// Получаем текущую локальную временную зону
	localZone, err := time.LoadLocation("Local")
	if err != nil {
		runner.logger.Println("Ошибка при загрузке локальной временной зоны:", err)
		return
	}

	for i, t := range schedule {
		// Парсим время из расписания
		targetTime, err := time.ParseInLocation("15:04", t, localZone)
		if err != nil {
			runner.logger.Println("Ошибка при парсинге времени:", err)
			return
		}
		if !tasks[i] {
			runner.logger.Println("Запуск задачи на время", i, targetTime.Hour(), targetTime.Minute())
			go task(targetTime, i)
		}
	}
}

func task(t time.Time, i int) {
	tasks[i] = true

	// Получаем текущее время
	now := time.Now()
	currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.Local)
	fmt.Println("Сейчвас", i, currentTime.Hour(), currentTime.Minute())

	// Вычисляем время до следующего запуска задачи
	duration := t.Sub(currentTime)
	fmt.Println("Разница расписание - сейчас 1", i, duration)

	if duration < 0 {
		// Если время уже прошло на сегодня, переходим к следующему
		duration = duration + time.Hour*24
		fmt.Println("Разница расписание - сейчас 2", i, duration)
	}

	fmt.Println("Задача будет запущена через", i, duration.Minutes())

	// Ожидаем до времени запуска задачи
	time.Sleep(duration)

	// Здесь можно вызвать функцию или выполнить нужную задачу
	loadUsr()
	loadMeterDevece()
	loadIndication()
	tasks[i] = false
	time.Sleep(time.Minute)
	loadAll()
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
				runner.logger.Printf("For meter device %d not found user %d\n", meter_id, lich_id)
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

func loadIndication() {
	/*
		`indication`
			`id` BIGINT(20),
			`data` VARCHAR(13),
			`tz` VARCHAR(9),
			`i_date` DATE,
			`vid_en` VARCHAR(5),
			`device_id` BIGINT(20),

			id, data, tz, i_date, vid_en, device_id
	*/
	timeStartDBload := time.Now()
	runner.logger.Printf("Start indication load: %s", timeStartDBload.Local())

	file_string := fmt.Sprintf("%s\\%s", config.path_string, config.indication_string)
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

		var meter_id, id uint64
		meter_id, _ = strconv.ParseUint(record[4], 10, 64)
		var data, _ = strconv.ParseFloat(record[0], 64)
		var dateList = strings.Split(record[2], ".")
		var year, _ = strconv.ParseInt(dateList[2], 10, 64)
		var month, _ = strconv.ParseInt(dateList[1], 10, 64)
		var day, _ = strconv.ParseInt(dateList[0], 10, 64)
		var i_date = time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
		if err == nil {
			sql_find_id := fmt.Sprintf("select id from meterdevice where meter_id=%d", meter_id)
			err2 := runner.db.QueryRow(sql_find_id).Scan(&id)
			if err2 != nil {
				runner.logger.Printf("For indication of meter device %d not found device %d\n", meter_id, id)
				continue
			} else {
				sql_find_id := fmt.Sprintf("select id from indication where i_date='%s' "+
					"and device_id=%d and tz='%s'",
					i_date.Format(time.DateOnly), id, record[1])
				err3 := runner.db.QueryRow(sql_find_id).Scan(&meter_id)
				if err3 == nil {
					runner.logger.Printf("Indication of date %v of meter device %d already found\n",
						i_date,
						meter_id)
					continue
				} else {
					countInsert += 1
					sql_str = fmt.Sprintf("INSERT INTO indication " +
						"(data, tz, i_date, vid_en, device_id)" +
						" VALUES (?, ?, ?, ?, ?);")
					stmt, err := runner.db.Prepare(sql_str)
					if err == nil {
						_, err = stmt.Exec(data, record[1], i_date, record[3], id)
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
	}
	runner.logger.Println("Indications insert: ", countInsert)
	runner.logger.Println(time.Until(timeStartDBload))
	runner.logger.Println("Indications load")
}
