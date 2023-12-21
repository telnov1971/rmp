// handlers.user.go

package main

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func showLoginPage(c *gin.Context) {
	// Call the render function with the name of the template to render
	render(c, gin.H{
		"title": "Login",
	}, "login.html")
}

func performLogin(c *gin.Context) {
	// Obtain the POSTed username and password values
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Check if the username/password combination is valid
	if isUserValid(username, password) {
		// If the username/password is valid set the token in a cookie
		token := generateSessionToken()
		c.SetCookie("token", token, 3600, "", "", false, true)
		c.SetCookie("username", username, 3600, "", "", false, true)
		c.Set("is_logged_in", true)
		c.Set("user_id", getUserId(username))

		c.Redirect(http.StatusTemporaryRedirect, "/")

	} else {
		// If the username/password combination is invalid,
		// show the error message on the login page
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "Invalid credentials provided"})
	}
}

func generateSessionToken() string {
	// We're using a random 16 character string as the session token
	// This is NOT a secure way of generating session tokens
	// DO NOT USE THIS IN PRODUCTION
	return strconv.FormatInt(rand.Int63(), 16)
}

func logout(c *gin.Context) {

	// Clear the cookie
	c.SetCookie("token", "", -1, "", "", false, true)
	c.SetCookie("username", "", -1, "", "", false, true)

	// Redirect to the home page
	c.Redirect(http.StatusTemporaryRedirect, "/")

	c.Set("is_logged_in", false)
	c.Set("user_id", 0)
}

func showRegistrationPage(c *gin.Context) {
	// Call the render function with the name of the template to render
	render(c, gin.H{
		"title": "Change Password"}, "register.html")
}

func register(c *gin.Context) {
	// Obtain the POSTed username and password values
	// username := c.PostForm("username")
	password := c.PostForm("password")
	passVerify := c.PostForm("passwordVerify")

	if password != passVerify {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"ErrorTitle":   "Password Changing Failed",
			"ErrorMessage": "Passwords don't match"})
		return
	}

	if username, err := c.Cookie("username"); err != nil {
		if _, err := changePassword(username, password); err == nil {
			// If the user is created, set the token in a cookie and log the user in
			token := generateSessionToken()
			c.SetCookie("token", token, 3600, "", "", false, true)
			c.Set("is_logged_in", true)
			c.Set("username", username)

			render(c, gin.H{
				"title": "Password changed"}, "login-successful.html")

		} else {
			// If the username/password combination is invalid,
			// show the error message on the login page
			c.HTML(http.StatusBadRequest, "register.html", gin.H{
				"ErrorTitle":   "Password Changing Failed",
				"ErrorMessage": err.Error()})

		}
	}

}
