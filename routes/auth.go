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
		var user models.User
		if err := c.Request.ParseForm(); err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to parse form.",
			})
			return
		}
		if err := c.ShouldBindWith(&user, binding.Form); err != nil {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": err.Error(),
			})
			return
		}
		if user := database.ReadUserByName(user.Username); user != nil {
			c.HTML(http.StatusForbidden, "errorT.html", gin.H{
				"error":   "403 Forbidden",
				"message": "Account already exists with the given username.",
			})
			return
		}
		user.CreatedAt = time.Now()
		user.Id = uuid.NewString()
		user.Verified = false
		user.HashPassword()
		if res := database.CreateUser(&user); !res {
			c.HTML(http.StatusForbidden, "errorT.html", gin.H{
				"error":   "403 Forbidden",
				"message": "Account already exists with the given email.",
			})
			return
		}
		// Set authorization token for user
		token, _ := middleware.CreateToken(user.Id)
		session := sessions.Default(c)
		session.Set("Authorization", token)
		session.Save()
		c.Redirect(http.StatusFound, "/auth/verify?signup=true")
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
		if !user.CheckPassword(login.Password) {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "401 Unauthorized",
				"message": "Incorrect password.",
			})
			return
		}
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
