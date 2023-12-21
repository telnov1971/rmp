// handlers.article.go

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func showIndexPage(c *gin.Context) {

	getIndications(c)

	// Call the render function with the name of the template to render
	render(c, gin.H{
		"title":           "Home Page",
		"personalAccount": ipa.personal_account,
		"address":         ipa.address,
		"fio":             ipa.fio,
		"table":           ipa.table},
		"index.html")
}

func getIndications(c *gin.Context) {
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
			rows, err := runner.db.Query(fmt.Sprintf("SELECT "+
				"m.nom_pu, m.marka, m.mt, m.koef, m.los_per, m.ktp, "+
				"i.data, i.tz, i.i_date, i.vid_en "+
				"FROM indication AS i, meterdevice AS m "+
				"WHERE i.device_id = m.id AND m.usr_id = %d "+
				"ORDER BY m.nom_pu, i.i_date", user_id))
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
