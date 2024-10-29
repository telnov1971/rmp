// models.user.go

package main

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Check if the username and password combination is valid
func isUserValid(username, password string) bool {
	return validateUser(username, password)
}

// Changing password for the given username
func changePassword(username, password string) error {
	var id int64
	if strings.TrimSpace(password) == "" {
		return errors.New("пароль не может быть пустым")
	}

	id = getUserId(username)

	if id != 0 {
		pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		sqlupdate := "UPDATE usr SET usr.password=? WHERE usr.id=?"
		stmt, err := runner.db.Prepare(sqlupdate)
		if err == nil {
			_, err := stmt.Exec(pass, id)
			if err != nil {
				return err
			}
		}
		defer stmt.Close()
	}
	return nil
}

func validateUser(user, pass string) bool {
	// Validate the username/password against the seed values defined above
	// In a production app,
	// users will most likely be authenticated directly against a database

	rowUsr, err := runner.db.Query(fmt.Sprintf("Select password from usr where username='%s'", user))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer rowUsr.Close()

	var password string
	for rowUsr.Next() {
		err := rowUsr.Scan(&password)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(pass)); err != nil {
			fmt.Println(err.Error())
			return false
		} else {
			return true
		}
	}
	return false
}

func getUserId(user string) int64 {
	rowUsr, err := runner.db.Query(fmt.Sprintf("Select id from usr where username='%s'", user))
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	defer rowUsr.Close()

	var id int64
	for rowUsr.Next() {
		err := rowUsr.Scan(&id)
		if err != nil {
			fmt.Println(err.Error())
			return 0
		} else {
			return id
		}
	}
	return 0
}
