package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Bhar8at/bhar8at.github.io/database"
	"github.com/Bhar8at/bhar8at.github.io/internal"
	"github.com/Bhar8at/bhar8at.github.io/middleware"
	"github.com/Bhar8at/bhar8at.github.io/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var config *oauth2.Config
var state string

func init() {
	state = os.Getenv("SECRET_KEY")
	config = &oauth2.Config{
		ClientID:     "708668997941-tsb8vf0j5olhckrum8ibr0mimhbjns6d.apps.googleusercontent.com",
		ClientSecret: "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",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

func GoogleSignUp(c *gin.Context) {
	config.RedirectURL = "http://localhost:8080/auth/google"
	c.Redirect(http.StatusFound, config.AuthCodeURL(state))
}

func GoogleLogin(c *gin.Context) {
	config.RedirectURL = "http://localhost:8080/auth/google?login=true"
	c.Redirect(http.StatusFound, config.AuthCodeURL(state))
}

func GoogleAuth(c *gin.Context) {
	if c.Query("state") != state {
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Invalid authorization URL.",
		})
		return
	}
	switch c.Query("login") {
	case "true":
		config.RedirectURL = "http://localhost:8080/auth/google?login=true"
	default:
		config.RedirectURL = "http://localhost:8080/auth/google"
	}
	token, err := config.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error": "400 Bad Request",
		})
		return
	}
	client := config.Client(context.Background(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to retrieve authorization response, try again later.",
		})
		return
	}
	defer response.Body.Close()
	var authUser models.GoogleUser
	if err := json.NewDecoder(response.Body).Decode(&authUser); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
			"error":   "400 Bad Request",
			"message": "Unable to parse authentication response, try again later.",
		})
		return
	}
	// Signup or login user
	exists := database.ReadUserByEmail(authUser.Email)
	switch c.Query("login") {
	case "true":
		if exists == nil {
			c.HTML(http.StatusUnauthorized, "errorT.html", gin.H{
				"error":   "401 Unauthorized",
				"message": "User does not exist.",
			})
			return
		}
		token, _ := middleware.CreateToken(exists.Id)
		session := sessions.Default(c)
		session.Set("Authorization", token)
		session.Save()
		c.Redirect(http.StatusFound, "/feed")
	default:
		if exists != nil {
			c.HTML(http.StatusForbidden, "errorT.html", gin.H{
				"error":   "403 Forbidden",
				"message": "Account already exists with the given email.",
			})
			return
		}
		var user models.User
		user.Username = authUser.Username
		// Update the username if it already exists in the database
		if result := database.ReadUserByName(user.Username); result != nil {
			user.Username += internal.RandomString(32 - len(authUser.Username))
		}
		user.CreatedAt = time.Now()
		user.Email = &authUser.Email
		user.Verified = authUser.Verified
		user.Id = uuid.NewString()
		// Generate a random password for oauth user
		user.Password = uuid.NewString()
		user.HashPassword()
		if authUser.Avatar != nil {
			user.Avatar = authUser.Avatar
		}
		if res := database.CreateUser(&user); !res {
			c.HTML(http.StatusBadRequest, "errorT.html", gin.H{
				"error":   "400 Bad Request",
				"message": "Unable to create account, try again later.",
			})
			return
		}
		// Add to table that identifies OAuth users
		database.CreateOAuthUser(user.Id)
		token, _ := middleware.CreateToken(user.Id)
		session := sessions.Default(c)
		session.Set("Authorization", token)
		session.Save()
		if user.Verified {
			c.Redirect(http.StatusFound, "/feed")
		} else {
			c.Redirect(http.StatusFound, "/auth/verify?signup=true")
		}
	}
}
