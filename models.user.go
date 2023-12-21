// models.user.go

package main

import (
	"errors"
	"fmt"
	"strings"
)

type user struct {
	Username string `json:"username"`
	Password string `json:"-"`
}

// Check if the username and password combination is valid
func isUserValid(username, password string) bool {
	return validateUser(username, password)
}

// Changing password for the given username
func changePassword(username, password string) (*user, error) {
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("The password can't be empty")
	}

	u := user{Username: username, Password: password}

	return &u, nil
}

func validateUser(user, pass string) bool {
	// Validate the username/password against the seed values defined above
	// In a production app,
	// users will most likely be authenticated directly against a database

	rowUsr, err := runner.db.Query(fmt.Sprintf("Select password from usr where username='%s'", user))
	if err != nil {
		runner.logger.Println(err.Error())
		return false
	}
	defer rowUsr.Close()

	var password string
	for rowUsr.Next() {
		err := rowUsr.Scan(&password)
		if err != nil {
			runner.logger.Println(err.Error())
			return false
		}
		if pass != password {
			runner.logger.Println(err.Error())
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
		runner.logger.Println(err.Error())
		return 0
	}
	defer rowUsr.Close()

	var id int64
	for rowUsr.Next() {
		err := rowUsr.Scan(&id)
		if err != nil {
			runner.logger.Println(err.Error())
			return 0
		} else {
			return id
		}
	}
	return 0
}
