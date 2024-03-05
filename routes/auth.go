package routes

import (
	"net/http"
	"os"
	"time"

	"github.com/Bhar8at/bhar8at.github.io/database"
	"github.com/Bhar8at/bhar8at.github.io/middleware"
	"github.com/Bhar8at/bhar8at.github.io/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	issuer    string
	secretKey []byte
)

func init() {
	// loads and gets the following variables from the .env file
	godotenv.Load(".env")
	issuer = os.Getenv("ISSUER")
	secretKey = []byte(os.Getenv("SECRET_KEY"))
}

func SignUp(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "authT.html", gin.H{
			"type": "signup",
		})
	case "POST":
		var user models.User // creating a user

		if err := c.Request.ParseForm(); err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to parse form.",
			})
			return
		}

		// binding parsed form data with the current logged in user
		if err := c.ShouldBindWith(&user, binding.Form); err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": err.Error(),
			})
			return
		}

		// Checking whether the current user is present in database or not
		if user := database.ReadUserByName(user.Username); user != nil {
			c.HTML(http.StatusForbidden, "errorT.html", gin.H{
				"error":   "403 Forbidden",
				"message": "Account already exists with the given username.",
			})
			return
		}

		user.CreatedAt = time.Now()
		user.Id = uuid.NewString()
		user.HashPassword()

		// storing user data in database
		if res := database.CreateUser(&user); !res {
			// error returned since email is the unique key
			c.HTML(http.StatusForbidden, "errorT.html", gin.H{
				"error":   "403 Forbidden",
				"message": "Account already exists with the given email.",
			})
			return
		}

		// Set authorization token for user
		token, _ := middleware.CreateToken(user.Id)
		// initializes a session for the current user
		session := sessions.Default(c)
		session.Set("Authorization", token)
		session.Save()
		c.Redirect(http.StatusFound, "/login")
	}
}

func Login(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "authT.html", gin.H{
			"type": "login",
		})
	case "POST":
		var login models.Login
		if err := c.Request.ParseForm(); err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to parse form.",
			})
			return
		}
		if err := c.ShouldBindWith(&login, binding.Form); err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "403 Forbidden",
				"message": err.Error(),
			})
			return
		}
		user := database.ReadUserByName(login.Username)
		if user == nil {
			c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
				"error":   "401 Unauthorized",
				"message": "User does not exist.",
			})
			return
		}
		// Checking if the passwords match
		if !user.CheckPassword(login.Password) {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "401 Unauthorized",
				"message": "Incorrect password.",
			})
			return
		}

		// initializing token
		token, _ := middleware.CreateToken(user.Id)
		session := sessions.Default(c)
		session.Set("Authorization", token)
		session.Save()
		c.Redirect(http.StatusFound, "/feed")
	}
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("userId")
	if id == nil {
		c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
			"error":   "401 Unauthorized",
			"message": "User not logged in.",
		})
		return
	}
	// Remove all session headers
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	c.HTML(http.StatusOK, "responseT.html", gin.H{
		"message": "Logged out successfully.",
	})
}
