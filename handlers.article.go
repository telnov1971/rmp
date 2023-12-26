// handlers.article.go

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var periodstart, periodstend, sort string

func filterIndexPage(c *gin.Context) {
	periodstart = c.PostForm("periodstart")
	periodstend = c.PostForm("periodend")
	sort = c.PostForm("sort")
	c.Request.Method = "GET"
	showIndexPage(c)
}

func showIndexPage(c *gin.Context) {

	getIndications(c)

	// Call the render function with the name of the template to render
	render(c, gin.H{
		"title":           "Home Page",
		"personalAccount": ipa.personal_account,
		"address":         ipa.address,
		"fio":             ipa.fio,
		"table":           ipa.table,
		"periodstart":     periodstart,
		"periodend":       periodstend,
		"sort":            sort,
	},
		"index.html")
}

func getIndications(c *gin.Context) {
	var sql1, sql2, sql3, sql string
	user, err := c.Cookie("username")
	if err == nil && user != "" {
		token, err := c.Cookie("token")
		if err == nil && token != "" {
			user_id := getUserId(user)
			rowUsr, err := runner.db.Query(fmt.Sprintf("SELECT u.personal_account, u.address, u.fio "+
				"FROM usr AS u WHERE id =%d", user_id))
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rowUsr.Close()

			for rowUsr.Next() {
				err := rowUsr.Scan(&ipa.personal_account, &ipa.address, &ipa.fio)
				if err != nil {
					runner.logger.Println(err)
					http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
					return
				}
				ipa.table = []IndicationRow{}
			}
			sql1 = fmt.Sprintf("SELECT "+
				"m.nom_pu, m.marka, m.mt, m.koef, m.los_per, m.ktp, "+
				"i.data, i.tz, i.i_date, i.vid_en "+
				"FROM indication AS i, meterdevice AS m "+
				"WHERE i.device_id = m.id AND m.usr_id = %d ", user_id)
			if periodstart != "" && periodstend != "" {
				sql2 = fmt.Sprintf("AND i.i_date BETWEEN '%s' AND '%s' ", periodstart, periodstend)
			} else {
				if periodstart != "" {
					sql2 = fmt.Sprintf("AND i.i_date > '%s' ", periodstart)
				}
				if periodstend != "" {
					sql2 = fmt.Sprintf("AND i.i_date < '%s' ", periodstend)
				}
			}
			switch sort {
			case "", "idate":
				sql3 = "ORDER BY i.i_date DESC, m.nom_pu ASC"
			case "number":
				sql3 = "ORDER BY m.nom_pu ASC, i.i_date DESC"
			}
			sql = sql1 + sql2 + sql3
			rows, err := runner.db.Query(sql)
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			count := 0
			var i_date_time time.Time
			for rows.Next() {
				var ir IndicationRow
				err := rows.Scan(&ir.Nom_pu, &ir.Marka, &ir.Mt, &ir.Koef, &ir.Los_per, &ir.Ktp,
					&ir.Data, &ir.Tz, &i_date_time, &ir.Vid_en)
				if err != nil {
					runner.logger.Println(err)
					http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
					return
				}
				ir.I_date = i_date_time.Format(time.DateOnly)
				ipa.table = append(ipa.table, ir)
				count += 1
			}
			if err = rows.Err(); err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			c.Redirect(http.StatusTemporaryRedirect, "/u/login")
		}
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/u/login")
	}
}
